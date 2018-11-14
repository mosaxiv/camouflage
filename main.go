package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/mosaxiv/camouflage/config"
	"github.com/mosaxiv/camouflage/hash"
	"github.com/mosaxiv/camouflage/middleware"
	"github.com/mosaxiv/camouflage/proxy"
)

type Param struct {
	digest string
	url    string
}

var hmac hash.HMAC
var conf config.Config
var pro proxy.Proxy

func main() {
	conf = config.NewConfig()
	hmac = hash.NewHMAC(conf.SharedKey)
	pro = proxy.NewProxy(conf)

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
		r.Use(middleware.SelfRequest(conf))
		r.Get("/{digest}", query)
		r.Get("/{digest}/{url}", param)
	})

	fmt.Printf("SSL-Proxy running on %s pid: %d\n", conf.Port, os.Getpid())

	if err := http.ListenAndServe(":"+conf.Port, r); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func notFound(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Expires", "0")
	h.Set("Cache-Control", "no-cache, no-store, private, must-revalidate")
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func query(w http.ResponseWriter, r *http.Request) {
	param := &Param{
		digest: chi.URLParam(r, "digest"),
		url:    r.URL.Query().Get("url"),
	}

	if param.url == "" {
		notFound(w)
		return
	}

	if err := param.proses(w, r); err != nil {
		notFound(w)
		return
	}
}

func param(w http.ResponseWriter, r *http.Request) {
	url, err := hex.DecodeString(chi.URLParam(r, "url"))
	if err != nil {
		notFound(w)
		return
	}

	param := &Param{
		digest: chi.URLParam(r, "digest"),
		url:    string(url),
	}

	if err := param.proses(w, r); err != nil {
		notFound(w)
		return
	}
}

func (p *Param) proses(w http.ResponseWriter, r *http.Request) error {
	if err := hmac.DigestCheck(p.digest, p.url); err != nil {
		return err
	}

	if err := pro.Proxy(w, r, p.url); err != nil {
		return err
	}

	return nil
}
