package trace

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"go-common/library/conf/dsn"
	"go-common/library/conf/env"
	"go-common/library/net/metadata"
	"go-common/library/net/trace/proto"
	xtime "go-common/library/time"

	"github.com/pkg/errors"
)

const (
	_maxLevel = 64

	// class
	_classComponent = 0
	_classService   = 1

	// event
	_client = 0
	_server = 1

	// protocol
	_separator = "\001"
	_spanProto = byte(1)
	// _annoProto = byte(2)
)

// Trace Key
const (
	KeyTraceID       = "x1-bilispy-id"
	KeyTraceSpanID   = "x1-bilispy-spanid"
	KeyTraceParentID = "x1-bilispy-parentid"
	KeyTraceSampled  = "x1-bilispy-sampled"
	KeyTraceLevel    = "x1-bilispy-lv"
	KeyTraceCaller   = "x1-bilispy-user"
)

type ctxKey string

var (
	_ctxkey ctxKey = "go-common/net/trace.trace"

	_traceDSN = "unixgram:///var/run/dapper-collect/dapper-collect.sock?timeout=20ms"
)

// Config config.
type Config struct {
	// Report network e.g. unixgram, tcp, udp
	Network string `dsn:"network"`
	// For TCP and UDP networks, the addr has the form "host:port".
	// For Unix networks, the address must be a file system path.
	Addr string `dsn:"address"`
	// DEPRECATED
	Proto string `dsn:"network"`
	// DEPRECATED
	Chan int `dsn:"query.chan,"`
	// Report timeout
	Timeout xtime.Duration `dsn:"query.timeout,20ms"`
	// DisableSample
	DisableSample bool `dsn:"query.disable_sample"`
}

func init() {
	if v := os.Getenv("TRACE"); v != "" {
		_traceDSN = v
	}
	flag.StringVar(&_traceDSN, "trace", _traceDSN, "trace report dsn, or use TRACE env.")
}

func parseDSN(rawdsn string) *Config {
	d, err := dsn.Parse(rawdsn)
	if err != nil {
		panic(errors.Wrapf(err, "trace: invalid dsn: %s", rawdsn))
	}
	conf := new(Config)
	if _, err = d.Bind(conf); err != nil {
		panic(errors.Wrapf(err, "trace: invalid dsn: %s", rawdsn))
	}
	return conf
}

// Init init trace.
func Init(conf *Config) {
	if conf != nil {
		// NOTE compatible proto field
		conf.Network = conf.Proto
		fmt.Fprintf(os.Stderr, "[deprecated] trace.Init() with conf is Deprecated, argument will be ignored. please use flag -trace or env TRACE to configure trace.\n")
	} else {
		if !flag.Parsed() {
			fmt.Fprint(os.Stderr, "[trace] please call flag.Parse() before Init trace, some configure may not effect\n")
		}
		conf = parseDSN(_traceDSN)
	}
	_tracer = &tracer2{
		cfg: conf,
		propagators: map[interface{}]propagator{
			HTTPFormat: httpPropagator{},
			GRPCFormat: grpcPropagator{},
		},
		report: newReport(conf.Network, conf.Addr, time.Duration(conf.Timeout)),
	}
}

// Trace trace common interface.
type Trace interface {
	// Fork fork a trace with client trace.
	Fork(family, title string) Trace

	// Follow
	Follow(family, title string) Trace

	// Finish when trace finish call it.
	Finish(err *error)

	// Scan scan trace into info.
	// Deprecated: method Scan is deprecated, use Inject instead of Scan
	// Scan(ti *Info)

	// Adds a tag to the trace.
	//
	// If there is a pre-existing tag set for `key`, it is overwritten.
	//
	// Tag values can be numeric types, strings, or bools. The behavior of
	// other tag value types is undefined at the OpenTracing level. If a
	// tracing system does not know how to handle a particular value type, it
	// may ignore the tag, but shall not panic.
	// NOTE current only support legacy tag: TagAnnotation TagAddress TagComment
	// other will be ignore
	SetTag(tags ...*proto.Tag) Trace

	// LogFields is an efficient and type-checked way to record key:value
	// NOTE current unsupport
	SetLog(logs ...*proto.Log) Trace

	// Visit visits the k-v pair in trace, calling fn for each.
	Visit(fn func(k, v string))
}

// FromContext returns the trace bound to the context, if any.
func FromContext(ctx context.Context) (t Trace, ok bool) {
	if v := metadata.Value(ctx, metadata.Trace); v != nil {
		t, ok = v.(Trace)
		return
	}
	t, ok = ctx.Value(_ctxkey).(Trace)
	return
}

// NewContext new a trace context.
// NOTE: This method is not thread safe.
func NewContext(ctx context.Context, t Trace) context.Context {
	if md, ok := metadata.FromContext(ctx); ok {
		md[metadata.Trace] = t
		return ctx
	}
	return context.WithValue(ctx, _ctxkey, t)
}

