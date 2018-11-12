package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"errors"
	"github.com/go-chi/chi"
	"github.com/mosaxiv/camouflage/config"
	"github.com/mosaxiv/camouflage/hash"
	"strconv"
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
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hwhat"))
	})
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/{digest}", query)
	r.Get("/{digest}/{url}", param)

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
		http.NotFound(w, r)
		return
	}

	if err := param.proses(w, r); err != nil {
		http.NotFound(w, r)
		return
	}
}

func param(w http.ResponseWriter, r *http.Request) {
	url, err := hex.DecodeString(chi.URLParam(r, "url"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	param := &Param{
		digest: chi.URLParam(r, "digest"),
		url:    string(url),
	}

	if err := param.proses(w, r); err != nil {
		http.NotFound(w, r)
		return
	}
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
		Timeout: conf.SocketTimeout,
		Transport: &http.Transport{
			DisableKeepAlives: !conf.KeepAlive,
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

	_, err = io.Copy(w, res.Body)
	if err != nil {
		return err
	}

	if h := res.Header.Get("Cache-Control"); h == "" {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	} else {
		w.Header().Set("Cache-Control", h)
	}

	w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", strconv.FormatInt(res.ContentLength, 10))
	w.Header().Set("Transfer-Encoding", res.Header.Get("Transfer-Encoding"))
	w.Header().Set("Content-Encoding", res.Header.Get("Content-Encoding"))
	w.Header().Set("Camo-Host", conf.HostName)
	w.Header().Set("X-Frame-Options", "deny")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	if h := res.Header.Get("etag"); h != "" {
		w.Header().Set("etag", h)
	}
	if h := res.Header.Get("expires"); h != "" {
		w.Header().Set("expires", h)
	}
	if h := res.Header.Get("last-modified"); h != "" {
		w.Header().Set("last-modified", h)
	}
	if conf.TimingAllowOrigin != "" {
		w.Header().Set("Timing-Allow-Origin", conf.TimingAllowOrigin)
	}

	return nil
}
