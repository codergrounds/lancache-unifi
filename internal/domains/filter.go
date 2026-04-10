package domains

import "log"

// Filter applies allowlist/blocklist logic to a set of domain entries.
//   - If allowlist is non-empty: keep only entries whose Group is in the allowlist.
//   - Else if blocklist is non-empty: remove entries whose Group is in the blocklist.
//   - Else: return all entries unchanged.
func Filter(entries []DomainEntry, allowlist, blocklist []string) []DomainEntry {
	if len(allowlist) > 0 {
		return filterByAllowlist(entries, allowlist)
	}
	if len(blocklist) > 0 {
		return filterByBlocklist(entries, blocklist)
	}
	return entries
}

func filterByAllowlist(entries []DomainEntry, allowlist []string) []DomainEntry {
	allowed := toSet(allowlist)
	var result []DomainEntry
	for _, e := range entries {
		if allowed[e.Group] {
			result = append(result, e)
		}
	}
	log.Printf("[INFO] allowlist filter: %d → %d domains", len(entries), len(result))
	return result
}

func filterByBlocklist(entries []DomainEntry, blocklist []string) []DomainEntry {
	blocked := toSet(blocklist)
	var result []DomainEntry
	for _, e := range entries {
		if !blocked[e.Group] {
			result = append(result, e)
		}
	}
	log.Printf("[INFO] blocklist filter: %d → %d domains", len(entries), len(result))
	return result
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
