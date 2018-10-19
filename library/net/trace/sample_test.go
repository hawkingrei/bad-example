package trace

import (
	"testing"

	"time"
)

func TestSampleRatio(t *testing.T) {
}

func TestSampleHits(t *testing.T) {
	s := newSampleHits(10, 100*time.Millisecond)
	tk := time.NewTicker(time.Millisecond)
	defer tk.Stop()
	var okCount int
	var sampleCount float32
	for i := 0; i < 1000; i++ {
		<-tk.C
		sample, ok := s("hello")
		if ok {
			sampleCount += sample
			okCount++
		}
	}
	if okCount < 90 || okCount > 110 {
		t.Errorf("expect okCount in [40~60] get %d", okCount)
	}
	t.Logf("avg sample %0.5f", sampleCount/float32(okCount))
}

func BenchmarkSampleRatio(b *testing.B) {
	s := sampleRatio(1)
	for i := 0; i < b.N; i++ {
		s("hell world")
	}
}

func BenchmarkSampleHits(b *testing.B) {
	s := newSampleHits(1, time.Second)
	for i := 0; i < b.N; i++ {
		s("hell world")
	}
}

func BenchmarkSampleRatioParallel(b *testing.B) {
	s := sampleRatio(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s("hell world")
		}
	})
}

func BenchmarkSampleHitsParallel(b *testing.B) {
	s := newSampleHits(1, time.Second)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s("hell world")
		}
	})
}
