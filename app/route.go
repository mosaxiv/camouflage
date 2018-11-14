package camouflage

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/mosaxiv/camouflage/app/middleware"
)

func newRoute(camo *Camouflage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.DefaultSecurityHeaders)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) { notFound(w) })

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hwhat"))
	})
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.SelfRequest(camo.conf))
		r.Get("/{digest}", camo.query)
		r.Get("/{digest}/{url}", camo.param)
	})

	return r
}
