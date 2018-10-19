package log

import (
	"bytes"
	"context"
	"encoding/json"
	stdlog "log"
	"net"
	"strconv"
	"sync"
	"time"

	"go-common/library/conf/env"
	xtime "go-common/library/time"
)

const (
	_agentTimeout = xtime.Duration(20 * time.Millisecond)
	_mergeWait    = 1 * time.Second
	_maxBuffer    = 10 * 1024 * 1024 // 10mb
	_defaultChan  = 2048

	_defaultAgentConfig = "unixgram:///var/run/lancer/collector.sock?timeout=100ms&chan=1024"
)

var (
	_logSeparator = []byte("\u0001")

	_defaultTaskIDs = map[string]string{
		env.DeployEnvFat1: "000069",
		env.DeployEnvUat:  "000069",
		env.DeployEnvPre:  "000161",
		env.DeployEnvProd: "000161",
	}
)

// AgentHandler agent struct.
type AgentHandler struct {
	c      *AgentConfig
	msgs   chan map[string]interface{}
	waiter sync.WaitGroup
	pool   sync.Pool
}

// AgentConfig agent config.
type AgentConfig struct {
	TaskID  string
	Buffer  int
	Proto   string         `dsn:"network"`
	Addr    string         `dsn:"address"`
	Chan    int            `dsn:"query.chan"`
	Timeout xtime.Duration `dsn:"query.timeout"`
}

// NewAgent a Agent.
func NewAgent(c *AgentConfig) (a *AgentHandler) {
	if c == nil {
		c = parseDSN(_agentDSN)
	}
	if len(c.TaskID) == 0 {
		c.TaskID = _defaultTaskIDs[env.DeployEnv]
	}
	a = &AgentHandler{
		c: c,
	}
	a.pool.New = func() interface{} {
		return make(map[string]interface{}, 20)
	}
	if c.Chan == 0 {
		c.Chan = _defaultChan
	}
	a.msgs = make(chan map[string]interface{}, c.Chan)
	if c.Timeout == 0 {
		c.Timeout = _agentTimeout
	}
	if c.Buffer == 0 {
		c.Buffer = 1
	}
	a.waiter.Add(1)
	go a.writeproc()
	return
}

func (h *AgentHandler) data() map[string]interface{} {
	return h.pool.Get().(map[string]interface{})
}

func (h *AgentHandler) free(d map[string]interface{}) {
	for k := range d {
		delete(d, k)
	}
	h.pool.Put(d)
}

// Log log to udp statsd daemon.
func (h *AgentHandler) Log(ctx context.Context, lv Level, args ...D) {
	if args == nil {
		return
	}
	d := h.data()
	for _, arg := range args {
		d[arg.Key] = arg.Value
	}
	// add extra fields
	addExtraField(ctx, d)
	fn := funcName()
	d[_levelValue] = lv
	d[_level] = lv.String()
	d[_time] = time.Now().Format(_timeFormat)
	d[_source] = fn
	c.filter(d)
	errIncr(lv, fn)
	select {
	case h.msgs <- d:
	default:
	}
}

// writeproc write data into connection.
func (h *AgentHandler) writeproc() {
	var (
		buf   bytes.Buffer
		conn  net.Conn
		err   error
		count int
		quit  bool
	)
	defer h.waiter.Done()
	taskID := []byte(h.c.TaskID)
	tick := time.NewTicker(_mergeWait)
	enc := json.NewEncoder(&buf)
	for {
		select {
		case d := <-h.msgs:
			if d == nil {
				quit = true
				goto DUMP
			}
			if buf.Len() >= _maxBuffer {
				buf.Reset() // avoid oom
			}
			now := time.Now()
			buf.Write(taskID)
			buf.Write([]byte(strconv.FormatInt(now.UnixNano()/1e6, 10)))
			enc.Encode(d)
			h.free(d)
			if count++; count < h.c.Buffer {
				buf.Write(_logSeparator)
				continue
			}
		case <-tick.C:
		}
		if conn == nil || err != nil {
			if conn, err = net.DialTimeout(h.c.Proto, h.c.Addr, time.Duration(h.c.Timeout)); err != nil {
				stdlog.Printf("net.DialTimeout(%s:%s) error(%v)\n", h.c.Proto, h.c.Addr, err)
				continue
			}
		}
	DUMP:
		if conn != nil && buf.Len() > 0 {
			count = 0
			if _, err = conn.Write(buf.Bytes()); err != nil {
				stdlog.Printf("conn.Write(%d bytes) error(%v)\n", buf.Len(), err)
				conn.Close()
			} else {
				// only succeed reset buffer, let conn reconnect.
				buf.Reset()
			}
		}
		if quit {
			if conn != nil && err == nil {
				conn.Close()
			}
			return
		}
	}
}

// Close close the connection.
func (h *AgentHandler) Close() (err error) {
	h.msgs <- nil
	h.waiter.Wait()
	return nil
}

// SetFormat .
func (h *AgentHandler) SetFormat(string) {
	// discard setformat
}
