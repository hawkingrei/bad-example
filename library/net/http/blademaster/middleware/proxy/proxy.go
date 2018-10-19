package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	bm "go-common/library/net/http/blademaster"

	"github.com/pkg/errors"
)

type endpoint struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

// New will create a reverse proxy handler
func New(rawurl string) bm.HandlerFunc {
	ep := newep(rawurl)
	return ep.serveHTTP
}

func newep(rawurl string) *endpoint {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(errors.Errorf("Invalid URL: %s", rawurl))
	}
	e := &endpoint{
		url: u,
	}
	e.proxy = &httputil.ReverseProxy{Director: e.director}
	return e
}

func (e *endpoint) director(req *http.Request) {
	req.URL.Scheme = e.url.Scheme
	req.URL.Host = e.url.Host
	req.URL.Path = e.url.Path
}

func (e *endpoint) serveHTTP(ctx *bm.Context) {
	e.proxy.ServeHTTP(ctx.Writer, ctx.Request)
}
