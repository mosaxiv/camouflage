package proxy

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mosaxiv/camouflage/app/config"
)

type Proxy struct {
	client *http.Client
	config config.Config
}

func NewProxy(conf config.Config) Proxy {
	c := &http.Client{
		Timeout: conf.Timeout,
		Transport: &http.Transport{
			DisableKeepAlives: conf.DisableKeepAlive,
		},
	}

	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= conf.MaxRedirects {
			return errors.New("too many redirects")
		}
		return nil
	}

	return Proxy{
		client: c,
		config: conf,
	}
}

func (p Proxy) request(url string, r *http.Request) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	h := req.Header
	h.Set("Via", p.config.HeaderVia)
	h.Set("User-Agent", p.config.HeaderVia)
	h.Set("X-Frame-Options", "deny")
	h.Set("X-XSS-Protection", "1; mode=block")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "default-src 'none'; img-src data:; style-src 'unsafe-inline'")
	if v := r.Header.Get("Accept"); v == "" {
		h.Set("Accept", "image/*")
	} else {
		h.Set("Accept", v)
	}
	if v := r.Header.Get("Accept-Encoding"); v == "" {
		h.Set("Accept-Encoding", "")
	} else {
		h.Set("Accept-Encoding", v)
	}

	return p.client.Do(req)
}

func (p Proxy) response(w http.ResponseWriter, res *http.Response) error {
	if res.ContentLength > p.config.LengthLimit {
		return errors.New("Content-Length exceeded")
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("no content-type returned")
	}

	arr := strings.Split(contentType, ";")
	s := strings.ToLower(arr[0])
	match := false
	for _, v := range p.config.MimeTypes {
		if v == s {
			match = true
			break
		}
	}
	if !match {
		return errors.New("Non-Image content-type returned " + arr[0])
	}

	h := w.Header()
	h.Set("Content-Type", res.Header.Get("Content-Type"))
	h.Set("Content-Length", strconv.FormatInt(res.ContentLength, 10))
	h.Set("Camo-Host", p.config.HostName)
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
	if p.config.TimingAllowOrigin != "" {
		h.Set("Timing-Allow-Origin", p.config.TimingAllowOrigin)
	}
	if v := res.Header.Get("Transfer-Encoding"); v != "" {
		h.Set("Transfer-Encoding", v)
	}
	if v := res.Header.Get("Content-Encoding"); v != "" {
		h.Set("Content-Encoding", v)
	}

	_, err := io.Copy(w, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func (p Proxy) Run(w http.ResponseWriter, r *http.Request, url string) error {
	res, err := p.request(url, r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return p.response(w, res)
}
