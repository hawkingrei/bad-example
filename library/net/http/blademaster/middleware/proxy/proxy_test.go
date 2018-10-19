package proxy

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"github.com/stretchr/testify/assert"
)

func init() {
	log.Init(nil)
}

func TestProxy(t *testing.T) {
	engine := bm.Default()
	engine.GET("/icon", New("http://api.hhhhhh.com/x/web-interface/index/icon"))

	go engine.Run(":18080")
	defer func() {
		engine.Server().Shutdown(context.TODO())
	}()
	time.Sleep(time.Second)

	req, err := http.NewRequest("GET", "http://127.0.0.1:18080/icon", nil)
	assert.NoError(t, err)
	req.Host = "api.hhhhhh.com"

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestProxyRace(t *testing.T) {
	engine := bm.Default()
	engine.GET("/icon", New("http://api.hhhhhh.com/x/web-interface/index/icon"))

	go engine.Run(":18080")
	defer func() {
		engine.Server().Shutdown(context.TODO())
	}()
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", "http://127.0.0.1:18080/icon", nil)
			assert.NoError(t, err)
			req.Host = "api.hhhhhh.com"

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, 200, resp.StatusCode)
		}()
	}
	wg.Wait()
}

func BenchmarkProxy(b *testing.B) {
	engine := bm.Default()
	engine.GET("/icon", New("http://api.hhhhhh.com/x/web-interface/index/icon"))

	go engine.Run(":18080")
	defer func() {
		engine.Server().Shutdown(context.TODO())
	}()
	time.Sleep(time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest("GET", "http://127.0.0.1:18080/icon", nil)
			assert.NoError(b, err)
			req.Host = "api.hhhhhh.com"

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(b, err)
			defer resp.Body.Close()
			assert.Equal(b, 200, resp.StatusCode)
		}
	})
}
