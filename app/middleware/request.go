package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/mosaxiv/camouflage/app/config"
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
