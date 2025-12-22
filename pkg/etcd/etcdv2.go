package etcd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/etcdfinder/etcdfinder/internal/customerrors"
	"github.com/etcdfinder/etcdfinder/pkg/common"
	"github.com/etcdfinder/etcdfinder/pkg/logger"
	etcdv2 "go.etcd.io/etcd/client/v2"
)

// ClientV2 wraps the etcd v2 client with custom functionality
type ClientV2 struct {
	client                etcdv2.KeysAPI
	watchEventChannelSize int64  // size of the watch event channel
	rootPrefixEtcd        string // prefix of the etcd keys to be watched
	numGetKeysLimit       int64  // number of keys to be returned in a single GetKeysWithPagination call
	EtcdAuditPeriod       time.Duration
	maxWatchRetries       int64    // maximum number of consecutive failures on the same ModRevision
	ExpectedModIndex      uint64   // expected modified index of the etcd keys
	endpoints             []string // endpoints for health checks
}

// NewClientV2 creates a new etcd v2 client
func NewClientV2(
	endpoints []string,
	watchEventChannelSize int64,
	rootPrefixEtcd string,
	numGetKeysLimit int64,
	etcdAuditPeriod int64,
	maxWatchRetries int64) (BaseClient, error) {
	if numGetKeysLimit <= 0 {
		return nil, fmt.Errorf("numGetKeysLimit must be greater than 0")
	}

	cfg := etcdv2.Config{
		Endpoints: endpoints,
		Transport: etcdv2.DefaultTransport,
	}

	cli, err := etcdv2.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd v2 client: %w", err)
	}

	return &ClientV2{
		client:                etcdv2.NewKeysAPI(cli),
		watchEventChannelSize: watchEventChannelSize,
		rootPrefixEtcd:        rootPrefixEtcd,
		numGetKeysLimit:       numGetKeysLimit,
		EtcdAuditPeriod:       time.Duration(etcdAuditPeriod) * time.Second,
		maxWatchRetries:       maxWatchRetries,
		ExpectedModIndex:      0,
		endpoints:             endpoints,
	}, nil
}

func (c *ClientV2) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key, nil)
	if err != nil {
		if etcdv2.IsKeyNotFound(err) {
			return "", customerrors.ErrKeyNotFound
		}
		return "", fmt.Errorf("failed to get key: %w", err)
	}

	if resp.Node == nil {
		return "", customerrors.ErrKeyNotFound
	}

	return resp.Node.Value, nil
}

func (c *ClientV2) Put(ctx context.Context, key string, value string) (string, error) {
	resp, err := c.client.Set(ctx, key, value, nil)
	if err != nil {
		return "", fmt.Errorf("failed to put key: %w", err)
	}
	if resp.Node == nil {
		return "", customerrors.ErrKeyNotPut
	}
	return resp.Node.Key, nil
}

func (c *ClientV2) Delete(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Delete(ctx, key, &etcdv2.DeleteOptions{})
	if err != nil {
		if etcdv2.IsKeyNotFound(err) {
			return "", nil // Already deleted
		}
		return "", fmt.Errorf("failed to delete key: %w", err)
	}
	if resp.Node == nil {
		return "", customerrors.ErrKeyNotDeleted
	}
	return resp.Node.Key, nil
}

// Watch watches for changes on keys
// Returns a channel of WatchEvents and an error channel
func (c *ClientV2) Watch(ctx context.Context) (<-chan WatchEvent, <-chan error) {
	eventCh := make(chan WatchEvent, c.watchEventChannelSize)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)
		watchIndexDiscrepancy := false
		var consecutiveFailureCount int64

		for {
			var watcher etcdv2.Watcher
			watchOpts := &etcdv2.WatcherOptions{
				Recursive: true,
			}

			if watchIndexDiscrepancy && c.ExpectedModIndex > 0 {
				watchOpts.AfterIndex = c.ExpectedModIndex - 1
			}

			watcher = c.client.Watcher(c.rootPrefixEtcd, watchOpts)

			for {
				resp, err := watcher.Next(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return // Context cancelled
					}
					errCh <- fmt.Errorf("watch error: %w", err)
					return
				}

				if resp.Node == nil {
					continue
				}

				// If this is the first event, set the expected modindex
				if c.ExpectedModIndex == 0 {
					c.ExpectedModIndex = resp.Node.ModifiedIndex
				}

				// Check if the modindex is not equal to the expected modindex
				if resp.Node.ModifiedIndex != c.ExpectedModIndex {
					consecutiveFailureCount++
					logger.Warnf("ModIndex mismatch: Consecutive failure #%d on ModIndex %d", consecutiveFailureCount, c.ExpectedModIndex)

					// If we've exceeded max retries on the same index, fail fast
					if consecutiveFailureCount >= c.maxWatchRetries {
						errCh <- fmt.Errorf("exceeded max watch retries (%d) on ModIndex %d - failing fast to prevent infinite loop", c.maxWatchRetries, c.ExpectedModIndex)
						return
					}
					watchIndexDiscrepancy = true
					break
				}

				// Successfully processed an event, reset failure counter
				consecutiveFailureCount = 0
				c.ExpectedModIndex = resp.Node.ModifiedIndex + 1

				watchEvent := WatchEvent{
					Key: resp.Node.Key,
				}

				switch resp.Action {
				case "set", "create", "update", "compareAndSwap":
					watchEvent.Type = "PUT"
					watchEvent.Value = resp.Node.Value
				case "delete", "expire", "compareAndDelete":
					watchEvent.Type = "DELETE"
				default:
					// Skip unknown actions
					continue
				}

				select {
				case eventCh <- watchEvent:
				case <-ctx.Done():
					return
				}

				watchIndexDiscrepancy = false
			}

			if !watchIndexDiscrepancy {
				break
			}
		}
	}()

	return eventCh, errCh
}

