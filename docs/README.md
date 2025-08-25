# Arrivo Transit API Documentation

## API Documentation

De Arrivo Transit API biedt uitgebreide documentatie via verschillende interfaces:

### Swagger UI (Interactieve Documentatie)

Bezoek de Swagger UI voor een interactieve API documentatie waar je direct API calls kunt testen:

- **Development**: http://localhost:8080/swagger/
- **Production**: https://api.goarrivo.nl/swagger/

### OpenAPI Specificatie

De volledige OpenAPI 3.0 specificatie is beschikbaar in verschillende formaten:

- **YAML**: [api.yaml](./api.yaml)
- **JSON**: http://localhost:8080/api/v1/swagger/doc.json
- **YAML Endpoint**: http://localhost:8080/api/v1/swagger/doc.yaml

## Snelstart

### 1. Start de API Server

```bash
# Development
go run cmd/api/main.go

# Of met Docker
docker-compose up api
```

### 2. Bekijk de Documentatie

Ga naar http://localhost:8080/swagger/ om de interactieve documentatie te bekijken.

### 3. Test een Endpoint

Probeer bijvoorbeeld de health check:

```bash
curl http://localhost:8080/health
```

Of zoek naar haltes:

```bash
curl "http://localhost:8080/api/v1/stops/search?q=Amsterdam"
```

## API Features

### 🚌 Transit Data
- Real-time vertrektijden
- Halte zoekfunctie
- GPS-gebaseerde nabije haltes
- Route informatie
- Voertuig tracking

### ⚡ Performance
- Multi-layer caching (In-Memory, Redis, CDN)
- Sub-100ms response times
- Rate limiting
- Graceful degradation

### 🔒 Security
- API key authenticatie
- Rate limiting per gebruiker
- CORS ondersteuning
- Request validation

## Endpoints Overzicht

| Endpoint | Beschrijving |
|----------|-------------|
| `GET /health` | Health check |
| `GET /api/v1/stops/search` | Zoek haltes |
| `GET /api/v1/stops/nearby` | Nabije haltes |
| `GET /api/v1/stops/{id}/departures` | Vertrektijden |
| `GET /api/v1/routes/search` | Zoek routes |
| `GET /api/v1/routes/{id}/vehicles` | Voertuigen op route |
| `GET /api/v1/vehicles/active` | Alle actieve voertuigen |

## Response Formaten

Alle responses zijn in JSON formaat met consistente error handling:

```json
{
  "data": { ... },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "cached": true,
    "cache_ttl": 300
  }
}
```

## Error Responses

```json
{
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "Parameter 'q' is required",
    "details": { ... }
  }
}
```

## Rate Limiting

- **Anoniem**: 100 requests/minuut
- **Met API Key**: 1000 requests/minuut  
- **Premium**: 10000 requests/minuut

Rate limit headers worden meegestuurd:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248600
```

## Caching

De API gebruikt een multi-layer caching strategie:

1. **L1 (In-Memory)**: 5 minuten - Snelste toegang
2. **L2 (Redis)**: 15 minuten - Gedeeld tussen instances
3. **L3 (CDN)**: 60 minuten - Edge caching

## Support

- **Email**: support@goarrivo.nl
- **Documentatie**: https://docs.goarrivo.nl
- **Status**: https://status.goarrivo.nl