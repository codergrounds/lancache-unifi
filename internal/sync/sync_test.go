package sync

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/codergrounds/lancache-unifi/internal/unifi"
)

func TestSyncer_GithubDown(t *testing.T) {
	// Create a test server that immediately returns 500 Internal Server Error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	// The domain fetching logic unfortunately hardcodes the URL to GitHub for now.
	// But let's assume we can simulate the HTTP client failing entirely.
	// We'll create a transport that always errors out.
	failingTransport := &errTransport{}
	httpClient := &http.Client{Transport: failingTransport}

	unifiClient := unifi.NewClient("https://127.0.0.1", "default", "key")

	syncer := NewSyncer(
		unifiClient,
		httpClient,
		"10.0.0.100",
		nil,
		nil,
		true, // dryRun
		0,    // ttl
	)
	syncer.backoffMaxDuration = 5 * time.Millisecond

	err := syncer.Run()
	if err == nil {
		t.Fatalf("expected error when github is down, got nil")
	}

	if !strings.Contains(err.Error(), "fetching domains") {
		t.Errorf("expected fetching error, got %v", err)
	}
}

type errTransport struct{}

func (t *errTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, http.ErrServerClosed
}