// GetKeysWithPagination retrieves keys with pagination support
func (c *ClientV2) GetKeysWithPagination(ctx context.Context, fromKey string) ([]common.KV, string, error) {
	// Always fetch from root to ensure we can traverse the tree
	getKey := c.rootPrefixEtcd

	opts := &etcdv2.GetOptions{
		Recursive: true,
		Sort:      true,
	}

	resp, err := c.client.Get(ctx, getKey, opts)
	if err != nil {
		if etcdv2.IsKeyNotFound(err) {
			return []common.KV{}, "", nil
		}
		return nil, "", fmt.Errorf("failed to get keys: %w", err)
	}

	keys := make([]common.KV, 0)

	// skipping indicates if we are currently skipping keys until we find fromKey
	skipping := fromKey != ""

	// Recursively collect all keys from the node tree
	var collectKeys func(node *etcdv2.Node)
	collectKeys = func(node *etcdv2.Node) {
		if node == nil {
			return
		}

		// Check if we've reached the limit
		if int64(len(keys)) >= c.numGetKeysLimit {
			return
		}

		if !node.Dir {
			// Leaf node logic
			if skipping {
				if node.Key == fromKey {
					skipping = false
				}
				// If we were skipping, we don't add the key (even the one that matched)
				// The next key encountered will be added.
				return
			}

			// Not skipping, add key
			keys = append(keys, common.KV{
				Key:   node.Key,
				Value: node.Value,
			})
			return
		}

		// directory node - process sorted children
		if node.Nodes != nil {
			for _, child := range node.Nodes {
				if int64(len(keys)) >= c.numGetKeysLimit {
					return
				}
				collectKeys(child)
			}
		}
	}

	collectKeys(resp.Node)

	if len(keys) == 0 {
		return keys, "", nil
	}

	// If result is full, return nextKey.
	// Note: If we reached exactly end of list and it's full, we still return nextKey.
	// The next call will return empty, which ends pagination.
	return keys, keys[len(keys)-1].Key, nil
}

// StartAuditor starts a background goroutine that checks etcd connection health every EtcdAuditPeriod
// Returns an error channel that will receive errors if the connection check fails
func (c *ClientV2) StartAuditor(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		ticker := time.NewTicker(c.EtcdAuditPeriod)
		defer ticker.Stop()

		logger.Infof("Starting etcd v2 connection auditor (checking every %s)", c.EtcdAuditPeriod)

		// Perform initial health check
		if err := c.checkConnection(ctx); err != nil {
			logger.Errorf("Initial etcd v2 connection check failed: %v", err)
			errCh <- fmt.Errorf("initial etcd v2 connection check failed: %w", err)
			return
		}
		logger.Infof("Initial etcd v2 connection check: OK")

		for {
			select {
			case <-ctx.Done():
				logger.Infof("Stopping etcd v2 connection auditor")
				return
			case <-ticker.C:
				if err := c.checkConnection(ctx); err != nil {
					logger.Errorf("Etcd v2 connection check failed: %v", err)
					errCh <- fmt.Errorf("etcd v2 connection check failed: %w", err)
					return
				}
				logger.Debugf("Etcd v2 connection check: OK")
			}
		}
	}()

	return errCh
}

// checkConnection performs a health check on the etcd v2 connection
func (c *ClientV2) checkConnection(ctx context.Context) error {
	// Create a timeout context for the health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get the version endpoint to verify connectivity
	// For v2, we can try to get a key that doesn't exist at a known path
	// If the service is unreachable, we'll get a network error instead of "key not found"
	_, err := c.client.Get(checkCtx, "/__etcdfinder_health_check__", nil)

	// If we get a key not found error, that's actually good - it means the service is reachable
	if err != nil && !etcdv2.IsKeyNotFound(err) {
		// Check if it's a network/connectivity error
		if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("failed to connect to etcd v2: %w", err)
		}
	}

	return nil
}

// Close closes the etcd v2 client connection
func (c *ClientV2) Close() error {
	// The etcd v2 client doesn't have an explicit Close method
	// The connection is managed by the HTTP transport
	return nil
}
