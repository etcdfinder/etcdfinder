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
> **Event Consistency Strategy**: The watch mechanism tracks the `ModRevision` of each event to detect gaps in the event stream. If a ModRevision mismatch is detected (meaning events were missed due to network issues or other failures), the watch automatically restarts from the last successfully processed revision using etcd's `WithRev()` option. This ensures no events are lost without requiring a full application restart. Only critical watch errors (e.g., etcd connection failures detected by the error channel) will cause the application to exit and perform a full re-sync.

### Consistency Guarantees

- **Writes**: Go to etcd first, then to search index
- **Reads**: Always from etcd (source of truth)
- **Search**: From Meilisearch (may lag slightly behind etcd)

### Connection Health Monitoring

To ensure the system maintains eventual consistency, we've implemented an **etcd connection auditor** that runs as a background goroutine. This auditor enhances our fail-fast strategy and improves overall system reliability.

#### Why This Improves Eventual Consistency

Connection failures are detected within 1 minute:
- The application exits and restarts automatically
- Upon restart, the full re-sync process ensures Meilisearch catches up with etcd's latest state
- This prevents prolonged periods of staleness in the search index

### Monitoring

Use `/v1/ingestion-delay` to check sync lag in milliseconds.
