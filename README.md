# Arrivo Transit API 🚌

[![CI](https://github.com/goarrivo/arrivo-transit-api/workflows/CI/badge.svg)](https://github.com/goarrivo/arrivo-transit-api/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/goarrivo/arrivo-transit-api)](https://goreportcard.com/report/github.com/goarrivo/arrivo-transit-api)
[![codecov](https://codecov.io/gh/goarrivo/arrivo-transit-api/branch/main/graph/badge.svg)](https://codecov.io/gh/goarrivo/arrivo-transit-api)
[![License: Proprietary](https://img.shields.io/badge/License-Proprietary-red.svg)]()
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)

Een moderne, high-performance real-time transit API platform gebouwd met Go microservices architectuur. Arrivo biedt real-time openbaar vervoer informatie met focus op lage latentie, schaalbaarheid en betrouwbaarheid.

🌐 **Website**: [goarrivo.nl](https://goarrivo.nl)

## ✨ Features

### 🚀 Core Functionaliteit
- **Real-time Vertrektijden**: Live vertrektijden voor alle haltes
- **Nabije Haltes**: GPS-gebaseerde zoekfunctie voor nabijgelegen haltes
- **Live Tracking**: Real-time voertuiglocaties en route-informatie
- **Intelligente Zoekfunctie**: Geavanceerde zoekfunctionaliteit voor haltes en routes
- **Multi-layer Caching**: LRU + Redis + Edge caching voor optimale performance

### 🏗️ Architectuur Highlights
- **Microservices**: Modulaire Go services met duidelijke scheiding van verantwoordelijkheden
- **High Performance**: Sub-100ms response times door intelligente caching strategieën
- **Observability**: Uitgebreide monitoring met Prometheus metrics en structured logging
- **Resilience**: Circuit breakers, retry logic en graceful degradation
- **Security**: API key authenticatie, rate limiting en mTLS voor interne communicatie

### 📊 Data Sources
- **GTFS Static**: Statische transit data (routes, stops, schedules)
- **GTFS Realtime**: Live updates voor vertrektijden en voertuigposities
- **OpenOV/OVapi**: Nederlandse openbaar vervoer data integratie

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### Installation

```bash
# Clone de repository
git clone https://github.com/goarrivo/arrivo-transit-api.git
cd arrivo-transit-api

# Start services met Docker Compose
docker-compose up -d

# Wacht tot services ready zijn
docker-compose logs -f api
```

### Development Setup

```bash
# Install dependencies
go mod download

# Run database migrations
go run cmd/migrate/main.go up

# Start API server
go run cmd/api/main.go

# Start GTFS ingestor (separate terminal)
go run cmd/gtfs-ingestor/main.go
```

## 📖 API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### 🚏 Stops (Haltes)

**Nabije haltes zoeken**
```http
GET /stops/nearby?lat=52.3676&lon=4.9041&radius=500
```

**Halte details**
```http
GET /stops/{stop_id}
```

**Haltes zoeken**
```http
GET /stops/search?q=centraal
```

#### 🚌 Routes (Lijnen)

**Routes zoeken**
```http
GET /routes/search?q=1
```

**Route details**
```http
GET /routes/{route_id}
```

#### ⏰ Real-time Data

**Vertrektijden per halte**
```http
GET /stops/{stop_id}/departures
```

**Live voertuig tracking**
```http
GET /routes/{route_id}/vehicles
```

### Response Format

Alle API responses volgen een consistente JSON structuur:

```json
{
  "data": [...],
  "meta": {
    "count": 10,
    "total": 150,
    "page": 1,
    "cached": true,
    "cache_ttl": 300
  },
  "links": {
    "self": "/api/v1/stops/nearby?lat=52.3676&lon=4.9041",
    "next": "/api/v1/stops/nearby?lat=52.3676&lon=4.9041&page=2"
  }
}
```

## 🏗️ Architectuur

### Service Overzicht

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │────│  Transit API    │────│   GTFS Ingestor │
│   (Cloudflare)  │    │   (Go Service)  │    │   (Go Service)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         └──────────────│  Load Balancer  │──────────────┘
                        │   (Traefik)     │
                        └─────────────────┘
                                 │
                    ┌─────────────────────────────┐
                    │                             │
            ┌───────▼────────┐           ┌───────▼────────┐
            │  PostgreSQL    │           │     Redis      │
            │  (Primary DB)  │           │   (Cache)      │
            └────────────────┘           └────────────────┘
```

### Caching Strategy

1. **L1 Cache (In-Memory LRU)**: 1000 items, 5 min TTL
2. **L2 Cache (Redis)**: 10000 items, 15 min TTL  
3. **L3 Cache (Edge/CDN)**: Cloudflare, 60 min TTL

### Database Schema

```sql
-- Core GTFS entities
CREATE TABLE agencies (...);     -- Transit agencies
CREATE TABLE routes (...);       -- Bus/tram routes  
CREATE TABLE stops (...);        -- Stop locations
CREATE TABLE stop_times (...);   -- Scheduled times
CREATE TABLE trips (...);        -- Individual trips

-- Real-time data
CREATE TABLE vehicle_positions (...);  -- Live vehicle locations
CREATE TABLE trip_updates (...);       -- Schedule deviations
```

## 🧪 Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

### Integration Tests
```bash
# Start test dependencies
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...
```

### Load Testing
```bash
# Install k6
brew install k6

# Run load tests
k6 run scripts/load-test.js
```

## 📊 Monitoring & Observability

### Metrics (Prometheus)
- Request latency (p50, p95, p99)
- Cache hit rates (LRU, Redis)
- Database connection pool stats
- GTFS data freshness
- Error rates per endpoint

### Logging (Structured JSON)
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "transit-api",
  "endpoint": "/api/v1/stops/nearby",
  "method": "GET",
  "status_code": 200,
  "duration_ms": 45,
  "cache_hit": true,
  "user_id": "anonymous"
}
```

### Health Checks
```http
GET /health        # Basic health check
GET /health/ready  # Readiness probe (K8s)
GET /health/live   # Liveness probe (K8s)
```

## 🚀 Deployment

### Docker
```bash
# Build images
docker build -t arrivo-api -f Dockerfile.api .
docker build -t arrivo-ingestor -f Dockerfile.ingestor .

# Run with compose
docker-compose up -d
```

### Kubernetes
```bash
# Apply manifests
kubectl apply -f k8s/

# Check deployment
kubectl get pods -l app=arrivo-api
```

### Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/arrivo
REDIS_URL=redis://localhost:6379/0

# API Configuration  
PORT=8080
API_KEY_REQUIRED=true
RATE_LIMIT_RPS=100

# GTFS Data Sources
GTFS_STATIC_URL=https://example.com/gtfs.zip
GTFS_REALTIME_URL=https://example.com/gtfs-rt
OVAPI_KEY=your-ovapi-key

# Monitoring
PROMETHEUS_ENABLED=true
LOG_LEVEL=info
TRACING_ENABLED=true
```

## 🤝 Contributing

Bijdragen zijn alleen mogelijk voor geautoriseerde team members. Zie [CONTRIBUTING.md](CONTRIBUTING.md) voor interne development guidelines.

### Development Workflow
1. Clone de repository (toegang vereist)
2. Maak een feature branch (`git checkout -b feature/amazing-feature`)
3. Commit je changes (`git commit -m 'Add amazing feature'`)
4. Push naar de branch (`git push origin feature/amazing-feature`)
5. Open een Pull Request voor review

### Code Style
- Gebruik `gofmt` voor formatting
- Run `golangci-lint` voor linting
- Schrijf tests voor nieuwe features
- Update documentatie waar nodig

## 📄 License

Dit is een proprietary/closed source project. Alle rechten voorbehouden aan Arrivo.

**Gebruik, distributie of modificatie zonder expliciete toestemming is niet toegestaan.**

## 🆘 Support

- 📧 Email: support@goarrivo.nl
- 💬 Discord: [Arrivo Community](https://discord.gg/arrivo)
- 🐛 Issues: [GitHub Issues](https://github.com/goarrivo/arrivo-transit-api/issues)
- 📖 Docs: [docs.goarrivo.nl](https://docs.goarrivo.nl)

## 🗺️ Roadmap

### Q1 2024
- [ ] GraphQL API endpoint
- [ ] WebSocket real-time updates
- [ ] Mobile SDK (iOS/Android)
- [ ] Advanced analytics dashboard

### Q2 2024
- [ ] Multi-region deployment
- [ ] AI-powered delay predictions
- [ ] Integration met meer EU transit APIs
- [ ] Performance optimizations

### Q3 2024
- [ ] User accounts en personalization
- [ ] Push notifications
- [ ] Offline-first mobile app
- [ ] Enterprise features

---

**Gebouwd met ❤️ in Nederland voor betere openbaar vervoer ervaring**
