package pipeline

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"go-common/library/net/metadata"
	xtime "go-common/library/time"
)

func TestPipeline(t *testing.T) {
	conf := &Config{
		MaxSize:  3,
		Interval: xtime.Duration(time.Millisecond * 20),
		Buffer:   3,
		Worker:   10,
	}
	type recv struct {
		mirror bool
		ch     int
		values map[string][]interface{}
	}
	var runs []recv
	do := func(c context.Context, ch int, values map[string][]interface{}) {
		runs = append(runs, recv{
			mirror: metadata.Bool(c, metadata.Mirror),
			values: values,
			ch:     ch,
		})
	}
	split := func(s string) int {
		n, _ := strconv.Atoi(s)
		return n
	}
	p := NewPipeline(conf)
	p.Do = do
	p.Split = split
	p.Start()
	p.Add(context.Background(), "1", 1)
	p.Add(context.Background(), "1", 2)
	p.Add(context.Background(), "11", 3)
	p.Add(context.Background(), "2", 3)
	time.Sleep(time.Millisecond * 60)
	mirrorCtx := metadata.NewContext(context.Background(), metadata.MD{metadata.Mirror: true})
	p.Add(mirrorCtx, "2", 3)
	time.Sleep(time.Millisecond * 60)
	p.SyncAdd(mirrorCtx, "5", 5)
	time.Sleep(time.Millisecond * 60)
	p.Close()
	expt := []recv{
		{
			mirror: false,
			ch:     1,
			values: map[string][]interface{}{
				"1":  {1, 2},
				"11": {3},
			},
		},
		{
			mirror: false,
			ch:     2,
			values: map[string][]interface{}{
				"2": {3},
			},
		},
		{
			mirror: true,
			ch:     2,
			values: map[string][]interface{}{
				"2": {3},
			},
		},
		{
			mirror: true,
			ch:     5,
			values: map[string][]interface{}{
				"5": {5},
			},
		},
	}
	if !reflect.DeepEqual(runs, expt) {
		t.Errorf("expect get %+v,\n got: %+v", expt, runs)
	}
}
