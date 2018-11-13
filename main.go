package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/mosaxiv/camouflage/config"
	"github.com/mosaxiv/camouflage/hash"
	"github.com/mosaxiv/camouflage/middleware"
)

type Param struct {
	digest string
	url    string
}

var hmac hash.HMAC
var conf config.Config

func main() {
	conf = config.NewConfig()
	hmac = hash.NewHMAC(conf.SharedKey)

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
		r.Use(middleware.SelfRequest)
		r.Get("/{digest}", query)
		r.Get("/{digest}/{url}", param)
	})

	fmt.Printf("SSL-Proxy running on %s pid: %d\n", conf.Port, os.Getpid())

	if err := http.ListenAndServe(":"+conf.Port, r); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
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

func notFound(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Expires", "0")
	h.Set("Cache-Control", "no-cache, no-store, private, must-revalidate")
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func (p *Param) proses(w http.ResponseWriter, r *http.Request) error {
	if err := hmac.DigestCheck(p.digest, p.url); err != nil {
		return err
	}

	if err := p.proxy(w, r); err != nil {
		return err
	}

	return nil
}

func (p *Param) proxy(w http.ResponseWriter, r *http.Request) error {
	req, err := http.NewRequest("GET", p.url, nil)
	if err != nil {
		return err
	}

	if h := r.Header.Get("Accept"); h == "" {
		req.Header.Set("Accept", "image/*")
	} else {
		req.Header.Set("Accept", h)
	}
	if h := r.Header.Get("Accept-Encoding"); h == "" {
		req.Header.Set("Accept-Encoding", "")
	} else {
		req.Header.Set("Accept-Encoding", h)
	}
	req.Header.Set("Via", conf.HeaderVia)
	req.Header.Set("User-Agent", conf.HeaderVia)
	req.Header.Set("X-Frame-Options", "deny")
	req.Header.Set("X-XSS-Protection", "1; mode=block")
	req.Header.Set("X-Content-Type-Options", "nosniff")
	req.Header.Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'")

	client := &http.Client{
		Timeout: conf.Timeout,
		Transport: &http.Transport{
			DisableKeepAlives: conf.DisableKeepAlive,
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.ContentLength > conf.LengthLimit {
		return errors.New("Content-Length exceeded")
	}

	h := w.Header()
	h.Set("Content-Type", res.Header.Get("Content-Type"))
	h.Set("Content-Length", strconv.FormatInt(res.ContentLength, 10))
	h.Set("Transfer-Encoding", res.Header.Get("Transfer-Encoding"))
	h.Set("Content-Encoding", res.Header.Get("Content-Encoding"))
	h.Set("Camo-Host", conf.HostName)
	if v := res.Header.Get("Cache-Control"); v == "" {
		h.Set("Cache-Control", "public, max-age=31536000")
	} else {
		h.Set("Cache-Control", v)
	}
	if v := res.Header.Get("Etag"); v != "" {
		h.Set("Etag", v)
	}
	if v := res.Header.Get("Expires"); v != "" {
		h.Set("Expires", v)
	}
	if v := res.Header.Get("Last-Modified"); v != "" {
		h.Set("Last-Modified", v)
	}
	if conf.TimingAllowOrigin != "" {
		h.Set("Timing-Allow-Origin", conf.TimingAllowOrigin)
	}

	_, err = io.Copy(w, res.Body)
	if err != nil {
		return err
	}

	return nil
}
