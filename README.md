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

### Option A: Using Precompiled Binaries

You can download the latest precompiled static binaries for Linux, macOS, and Windows directly from the [GitHub Releases page](https://github.com/codergrounds/lancache-unifi/releases). Simply download, extract, and execute:

```bash
# Copy and configure the environment variables
cp .env.example .env

# Run the application (it will automatically read the .env file)
./lancache-unifi
```

### Option B: Using Prebuilt Docker Containers

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


### Option A: Building Natively (Go)

If you have Go installed on your machine, you can launch or compile the application natively:

First, clone the repository and configure your local environment variables:
```bash
git clone https://github.com/codergrounds/lancache-unifi.git
cd lancache-unifi
cp .env.example .env

# Run the code instantly without compiling (it will securely read .env locally)
make run

# Or explicitly compile the production standalone binary (outputs to dist/)
make build
./dist/lancache-unifi
```

### Option B: Building via Docker

If you don't have Go installed, or prefer deploying via containers, Docker isolates all dependencies perfectly using a multi-stage builder. *(Note: Our Makefile logic dynamically tags the container with your current active Git branch name!)*

```bash
make docker
```

```bash
docker run -d \
  --name lancache-unifi \
  --restart unless-stopped \
  -e UNIFI_HOST=https://192.168.1.1 \
  -e UNIFI_API_KEY=your-api-key \
  -e LANCACHE_IP=192.168.1.100 \
  -e SERVICE_ALLOWLIST=steam,epicgames,blizzard \
  lancache-unifi:<branch-name>
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

### Obtaining a UniFi API Key

1. Log into your **UniFi OS Console**
2. Open **Settings**
3. Navigate to **Control Plane** -> **Integrations**
4. Fill out "**API Key Name** and click **Create API Key**.
5. Copy the generated API key and paste it into the `UNIFI_API_KEY` environment variable.

### Filtering Logic

- If `SERVICE_ALLOWLIST` is set, **only** those groups are synced
- If only `SERVICE_BLOCKLIST` is set, those groups are **excluded**
- If neither is set, **all** groups are synced

Available services can be found [here](https://github.com/uklans/cache-domains/blob/master/cache_domains.json).

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
