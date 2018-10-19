package rand

import (
	"math/rand"
	"sync"
)

// NewSource returns a new pseudo-random Source seeded with the given value.
// this source is safe for concurrent use by multiple goroutines.
func NewSource(seed int64) Source {
	return &lockedSource{src: rand.NewSource(seed).(rand.Source64)}
}

// A Source represents a source of uniformly-distributed
// pseudo-random int64 values in the range [0, 1<<63).
type Source interface {
	rand.Source
}

type lockedSource struct {
	lk  sync.Mutex
	src rand.Source64
}

func (r *lockedSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Uint64() (n uint64) {
	r.lk.Lock()
	n = r.src.Uint64()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}

// New returns a new Rand that uses random values from src
// to generate other random values.
func New(s Source) *Rand {
	return &Rand{r: rand.New(&lockedSource{src: s.(rand.Source64)})}
}

// A Rand is a source of random numbers.
type Rand struct {
	r *rand.Rand
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (r *Rand) Int63() int64 { return r.r.Int63() }

// Uint32 returns a pseudo-random 32-bit value as a uint32.
func (r *Rand) Uint32() uint32 { return r.r.Uint32() }

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func (r *Rand) Uint64() uint64 {
	return r.r.Uint64()
}

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func (r *Rand) Int31() int32 { return r.r.Int31() }

// Int returns a non-negative pseudo-random int.
func (r *Rand) Int() int { return r.r.Int() }

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (r *Rand) Int63n(n int64) int64 { return r.r.Int63n(n) }

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (r *Rand) Int31n(n int32) int32 { return r.r.Int31n(n) }

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (r *Rand) Intn(n int) int { return r.r.Intn(n) }

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func (r *Rand) Float64() float64 { return r.r.Float64() }

// Float32 returns, as a float32, a pseudo-random number in [0.0,1.0).
func (r *Rand) Float32() float32 { return r.r.Float32() }

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers [0,n).
func (r *Rand) Perm(n int) []int { return r.r.Perm(n) }
