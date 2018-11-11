package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/mosaxiv/camouflage/config"
	"github.com/mosaxiv/camouflage/hash"
)

type Param struct {
	digest string
	url    string
}

var hmac hash.HMAC

func main() {
	c := config.NewConfig()
	hmac = hash.NewHMAC(c.SharedKey)

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

	fmt.Printf("SSL-Proxy running on %s pid: %d\n", c.Port, os.Getpid())

	if err := http.ListenAndServe(":"+c.Port, r); err != nil {
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

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.Copy(w, res.Body)
	if err != nil {
		return err
	}

	return nil
}
