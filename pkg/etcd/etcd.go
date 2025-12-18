package etcd

import (
	"context"
	"fmt"

	"github.com/etcdfinder/etcdfinder/internal/customerrors"
	"github.com/etcdfinder/etcdfinder/pkg/common"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type BaseClient interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key string, value string) error
	Delete(ctx context.Context, key string) error
	Watch(ctx context.Context) (<-chan WatchEvent, <-chan error)
	GetKeysWithPagination(ctx context.Context, fromKey string) ([]common.KV, string, error)
	Close() error
}

// Client wraps the etcd client with custom functionality
type Client struct {
	client                *clientv3.Client
	watchEventChannelSize int64  // size of the watch event channel
	rootPrefixEtcd        string // prefix of the etcd keys to be watched
	numGetKeysLimit       int64  // number of keys to be returned in a single GetKeysWithPagination call
}

// WatchEvent represents a change event from etcd
type WatchEvent struct {
	Type  string
	Key   string
	Value string
}

// NewClient creates a new etcd client
func NewClient(
	endpoints []string,
	watchEventChannelSize int64,
	rootPrefixEtcd string,
	numGetKeysLimit int64) (BaseClient, error) {
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

func (c *Client) Put(ctx context.Context, key string, value string) error {
	_, err := c.client.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put key: %w", err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// WatchPrefix watches for changes on keys
// Returns a channel of WatchEvents and an error channel
func (c *Client) Watch(ctx context.Context) (<-chan WatchEvent, <-chan error) {
	eventCh := make(chan WatchEvent, c.watchEventChannelSize)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)

		watchChan := c.client.Watch(ctx, c.rootPrefixEtcd, clientv3.WithPrefix())

		for watchResp := range watchChan {
			if watchResp.Err() != nil {
				errCh <- fmt.Errorf("watch error: %w", watchResp.Err())
				return
			}

			for _, event := range watchResp.Events {
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

// Close closes the etcd client connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
