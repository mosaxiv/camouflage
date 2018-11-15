package camouflage

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/mosaxiv/camouflage/app/config"
	"github.com/mosaxiv/camouflage/app/hash"
	"github.com/mosaxiv/camouflage/app/proxy"
)

type Param struct {
	digest string
	url    string
}

type Camouflage struct {
	hmac  hash.HMAC
	conf  config.Config
	proxy proxy.Proxy
	p     *Param
}

func Sever() {
	conf := config.NewConfig()
	camo := &Camouflage{
		conf:  conf,
		hmac:  hash.NewHMAC(conf.SharedKey),
		proxy: proxy.NewProxy(conf),
	}

	fmt.Printf("SSL-Proxy running on %s with pid: %d\n", conf.Port, os.Getpid())

	if err := http.ListenAndServe(":"+conf.Port, newRoute(camo)); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func notFound(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Expires", "0")
	h.Set("Cache-Control", "no-cache, no-store, private, must-revalidate")
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func (camo *Camouflage) query(w http.ResponseWriter, r *http.Request) {
	camo.p = &Param{
		digest: chi.URLParam(r, "digest"),
		url:    r.URL.Query().Get("url"),
	}

	if camo.p.url == "" {
		notFound(w)
		return
	}

	if err := camo.proses(w, r); err != nil {
		notFound(w)
		return
	}
}

func (camo *Camouflage) param(w http.ResponseWriter, r *http.Request) {
	url, err := hex.DecodeString(chi.URLParam(r, "url"))
	if err != nil {
		notFound(w)
		return
	}

	camo.p = &Param{
		digest: chi.URLParam(r, "digest"),
		url:    string(url),
	}

	if err := camo.proses(w, r); err != nil {
		notFound(w)
		return
	}
}

func (camo *Camouflage) proses(w http.ResponseWriter, r *http.Request) error {
	if err := camo.hmac.DigestCheck(camo.p.digest, camo.p.url); err != nil {
		return err
	}

	if err := camo.proxy.Run(w, r, camo.p.url); err != nil {
		return err
	}

	return nil
}
