package trace_test

import (
	"context"

	"net/http"

	"go-common/library/net/trace"
)

func ExampleFromContext() {
	ctx := context.Background()
	t := trace.New("test")

	// save trace into context
	ctx = trace.NewContext(ctx, t)

	// get trace from context
	t2, ok := trace.FromContext(ctx)

	if !ok {
		// if trace not exists in context, you may want create new trace
		t2 = trace.New("xxxx")
	}

	// do something
	t2.Finish(nil)
	t.Finish(nil)
}

// ExampleHTTPFormat show how to use trace in http
func ExampleHTTPFormat() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get trace from http.header
		t, err := trace.Extract(trace.HTTPFormat, r.Header)
		if err != nil {
			if err != trace.ErrTraceNotFound {
				// you may want to log something, if you care
			}
			t = trace.New(r.URL.Path)
		}

		url := "http://api.example.com/ping"
		// propagation trace through http
		t2 := t.Fork("http_client", "/ping")
		t2.SetTag(trace.String(trace.TagHTTPMethod, http.MethodGet))
		t2.SetTag(trace.String(trace.TagHTTPURL, url))
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		if err = trace.Inject(t2, trace.HTTPFormat, req.Header); err != nil {
			// you may want to do something if err happend or just ignored
		}

		resp, err := http.DefaultClient.Do(req)
		t2.Finish(&err)
		if err != nil {
			return
		}
		t2.SetTag(trace.Int(trace.TagHTTPStatusCode, resp.StatusCode))
		// do someting with resp
		resp.Body.Close()

		t.Finish(nil)
	})
}
