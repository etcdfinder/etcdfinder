package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/etcdfinder/etcdfinder/internal/customerrors"
	"github.com/etcdfinder/etcdfinder/pkg/common"
	"github.com/etcdfinder/etcdfinder/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client wraps the etcd client with custom functionality
type Client struct {
	client                *clientv3.Client
	watchEventChannelSize int64  // size of the watch event channel
	rootPrefixEtcd        string // prefix of the etcd keys to be watched
	numGetKeysLimit       int64  // number of keys to be returned in a single GetKeysWithPagination call
	EtcdAuditPeriod       time.Duration
	maxWatchRetries       int64 // maximum number of consecutive failures on the same ModRevision
	ExpectedModRevision   int64 // expected modified revision of the etcd keys
}

// WatchEvent represents a change event from etcd
type WatchEvent struct {
	Type  string
	Key   string
	Value string
}

// NewClient creates a new etcd client
func NewClientV3(
	endpoints []string,
	watchEventChannelSize int64,
	rootPrefixEtcd string,
	numGetKeysLimit int64,
	etcdAuditPeriod int64,
	maxWatchRetries int64) (BaseClient, error) {
	if numGetKeysLimit <= 0 {
		return nil, fmt.Errorf("numGetKeysLimit must be greater than 0")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Client{
		client:                cli,
		watchEventChannelSize: watchEventChannelSize,
		rootPrefixEtcd:        rootPrefixEtcd,
		numGetKeysLimit:       numGetKeysLimit,
		EtcdAuditPeriod:       time.Duration(etcdAuditPeriod) * time.Second,
		maxWatchRetries:       maxWatchRetries,
		ExpectedModRevision:   -1,
	}, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return "", customerrors.ErrKeyNotFound
	}

	return string(resp.Kvs[0].Value), nil
}

func (c *Client) Put(ctx context.Context, key string, value string) (string, error) {
	_, err := c.client.Put(ctx, key, value)
	if err != nil {
		return "", fmt.Errorf("failed to put key: %w", err)
	}
	return key, nil
}

func (c *Client) Delete(ctx context.Context, key string) (string, error) {
	_, err := c.client.Delete(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to delete key: %w", err)
	}
	return key, nil
}

// WatchPrefix watches for changes on keys
// Returns a channel of WatchEvents and an error channel
func (c *Client) Watch(ctx context.Context) (<-chan WatchEvent, <-chan error) {
	eventCh := make(chan WatchEvent, c.watchEventChannelSize)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)
		watchRevisionDiscrepancy := false
		var consecutiveFailureCount int64

		for {
			var watchChan clientv3.WatchChan
			if watchRevisionDiscrepancy {
				watchChan = c.client.Watch(
					ctx,
					c.rootPrefixEtcd,
					clientv3.WithPrefix(),
					clientv3.WithRev(c.ExpectedModRevision))
			} else {
				watchChan = c.client.Watch(
					ctx,
					c.rootPrefixEtcd,
					clientv3.WithPrefix())
			}

			for watchResp := range watchChan {
				if watchResp.Err() != nil {
					errCh <- fmt.Errorf("watch error: %w", watchResp.Err())
					return
				}

				for _, event := range watchResp.Events {
					// if this is the first event, set the expected modrevision to the current modrevision
					if c.ExpectedModRevision == -1 {
						c.ExpectedModRevision = event.Kv.ModRevision
					}
					// check if the modrevision is not equal to the expected modrevision
					// which is the last modrevision + 1
					// it means that some event must have been missed due to some network issues
					// so it will break the loop and restart the watch to ensure consistency
					if event.Kv.ModRevision != c.ExpectedModRevision {
						consecutiveFailureCount++
						logger.Warnf("ModRevision mismatch: Consecutive failure #%d on ModRevision %d", consecutiveFailureCount, c.ExpectedModRevision)

						// If we've exceeded max retries on the same revision, fail fast
						if consecutiveFailureCount >= c.maxWatchRetries {
							errCh <- fmt.Errorf("exceeded max watch retries (%d) on ModRevision %d - failing fast to prevent infinite loop", c.maxWatchRetries, c.ExpectedModRevision)
							return
						}
						watchRevisionDiscrepancy = true
						break
					}

					// Successfully processed an event, reset failure counter
					consecutiveFailureCount = 0
					c.ExpectedModRevision = event.Kv.ModRevision + 1

					watchEvent := WatchEvent{
						Key: string(event.Kv.Key),
					}

					switch event.Type {
					case clientv3.EventTypePut:
						watchEvent.Type = "PUT"
						watchEvent.Value = string(event.Kv.Value)
					case clientv3.EventTypeDelete:
						watchEvent.Type = "DELETE"
					}				

					select {
					case eventCh <- watchEvent:
					case <-ctx.Done():
						return
					}
				}

				if watchRevisionDiscrepancy {
					break
				}
			}
		}
	}()

	return eventCh, errCh
}

// GetKeysWithPagination retrieves keys with pagination support
func (c *Client) GetKeysWithPagination(ctx context.Context, fromKey string) ([]common.KV, string, error) {

	opts := []clientv3.OpOption{
		clientv3.WithLimit(c.numGetKeysLimit),
	}
	var key string
	if fromKey != "" {
		key = fromKey
		opts = append(opts, clientv3.WithFromKey())
	} else {
		key = c.rootPrefixEtcd
		opts = append(opts, clientv3.WithPrefix())
	}

	resp, err := c.client.Get(ctx, key, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get keys: %w", err)
	}

	keys := make([]common.KV, 0)

	for _, kv := range resp.Kvs {
		if string(kv.Key) == fromKey {
			continue
		}

		keys = append(keys, common.KV{
			Key:   string(kv.Key),
			Value: string(kv.Value),
		})
	}

	if len(keys) == 0 {
		return keys, "", nil
	}

	return keys, keys[len(keys)-1].Key, nil
}

// StartAuditor starts a background goroutine that checks etcd connection health every EtcdAuditPeriod
// Returns an error channel that will receive errors if the connection check fails
func (c *Client) StartAuditor(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		ticker := time.NewTicker(c.EtcdAuditPeriod)
		defer ticker.Stop()

		logger.Infof("Starting etcd connection auditor (checking every %s)", c.EtcdAuditPeriod)

		// Perform initial health check
		if err := c.checkConnection(ctx); err != nil {
			logger.Errorf("Initial etcd connection check failed: %v", err)
			errCh <- fmt.Errorf("initial etcd connection check failed: %w", err)
			return
		}
		logger.Infof("Initial etcd connection check: OK")

		for {
			select {
			case <-ctx.Done():
				logger.Infof("Stopping etcd connection auditor")
				return
			case <-ticker.C:
				if err := c.checkConnection(ctx); err != nil {
					logger.Errorf("Etcd connection check failed: %v", err)
					errCh <- fmt.Errorf("etcd connection check failed: %w", err)
					return
				}
				logger.Debugf("Etcd connection check: OK")
			}
		}
	}()

	return errCh
}

// checkConnection performs a health check on the etcd connection
func (c *Client) checkConnection(ctx context.Context) error {
	// Create a timeout context for the health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get the status from etcd to verify connectivity
	_, err := c.client.Status(checkCtx, c.client.Endpoints()[0])
	if err != nil {
		return fmt.Errorf("failed to get etcd status: %w", err)
	}

	return nil
}

// Close closes the etcd client connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
