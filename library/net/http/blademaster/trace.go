package blademaster

import (
	"context"
	"net/http"

	"go-common/library/net/trace"
)

const _component = "http"

// Trace is trace middleware
func Trace() HandlerFunc {
	return func(c *Context) {
		// handle http request
		// get derived trace from http request header
		t, err := trace.Extract(trace.HTTPFormat, c.Request.Header)
		if err != nil {
			t = trace.New(c.Request.URL.Path)
		}
		t.SetTag(trace.String(trace.TagComponent, _component))
		t.SetTag(trace.String(trace.TagHTTPMethod, c.Request.Method))
		t.SetTag(trace.String(trace.TagHTTPURL, c.Request.URL.String()))
		t.SetTag(trace.String(trace.TagSpanKind, "server"))
		c.Context = trace.NewContext(c.Context, t)
		c.Next()
		t.Finish(&c.Error)
	}
}

// traceClient attach trace to http.Request
func traceClient(ctx context.Context, req *http.Request) (trace.Trace, bool) {
	urlNoQuery := req.URL.Scheme + "://" + req.Host + req.URL.Path
	t, ok := trace.FromContext(ctx)
	if !ok {
		return nil, false
	}
	t = t.Fork(_component, urlNoQuery)
	t.SetTag(trace.String(trace.TagComponent, _component))
	t.SetTag(trace.String(trace.TagHTTPMethod, req.Method))
	t.SetTag(trace.String(trace.TagHTTPURL, req.URL.String()))
	t.SetTag(trace.String(trace.TagSpanKind, "client"))
	trace.Inject(t, trace.HTTPFormat, req.Header)
	return t, true
}
