// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package redis

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"go-common/library/container/pool"
	"go-common/library/net/trace"
	"go-common/library/stat"
	xtime "go-common/library/time"
)

var stats = stat.Cache
var beginTime, _ = time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")

var (
	errConnClosed = errors.New("redigo: connection closed")
)

const _family = "redis"

// Pool .
type Pool struct {
	*pool.Slice
	// config
	c *Config
}

// Config client settings.
type Config struct {
	*pool.Config

	Name         string // redis name, for trace
	Proto        string
	Addr         string
	Auth         string
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

// NewPool creates a new pool.
func NewPool(c *Config) (p *Pool) {
	if c.DialTimeout <= 0 || c.ReadTimeout <= 0 || c.WriteTimeout <= 0 {
		panic("must config redis timeout")
	}
	p1 := pool.NewSlice(c.Config)
	cnop := DialConnectTimeout(time.Duration(c.DialTimeout))
	rdop := DialReadTimeout(time.Duration(c.ReadTimeout))
	wrop := DialWriteTimeout(time.Duration(c.WriteTimeout))
	auop := DialPassword(c.Auth)
	// new pool
	p1.New = func(ctx context.Context) (io.Closer, error) {
		return Dial(c.Proto, c.Addr, cnop, rdop, wrop, auop)
	}

	p = &Pool{Slice: p1, c: c}
	return
}

// Get gets a connection. The application must close the returned connection.
// This method always returns a valid connection so that applications can defer
// error handling to the first use of the connection. If there is an error
// getting an underlying connection, then the connection Err, Do, Send, Flush
// and Receive methods return that error.
func (p *Pool) Get(ctx context.Context) Conn {
	c, err := p.Slice.Get(ctx)
	if err != nil {
		return errorConnection{err}
	}
	c1, _ := c.(Conn)
	return &pooledConnection{p: p, c: c1, ctx: ctx, now: beginTime}
}

// Close releases the resources used by the pool.
func (p *Pool) Close() error {
	return p.Slice.Close()
}

type pooledConnection struct {
	p     *Pool
	c     Conn
	state int

	now  time.Time
	cmds []string
	t    trace.Trace
	ctx  context.Context
}

var (
	sentinel     []byte
	sentinelOnce sync.Once
)

func initSentinel() {
	p := make([]byte, 64)
	if _, err := rand.Read(p); err == nil {
		sentinel = p
	} else {
		h := sha1.New()
		io.WriteString(h, "Oops, rand failed. Use time instead.")
		io.WriteString(h, strconv.FormatInt(time.Now().UnixNano(), 10))
		sentinel = h.Sum(nil)
	}
}

func pstat(cmd string, t time.Time, err error) {
	stats.Timing(fmt.Sprintf("redis:%s", cmd), int64(time.Since(t)/time.Millisecond))
	if err != nil {
		if msg := formatErr(err); msg != "" {
			stats.Incr("redis", msg)
		}
	}
}

func (pc *pooledConnection) Close() error {
	c := pc.c
	if _, ok := c.(errorConnection); ok {
		return nil
	}
	pc.c = errorConnection{errConnClosed}

	if pc.state&MultiState != 0 {
		c.Send("DISCARD")
		pc.state &^= (MultiState | WatchState)
	} else if pc.state&WatchState != 0 {
		c.Send("UNWATCH")
		pc.state &^= WatchState
	}
	if pc.state&SubscribeState != 0 {
		c.Send("UNSUBSCRIBE")
		c.Send("PUNSUBSCRIBE")
		// To detect the end of the message stream, ask the server to echo
		// a sentinel value and read until we see that value.
		sentinelOnce.Do(initSentinel)
		c.Send("ECHO", sentinel)
		c.Flush()
		for {
			p, err := c.Receive()
			if err != nil {
				break
			}
			if p, ok := p.([]byte); ok && bytes.Equal(p, sentinel) {
				pc.state &^= SubscribeState
				break
			}
		}
	}
	_, err := c.Do("")
	pc.p.Slice.Put(context.Background(), c, pc.state != 0 || c.Err() != nil)
	return err
}

func (pc *pooledConnection) Err() error {
	return pc.c.Err()
}

func key(args interface{}) (key string) {
	keys, _ := args.([]interface{})
	if len(keys) > 0 {
		key, _ = keys[0].(string)
	}
	return
}

func (pc *pooledConnection) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	now := time.Now()
	if t, ok := trace.FromContext(pc.ctx); ok {
		t = t.Fork(_family, commandName)
		t.SetTag(trace.String(trace.TagAddress, pc.p.c.Addr), trace.String(trace.TagComment, key(args)))
		defer t.Finish(&err)
	}
	ci := LookupCommandInfo(commandName)
	pc.state = (pc.state | ci.Set) &^ ci.Clear
	reply, err = pc.c.Do(commandName, args...)
	pstat(commandName, now, err)
	return
}

func (pc *pooledConnection) Send(commandName string, args ...interface{}) (err error) {
	if pc.t == nil {
		if t, ok := trace.FromContext(pc.ctx); ok {
			pc.t = t.Fork(_family, "pipeline")
			pc.t.SetTag(trace.String(trace.TagAddress, pc.p.c.Addr), trace.String(trace.TagComment, ""))
			pc.t.SetTag(trace.String(trace.TagAnnotation, fmt.Sprintf("%s %s", commandName, key(args))))
		}
	} else {
		pc.t.SetTag(trace.String(trace.TagAnnotation, fmt.Sprintf("%s %s", commandName, key(args))))
	}
	ci := LookupCommandInfo(commandName)
	pc.state = (pc.state | ci.Set) &^ ci.Clear
	if pc.now.Equal(beginTime) {
		// mark first send time
		pc.now = time.Now()
	}
	pc.cmds = append(pc.cmds, commandName)
	return pc.c.Send(commandName, args...)
}

func (pc *pooledConnection) Flush() error {
	return pc.c.Flush()
}

func (pc *pooledConnection) Receive() (reply interface{}, err error) {
	if pc.t != nil {
		defer pc.t.Finish(&err)
		pc.t = nil
	}
	reply, err = pc.c.Receive()
	if len(pc.cmds) > 0 {
		cmd := pc.cmds[0]
		pc.cmds = pc.cmds[1:]
		pstat(cmd, pc.now, err)
	}
	return
}

type errorConnection struct{ err error }

func (ec errorConnection) Do(string, ...interface{}) (interface{}, error) {
	return nil, ec.err
}
func (ec errorConnection) Send(string, ...interface{}) error { return ec.err }
func (ec errorConnection) Err() error                        { return ec.err }
func (ec errorConnection) Close() error                      { return ec.err }
func (ec errorConnection) Flush() error                      { return ec.err }
func (ec errorConnection) Receive() (interface{}, error)     { return nil, ec.err }
