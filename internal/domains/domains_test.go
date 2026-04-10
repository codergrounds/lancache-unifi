package domains

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseDomainLines(t *testing.T) {
	input := `
# A comment
lancache.steamcontent.com
*.cdn.blizzard.com


dyn.riotcdn.net
`

	want := []string{
		"lancache.steamcontent.com",
		"*.cdn.blizzard.com",
		"dyn.riotcdn.net",
	}

	got, err := parseDomainLines(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDomainLines() = %v, want %v", got, want)
	}
}
