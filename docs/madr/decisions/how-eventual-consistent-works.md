# How Eventual Consistency Works

## Status

Accepted

## Implementation

etcd is the source of truth. Meilisearch is the search index that stays eventually consistent.

### Startup Flow

1. **Index Recreation** - The Meilisearch client is initialized, and the index is deleted and recreated. This ensures no stale records remain from periods when the application was not running.
2. **Watch Goroutine Starts** - Begins listening to etcd events but waits for the initial sync to complete.
3. **Initial Sync** - Fetches all existing keys from etcd using pagination and writes them to Meilisearch in batches.
4. **Watch Activated** - After the sync completes, the watch goroutine starts processing new events.

This order prevents race conditions where new events could be missed during the initial sync.

### Ongoing Sync

The watch goroutine continuously monitors etcd for changes and applies them (put/delete) to Meilisearch in near real-time.

> [!NOTE]
> **Error Handling Strategy**: If the application fails to handle any watch event, it will exit and restart. This fail-fast mechanism prevents data inconsistency by forcing a full re-sync upon restart. We recognize this as a current limitation and welcome suggestions for more resilient error handling strategies (please open an issue).

### Consistency Guarantees

- **Writes**: Go to etcd first, then to search index
- **Reads**: Always from etcd (source of truth)
- **Search**: From Meilisearch (may lag slightly behind etcd)

### Monitoring

Use `/v1/ingestion-delay` to check sync lag in milliseconds.
