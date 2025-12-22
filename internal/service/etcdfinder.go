package service

import (
	"context"

	"github.com/etcdfinder/etcdfinder/internal/ingestor"
	"github.com/etcdfinder/etcdfinder/pkg/etcd"
	"github.com/etcdfinder/etcdfinder/pkg/kvstore"
)

type Etcdfinder interface {
	GetKey(ctx context.Context, key string) (string, error)
	SearchKeys(ctx context.Context, searchStr string) ([]string, error)
	PutKey(ctx context.Context, key string, value string) error
	DeleteKey(ctx context.Context, key string) error
	GetIngestionDelay(ctx context.Context) int
}

type DefaultEtcdfinder struct {
	etcdClt     etcd.BaseClient
	kvStore     kvstore.KVStore
	ingestorClt ingestor.Base
}

func NewDefaultEtcdfinder(etcdClt etcd.BaseClient, kvStore kvstore.KVStore, ingestorClt ingestor.Base) Etcdfinder {
	return &DefaultEtcdfinder{
		etcdClt:     etcdClt,
		kvStore:     kvStore,
		ingestorClt: ingestorClt,
	}
}

func (d *DefaultEtcdfinder) GetKey(ctx context.Context, key string) (string, error) {
	// Always Read from kvStore
	return d.etcdClt.Get(ctx, key)
}

func (d *DefaultEtcdfinder) SearchKeys(ctx context.Context, searchStr string) ([]string, error) {
	var keys []string
	kvs, err := d.kvStore.Search(ctx, searchStr)
	if err != nil {
		return nil, err
	}
	for _, kv := range kvs {
		keys = append(keys, kv.Key)
	}
	return keys, nil
}

func (d *DefaultEtcdfinder) PutKey(ctx context.Context, key string, value string) error {
	key, err := d.etcdClt.Put(ctx, key, value)
	if err != nil {
		return err
	}
	return d.kvStore.Put(ctx, key, value)
}

func (d *DefaultEtcdfinder) DeleteKey(ctx context.Context, key string) error {
	key, err := d.etcdClt.Delete(ctx, key)
	if err != nil {
		return err
	}
	return d.kvStore.Delete(ctx, key)
}

func (d *DefaultEtcdfinder) GetIngestionDelay(ctx context.Context) int {
	return d.ingestorClt.GetIngestionDelay(ctx)
}
