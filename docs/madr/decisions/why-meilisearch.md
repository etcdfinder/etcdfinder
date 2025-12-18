# Why Meilisearch

## Status

Accepted

## Alternatives Considered

**Elasticsearch**
- Search wasn't working correctly for our use case because fuzziness parameter uses Levenshtein edit distance with a max of 2 edits, which is intentionally limited for performance
- This restriction caused issues with our fuzzy search requirements

**Bleve**
- No batch write support
- Made initial sync very slow when loading thousands of keys from etcd

## Decision

Use Meilisearch as the search backend.
- Good fuzzy search out of the box
- Supports batch operations for efficient bulk indexing

## Notes

The `KVStore` interface in `pkg/kvstore/` is designed to be pluggable, so we can swap to a different backend if requirements change.
