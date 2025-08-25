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
```