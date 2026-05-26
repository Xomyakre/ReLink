# 🔗 ReLink - URL Shortener with Analytics

Modern URL shortening service with analytics, built with Go.



##  Quick Start

### Installation

```bash

git clone <your-repo-url>
cd ReLink

# Install dependencies
go mod tidy

# Run the server
go run main.go
```

Server starts at `http://localhost:8080`

### Using the Web Interface

Open `http://localhost:8080` in your browser to access the modern web UI:

- **Shorten URLs** - Paste long URL, get short link instantly
- **View Analytics** - Check click counts and timestamps
- **Copy Links** - One-click copy to clipboard

##  API Endpoints

### Create Short URL
```bash
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url"
}

# Response
{
  "short_url": "http://localhost:8080/abc123",
  "short_code": "abc123"
}
```

### Redirect
```bash
GET /{shortCode}
# Returns HTTP 302 redirect to original URL
```

### Get Analytics
```bash
GET /api/analytics/{shortCode}

# Response
{
  "short_code": "abc123",
  "click_count": 42,
  "first_access": "2024-01-01T00:00:00Z",
  "last_access": "2024-01-15T12:30:00Z"
}
```

### Health Check
```bash
GET /health

# Response
{
  "status": "healthy",
  "database": "connected",
  "analytics_queue": 0
}
```

## ⚙️ Configuration

Configure via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `BASE_URL` | Base URL for short links | `http://localhost:8080` |
| `DB_PATH` | BadgerDB data directory | `./data/badger` |
| `SHORT_CODE_LENGTH` | Short code length (6-10) | `7` |
| `ANALYTICS_QUEUE_SIZE` | Analytics queue buffer size | `1000` |



##  Architecture

```
ReLink/
├── internal/
│   ├── config/         # Configuration management
│   ├── handler/        # HTTP handlers
│   ├── service/        # Business logic
│   ├── repository/     # Data access layer
│   └── models/         # Data models
├── static/             # Web UI (HTML/CSS/JS)
├── data/               # BadgerDB storage (auto-created)
└── main.go             # Application entry point
```



##  Tech Stack

- **Go 1.26** - Programming language
- **BadgerDB v4** - Embedded key-value database
- **Gorilla Mux** - HTTP router
- **Validator v10** - Request validation






```

## 📝 License

MIT
