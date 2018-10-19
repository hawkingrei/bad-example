package trace

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"go-common/library/stat/counter"
)

// sampleFunc trace while be ignored if any filter return false.
type sampleFunc func(title string) (float32, bool)

var (
	sampleFuncs []sampleFunc
)

func init() {
	rand.Seed(time.Now().UnixNano())

	sampleFuncs = append(sampleFuncs,
		newSampleHits(1, 5*time.Second),
		sampleTitle("/metrics", "/monitor/ping"), // prom & readness title
	)
}

func sampleTrace(title string) (float32, bool) {
	var sample float32 = 1
	for _, sf := range sampleFuncs {
		s, ok := sf(title)
		if !ok {
			return 0, false
		}
		sample = sample * s
	}
	return sample, true
}

// sampleTitle filter titles.
func sampleTitle(titles ...string) sampleFunc {
	tm := make(map[string]bool, len(titles))
	for _, title := range titles {
		tm[title] = true
	}
	return func(title string) (float32, bool) {
		return 1, !tm[title]
	}
}

// sampleRatio aggretive sample by except every second.
func sampleRatio(eps float32) sampleFunc {
	if eps < 0 {
		panic("trace: non-positive eps")
	}
	c := &counter.Group{
		New: func() counter.Counter {
			return counter.NewRolling(time.Second, 10)
		},
	}
	return func(title string) (float32, bool) {
		c.Add(title, 1)
		sample := eps / float32(c.Value(title))
		return sample, rand.Float32() < sample
	}
}

func newSampleHits(hits int, interval time.Duration) sampleFunc {
	sh := &sampleHits{
		interval: int64(interval),
		titleMap: make(map[string]*titleBundle),
		hits:     int32(hits),
	}
	return sh.sample
}

type titleBundle struct {
	interval   int32
	leftover   int32
	val        int32
	qps        int32
	lastModify int64
}

type sampleHits struct {
	rmx      sync.RWMutex
	titleMap map[string]*titleBundle
	interval int64
	hits     int32
}

func (s *sampleHits) sample(title string) (float32, bool) {
	s.rmx.RLock()
	h, ok := s.titleMap[title]
	s.rmx.RUnlock()
	if !ok {
		s.rmx.Lock()
		h = &titleBundle{
			leftover: s.hits,
		}
		s.titleMap[title] = h
		s.rmx.Unlock()
	}

	now := time.Now().UnixNano()
	lm := atomic.LoadInt64((&h.lastModify))

	var qps int32
	if now-lm > s.interval {
		qps = atomic.SwapInt32(&h.qps, 1)
		atomic.StoreInt32(&h.interval, qps/s.hits)
		atomic.SwapInt64(&h.lastModify, now)
		atomic.SwapInt32(&h.leftover, s.hits)
	} else {
		qps = atomic.AddInt32(&h.qps, 1)
	}
	interval := atomic.LoadInt32(&h.interval)
	val := atomic.AddInt32(&h.val, 1)
	if val < interval {
		return 0, false
	}
	atomic.StoreInt32(&h.val, 0)
	leftover := atomic.LoadInt32(&h.leftover)
	if leftover > 0 {
		atomic.AddInt32(&h.leftover, -1)
		if qps == 0 {
			return 1, true
		}
		return float32(s.hits) / float32(qps), true
	}
	return 0, false
}
