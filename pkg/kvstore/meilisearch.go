package kvstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cespare/xxhash/v2"
	"github.com/etcdfinder/etcdfinder/internal/lib"
	"github.com/etcdfinder/etcdfinder/pkg/common"
	"github.com/etcdfinder/etcdfinder/pkg/logger"
	"github.com/meilisearch/meilisearch-go"
)

// MeilisearchStore implements the KVStore interface using Meilisearch
type MeilisearchStore struct {
	client           meilisearch.ServiceManager
	indexName        string
	matchingStrategy meilisearch.MatchingStrategy
}

func makeID(key string) string {
	return strconv.FormatUint(xxhash.Sum64String(key), 36)
}

func createDocument(key string, value string) map[string]any {
	return map[string]any{
		lib.ID_CONSTANT:    makeID(key), // Meilisearch uses 'id' as the default primary key
		lib.KEY_CONSTANT:   key,
		lib.VALUE_CONSTANT: value,
	}
}

// NewMeilisearchStore creates a new Meilisearch-backed KVStore
func NewMeilisearchStore(host, indexName, matchingStrategy string) (KVStore, error) {
	client := meilisearch.New(host)

	// Delete existing index to start fresh
	if _, err := client.DeleteIndex(indexName); err != nil {
		// Ignore error as index might not exist
		// In a production app we might want to check the error type
		logger.Errorf("Failed to delete existing index: %v", err)
		return nil, err
	}

	_, err := client.Index(indexName).UpdateSettings(&meilisearch.Settings{
		RankingRules: []string{
			"words",
			"exactness",
			"typo",
			"proximity",
			"attribute",
			"sort",
		},
		SearchableAttributes: []string{
			lib.KEY_CONSTANT,
		},
		FilterableAttributes: []string{
			lib.KEY_CONSTANT,
		},
	})
	if err != nil {
		logger.Errorf("Failed to configure index settings: %v", err)
		return nil, err
	}

	return &MeilisearchStore{
		client:           client,
		indexName:        indexName,
		matchingStrategy: meilisearch.MatchingStrategy(matchingStrategy),
	}, nil
}

// Get retrieves the value for a given key
func (ms *MeilisearchStore) Get(ctx context.Context, key string) (string, error) {
	var doc map[string]any
	// Meilisearch GetDocument uses the primary key (id) to retrieve the document
	err := ms.client.Index(ms.indexName).GetDocument(makeID(key), &meilisearch.DocumentQuery{}, &doc)
	if err != nil {
		return "", fmt.Errorf("failed to get document: %w", err)
	}

	val, ok := doc[lib.VALUE_CONSTANT].(string)
	if !ok {
		return "", fmt.Errorf("value field not found or not a string for key: %s", key)
	}
	return val, nil
}

// Put stores or updates a key-value pair
func (ms *MeilisearchStore) Put(ctx context.Context, key string, value string) error {
	doc := createDocument(key, value)
	_, err := ms.client.Index(ms.indexName).AddDocuments([]map[string]any{doc}, nil)
	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}
	return nil
}

// PutBatch stores or updates a batch of key-value pairs
func (ms *MeilisearchStore) PutBatch(ctx context.Context, kvs []common.KV) error {
	items := []map[string]any{}
	for _, kv := range kvs {
		items = append(items, createDocument(kv.Key, kv.Value))
	}
	_, err := ms.client.Index(ms.indexName).AddDocuments(items, nil)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	return nil
}

// Search searches for keys or values matching the search string
func (ms *MeilisearchStore) Search(ctx context.Context, searchStr string) ([]common.KV, error) {
	searchRes, err := ms.client.Index(ms.indexName).Search(searchStr, &meilisearch.SearchRequest{
		Limit:            100, // Set a reasonable limit
		MatchingStrategy: ms.matchingStrategy,
	})
	if err != nil {
		if msErr, ok := err.(*meilisearch.Error); ok && msErr.StatusCode == 404 {
			logger.Infof("Index not found during search, returning empty results: %v", err)
			return []common.KV{}, nil
		}
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var kvs []common.KV
	for _, hit := range searchRes.Hits {
		// hit is map[string]json.RawMessage
		var key, value string

		if rawKey, ok := hit[lib.KEY_CONSTANT]; ok {
			if err := json.Unmarshal(rawKey, &key); err != nil {
				// skip if key cannot be unmarshaled
				continue
			}
		}
		if rawValue, ok := hit[lib.VALUE_CONSTANT]; ok {
			if err := json.Unmarshal(rawValue, &value); err != nil {
				// skip if value cannot be unmarshaled
				continue
			}
		}

		if key != "" && value != "" {
			kvs = append(kvs, common.KV{
				Key:   key,
				Value: value,
			})
		}
	}

	return kvs, nil
}

// Delete removes a key-value pair
func (ms *MeilisearchStore) Delete(ctx context.Context, key string) error {
	_, err := ms.client.Index(ms.indexName).DeleteDocument(makeID(key))
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// Close closes the Meilisearch client
func (ms *MeilisearchStore) Close(ctx context.Context) error {
	// Meilisearch client doesn't need explicit closing as it uses http.Client
	return nil
}
