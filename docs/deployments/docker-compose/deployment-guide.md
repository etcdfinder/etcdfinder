# Docker Compose Deployment

Deploy etcdfinder with your existing etcd cluster using Docker Compose.

## Overview

This directory contains two docker-compose files to support both etcd v2 and v3:

- **`docker-compose-v2.yaml`**: For etcd v2 clusters
- **`docker-compose-v3.yaml`**: For etcd v3 clusters (recommended)

Choose the file that matches your etcd cluster version.

> [!IMPORTANT]
> If you don't have an existing etcd cluster and want to test etcdfinder with a local etcd instance, use the **Quickstart Deployment** instead:
> ```bash
> cd ../quickstart && docker-compose up -d
> ```

## Quick Start

### For etcd v3 (Recommended)

```bash
# Run with a single endpoint
ETCD_ENDPOINTS=http://your-etcd-host:2379 docker-compose -f docker-compose-v3.yaml up -d

# Run with multiple endpoints
ETCD_ENDPOINTS=http://etcd1:2379,http://etcd2:2379 docker-compose -f docker-compose-v3.yaml up -d
```

### For etcd v2

```bash
# Run with a single endpoint
ETCD_ENDPOINTS=http://your-etcd-host:2379 docker-compose -f docker-compose-v2.yaml up -d

# Run with multiple endpoints
ETCD_ENDPOINTS=http://etcd1:2379,http://etcd2:2379 docker-compose -f docker-compose-v2.yaml up -d
```

## Configuration

The compose setup uses default configuration with the etcd version set via the `ETCD_VERSION` environment variable. To customize further:

1. Edit `internal/config/config.yaml` to change default settings
2. Or mount a custom config:

```yaml
services:
  etcdfinder:
    volumes:
      - ./custom-config.yaml:/app/config.yaml
    command: ["./etcdfinder", "--config", "/app/config.yaml"]
```

## Services

Both compose files include:

- **etcdfinder-ui**: Web UI on port 3000
- **etcdfinder**: Backend API on port 8080
- **meilisearch**: Search engine on port 7700

## Accessing the Services

### Web UI

Open your browser to `http://localhost:3000` to access the etcdfinder web interface.

### API

```bash
# Search for keys
curl -X POST http://localhost:8080/v1/search-keys \
  -H "Content-Type: application/json" \
  -d '{"search_str": "config"}'

# Put a key
curl -X POST http://localhost:8080/v1/put-key \
  -H "Content-Type: application/json" \
  -d '{"key": "/app/test", "value": "hello"}'

# Get a key
curl -X GET http://localhost:8080/v1/get-key \
  -H "Content-Type: application/json" \
  -d '{"key": "/app/test"}'

# Search again
curl -X POST http://localhost:8080/v1/search-keys \
  -H "Content-Type: application/json" \
  -d '{"search_str": "test"}'
```

## Stopping the Services

```bash
# For v3
ETCD_ENDPOINTS=http://your-etcd-host:2379 docker-compose -f docker-compose-v3.yaml down

# For v2
ETCD_ENDPOINTS=http://your-etcd-host:2379 docker-compose -f docker-compose-v2.yaml down

# To also remove volumes
ETCD_ENDPOINTS=http://your-etcd-host:2379 docker-compose -f docker-compose-v3.yaml down -v
```
