package etcd

import (
	"context"

	"github.com/etcdfinder/etcdfinder/pkg/common"
)

type BaseClient interface {
	// returns the value of the key and error if any
	Get(ctx context.Context, key string) (string, error)
	// returns the key that was put and error if any
	Put(ctx context.Context, key string, value string) (string, error)
	// returns the key that was deleted and error if any
	Delete(ctx context.Context, key string) (string, error)
	// returns the channel of watch events and error channel
	Watch(ctx context.Context) (<-chan WatchEvent, <-chan error)
	// returns the list of keys and the next key to be fetched and error if any
	GetKeysWithPagination(ctx context.Context, fromKey string) ([]common.KV, string, error)
	// returns the error channel
	StartAuditor(ctx context.Context) <-chan error
	// closes the client
	Close() error
}