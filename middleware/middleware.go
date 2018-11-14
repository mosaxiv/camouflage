package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/mosaxiv/camouflage/config"
)

func SelfRequest(c config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Via"), c.HeaderVia) {
				log.Println("Requesting from self")
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func DefaultSecurityHeaders(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Frame-Options", "deny")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
