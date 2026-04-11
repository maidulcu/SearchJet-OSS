# SearchJet OSS

A cloud-native, bilingual (Arabic/English) search engine designed for the UAE market. Built in Go with PDPL compliance built-in.

## Features

- **Bilingual Search** — Arabic/English with UAE dialect awareness
- **PDPL Compliant** — Consent logging, audit trail, data retention
- **UAE Connectors** — Open Data Portal, Business Directory integration
- **Geo-Aware Ranking** — Emirate-based boosting
- **Pluggable Backend** — Meilisearch (default), ZincSearch support

## Quick Start

```bash
# Clone and run
make dev
curl "http://localhost:8080/v1/search?q=مطعم dubai&emirate=dubai"
```

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/search` | GET | Search with query params |
| `/v1/index` | POST | Index documents |
| `/v1/index/:id` | DELETE | Delete document |
| `/v1/health` | GET | Health check |

## Configuration

YAML or env vars. See `config.example.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

meilisearch:
  host: "http://localhost:7700"
  api_key: ""
  index: "uae-search"

pdpl:
  retention_days: 90
  analytics_days: 365
```

## Development

```bash
make test      # Run tests
make lint     # Run linter
make build    # Build binary
make docker   # Docker build
```

## License

MIT — free forever. See [LICENSE](LICENSE).