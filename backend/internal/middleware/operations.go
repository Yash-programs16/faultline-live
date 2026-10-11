package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"faultline-live/backend/internal/response"
)

// LocalDemo rejects non-loopback peers and cross-site browser mutations,
// even when a local server is reached through a proxy. Forwarded IP headers
// are deliberately not trusted.
func LocalDemo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)

		if err != nil || ip == nil || !ip.IsLoopback() {
			response.Forbidden(w, "local demo requires a loopback peer")
			return
		}

		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)

			if err != nil || u == nil ||
				(u.Scheme != "http" && u.Scheme != "https") {
				response.Forbidden(w, "local demo rejects non-local browser origins")
				return
			}

			hostname := u.Hostname()
			originIP := net.ParseIP(hostname)

			if hostname != "localhost" &&
				(originIP == nil || !originIP.IsLoopback()) {
				response.Forbidden(w, "local demo rejects non-local browser origins")
				return
			}
		}

		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			response.Forbidden(w, "cross-site demo request rejected")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func uHostname(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.Hostname()
}

// Customer JWTs are not operations authorization.
// Operators must be explicitly allowlisted by ID.
// An empty list denies access.
func Operators(ids string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)

	for _, id := range strings.Split(ids, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := GetUserID(r)

			if !ok || !allowed[strconv.FormatInt(id, 10)] {
				response.Forbidden(
					w,
					"operations access requires an allowlisted operator",
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
