package unifi

// DNSRecord represents a static DNS record on a UniFi controller.
type DNSRecord struct {
	ID         string `json:"_id,omitempty"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	RecordType string `json:"record_type"`
	Enabled    bool   `json:"enabled"`
	TTL        int    `json:"ttl,omitempty"`
}
