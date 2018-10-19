package trace

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"go-common/library/conf/env"
	"go-common/library/net/trace/proto"
)

type trace2 struct {
	report reporter

	proto.Span
	sampled bool
}

func (t *trace2) Visit(fn func(key, val string)) {
	fn(KeyTraceID, strconv.FormatUint(t.TraceId, 10))
	fn(KeyTraceSpanID, strconv.FormatUint(t.SpanId, 10))
	fn(KeyTraceParentID, strconv.FormatUint(t.ParentSpanId, 10))
	fn(KeyTraceSampled, strconv.FormatBool(t.sampled))
	fn(KeyTraceLevel, strconv.Itoa(int(t.Level)))
	fn(KeyTraceCaller, t.Caller)
}

func (t *trace2) Fork(family, title string) Trace {
	t1 := getTrace2()
	t1.report = t.report

	t1.StartAt = time.Now().UnixNano()
	t1.TraceId = t.TraceId
	if t1.Level = t.Level + 1; t1.Level > _maxLevel {
		t1.sampled = false
	} else {
		t1.sampled = t.sampled
	}
	if t1.sampled {
		t1.SpanId = genID()
	}
	t1.ParentSpanId = t.SpanId
	t1.Family = family
	t1.Title = title
	t1.Caller = env.AppID
	// TODO external set tags: _client/component
	t1.SetTag(
		String(TagSpanKind, "client"),
		String(TagComponent, "component"),
	)
	return t1
}

func (t *trace2) Follow(family, title string) Trace {
	t1 := getTrace2()
	t1.report = t.report

	t1.StartAt = time.Now().UnixNano()
	t1.TraceId = t.TraceId
	if t1.Level = t.Level; t1.Level > _maxLevel {
		t1.sampled = false
	} else {
		t1.sampled = t.sampled
	}
	if t1.sampled {
		t1.SpanId = genID()
	}
	t1.ParentSpanId = t.SpanId
	t1.Family = family
	t1.Title = title
	t1.Caller = env.AppID
	// TODO external set tags: _client/component
	t1.SetTag(
		String(TagSpanKind, "client"),
		String(TagComponent, "component"),
	)
	return t1
}

func (t *trace2) SetTag(tags ...*proto.Tag) Trace {
	t.Tags = append(t.Tags, tags...)
	return t
}

func (t *trace2) SetLog(logs ...*proto.Log) Trace {
	t.Logs = append(t.Logs, logs...)
	return t
}

func (t *trace2) Finish(perr *error) {
	if t == nil {
		return
	}
	if !t.sampled {
		return
	}
	t.FinishAt = time.Now().UnixNano()
	if perr != nil && *perr != nil {
		err := *perr
		t.SetTag(Bool(TagError, true))
		t.SetLog(Log(LogMessage, err.Error()))
		if err, ok := err.(stackTracer); ok {
			t.SetLog(Log(LogStack, fmt.Sprintf("%+v", err.StackTrace())))
		}
	}
	t.report.Report(t)
	putTrace2(t)
}

func (t *trace2) String() string {
	return strconv.FormatUint(t.TraceId, 10)
}

// WriteTo write trace into a buffer.
func (t *trace2) WriteTo(w io.Writer) (n int64, err error) {
	b, err := t.Span.Marshal()
	if err != nil {
		return
	}
	m, err := w.Write(b)
	if err != nil {
		return
	}
	n = int64(m)
	return
}
