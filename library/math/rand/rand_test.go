package rand

import (
	"sync"
	"testing"
	"time"
)

func TestNewRand(t *testing.T) {
	r := New(NewSource(time.Now().UnixNano()))
	wait := sync.WaitGroup{}
	for i := 0; i < 1e3; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for k := 0; k < 1e4; k++ {
				res := r.Intn(30)
				if res >= 30 || res < 0 {
					t.Fatalf("must less than 30")
				}
				f := r.Float64()
				if f < 0 || f > 1.0 {
					t.Fatalf("f less than 1.0")
				}
			}
		}()
	}
	wait.Wait()
}
