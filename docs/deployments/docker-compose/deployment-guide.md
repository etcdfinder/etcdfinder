# Docker Compose Deployment

Deploy etcdfinder locally with etcd and Meilisearch using Docker Compose.

## Update your etcd configuration

To use your etcd cluster, update the config:

```yaml
etcd:
  endpoints: your-etcd-host:2379  # Point to your etcd
```

## Quick Start

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f

# Stop services
docker-compose down
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
