package ingestor

import (
	"context"

	"github.com/etcdfinder/etcdfinder/pkg/etcd"
	"github.com/etcdfinder/etcdfinder/pkg/kvstore"
	"github.com/etcdfinder/etcdfinder/pkg/logger"
)

type Base interface {
	InitKVStore(context.Context) error
	ChangeUpdater(context.Context) error
	GetIngestionDelay(context.Context) int
}

type Ingestor struct {
	kvStore    kvstore.KVStore
	etcdClt    etcd.BaseClient
	watchChan  <-chan etcd.WatchEvent
	initDoneCh chan struct{}
}

func NewIngestor(kvStore kvstore.KVStore, etcdClt etcd.BaseClient) Base {
	return &Ingestor{
		kvStore:    kvStore,
		etcdClt:    etcdClt,
		initDoneCh: make(chan struct{}),
	}
}

func (i *Ingestor) InitKVStore(ctx context.Context) error {
	defer close(i.initDoneCh)
	nextKey := ""

	for {
		keys, returnedNextKey, err := i.etcdClt.GetKeysWithPagination(ctx, nextKey)
		if err != nil {
			return err
		}
		// If no keys returned, we've reached the end
		if len(keys) == 0 {
			break
		}
		// Insert all keys from this page into the KVStore
		if err := i.kvStore.PutBatch(ctx, keys); err != nil {
			return err
		}
		logger.Debugf("Inserting %d keys into KVStore", len(keys))
		// Update nextKey for the next iteration
		nextKey = returnedNextKey
		// If no nextKey is returned, we've reached the end
		if nextKey == "" {
			break
		}
	}

	return nil
}

func (i *Ingestor) ChangeUpdater(ctx context.Context) error {
	// Get the watch channel and error channel from etcd
	eventCh, errCh := i.etcdClt.Watch(ctx)
	i.watchChan = eventCh

	// Wait for initialization to complete
	select {
	case <-i.initDoneCh:
	case <-ctx.Done():
		return ctx.Err()
	case err, ok := <-errCh:
		if !ok {
			return nil
		}
		return err
	}

	// Start a goroutine to listen to watch events
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed, exit
				return nil
			}

			logger.Debugf("Received event %s for key %s", event.Type, event.Key)
			// Handle the event based on type
			switch event.Type {
			case "PUT":
				// Update the kvstore with the new/updated key-value
				if err := i.kvStore.Put(ctx, event.Key, event.Value); err != nil {
					// return as it will lead to inconsistent state
					return err
				}
			case "DELETE":
				// Remove the key from kvstore
				if err := i.kvStore.Delete(ctx, event.Key); err != nil {
					// return as it will lead to inconsistent state
					return err
				}
			}

		case err, ok := <-errCh:
			if !ok {
				// Error channel closed, exit
				return nil
			}
			// Return watch error
			return err

		case <-ctx.Done():
			// Context cancelled, exit
			return ctx.Err()
		}
	}

}

func (i *Ingestor) GetIngestionDelay(ctx context.Context) int {
	return len(i.watchChan)
}
