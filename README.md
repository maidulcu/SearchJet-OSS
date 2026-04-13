# SearchJet OSS

A cloud-native, bilingual (Arabic/English) search engine designed for the UAE market. Built in Go with PDPL compliance built-in.

## Vision

"UAE Search Engine OSS" — A cloud-native, bilingual (Arabic/English), PDPL-compliant search platform built in Go. Drop-in alternative to commercial search SaaS, optimized for UAE data, open for global contribution.

> "Built in the UAE, for the UAE, open to the world."

## Features

- **Bilingual Search** — Arabic/English with UAE dialect awareness
- **PDPL Compliant** — Consent logging, audit trail, data retention policies
- **UAE Connectors** — Open Data Portal, Business Directory integration
- **Geo-Aware Ranking** — Emirate-based boosting with distance calculations
- **Arabic NLP** — UAE dialect tokenizer, query expansion, transliteration
- **Prayer Time Context** — Optional ranking signal for time-sensitive queries
- **Pluggable Backend** — Meilisearch (default), ready for ZincSearch/Bleve
- **gRPC Bridge** — Ready for Python AraBERT integration

## Quick Start

```bash
# Clone and run with Docker Compose
make dev

# Search API
curl "http://localhost:8080/v1/search?q=مطعم+dubai&emirate=dubai&lang=ar"

# Index documents
curl -X POST http://localhost:8080/v1/index \
  -H "Content-Type: application/json" \
  -d '{"documents":[{"id":"1","title":"Best Restaurant","body":"Great food","lang":"en"}]}'
```

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/search` | GET | Search with query params (q, lang, emirate, lat, lng, page, limit) |
| `/v1/index` | POST | Index documents (bulk) |
| `/v1/index/:id` | DELETE | Delete document |
| `/v1/health` | GET | Health check |

## Feature Matrix

| Feature | SearchJet OSS | Commercial SaaS |
|---------|--------------|-----------------|
| Full-Text Search | ✅ Meilisearch | ✅ |
| Semantic/Vector | ✅ gRPC bridge ready | ✅ |
| Real-Time Indexing | ✅ Webhook + worker | ✅ |
| Geosearch | ✅ UAE geo ranking | ✅ |
| Faceted Search | ✅ Dynamic filters | ✅ |
| Analytics Dashboard | ✅ Grafana + Prometheus | ✅ |
| Multi-Source/Federated | ✅ Colly crawler | ✅ |
| Arabic NLP | ✅ UAE dialect support | Limited |
| PDPL Compliance | ✅ Built-in | GDPR only |
| Self-Hosted | ✅ 100% | Partial |

## Performance Targets

| Metric | Target | Status |
|--------|--------|--------|
| QPS | 10,000 | Roadmap |
| p95 Latency | <100ms | ✅ |
| p99 Latency | <200ms | ✅ |
| Error Rate | <1% | ✅ |

Run benchmarks:

```bash
k6 run benchmarks/search.js
```

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Fiber API │────▶│  Search Engine  │────▶│ Meilisearch │
│   (Go)      │     │   Interface    │     │  (Backend)  │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                    │
       ▼                   ▼                    ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ PDPL Middle │     │   Arabic NLP │     │  Prometheus │
│   (Audit)   │     │  (Tokenizer) │     │  (Metrics)  │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Configuration

YAML or environment variables:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

meilisearch:
  host: "http://localhost:7700"
  api_key: "masterKey"
  index: "uae-search"

pdpl:
  retention_days: 90
  analytics_days: 365
```

Environment variables: `SEARCHJET_*` prefix (e.g., `SEARCHJET_SERVER_PORT=8080`)

## Development

```bash
make help        # Show available commands
make dev         # Start local Docker Compose
make build      # Build binary
make test       # Run tests
make lint       # Run linter
make docker     # Build Docker image
```

## Deployment

### Docker Compose (Recommended)

```bash
docker compose -f deploy/docker-compose.yml up -d
```

Services: SearchJet, Meilisearch, Redis, Prometheus, Grafana

### Kubernetes

```bash
helm install searchjet ./deploy/helm
```

## Roadmap

- [x] v0.1.0 - Core engine with API
- [x] v0.1.0 - Arabic NLP + connectors
- [ ] v0.2.0 - Vector search with Qdrant
- [ ] v0.3.0 - Admin UI (Next.js)
- [ ] v1.0.0 - Production-ready

## Community

- Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)
- Report issues at https://github.com/uae-search-oss/uae-search-oss/issues

## License

MIT — free forever. Enterprise support/features optional.

---

Built with ❤️ in the UAE
