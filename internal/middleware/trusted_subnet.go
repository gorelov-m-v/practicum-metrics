package middleware

import (
	"net"
	"net/http"
	"strings"
)

const headerXRealIP = "X-Real-IP"

// TrustedSubnet returns a middleware that allows metrics only from the trusted subnet.
func TrustedSubnet(trustedSubnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := net.ParseIP(strings.TrimSpace(r.Header.Get(headerXRealIP)))
			if ip == nil || !trustedSubnet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
