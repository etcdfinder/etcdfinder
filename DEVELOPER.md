# Development Guide

This guide will help you set up your development environment and start contributing to **etcdfinder**.

## Prerequisites

Before you begin, ensure you have the following installed:
- [Go](https://golang.org/doc/install) (version 1.25 or later recommended)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Getting Started

Follow these steps to get the project running locally:

### 1. Start Infrastructure

The project relies on **etcdfinder-ui**, **etcd** and **Meilisearch**. You can start these services using Docker Compose from the root directory:

```bash
docker-compose up -d
```

This will start:
- **etcdfinder-ui**: Accessible at `http://localhost:3000`
- **etcd**: Accessible at `http://localhost:22379`
- **Meilisearch**: Accessible at `http://localhost:7700`

### 2. Run the Application

Once the infrastructure is up, you can start the **etcdfinder** backend:

```bash
go run main.go
```

By default, the API will be available at `http://localhost:8080`.

And UI will be available at `http://localhost:3000`.

## Configuration

The application's default configuration is located at `internal/config/config.yaml`. You can modify this file to change ports, endpoints, or other settings.

Environment variables can also be used to override these settings (e.g., `ETCDF_SERVER_PORT=9090`).

## Adding Data to etcd

To test the search functionality, you can add some sample data to etcd using `etcdctl` (if installed) or by using the etcdctl-ui at `http://localhost:3000`:

```bash
# Example using etcdctl (pointing to the local docker instance)
etcdctl --endpoints=localhost:22379 put /config/app/database "postgres://user:pass@localhost:5432/db"
etcdctl --endpoints=localhost:22379 put /config/app/api_key "secret-key-123"
```
