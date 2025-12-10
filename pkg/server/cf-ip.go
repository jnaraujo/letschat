package server

import (
	"net"
	"net/http"
	"strings"
)

func getRealIP(r *http.Request) string {
	// 1. Priority: specific Cloudflare header
	cfIP := r.Header.Get("CF-Connecting-IP")
	if cfIP != "" {
		return cfIP
	}

	// 2. Fallback: X-Forwarded-For
	// This header can contain a list of IPs (client, proxy1, proxy2...)
	// usually the first one is the client.
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Get the first IP in the list
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// 3. Last resort: Direct connection IP (will be Cloudflare's IP if proxied)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
