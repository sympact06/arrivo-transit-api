# Arrivo Transit API - Todo Lijst

Deze todo-lijst is gebaseerd op het `project.txt` document en is opgedeeld in fases en componenten voor een gestructureerde aanpak.

## Fase 0: Setup (1-3 dagen)

- [ ] **Infrastructuur opzetten:**
    - [ ] Maak Hetzner Managed Postgres (met PostGIS) aan.
    - [ ] Maak Hetzner Managed Redis 7 aan.
    - [ ] Configureer Cloudflare: DNS, proxy, WAF, en cache rules (SWR=60).
- [ ] **Project Initialisatie:**
    - [ ] Maak een nieuw Go project (Go 1.22+).
    - [ ] Initialiseer Git repository.
    - [ ] Maak een basis `README.md` aan.
- [ ] **Coolify Setup:**
    - [ ] Maak Coolify apps aan: `arrivo-api`, `arrivo-worker-realtime`, `arrivo-worker-gtfs`, `telemetry`.
    - [ ] Configureer environment variables met DSN's voor Postgres en Redis.
- [ ] **Observability Stack:**
    - [ ] Zet Prometheus, Grafana, Loki, en Tempo op (via Coolify of een aparte VM).
    - [ ] Configureer OpenTelemetry in de Go applicatie.

## Fase 1: MVP (1-2 weken)

- [ ] **GTFS Ingestor:**
    - [ ] Implementeer een worker die periodiek GTFS data downloadt.
    - [ ] Schrijf een ETL-proces om GTFS data naar Postgres + PostGIS te importeren.
    - [ ] Zorg voor de juiste indexen op de geo-data.
- [ ] **Realtime Worker:**
    - [ ] Implementeer een worker die de OpenOV/OVapi pollt.
    - [ ] Normaliseer de realtime data.
    - [ ] Vul de Redis cache met genormaliseerde data.
    - [ ] Implementeer circuit breakers (sony/gobreaker) rond de OVapi calls.
- [ ] **API Gateway (Read-Only):**
    - [ ] Bouw de API met Go en de `chi` router.
    - [ ] Implementeer endpoints voor het opvragen van transit data.
    - [ ] Integreer de multi-layer cache (in-memory LRU, Redis).
    - [ ] Implementeer health checks en metrics endpoints.
    - [ ] Maak Grafana dashboards voor de belangrijkste metrics.
- [ ] **Testen:**
    - [ ] Schrijf unit tests voor de data mapping en validatie.
    - [ ] Zet integration tests op met testcontainers voor Postgres en Redis.
    - [ ] Voer load tests uit met k6 om de performance te meten en TTL's/breakers te tunen.
    - [ ] Voer chaos tests uit met Toxiproxy om de veerkracht te testen.

## Fase 2: SLA 99.9% (2-4 weken)

- [ ] **Schaalbaarheid en Betrouwbaarheid:**
    - [ ] Configureer de API om met meerdere instances te draaien.
    - [ ] Zet een load balancer op.
    - [ ] Implementeer cache warmers om de cache proactief te vullen.
    - [ ] Verfijn de Cloudflare edge rules.
- [ ] **Monitoring en Alerts:**
    - [ ] Stel alerts in voor de gedefinieerde SLO's (p99 latency, cache hit ratio, etc.).
    - [ ] Maak een statuspagina aan.
    - [ ] Schrijf incident runbooks.

## Fase 3: v1.5 (later)

- [ ] **AI-functionaliteit:**
    - [ ] Ontwikkel een NLU microservice voor zoekopdrachten in natuurlijke taal.
    - [ ] Implementeer een reliability score voor ETA's.
- [ ] **Premium Features:**
    - [ ] Voeg een SSE/WebSocket stream toe voor realtime updates.
- [ ] **Deployment:**
    - [ ] Implementeer Canary/Blue-Green deployments.
    - [ ] Onderzoek de mogelijkheden voor een multi-regio setup.

## Overig

- [ ] **Security:**
    - [ ] Implementeer API keys met quota's en rate limiting.
    - [ ] Beheer secrets via Coolify/Dokploy's secret manager.
    - [ ] Voer regelmatig backup- en restore-drills uit.
- [ ] **CI/CD:**
    - [ ] Zet een CI/CD pipeline op met GitHub Actions.
    - [ ] Automatiseer linting, testing, building, en deployment.
- [ ] **Documentatie:**
    - [ ] Schrijf API documentatie.
    - [ ] Houd een changelog bij.