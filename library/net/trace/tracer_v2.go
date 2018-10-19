package trace

import (
	"strconv"
	"time"

	"go-common/library/conf/env"
)

type tracer2 struct {
	cfg         *Config
	propagators map[interface{}]propagator
	report      reporter
}

func (d *tracer2) New(title string) Trace {
	t := getTrace2()
	t.report = d.report

	t.StartAt = time.Now().UnixNano()
	t.TraceId = genID()
	t.Level = 1
	t.SpanId = t.TraceId
	t.ParentSpanId = 0
	if d.cfg != nil && d.cfg.DisableSample {
		t.Sample, t.sampled = 1, true
	} else {
		t.Sample, t.sampled = sampleTrace(title)
	}
	t.Family = env.AppID
	t.Title = title
	// TODO tags: _server/service
	t.SetTag(
		String(TagSpanKind, "server"),
		String(TagComponent, "service"),
	)
	return t
}

func (d *tracer2) Inject(t Trace, format interface{}, carrier interface{}) error {
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

func (d *tracer2) Extract(format interface{}, carrier interface{}) (Trace, error) {
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

	t := getTrace2()
	t.report = d.report
	t.TraceId = traceID
	t.sampled, _ = strconv.ParseBool(carr.Get(KeyTraceSampled))
	t.SpanId = genID()
	t.ParentSpanId, _ = strconv.ParseUint(carr.Get(KeyTraceSpanID), 10, 64)
	lv, _ := strconv.ParseInt(carr.Get(KeyTraceLevel), 10, 64)
	t.Level = int32(lv)
	t.Caller = carr.Get(KeyTraceCaller)
	t.StartAt = time.Now().UnixNano()
	t.Family = env.AppID
	// TODO _server/service
	t.SetTag(
		String(TagSpanKind, "server"),
		String(TagComponent, "service"),
	)
	return t, nil
}

func (d *tracer2) Close() error {
	return d.report.Close()
}
