package internal

import (
	"net"
	"net/http"
)

func ResolveClient(r *http.Request) string {
	apiKey := r.Header.Get("X-API-Key")

	if apiKey != "" {
		return "api:" + apiKey
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	return "ip:" + ip
}