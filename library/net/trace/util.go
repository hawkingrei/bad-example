package trace

import (
	"math"
	"sync"

	"github.com/dgryski/go-farm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

var _tracePool = sync.Pool{New: func() interface{} { return new(trace) }}

func putTrace(t *trace) {
	_tracePool.Put(t)
}

func getTrace() *trace {
	t := _tracePool.Get().(*trace)
	*t = trace{}
	return t
}

var _tracePool2 = sync.Pool{New: func() interface{} { return new(trace2) }}

func putTrace2(t *trace2) {
	_tracePool2.Put(t)
}

func getTrace2() *trace2 {
	t := _tracePool2.Get().(*trace2)
	*t = trace2{}
	return t
}

func genID() uint64 {
	i := [16]byte(uuid.NewV1())
	return farm.Hash64(i[:]) % math.MaxInt64
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
