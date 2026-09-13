package internal

import (
	"net/http"
	"testing"
)

func TestResolveClientUsesAPIKey(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost/api/data", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-API-Key", "alice")
	got := ResolveClient(req)
	want := "api:alice"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveClientFallsBackToIP(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost/api/data",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.RemoteAddr = "127.0.0.1:5000"
	got := ResolveClient(req)
	want := "ip:127.0.0.1"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveClientPrefersAPIKeyOverIP(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost/api/data", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-API-Key", "alice")
	req.RemoteAddr = "127.0.0.1:5000"
	got := ResolveClient(req)
	want := "api:alice"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
