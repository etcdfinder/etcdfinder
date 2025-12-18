# API Reference

All endpoints accept and return JSON. Requests use POST method with JSON body.

## Search Keys

**POST** `/v1/search-keys`

Search for keys using full-text search.

**Request:**
```json
{
  "search_str": "config"
}
```

**Response:**
```json
{
  "keys": [
    "/app/config/database",
    "/app/config/cache"
  ]
}
```

## Get Key

**POST** `/v1/get-key`

Retrieve value for a specific key from etcd.

**Request:**
```json
{
  "key": "/app/config/database"
}
```

**Response:**
```json
{
  "key": "/app/config/database",
  "value": "postgresql://..."
}
```

## Put Key

**POST** `/v1/put-key`

Create or update a key-value pair.

**Request:**
```json
{
  "key": "/app/config/database",
  "value": "postgresql://..."
}
```

**Response:**
```json
{
  "key": "/app/config/database",
  "value": "postgresql://..."
}
```

## Delete Key

**POST** `/v1/delete-key`

Delete a key from etcd.

**Request:**
```json
{
  "key": "/app/config/database"
}
```

**Response:**
```json
{
  "key": "/app/config/database"
}
```

## Get Ingestion Delay

**GET** `/v1/ingestion-delay`

Returns the current ingestion delay in milliseconds (time lag between etcd and search index).

**Response:**
```json
{
  "ingestion_delay": 42
}
```

## Error Responses

All endpoints return standard error format:

```json
{
  "error": "KEY_NOT_FOUND",
  "message": "Key not found in etcd"
}
```

Common error codes:
- `BAD_REQUEST` - Invalid request format
- `KEY_NOT_FOUND` - Key does not exist
- `INTERNAL_ERROR` - Server error
