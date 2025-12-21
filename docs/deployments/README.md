# Deployments

## Quickstart

To try out etcdfinder, use the quickstart deployment. It runs with etcd and Meilisearch.

See [quickstart deployment guide](quickstart/deployment-guide.md)

## Docker Compose

Quick docker compose deployment with Meilisearch.

See [docker-compose deployment guide](docker-compose/deployment-guide.md)

## Kubernetes

Coming soon...

## Configuration

Use a custom config file:
```bash
./etcdfinder --config /etc/etcdfinder/config.yaml
```

## Optimization Tips

Adjust these settings based on your use case:

- `pagination_limit`: Higher values speed up initial sync but use more memory (default: 10000)
- `watch_event_channel_size`: Increase if you have high write throughput to avoid backpressure (default: 100)
- `log.level`: Use `info` in production; `debug` for troubleshooting
