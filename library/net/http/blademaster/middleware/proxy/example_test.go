package proxy_test

import (
	"go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/proxy"
)

// This example create several reverse proxy to show how to use proxy middleware.
// We proxy three path to `api.hhhhhh.com` and return response without any changes.
func Example() {
	proxies := map[string]string{
		"/index":        "http://api.hhhhhh.com/html/index",
		"/ping":         "http://api.hhhhhh.com/api/ping",
		"/api/versions": "http://api.hhhhhh.com/api/web/versions",
	}

	engine := blademaster.Default()
	for path, ep := range proxies {
		engine.GET(path, proxy.New(ep))
	}
	engine.Run(":18080")
}
