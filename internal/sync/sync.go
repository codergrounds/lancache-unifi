package sync

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/codergrounds/lancache-unifi/internal/domains"
	"github.com/codergrounds/lancache-unifi/internal/unifi"
)

// Syncer orchestrates fetching domains and upserting/cleaning DNS records.
type Syncer struct {
	unifiClient     *unifi.Client
	httpClient      *http.Client
	lancacheIP      string
	serviceAllowlist   []string
	serviceBlocklist   []string
	dryRun             bool
	ttl                int
	backoffMaxDuration time.Duration
}

// NewSyncer creates a new Syncer instance.
func NewSyncer(
	unifiClient *unifi.Client,
	httpClient *http.Client,
	lancacheIP string,
	allowlist, blocklist []string,
	dryRun bool,
	ttl int,
) *Syncer {
	return &Syncer{
		unifiClient:     unifiClient,
		httpClient:      httpClient,
		lancacheIP:         lancacheIP,
		serviceAllowlist:   allowlist,
		serviceBlocklist:   blocklist,
		dryRun:             dryRun,
		ttl:                ttl,
		backoffMaxDuration: 6 * time.Hour,
	}
}

// Run executes a full sync cycle:
//  1. Fetch domain entries from GitHub
//  2. Apply allowlist/blocklist filtering
//  3. Fetch existing DNS records from UniFi
//  4. Create, update, or delete records as needed
func (s *Syncer) Run() error {
	log.Println("[INFO] === starting sync run ===")

	// 1. Fetch all domain entries from GitHub (unfiltered) with exponential backoff
	var allEntries []domains.DomainEntry
	var err error

	startTime := time.Now()
	delay := 10 * time.Second
	maxDelay := 30 * time.Minute

	for {
		allEntries, err = domains.FetchAll(s.httpClient)
		if err == nil {
			break
		}

		if time.Since(startTime) > s.backoffMaxDuration {
			return fmt.Errorf("fetching domains failed after %v: %w", s.backoffMaxDuration, err)
		}

		log.Printf("[WARN] failed to fetch domains from GitHub (retrying in %v): %v", delay, err)
		time.Sleep(delay)

		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}

	// Build a set of ALL cache domains known to the GitHub repo.
	// We use this for safe stateless cleanup later.
	allGithubDomains := make(map[string]bool, len(allEntries))
	for _, e := range allEntries {
		allGithubDomains[e.Domain] = true
	}

	// 2. Apply allowlist/blocklist filters
	filteredEntries := domains.Filter(allEntries, s.serviceAllowlist, s.serviceBlocklist)
	if len(filteredEntries) == 0 && len(allEntries) > 0 {
		log.Println("[WARN] no domains remaining after filtering, nothing to sync")
	}

	// 3. Build desired state: set of domain keys we actually want to sync
	desired := make(map[string]bool, len(filteredEntries))
	for _, e := range filteredEntries {
		desired[e.Domain] = true
	}
	log.Printf("[INFO] desired state: %d unique domains → %s", len(desired), s.lancacheIP)

	// 4. Fetch existing records from UniFi
	existing, err := s.unifiClient.ListDNSRecords()
	if err != nil {
		return err
	}
	log.Printf("[INFO] found %d existing DNS records on UniFi", len(existing))

	// Build lookup map: key → record
	existingMap := make(map[string]unifi.DNSRecord, len(existing))
	for _, rec := range existing {
		existingMap[rec.Key] = rec
	}

	// 5. Upsert: create or update records
	var created, updated, skipped int
	for domain := range desired {
		rec, exists := existingMap[domain]
		if exists {
			if rec.Value == s.lancacheIP && rec.RecordType == "A" && rec.Enabled {
				if s.ttl == 0 || rec.TTL == s.ttl {
					skipped++
					continue
				}
			}
			// Update existing record
			if s.dryRun {
				log.Printf("[DRY-RUN] would update %q: %s → %s", domain, rec.Value, s.lancacheIP)
			} else {
				err := s.unifiClient.UpdateDNSRecord(rec.ID, unifi.DNSRecord{
					Key:        domain,
					Value:      s.lancacheIP,
					RecordType: "A",
					Enabled:    true,
					TTL:        s.ttl,
				})
				if err != nil {
					log.Printf("[ERROR] failed to update %q: %v", domain, err)
					continue
				}
				log.Printf("[INFO] updated %q → %s", domain, s.lancacheIP)
			}
			updated++
		} else {
			// Create new record
			if s.dryRun {
				log.Printf("[DRY-RUN] would create %q → %s", domain, s.lancacheIP)
			} else {
				err := s.unifiClient.CreateDNSRecord(unifi.DNSRecord{
					Key:        domain,
					Value:      s.lancacheIP,
					RecordType: "A",
					Enabled:    true,
					TTL:        s.ttl,
				})
				if err != nil {
					log.Printf("[ERROR] failed to create %q: %v", domain, err)
					continue
				}
				log.Printf("[INFO] created %q → %s", domain, s.lancacheIP)
			}
			created++
		}
	}

	// 6. Cleanup: delete stateless records
	// To safely do this statelessly, we look for records that:
	//   a) Point to LANCACHE_IP (so we don't touch random other records)
	//   b) Exist in the UNFILTERED cache_domains set from GitHub (so we confirm it's a lancache domain)
	//   c) DO NOT exist in our FILTERED desired set (so we no longer want it)
	var deleted int
	for _, rec := range existing {
		if rec.RecordType != "A" || rec.Value != s.lancacheIP {
			continue // Not managed by this tool, or changed to point somewhere else
		}
		
		isCacheDomain := allGithubDomains[rec.Key]
		isDesired := desired[rec.Key]
		
		if isCacheDomain && !isDesired {
			if s.dryRun {
				log.Printf("[DRY-RUN] would delete stale record %q (id=%s)", rec.Key, rec.ID)
			} else {
				err := s.unifiClient.DeleteDNSRecord(rec.ID)
				if err != nil {
					log.Printf("[ERROR] failed to delete stale record %q: %v", rec.Key, err)
					continue
				}
				log.Printf("[INFO] deleted stale record %q (id=%s)", rec.Key, rec.ID)
			}
			deleted++
		}
	}

	log.Printf("[INFO] === sync complete: created=%d updated=%d skipped=%d deleted=%d ===",
		created, updated, skipped, deleted)
	return nil
}
