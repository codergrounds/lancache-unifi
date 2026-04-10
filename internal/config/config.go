package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration from environment variables.
type Config struct {
	// UniFi controller URL, e.g. "https://192.168.1.1"
	UniFiHost string
	// API key from UniFi Console (Settings → Control Plane → Integrations)
	UniFiAPIKey string
	// UniFi site name (default: "default")
	UniFiSite string
	// IP address of the lancache server
	LancacheIP string
	// Comma-separated list of cache-domain names to include. If set, only these are synced.
	ServiceAllowlist []string
	// Comma-separated list of cache-domain names to exclude. Ignored if allowlist is set.
	ServiceBlocklist []string
	// If true, log what would be done without making changes
	DryRun bool
	// Cron expression for scheduling (e.g. "0 3 * * *"). If empty, a random daily time is generated.
	CronSchedule string
	// Optional TTL for DNS records
	TTL int
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{
		UniFiHost:   os.Getenv("UNIFI_HOST"),
		UniFiAPIKey: os.Getenv("UNIFI_API_KEY"),
		UniFiSite:   os.Getenv("UNIFI_SITE"),
		LancacheIP:  os.Getenv("LANCACHE_IP"),
		DryRun:      strings.EqualFold(os.Getenv("DRY_RUN"), "true"),
		CronSchedule: os.Getenv("CRON_SCHEDULE"),
	}

	if ttlStr := os.Getenv("DNS_TTL"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil && ttl > 0 && ttl <= 86400 {
			cfg.TTL = ttl
		} else {
			return nil, fmt.Errorf("invalid DNS_TTL value: %s (must be between 1 and 86400)", ttlStr)
		}
	}

	if cfg.UniFiSite == "" {
		cfg.UniFiSite = "default"
	}

	if allowlist := os.Getenv("SERVICE_ALLOWLIST"); allowlist != "" {
		cfg.ServiceAllowlist = parseCSV(allowlist)
	}

	if blocklist := os.Getenv("SERVICE_BLOCKLIST"); blocklist != "" {
		cfg.ServiceBlocklist = parseCSV(blocklist)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.UniFiHost == "" {
		return fmt.Errorf("UNIFI_HOST is required")
	}
	if c.UniFiAPIKey == "" {
		return fmt.Errorf("UNIFI_API_KEY is required")
	}
	if c.LancacheIP == "" {
		return fmt.Errorf("LANCACHE_IP is required")
	}
	// Strip trailing slash from host
	c.UniFiHost = strings.TrimRight(c.UniFiHost, "/")
	return nil
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, strings.ToLower(trimmed))
		}
	}
	return result
}
