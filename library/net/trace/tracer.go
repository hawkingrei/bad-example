package trace

import (
	"io"
	"strconv"
	"time"

	"go-common/library/conf/env"
	"go-common/library/net/trace/proto"
)

var (
	_ Tracer = &tracer{}

	// global tracer
	_tracer Tracer = nooptracer{}
)

// Tracer is a simple, thin interface for Trace creation and propagation.
type Tracer interface {
	// New trace instance with given title.
	New(title string) Trace
	// Inject takes the Trace instance and injects it for
	// propagation within `carrier`. The actual type of `carrier` depends on
	// the value of `format`.
	Inject(t Trace, format interface{}, carrier interface{}) error
	// Extract returns a Trace instance given `format` and `carrier`.
	// return `ErrTraceNotFound` if trace not found.
	Extract(format interface{}, carrier interface{}) (Trace, error)
}

// New trace instance with given title.
func New(title string) Trace {
	return _tracer.New(title)
}

// Inject takes the Trace instance and injects it for
// propagation within `carrier`. The actual type of `carrier` depends on
// the value of `format`.
func Inject(t Trace, format interface{}, carrier interface{}) error {
	return _tracer.Inject(t, format, carrier)
}

// Extract returns a Trace instance given `format` and `carrier`.
// return `ErrTraceNotFound` if trace not found.
func Extract(format interface{}, carrier interface{}) (Trace, error) {
	return _tracer.Extract(format, carrier)
}

// Close trace flush data.
func Close() error {
	if closer, ok := _tracer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type tracer struct {
	propagators map[interface{}]propagator
	report      reporter
}

func (d *tracer) New(title string) Trace {
	t := getTrace()
	t.report = d.report

	t.startTime = time.Now().UnixNano()
	t.id = genID()
	t.level = 1
	t.spanID = t.id
	t.parentID = 0
	t.sample, t.sampled = sampleTrace(title)
	t.family = env.AppID
	t.title = title
	t.event = _server
	t.class = _classService
	return t
}

func (d *tracer) Inject(t Trace, format interface{}, carrier interface{}) error {
	// if carrier implement Carrier use direct, ignore format
	carr, ok := carrier.(Carrier)
	if ok {
		t.Visit(carr.Set)
		return nil
	}
	// use Built-in propagators
	pp, ok := d.propagators[format]
	if !ok {
		return ErrUnsupportedFormat
	}
	carr, err := pp.Inject(carrier)
	if err != nil {
		return err
	}
	if t != nil {
		t.Visit(carr.Set)
	}
	return nil
}

func (d *tracer) Extract(format interface{}, carrier interface{}) (Trace, error) {
	// if carrier implement Carrier use direct, ignore format
	carr, ok := carrier.(Carrier)
	if !ok {
		// use Built-in propagators
		pp, ok := d.propagators[format]
		if !ok {
			return nil, ErrUnsupportedFormat
		}
		var err error
		if carr, err = pp.Extract(carrier); err != nil {
			return nil, err
		}
	}

	traceIDstr := carr.Get(KeyTraceID)
	if traceIDstr == "" {
		return nil, ErrTraceNotFound
	}
	traceID, err := strconv.ParseUint(traceIDstr, 10, 64)
	if err != nil {
		return nil, ErrTraceCorrupted
	}

	t := getTrace()
	t.report = d.report
	t.id = traceID
	t.sampled, _ = strconv.ParseBool(carr.Get(KeyTraceSampled))
	t.spanID = genID()
	t.parentID, _ = strconv.ParseUint(carr.Get(KeyTraceSpanID), 10, 64)
	lv, _ := strconv.ParseInt(carr.Get(KeyTraceLevel), 10, 64)
	t.level = int32(lv)
	t.caller = carr.Get(KeyTraceCaller)
	t.startTime = time.Now().UnixNano()
	t.family = env.AppID
	t.event = _server
	t.class = _classService
	return t, nil
}

func (d *tracer) Close() error {
	return d.report.Close()
}

var (
	_ Tracer = nooptracer{}
)

type nooptracer struct{}

func (n nooptracer) New(title string) Trace {
	return nooptrace{}
}

func (n nooptracer) Inject(t Trace, format interface{}, carrier interface{}) error {
	return nil
}

func (n nooptracer) Extract(format interface{}, carrier interface{}) (Trace, error) {
	return nooptrace{}, nil
}

type nooptrace struct{}

func (n nooptrace) Fork(family string, title string) Trace {
	return nooptrace{}
}

func (n nooptrace) Follow(family string, title string) Trace {
	return nooptrace{}
}

func (n nooptrace) Finish(err *error) {}

func (n nooptrace) SetTag(tags ...*proto.Tag) Trace {
	return nooptrace{}
}

func (n nooptrace) SetLog(logs ...*proto.Log) Trace {
	return nooptrace{}
}

func (n nooptrace) Visit(func(k, v string)) {}
