package memcache

import (
	"context"
	"io"
	"strings"
	"time"

	"go-common/library/container/pool"
	"go-common/library/net/trace"
	"go-common/library/stat"
	xtime "go-common/library/time"
)

var stats = stat.Cache

const (
	_family = "memcache"
)

// Config memcache config.
type Config struct {
	*pool.Config

	Name         string // memcache name, for trace
	Proto        string
	Addr         string
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

// Pool memcache connection pool struct.
type Pool struct {
	p pool.Pool
	c *Config
}

// NewPool new a memcache conn pool.
func NewPool(c *Config) (p *Pool) {
	if c.DialTimeout <= 0 || c.ReadTimeout <= 0 || c.WriteTimeout <= 0 {
		panic("must config memcache timeout")
	}
	p1 := pool.NewList(c.Config)
	cnop := DialConnectTimeout(time.Duration(c.DialTimeout))
	rdop := DialReadTimeout(time.Duration(c.ReadTimeout))
	wrop := DialWriteTimeout(time.Duration(c.WriteTimeout))
	p1.New = func(ctx context.Context) (io.Closer, error) {
		return Dial(c.Proto, c.Addr, cnop, rdop, wrop)
	}

	p = &Pool{p: p1, c: c}
	return
}

// Get gets a connection. The application must close the returned connection.
// This method always returns a valid connection so that applications can defer
// error handling to the first use of the connection. If there is an error
// getting an underlying connection, then the connection Err, Do, Send, Flush
// and Receive methods return that error.
func (p *Pool) Get(ctx context.Context) Conn {
	c, err := p.p.Get(ctx)
	if err != nil {
		return errorConnection{err}
	}
	c1, _ := c.(Conn)
	return &pooledConnection{p: p, c: c1, ctx: ctx}
}

// Close release the resources used by the pool.
func (p *Pool) Close() error {
	return p.p.Close()
}

type pooledConnection struct {
	p   *Pool
	c   Conn
	ctx context.Context
}

func pstat(key string, t time.Time, err error) {
	stats.Timing(key, int64(time.Since(t)/time.Millisecond))
	if err != nil {
		if msg := formatErr(err); msg != "" {
			stats.Incr("memcache", msg)
		}
	}
}

func (pc *pooledConnection) setTrace(action, statement string) func(*error) {
	t, ok := trace.FromContext(pc.ctx)
	if !ok {
		return func(*error) {}
	}
	t = t.Fork(_family, action)
	t.SetTag(trace.String(trace.TagSpanKind, "client"))
	t.SetTag(trace.String(trace.TagComponent, "memcache"))
	t.SetTag(trace.String(trace.TagPeerAddress, pc.p.c.Addr))
	t.SetTag(trace.String(trace.TagDBStatement, action+" "+statement))
	return func(perr *error) {
		t.Finish(perr)
	}
}

func (pc *pooledConnection) Close() error {
	c := pc.c
	if _, ok := c.(errorConnection); ok {
		return nil
	}
	pc.c = errorConnection{ErrConnClosed}
	pc.p.p.Put(context.Background(), c, c.Err() != nil)
	return nil
}

func (pc *pooledConnection) Err() error {
	return pc.c.Err()
}

func (pc *pooledConnection) Set(item *Item) (err error) {
	defer pc.setTrace("set", item.Key)(&err)
	now := time.Now()
	err = pc.c.Set(item)
	pstat("memcache:set", now, err)
	return
}

func (pc *pooledConnection) Add(item *Item) (err error) {
	defer pc.setTrace("add", item.Key)(&err)
	now := time.Now()
	err = pc.c.Add(item)
	pstat("memcache:add", now, err)
	return
}

func (pc *pooledConnection) Replace(item *Item) (err error) {
	defer pc.setTrace("replace", item.Key)(&err)
	now := time.Now()
	err = pc.c.Replace(item)
	pstat("memcache:replace", now, err)
	return
}

func (pc *pooledConnection) CompareAndSwap(item *Item) (err error) {
	defer pc.setTrace("cas", item.Key)(&err)
	now := time.Now()
	err = pc.c.CompareAndSwap(item)
	pstat("memcache:cas", now, err)
	return
}

func (pc *pooledConnection) Get(key string) (r *Item, err error) {
	defer pc.setTrace("get", key)(&err)
	now := time.Now()
	r, err = pc.c.Get(key)
	pstat("memcache:get", now, err)
	return
}

func (pc *pooledConnection) GetMulti(keys []string) (res map[string]*Item, err error) {
	defer pc.setTrace("gets", strings.Join(keys, ","))(&err)
	now := time.Now()
	res, err = pc.c.GetMulti(keys)
	pstat("memcache:gets", now, err)
	return
}

func (pc *pooledConnection) Touch(key string, timeout int32) (err error) {
	defer pc.setTrace("touch", key)(&err)
	now := time.Now()
	err = pc.c.Touch(key, timeout)
	pstat("memcache:touch", now, err)
	return
}

func (pc *pooledConnection) Scan(item *Item, v interface{}) error {
	return pc.c.Scan(item, v)
}

func (pc *pooledConnection) Delete(key string) (err error) {
	defer pc.setTrace("delete", key)(&err)
	now := time.Now()
	err = pc.c.Delete(key)
	pstat("memcache:delete", now, err)
	return
}

func (pc *pooledConnection) Increment(key string, delta uint64) (newValue uint64, err error) {
	defer pc.setTrace("incr", key)(&err)
	now := time.Now()
	newValue, err = pc.c.Increment(key, delta)
	pstat("memcache:increment", now, err)
	return
}

func (pc *pooledConnection) Decrement(key string, delta uint64) (newValue uint64, err error) {
	defer pc.setTrace("decr", key)(&err)
	now := time.Now()
	newValue, err = pc.c.Decrement(key, delta)
	pstat("memcache:decrement", now, err)
	return
}

type errorConnection struct{ err error }

func (ec errorConnection) Err() error                                         { return ec.err }
func (ec errorConnection) Close() error                                       { return ec.err }
func (ec errorConnection) Add(item *Item) error                               { return ec.err }
func (ec errorConnection) Set(item *Item) error                               { return ec.err }
func (ec errorConnection) Replace(item *Item) error                           { return ec.err }
func (ec errorConnection) CompareAndSwap(item *Item) error                    { return ec.err }
func (ec errorConnection) Get(key string) (*Item, error)                      { return nil, ec.err }
func (ec errorConnection) GetMulti(keys []string) (map[string]*Item, error)   { return nil, ec.err }
func (ec errorConnection) Touch(key string, timeout int32) error              { return ec.err }
func (ec errorConnection) Delete(key string) error                            { return ec.err }
func (ec errorConnection) Increment(key string, delta uint64) (uint64, error) { return 0, ec.err }
func (ec errorConnection) Decrement(key string, delta uint64) (uint64, error) { return 0, ec.err }
func (ec errorConnection) Scan(item *Item, v interface{}) error               { return ec.err }
