package middleware

import "net/http"

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
