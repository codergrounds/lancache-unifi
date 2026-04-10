package domains

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	entries := []DomainEntry{
		{Group: "steam", Domain: "lancache.steamcontent.com"},
		{Group: "blizzard", Domain: "cdn.blizzard.com"},
		{Group: "riot", Domain: "dyn.riotcdn.net"},
	}

	tests := []struct {
		name      string
		allowlist []string
		blocklist []string
		want      []DomainEntry
	}{
		{
			name:      "no filters",
			allowlist: nil,
			blocklist: nil,
			want:      entries,
		},
		{
			name:      "allowlist only",
			allowlist: []string{"steam", "riot"},
			blocklist: nil,
			want: []DomainEntry{
				{Group: "steam", Domain: "lancache.steamcontent.com"},
				{Group: "riot", Domain: "dyn.riotcdn.net"},
			},
		},
		{
			name:      "blocklist only",
			allowlist: nil,
			blocklist: []string{"steam"},
			want: []DomainEntry{
				{Group: "blizzard", Domain: "cdn.blizzard.com"},
				{Group: "riot", Domain: "dyn.riotcdn.net"},
			},
		},
		{
			name:      "allowlist overrides blocklist",
			allowlist: []string{"blizzard"},
			blocklist: []string{"blizzard"}, // If allowlist is set, blocklist is ignored
			want: []DomainEntry{
				{Group: "blizzard", Domain: "cdn.blizzard.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(entries, tt.allowlist, tt.blocklist)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
