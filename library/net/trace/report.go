package trace

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

const (
	// MaxPackageSize .
	_maxPackageSize = 4096
	// safe udp package size
	// MaxPackageSize = 508
	_dataChSize = 1024
)

// reporter trace reporter.
type reporter interface {
	Report(io.WriterTo)
	Close() error
}

// newReport with network address
func newReport(network, address string, timeout time.Duration) reporter {
	if timeout == 0 {
		timeout = 20 * time.Millisecond
	}
	report := &connReport{
		network: network,
		address: address,
		dataCh:  make(chan *bytes.Buffer, _dataChSize),
		done:    make(chan struct{}),
		timeout: timeout,
		bufPool: sync.Pool{New: func() interface{} { return &bytes.Buffer{} }},
	}
	go report.daemon()
	return report
}

type connReport struct {
	rmx    sync.RWMutex
	closed bool

	network, address string

	dataCh chan *bytes.Buffer

	conn net.Conn

	done chan struct{}

	timeout time.Duration

	bufPool sync.Pool
}

func (c *connReport) daemon() {
	for b := range c.dataCh {
		c.send(b)
	}
	c.done <- struct{}{}
}

func (c *connReport) Report(wt io.WriterTo) {
	c.rmx.RLock()
	defer c.rmx.RUnlock()
	if c.closed {
		return
	}

	buf := c.getBuf()
	n, err := wt.WriteTo(buf)
	if err != nil {
		c.Errorf("write to buf error: %s", err)
		c.putBuf(buf)
		return
	}

	if n > _maxPackageSize {
		c.Errorf("package too large length %d > %d", n, _maxPackageSize)
		return
	}

	select {
	case c.dataCh <- buf:
		return
	case <-time.After(100 * time.Millisecond):
		c.Errorf("write to data channel timeout")
	}
}

func (c *connReport) Close() error {
	c.rmx.Lock()
	c.closed = true
	c.rmx.Unlock()

	t := time.NewTimer(time.Second)
	close(c.dataCh)
	select {
	case <-t.C:
		c.closeConn()
		return fmt.Errorf("close report timeout force close")
	case <-c.done:
		return c.closeConn()
	}
}

func (c *connReport) send(buf *bytes.Buffer) {
	defer c.putBuf(buf)
	if c.conn == nil {
		if err := c.reconnect(); err != nil {
			c.Errorf("connect error: %s retry after second", err)
			time.Sleep(time.Second)
			return
		}
	}
	c.conn.SetWriteDeadline(time.Now().Add(100 * time.Microsecond))
	if _, err := buf.WriteTo(c.conn); err != nil {
		c.Errorf("write to conn error: %s, close connect", err)
		c.conn.Close()
		c.conn = nil
	}
}

func (c *connReport) reconnect() (err error) {
	c.conn, err = net.DialTimeout(c.network, c.address, c.timeout)
	return
}

func (c *connReport) closeConn() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *connReport) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (c *connReport) getBuf() *bytes.Buffer {
	return c.bufPool.Get().(*bytes.Buffer)
}

func (c *connReport) putBuf(buf *bytes.Buffer) {
	buf.Reset()
	c.bufPool.Put(buf)
}
