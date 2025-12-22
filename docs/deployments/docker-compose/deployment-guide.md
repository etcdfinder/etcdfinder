# Docker Compose Deployment

Deploy etcdfinder locally with etcd and Meilisearch using Docker Compose.

## Update your etcd configuration

To use your etcd cluster, update the config:

```yaml
etcd:
  endpoints: your-etcd-host:2379  # Point to your etcd
```

## Quick Start

This deployment requires you to provide your own etcd endpoints via the `ETCD_ENDPOINTS` environment variable.

```bash
# Run with a single endpoint
ETCD_ENDPOINTS=http://your-etcd-host:2379 docker-compose up -d

# Run with multiple endpoints
ETCD_ENDPOINTS=http://etcd1:2379,http://etcd2:2379 docker-compose up -d
```

If you don't have an existing etcd cluster and want to test etcdfinder with a local etcd instance, use the **Quickstart Deployment** instead:

```bash
cd ../quickstart && docker-compose up -d
```

## Configuration

The compose setup uses default configuration. To customize:

1. Edit `internal/config/config.yaml`
2. Or mount a custom config:

```yaml
services:
  etcdfinder:
    volumes:
      - ./custom-config.yaml:/app/config.yaml
    command: ["./etcdfinder", "--config", "/app/config.yaml"]
```

## Services

The compose file includes:

- **meilisearch**: Port 7700
- **etcdfinder**: Port 8080

## Accessing the API

```bash
# Search for keys
curl -X POST http://localhost:8080/v1/search-keys \
  -H "Content-Type: application/json" \
  -d '{"search_str": "config"}'

# Put a key
curl -X POST http://localhost:8080/v1/put-key \
  -H "Content-Type: application/json" \
  -d '{"key": "/app/test", "value": "hello"}'

# Search again
curl -X POST http://localhost:8080/v1/search-keys \
  -H "Content-Type: application/json" \
  -d '{"search_str": "test"}'
```
