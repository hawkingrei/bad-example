package trace

import (
	"net/http"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestHttpPropagator(t *testing.T) {
	tra := &tracer{
		propagators: map[interface{}]propagator{
			HTTPFormat: httpPropagator{},
			GRPCFormat: grpcPropagator{},
		},
	}
	header := http.Header{}
	t1 := tra.New("test")
	if err := tra.Inject(t1, HTTPFormat, header); err != nil {
		t.Error(err)
	}
	t2, err := tra.Extract(HTTPFormat, header)
	if err != nil {
		t.Error(err)
	}
	rt1 := t1.(*trace)
	rt2 := t2.(*trace)
	if rt1.id != rt2.id || rt1.spanID != rt2.parentID || rt1.level != rt2.level || rt1.caller != rt2.caller || rt1.sampled != rt2.sampled {
		t.Errorf("error trace %+v => %+v", rt1, rt2)
	}
}

func TestGRPCPropagator(t *testing.T) {
	tra := &tracer{
		propagators: map[interface{}]propagator{
			HTTPFormat: httpPropagator{},
			GRPCFormat: grpcPropagator{},
		},
	}
	md := metadata.MD{}
	t1 := tra.New("test")
	if err := tra.Inject(t1, GRPCFormat, md); err != nil {
		t.Error(err)
	}
	t2, err := tra.Extract(GRPCFormat, md)
	if err != nil {
		t.Error(err)
	}
	rt1 := t1.(*trace)
	rt2 := t2.(*trace)
	if rt1.id != rt2.id || rt1.spanID != rt2.parentID || rt1.level != rt2.level || rt1.caller != rt2.caller || rt1.sampled != rt2.sampled {
		t.Errorf("error trace %+v => %+v", rt1, rt2)
	}
}