// trace is server and client called trace info.
type trace struct {
	report reporter

	// trace
	id       uint64
	spanID   uint64
	parentID uint64
	level    int32
	sampled  bool
	caller   string
	sample   float32

	// data
	startTime int64  // startTime timestamp
	endTime   int64  // endTime timestamp
	address   string // ip:port address
	family    string // project name, service name
	title     string // method name, rpc name
	comment   string // comment
	class     int8   // class type
	err       error  // trace error info

	// event
	event int8 // trace event
}

func (t *trace) Visit(fn func(key, val string)) {
	fn(KeyTraceID, strconv.FormatUint(t.id, 10))
	fn(KeyTraceSpanID, strconv.FormatUint(t.spanID, 10))
	fn(KeyTraceParentID, strconv.FormatUint(t.parentID, 10))
	fn(KeyTraceSampled, strconv.FormatBool(t.sampled))
	fn(KeyTraceLevel, strconv.Itoa(int(t.level)))
	fn(KeyTraceCaller, t.caller)
}

func (t *trace) Fork(family, title string) Trace {
	t1 := getTrace()
	t1.report = t.report

	t1.startTime = time.Now().UnixNano()
	t1.id = t.id
	if t1.level = t.level + 1; t1.level > _maxLevel {
		t1.sampled = false
	} else {
		t1.sampled = t.sampled
	}
	if t1.sampled {
		t1.spanID = genID()
	}
	t1.parentID = t.spanID
	t1.family = family
	t1.title = title
	t1.event = _client
	t1.class = _classComponent
	t1.caller = env.AppID
	return t1
}

func (t *trace) Follow(family, title string) Trace {
	t1 := getTrace()
	t1.report = t.report

	t1.startTime = time.Now().UnixNano()
	t1.id = t.id
	if t1.level = t.level; t1.level > _maxLevel {
		t1.sampled = false
	} else {
		t1.sampled = t.sampled
	}
	if t1.sampled {
		t1.spanID = genID()
	}
	t1.parentID = t.spanID
	t1.family = family
	t1.title = title
	// TODO: event ?
	t1.event = _client
	t1.class = _classComponent
	t1.caller = env.AppID
	return t1
}

func (t *trace) SetTag(tags ...*proto.Tag) Trace {
	for _, tag := range tags {
		switch tag.Key {
		case TagAnnotation:
			if tag.Kind == proto.Tag_STRING {
				t.comment = string(tag.Value)
				t.event = _client
			}
		case TagAddress:
			if tag.Kind == proto.Tag_STRING {
				t.address = string(tag.Value)
			}
		case TagComment:
			if tag.Kind == proto.Tag_STRING {
				t.comment = string(tag.Value)
			}
		}
	}
	return t
}

func (t *trace) SetLog(logs ...*proto.Log) Trace {
	// FIXME(weicheng) implement log
	return t
}

func (t *trace) Finish(err *error) {
	if t == nil {
		return
	}
	t.endTime = time.Now().UnixNano()
	if err != nil {
		t.err = errors.Cause(*err)
	}
	if t.sampled {
		t.report.Report(t)
	}
	putTrace(t)
}

func (t *trace) String() string {
	return strconv.FormatUint(t.id, 10)
}

// WriteTo write trace into a buffer.
func (t *trace) WriteTo(w io.Writer) (n int64, err error) {
	if strings.Contains(t.comment, _separator) {
		t.comment = strings.Replace(t.comment, _separator, "", -1)
	}

	// write trace header
	if err = writeStringWithSep(w, &n, string(_spanProto)); err != nil {
		return
	}

	// write trace body
	if err = writeStringWithSep(w, &n, strconv.FormatInt(t.startTime, 10)); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.FormatInt(t.endTime, 10)); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.FormatUint(t.id, 10)); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.FormatUint(t.spanID, 10)); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.FormatUint(t.parentID, 10)); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.Itoa(int(t.event))); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.Itoa(int(t.level))); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.Itoa(int(t.class))); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, strconv.FormatFloat(float64(t.sample), 'E', -1, 32)); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, t.address); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, t.family); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, t.title); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, t.comment); err != nil {
		return
	}
	if err = writeStringWithSep(w, &n, t.caller); err != nil {
		return
	}
	// write err message
	var msg string
	if t.err != nil {
		msg = t.err.Error()
	}
	err = writeString(w, &n, msg)
	return
}

func writeString(w io.Writer, count *int64, data string) (err error) {
	var n int
	n, err = w.Write([]byte(data))
	*count += int64(n)
	return
}

func writeStringWithSep(w io.Writer, count *int64, data string) (err error) {
	var n int
	n, err = w.Write([]byte(data))
	*count += int64(n)
	if err == nil {
		n, err = w.Write([]byte(_separator))
		*count += int64(n)
	}
	return
}
