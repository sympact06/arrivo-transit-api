# Contributing to Arrivo Transit API

🚌 Bedankt voor je interesse in het bijdragen aan Arrivo Transit API!

## 📋 Voordat je begint

Dit is een **proprietary/closed source** project van Arrivo. Bijdragen zijn alleen toegestaan voor geautoriseerde teamleden.

## 🔐 Toegang

Om bij te dragen aan dit project heb je nodig:
- Toegang tot de private repository op GitHub
- Arrivo teamlidmaatschap
- Ondertekende NDA (Non-Disclosure Agreement)

## 🛠️ Development Setup

### Vereisten
- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+
- Git

### Lokale Setup

```bash
# Clone de repository
git clone https://github.com/goarrivo/transit-api.git
cd transit-api

# Start dependencies
docker-compose up -d postgres redis

# Install dependencies
go mod download

# Run migrations
./run_migrations.sh

# Start de API
go run cmd/api/main.go
```

## 📝 Code Guidelines

### Go Code Style
- Volg de [Effective Go](https://golang.org/doc/effective_go.html) richtlijnen
- Gebruik `gofmt` en `goimports` voor formatting
- Run `go vet` en `staticcheck` voor linting
- Schrijf tests voor nieuwe functionaliteit
- Documenteer publieke functies en types

### Commit Messages
Gebruik [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: voeg realtime tracking toe voor bussen
fix: los race condition op in cache layer
docs: update API documentatie voor nieuwe endpoints
test: voeg integration tests toe voor GTFS ingestor
```

### Branch Naming
```
feature/realtime-tracking
bugfix/cache-race-condition
hotfix/critical-security-patch
docs/api-documentation-update
```

## 🔄 Development Workflow

### 1. Issue Creation
- Gebruik de issue templates voor bug reports en feature requests
- Label issues appropriaat (bug, enhancement, documentation, etc.)
- Assign aan jezelf als je eraan gaat werken

### 2. Branch Creation
```bash
# Maak een nieuwe branch vanaf main
git checkout main
git pull origin main
git checkout -b feature/nieuwe-functionaliteit
```

### 3. Development
- Schrijf clean, geteste code
- Volg de bestaande architectuur patronen
- Update documentatie waar nodig
- Test lokaal voordat je pusht

### 4. Testing
```bash
# Run alle tests
go test ./...

# Run tests met coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linting
go vet ./...
staticcheck ./...
```

### 5. Pull Request
- Gebruik de PR template
- Zorg voor een duidelijke beschrijving
- Link gerelateerde issues
- Request review van teamleden
- Zorg dat alle CI checks slagen

## 🏗️ Architecture Guidelines

### Microservices Patterns
- **API Gateway**: Centrale entry point voor alle requests
- **Service Discovery**: Gebruik Consul of Kubernetes DNS
- **Circuit Breakers**: Implementeer met hystrix-go
- **Distributed Tracing**: OpenTelemetry met Jaeger

### Database Guidelines
- Gebruik migrations voor schema changes
- Implementeer proper indexing voor performance
- Gebruik connection pooling (pgxpool)
- Volg database naming conventions

### Caching Strategy
- **L1**: In-memory cache (15s TTL)
- **L2**: Redis cache (60s TTL)
- **L3**: Edge cache via Cloudflare (30s TTL)

### Error Handling
```go
// Gebruik structured errors
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

// Wrap errors met context
if err != nil {
    return fmt.Errorf("failed to fetch stops: %w", err)
}
```

## 📊 Monitoring & Observability

### Metrics
- Gebruik Prometheus voor metrics
- Implementeer custom business metrics
- Monitor SLA compliance (99.9% uptime)

### Logging
```go
// Gebruik structured logging
log.Info("processing GTFS data",
    "feed_id", feedID,
    "records", recordCount,
    "duration", duration,
)
```

### Tracing
- Implementeer distributed tracing
- Trace kritieke user journeys
- Monitor database query performance

## 🚀 Deployment

### Environments
- **Development**: Lokale development
- **Staging**: `staging-api.goarrivo.nl`
- **Production**: `api.goarrivo.nl`

### CI/CD Pipeline
1. **Test**: Unit tests, integration tests, linting
2. **Build**: Compile binaries, build Docker images
3. **Security**: Vulnerability scanning met Trivy
4. **Deploy**: Automated deployment via GitHub Actions

## 🔒 Security Guidelines

### API Security
- Implementeer rate limiting per API key
- Gebruik HTTPS everywhere
- Valideer alle input parameters
- Implementeer proper authentication

### Data Protection
- Geen PII in logs
- Encrypt sensitive data at rest
- Gebruik secrets management (Kubernetes secrets)
- Regular security audits

## 📚 Documentation

### Code Documentation
- Documenteer alle publieke APIs
- Gebruik godoc comments
- Schrijf README voor elke service

### API Documentation
- Maintain OpenAPI/Swagger specs
- Include request/response examples
- Document error codes en responses

## 🐛 Bug Reports

Gebruik de bug report template en include:
- Stappen om te reproduceren
- Verwacht vs actueel gedrag
- Environment details
- Logs en error messages
- Screenshots indien relevant

## 💡 Feature Requests

Gebruik de feature request template en include:
- Probleem beschrijving
- Voorgestelde oplossing
- Alternatieven overwogen
- Business impact
- Technical considerations

## 📞 Contact

Voor vragen over contributing:
- 💬 Slack: #arrivo-dev
- 📧 Email: dev@goarrivo.nl
- 🐛 Issues: GitHub Issues

## 📄 License

Door bij te dragen ga je akkoord met de proprietary license van dit project. Alle bijdragen worden eigendom van Arrivo.

---

**Belangrijk**: Dit project bevat vertrouwelijke informatie. Deel geen code, documentatie of andere projectgerelateerde informatie buiten het geautoriseerde team.