# Configuration Reference

This document describes all available configuration options for etcdfinder. Configuration can be provided via YAML file or environment variables.

## Configuration Precedence

Configuration values are loaded in the following order (highest priority first):
1. **Environment Variables** - Override any YAML configuration
2. **YAML Configuration File** - Default configuration file

---

## Server Configuration

HTTP server settings.

| YAML Path | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `server.port` | `SERVER_PORT` | string | `8080` | HTTP server port |

**Example YAML:**
```yaml
server:
  port: 8080
```

**Example Environment Variable:**
```bash
export SERVER_PORT=9000
```

---

## Logging Configuration

Application logging settings.

| YAML Path | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `log.level` | `LOG_LEVEL` | string | `info` | Log level (`debug`, `info`) |

**Example YAML:**
```yaml
log:
  level: info
```

**Example Environment Variable:**
```bash
export LOG_LEVEL=debug
```

---

## Etcd Configuration

Etcd client and behavior settings.

| YAML Path | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `etcd.version` | `ETCD_VERSION` | string | `v3` | Etcd API version (`v2`, `v3`) |
| `etcd.endpoints` | `ETCD_ENDPOINTS` | string | `http://localhost:22379` | Comma-separated etcd endpoints |
| `etcd.root_etcd_prefix` | `ETCD_ROOT_ETCD_PREFIX` | string | `""` | Root prefix for etcd keys to watch/index |
| `etcd.watch_event_channel_size` | `ETCD_WATCH_EVENT_CHANNEL_SIZE` | int64 | `100` | Buffer size for watch event channel (if addition/change in etcd values is very frequent, consider increasing this value) |
| `etcd.pagination_limit` | `ETCD_PAGINATION_LIMIT` | int64 | `10000` | Maximum keys to fetch per pagination request |
| `etcd.etcd_audit_period` | `ETCD_ETCD_AUDIT_PERIOD` | int64 | `60` | Period (in seconds) for etcd connection audit sync |
| `etcd.max_watch_retries` | `ETCD_MAX_WATCH_RETRIES` | int64 | `5` | Maximum consecutive watch retry attempts for expected modindex before exiting |

**Example YAML:**
```yaml
etcd:
  version: v3
  endpoints: http://localhost:22379
  root_etcd_prefix: ""
  watch_event_channel_size: 100
  pagination_limit: 10000
  etcd_audit_period: 60
  max_watch_retries: 5
```

**Example Environment Variables:**
```bash
export ETCD_VERSION=v3
export ETCD_ENDPOINTS=http://etcd-1:2379,http://etcd-2:2379
export ETCD_ROOT_ETCD_PREFIX=/myapp
export ETCD_WATCH_EVENT_CHANNEL_SIZE=200
export ETCD_PAGINATION_LIMIT=5000
export ETCD_ETCD_AUDIT_PERIOD=120
export ETCD_MAX_WATCH_RETRIES=10
```

---

## Datastore Configuration

Search backend configuration (Meilisearch).

| YAML Path | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `datastore.type` | `DATASTORE_TYPE` | string | `meilisearch` | Datastore type (currently `meilisearch`) |
| `datastore.meilisearch.host` | `DATASTORE_MEILISEARCH_HOST` | string | `http://localhost:7700` | Meilisearch server URL |
| `datastore.meilisearch.index_name` | `DATASTORE_MEILISEARCH_INDEX_NAME` | string | `etcd-keys` | Meilisearch index name |
| `datastore.meilisearch.matching_strategy` | `DATASTORE_MEILISEARCH_MATCHING_STRATEGY` | string | `frequency` | Meilisearch matching strategy |

**Example YAML:**
```yaml
datastore:
  type: meilisearch
  meilisearch:
    host: http://localhost:7700
    index_name: etcd-keys
    matching_strategy: frequency
```

**Example Environment Variables:**
```bash
export DATASTORE_TYPE=meilisearch
export DATASTORE_MEILISEARCH_HOST=http://meilisearch:7700
export DATASTORE_MEILISEARCH_INDEX_NAME=my-etcd-index
export DATASTORE_MEILISEARCH_MATCHING_STRATEGY=all
```
