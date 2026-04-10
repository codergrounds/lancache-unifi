package domains

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	cacheDomainsURL = "https://raw.githubusercontent.com/uklans/cache-domains/master/cache_domains.json"
	domainFileBase  = "https://raw.githubusercontent.com/uklans/cache-domains/master/"
)

// CacheDomainsJSON is the top-level structure of cache_domains.json.
type CacheDomainsJSON struct {
	CacheDomains []CacheDomainGroup `json:"cache_domains"`
}

// CacheDomainGroup is a single group (e.g. "steam", "blizzard").
type CacheDomainGroup struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	DomainFiles  []string `json:"domain_files"`
	Notes        string   `json:"notes,omitempty"`
	MixedContent bool     `json:"mixed_content,omitempty"`
}

// DomainEntry represents a single domain belonging to a named group.
type DomainEntry struct {
	Group  string // e.g. "steam"
	Domain string // e.g. "lancache.steamcontent.com"
}

// FetchAll downloads cache_domains.json and all referenced domain files,
// returning a flat list of DomainEntry values.
func FetchAll(client *http.Client) ([]DomainEntry, error) {
	groups, err := fetchIndex(client)
	if err != nil {
		return nil, fmt.Errorf("fetching cache_domains.json: %w", err)
	}

	var entries []DomainEntry
	for _, g := range groups {
		for _, file := range g.DomainFiles {
			domains, err := fetchDomainFile(client, file)
			if err != nil {
				log.Printf("[WARN] failed to fetch domain file %s for group %s: %v", file, g.Name, err)
				continue
			}
			for _, d := range domains {
				entries = append(entries, DomainEntry{Group: g.Name, Domain: d})
			}
		}
	}

	log.Printf("[INFO] fetched %d domains across %d groups", len(entries), len(groups))
	return entries, nil
}

// FetchGroups returns the raw group metadata without fetching domain files.
func FetchGroups(client *http.Client) ([]CacheDomainGroup, error) {
	return fetchIndex(client)
}

func fetchIndex(client *http.Client) ([]CacheDomainGroup, error) {
	resp, err := client.Get(cacheDomainsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from cache_domains.json", resp.StatusCode)
	}

	var data CacheDomainsJSON
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}

	return data.CacheDomains, nil
}

func fetchDomainFile(client *http.Client, filename string) ([]string, error) {
	url := domainFileBase + filename
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, filename)
	}

	return parseDomainLines(resp.Body)
}

func parseDomainLines(r io.Reader) ([]string, error) {
	var domains []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		domains = append(domains, strings.ToLower(line))
	}
	return domains, scanner.Err()
}
