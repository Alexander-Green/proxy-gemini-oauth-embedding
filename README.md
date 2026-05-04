# Google Embeddings Proxy

OpenAI-compatible proxy for Google Gemini Embeddings API. Allows PostClaw to use Google embeddings without modification.

## Architecture

Based on the finance-service project structure:
- `cmd/main.go` - entry point with zap logger and graceful shutdown
- `internal/config` - configuration from `config.json` file
- `internal/client` - Google API client
- `internal/handler` - HTTP handlers
- `internal/model` - request/response types

## Configuration

Configuration is loaded from `config.json` file. Copy the example and fill in your values:

```bash
cp config.example.json config.json
# Edit config.json with your Google API key
```

**config.json:**
```json
{
  "google_api_key": "your-google-api-key-here",
  "google_embedding_model": "gemini-embedding-2",
  "embedding_dim": 768,
  "port": "1234",
  "log_level": "INFO",
  "log_payloads": false,
  "shutdown_timeout_sec": 30,
  "read_timeout_sec": 30,
  "write_timeout_sec": 30
}
```

Required fields:
- `google_api_key` - Google API key

Optional fields (with defaults):
- `google_embedding_model` - Model name (default: gemini-embedding-2)
- `embedding_dim` - Embedding dimension (default: 768)
- `port` - Service port (default: 1234)
- `log_payloads` - Enable payload logging (default: false)
- `shutdown_timeout_sec` - Graceful shutdown timeout (default: 30)
- `read_timeout_sec` - HTTP read timeout (default: 30)
- `write_timeout_sec` - HTTP write timeout (default: 30)

## Usage

```bash
go run cmd/main.go
```

## Testing

### Test health endpoint:
```bash
curl http://localhost:1234/health
```

**Response:**
```json
{
  "status": "ok",
  "model": "gemini-embedding-2",
  "dim": 768
}
```

### Test embeddings endpoint (single text):
```bash
curl -X POST http://localhost:1234/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello, world!",
    "model": "gemini-embedding-2"
  }'
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.1, 0.2, ...],
      "index": 0
    }
  ],
  "model": "gemini-embedding-2",
  "usage": {
    "prompt_tokens": 11,
    "total_tokens": 11
  }
}
```

### Test embeddings endpoint (multiple texts):
```bash
curl -X POST http://localhost:1234/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": ["Hello, world!", "How are you?"],
    "model": "gemini-embedding-2"
  }'
```

## API Endpoints

### POST /v1/embeddings

OpenAI-compatible embeddings endpoint.

**Request:**
```json
{
  "input": "text to embed",
  "model": "gemini-embedding-2"
}
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.1, 0.2, ...],
      "index": 0
    }
  ],
  "model": "gemini-embedding-2",
  "usage": {
    "prompt_tokens": 11,
    "total_tokens": 11
  }
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "model": "gemini-embedding-2",
  "dim": 768
}
```

## Integration with PostClaw

1. Update `~/.openclaw/openclaw.json`:
```json
{
  "agents": {
    "defaults": {
      "memorySearch": {
        "model": "gemini-embedding-2",
        "remote": {
          "baseUrl": "http://localhost:1234/v1"
        }
      }
    }
  }
}
```

2. Update PostClaw database:
```sql
UPDATE agents 
SET embedding_dimensions = 768 
WHERE agent_id = 'main';
```

3. Restart OpenClaw Gateway:
```bash
openclaw gateway restart
```

## Building

```bash
go build -o google-embeddings-proxy cmd/main.go
```

## Systemd Service

```ini
[Unit]
Description=Google Embeddings Proxy
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/to/proxy
ExecStart=/usr/local/bin/google-embeddings-proxy
Restart=always

[Install]
WantedBy=multi-user.target
```
