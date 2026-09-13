package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareEnforcesRateLimit(t *testing.T) {
	rl := NewratelimiterManager()
	handlerCalls := 0
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalls++
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Middleware(1, 0.000001)(protectedHandler)

	firstRequest := httptest.NewRequest(
		http.MethodGet,
		"http://localhost/api/data",
		nil,
	)
	firstRequest.Header.Set("X-API-Key", "alice")

	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, firstRequest)

	if firstResponse.Code != http.StatusOK {
		t.Fatalf("first request: expected status 200, got %d", firstResponse.Code)
	}

	if firstResponse.Header().Get("RateLimit-Limit") != "1" {
		t.Fatalf(
			"first request: expected RateLimit-Limit 1, got %q",
			firstResponse.Header().Get("RateLimit-Limit"),
		)
	}

	secondRequest := httptest.NewRequest(
		http.MethodGet,
		"http://localhost/api/data",
		nil,
	)
	secondRequest.Header.Set("X-API-Key", "alice")

	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, secondRequest)

	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"second request: expected status 429, got %d",
			secondResponse.Code,
		)
	}

	if secondResponse.Header().Get("Retry-After") == "" {
		t.Fatal("second request: expected Retry-After header")
	}

	if handlerCalls != 1 {
		t.Fatalf(
			"expected protected handler to run once, ran %d times",
			handlerCalls,
		)
	}
}
