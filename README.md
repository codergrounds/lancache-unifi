# Lancache UniFi DNS Sync

Automatically syncs [lancache cache-domains](https://github.com/uklans/cache-domains) to your UniFi router's DNS records, redirecting game/CDN downloads to your local lancache server.

## Features

- Fetches domain lists from the [uklans/cache-domains](https://github.com/uklans/cache-domains) repository
- Upserts DNS A records on your UniFi controller via the static-dns API
- Filters domains using `SERVICE_ALLOWLIST` and `SERVICE_BLOCKLIST` environment variables
- Cleans up stale DNS records that are no longer in the domain lists
- Configurable cron schedule (random daily time by default)
- Dry-run mode for safe testing
- Lightweight Docker image

## Running the Application

### 1. Using Precompiled Binaries

You can download the latest precompiled static binaries for Linux, macOS, and Windows directly from the [GitHub Releases page](https://github.com/codergrounds/lancache-unifi/releases). Simply download, extract, and execute:

```bash
# Copy and configure the environment variables
cp .env.example .env

# Run the application (it will automatically read the .env file)
./lancache-unifi
```

### 2. Using Prebuilt Docker Containers

Pre-assembled multi-architecture Docker containers are distributed instantly on tag releases via the GitHub Container Registry (GHCR). 

```bash
docker run -d \
  --name lancache-unifi \
  --restart unless-stopped \
  -e UNIFI_HOST=https://192.168.1.1 \
  -e UNIFI_API_KEY=your-api-key \
  -e LANCACHE_IP=192.168.1.100 \
  -e SERVICE_ALLOWLIST=steam,epicgames,blizzard \
  ghcr.io/codergrounds/lancache-unifi:latest
```

## Building from Source

If you want to review the code, build the application manually, or deploy test branches locally, all core commands are mapped directly to the `Makefile`:

```bash
# Clone the repo and configure your environment
git clone https://github.com/codergrounds/lancache-unifi.git
cd lancache-unifi
cp .env.example .env

# 1. Run the code directly via Go (reads .env automatically)
make run

# 2. Compile the binary locally (outputs to dist/lancache-unifi)
make build
./dist/lancache-unifi

# 3. Build the Docker container locally
make docker
```

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `UNIFI_HOST` | ✅ | — | UniFi controller URL |
| `UNIFI_API_KEY` | ✅ | — | API key from UniFi Console |
| `LANCACHE_IP` | ✅ | — | Target IP for DNS records |
| `UNIFI_SITE` | ❌ | `default` | UniFi site name |
| `DNS_TTL` | ❌ | — | DNS record TTL in seconds (maximum 86400) |
| `SERVICE_ALLOWLIST` | ❌ | — | Comma-separated groups to include |
| `SERVICE_BLOCKLIST` | ❌ | — | Comma-separated groups to exclude |
| `DRY_RUN` | ❌ | `false` | Log actions without making changes |
| `CRON_SCHEDULE` | ❌ | random daily | Standard 5-field cron expression |

### Filtering Logic

- If `SERVICE_ALLOWLIST` is set, **only** those groups are synced
- If only `SERVICE_BLOCKLIST` is set, those groups are **excluded**
- If neither is set, **all** groups are synced

### Available Domain Groups

`arenanet`, `blizzard`, `bsg`, `cityofheroes`, `cod`, `daybreak`, `epicgames`, `frontier`, `neverwinter`, `nexusmods`, `nintendo`, `origin`, `pathofexile`, `renegadex`, `riot`, `rockstar`, `sony`, `square`, `steam`, `teso`, `test`, `uplay`, `warframe`, `wargaming`, `wsus`, `xboxlive`

## How It Works

1. On startup, runs an initial sync immediately
2. Fetches `cache_domains.json` from GitHub
3. Downloads each domain group's `.txt` file(s)
4. Applies allowlist/blocklist filtering
5. Compares desired state against existing UniFi DNS records
6. Creates, updates, or deletes records as needed
7. Starts the cron scheduler for recurring syncs

### Cleanup Behaviour

Records pointing to your `LANCACHE_IP` that are no longer in the active domain lists are automatically deleted. Records pointing to other IPs are never touched.
