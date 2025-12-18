package kvstore

import (
	"context"

	"github.com/etcdfinder/etcdfinder/pkg/common"
)

type KVStore interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	PutBatch(ctx context.Context, kvs []common.KV) error
	Search(ctx context.Context, searchStr string) ([]common.KV, error)
	Delete(ctx context.Context, key string) error
	Close(ctx context.Context) error
}
