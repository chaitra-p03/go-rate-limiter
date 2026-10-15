package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"rateLimiter/internal"
)

func TestRouterLimitsDataButKeepsStatsAvailable(t *testing.T) {
	limiter := internal.NewratelimiterManager()

	// One request is allowed. The tiny refill rate keeps the test
	// from depending on timing.
	handler := newRouter(limiter, 1, 0.000001)

	makeRequest := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-API-Key", "alice")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}

	firstDataResponse := makeRequest("/api/data")
	if firstDataResponse.Code != http.StatusOK {
		t.Fatalf("first /api/data request: got %d, want %d",
			firstDataResponse.Code, http.StatusOK)
	}

	secondDataResponse := makeRequest("/api/data")
	if secondDataResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("second /api/data request: got %d, want %d",
			secondDataResponse.Code, http.StatusTooManyRequests)
	}

	statsResponse := makeRequest("/stats")
	if statsResponse.Code != http.StatusOK {
		t.Fatalf("/stats after quota is exhausted: got %d, want %d",
			statsResponse.Code, http.StatusOK)
	}
}

func TestCheckRouteIsNotRegistered(t *testing.T) {
	limiter := internal.NewratelimiterManager()
	handler := newRouter(limiter, 10, 1)

	req := httptest.NewRequest(http.MethodPost, "/check", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)

	if response.Code != http.StatusNotFound {
		t.Fatalf("/check: got %d, want %d",
			response.Code, http.StatusNotFound)
	}
}
