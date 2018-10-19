package blademaster

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-common/library/net/trace"
)

func TestTrace(t *testing.T) {
	wait := make(chan bool, 1)
	eng := New()
	eng.Use(Trace())
	eng.GET("/test-trace", func(c *Context) {
		if _, ok := trace.FromContext(c.Context); !ok {
			t.Errorf("expect get trace from context")
		}
		c.Writer.Write([]byte("pong"))
		wait <- true
	})
	go eng.Run("127.0.0.1:28080")
	http.Get("http://127.0.0.1:28080/test-trace")
	<-wait
}

func TestTraceClient(t *testing.T) {
	wait := make(chan bool, 1)
	trace.Init(nil)
	root := trace.New("test-title")
	ctx := trace.NewContext(context.Background(), root)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys := []string{trace.KeyTraceID, trace.KeyTraceLevel, trace.KeyTraceSampled, trace.KeyTraceSpanID, trace.KeyTraceParentID}
		for _, key := range keys {
			if r.Header.Get(key) == "" {
				t.Errorf("empty key: %s", key)
			}
		}
		wait <- true
	}))
	defer srv.Close()
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := traceClient(ctx, req)
	if !ok {
		t.Fatalf("inject trace fail")
	}
	if _, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	<-wait
}
