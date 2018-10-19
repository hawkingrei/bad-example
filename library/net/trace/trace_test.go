package trace

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"go-common/library/net/metadata"
	xtime "go-common/library/time"

	"github.com/pkg/errors"
)

var testtra Tracer
var testtra2 Tracer

func TestMain(m *testing.M) {
	testtra = &tracer{
		propagators: map[interface{}]propagator{
			HTTPFormat: httpPropagator{},
			GRPCFormat: grpcPropagator{},
		},
		report: newReport("unixgram", "/tmp/collect/dapper-collector.sock", time.Second),
	}
	testtra2 = &tracer2{
		propagators: map[interface{}]propagator{
			HTTPFormat: httpPropagator{},
			GRPCFormat: grpcPropagator{},
		},
		report: newReport("unixgram", "/tmp/collect/dapper-collector.sock", time.Second),
	}
	os.Exit(m.Run())
}

func TestTrace(t *testing.T) {
	t1 := testtra.New("test")
	t2 := t1.Fork("test f1", "test t1")
	t3 := t2.Fork("test f2", "test t2")
	t3.Finish(nil)
	t2.Finish(nil)
	t1.Finish(nil)
}

func TestTrace2(t *testing.T) {
	t1 := testtra2.New("test2")
	t2 := t1.Fork("test2 f1", "test2 t1")
	t3 := t2.Fork("test2 f2", "test2 t2")
	t3.Finish(nil)
	t2.Finish(nil)
	t1.Finish(nil)
}

type testWriter int

func (t *testWriter) Write(p []byte) (int, error) {
	n := bytes.Count(p, []byte(_separator))
	// 14 client/server
	// 7 annotation
	if n != 14 && n != 7 {
		panic(fmt.Errorf("trace separator len: %d", n))
	}
	log.Printf("Trace write: %s", bytes.Replace(p, []byte(_separator), []byte("|"), -1))
	return len(p), nil
}

func TestContext(t *testing.T) {
	t1 := getTrace()
	ctx := NewContext(context.Background(), t1)
	t2, ok := FromContext(ctx)
	if !ok {
		t.FailNow()
	}
	if t1 != t2 {
		t.FailNow()
	}
}

func TestTrace_SetInfo(t *testing.T) {
	tr := testtra.New("t")
	tr.SetTag(String(TagAddress, "a"), String(TagComment, "c"))
	tr1 := tr.(*trace)
	if tr1.address != "a" || tr1.comment != "c" {
		t.FailNow()
	}
}

func TestTrace_Finish(t *testing.T) {
	err := errors.New("test")
	tr := testtra.New("t")
	tr.Finish(&err)
}

func TestTrace_Annotation(t *testing.T) {
	tr := testtra.New("t")
	tr.SetTag(String(TagAnnotation, "test"))
	tr1 := tr.(*trace)
	if tr1.event != _client || tr1.comment != "test" {
		t.Errorf("tr1.event != _client || tr1.comment != \"test\"")
	}
}

func TestTrace_Fork(t *testing.T) {
	tr := testtra.New("t")
	tr1 := tr.Fork("f1", "t1")
	tr2 := tr.(*trace)
	tr3 := tr1.(*trace)
	if tr3.family != "f1" || tr3.title != "t1" || tr2.id != tr3.id || tr2.spanID == tr3.spanID || tr3.parentID != tr2.spanID || tr3.class != _classComponent || tr2.level != (tr3.level-1) || tr2.sampled != tr3.sampled {
		t.FailNow()
	}
}

func TestWithMetaCtx(t *testing.T) {
	t1 := testtra.New("t")

	md := metadata.New(map[string]interface{}{})
	mctx := metadata.NewContext(context.Background(), md)

	ctx := NewContext(mctx, t1)
	t2, ok := FromContext(ctx)
	if !ok || t1 != t2 {
		t.Errorf("Failed to retrive trace from metadata context")
	}
}

func TestWithMetaCtx1(t *testing.T) {
	t1 := testtra.New("t")

	md := metadata.New(map[string]interface{}{
		"trace": t1,
	})
	mctx := metadata.NewContext(context.Background(), md)

	t2, ok := FromContext(mctx)
	if !ok || t1 != t2 {
		t.Fatal("Failed to retrive trace from metadata context")
	}
}

func BenchmarkNewTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testtra.New("t")
	}
}

func Test_parseDSN(t *testing.T) {
	type args struct {
		dsn string
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		{
			name: "naked dsn",
			args: args{
				dsn: "unixgram:///var/run/dapper-collect/dapper-collect.sock",
			},
			want: &Config{
				Addr:    "/var/run/dapper-collect/dapper-collect.sock",
				Network: "unixgram",
				Proto:   "unixgram",
				Timeout: xtime.Duration(20 * time.Millisecond),
			},
		},
		{
			name: "default dsn",
			args: args{
				dsn: "unixgram:///var/run/dapper-collect/dapper-collect.sock?timeout=20ms",
			},
			want: &Config{
				Addr:    "/var/run/dapper-collect/dapper-collect.sock",
				Network: "unixgram",
				Proto:   "unixgram",
				Timeout: xtime.Duration(20 * time.Millisecond),
			},
		},
		{
			name: "set timeout",
			args: args{
				dsn: "unixgram:///var/run/dapper-collect/dapper-collect.sock?timeout=1s",
			},
			want: &Config{
				Addr:    "/var/run/dapper-collect/dapper-collect.sock",
				Network: "unixgram",
				Proto:   "unixgram",
				Timeout: xtime.Duration(time.Second),
			},
		},
		{
			name: "tcp dsn",
			args: args{
				dsn: "tcp://10.0.0.1:2233?timeout=1s",
			},
			want: &Config{
				Addr:    "10.0.0.1:2233",
				Network: "tcp",
				Proto:   "tcp",
				Timeout: xtime.Duration(time.Second),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDSN(tt.args.dsn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}
