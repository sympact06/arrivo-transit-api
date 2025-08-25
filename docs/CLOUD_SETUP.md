# Cloud Setup Gids

Dit document beschrijft de stappen om de cloud infrastructuur voor de Arrivo Transit API op te zetten. Vul de details in waar aangegeven.

## 1. Hetzner

### Managed Postgres Database

1.  **Maak een nieuwe Managed Database aan** in de Hetzner Cloud Console.
2.  **Kies de gewenste specificaties.**
3.  **Voeg een database en gebruiker toe.**
4.  **Installeer de PostGIS extensie.** Dit kan meestal met een commando zoals `CREATE EXTENSION postgis;`.
5.  **Noteer de connection string (DSN).**
    -   `HETZNER_POSTGRES_DSN=`

### Managed Redis

1.  **Maak een nieuwe Managed Redis instance aan.**
2.  **Kies de gewenste specificaties.**
3.  **Noteer de connection string (DSN).**
    -   `HETZNER_REDIS_DSN=`

## 2. Cloudflare

1.  **Voeg je domein toe aan Cloudflare.**
2.  **Configureer de DNS records** om naar je Hetzner server (waar Coolify/Dokploy draait) te wijzen. Zorg ervoor dat de proxy-status (oranje wolkje) aan staat.
3.  **Activeer de Web Application Firewall (WAF)** met de standaard aanbevolen regels.
4.  **Stel een Cache Rule in:**
    -   **URL:** `api.jouwdomein.nl/*`
    -   **Cache Level:** `Cache Everything`
    -   **Edge TTL:** `1 minute`
    -   **Browser TTL:** `1 minute`
    -   **Stale-while-revalidate:** `60 seconds`

## 3. Coolify

1.  **Installeer Coolify** op een nieuwe Hetzner Cloud server.
2.  **Maak een nieuw project aan.**
3.  **Voeg de externe database en Redis toe** als services, gebruikmakend van de DSN's die je eerder hebt genoteerd.
4.  **Maak de volgende applicaties aan** (als Git-gebaseerde deployments):
    -   `arrivo-api`
    -   `arrivo-worker-realtime`
    -   `arrivo-worker-gtfs`
5.  **Configureer de environment variables** voor elke applicatie. Zorg ervoor dat ze de DSN's van de externe database en Redis gebruiken.
6.  **Zet de deployments op.**