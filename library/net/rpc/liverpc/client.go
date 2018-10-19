package liverpc

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"go-common/library/conf/env"
	"go-common/library/naming"
	"go-common/library/naming/discovery"
	"go-common/library/net/trace"
	"go-common/library/stat"
	xtime "go-common/library/time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// Key is ContextKey
type Key int

const (
	_ Key = iota
	// KeyHeader use this in context to pass rpc header field
	KeyHeader
	// KeyHTTP use this in context to pass rpc http field
	KeyHTTP
	// KeyTimeout use this to pass rpc timeout
	KeyTimeout
)

const (
	_scheme      = "liverpc"
	_dialRetries = 3
)

// Get Implement tracer carrier interface
func (m *Header) Get(key string) string {
	if key == trace.KeyTraceID {
		return m.TraceId
	}
	return ""
}

// Set Implement tracer carrier interface
func (m *Header) Set(key string, val string) {
	if key == trace.KeyTraceID {
		m.TraceId = val
	}
}

var (
	// ErrNoClient no RPC client.
	errNoClient     = errors.New("no rpc client")
	errGroupInvalid = errors.New("invalid group")

	stats = stat.RPCClient

	_defaultGroup = "default"
)

// GroupAddrs a map struct storing addrs vary groups
type GroupAddrs map[string][]string

// ClientConfig client config.
type ClientConfig struct {
	AppID       string
	Group       string
	Timeout     xtime.Duration
	ConnTimeout xtime.Duration
	Addr        string // if addr is provided, it will use add, else, use discovery
}

// Client is a RPC client.
type Client struct {
	conf         *ClientConfig
	dis          naming.Resolver
	addrs        atomic.Value // GroupAddrs
	defaultGroup string
	addrsIdx     int64
}

// NewClient new a RPC client with discovery.
func NewClient(c *ClientConfig) *Client {
	if c.Timeout <= 0 {
		c.Timeout = xtime.Duration(time.Second)
	}
	if c.ConnTimeout <= 0 {
		c.ConnTimeout = xtime.Duration(time.Second)
	}
	cli := &Client{
		conf:         c,
		defaultGroup: getGroup(),
	}
	if c.Addr != "" {
		groupAddrs := make(GroupAddrs)
		groupAddrs[cli.defaultGroup] = []string{c.Addr}
		cli.addrs.Store(groupAddrs)
		return cli
	}

	cli.dis = discovery.Build(c.AppID)
	// discovery watch & fetch nodes
	event := cli.dis.Watch()
	select {
	case _, ok := <-event:
		if !ok {
			panic("刚启动就从discovery拉到了关闭的event")
		}
		cli.disFetch()
		fmt.Printf("开始创建：%s 的liverpc client，等待从discovery拉取节点：%s\n", c.AppID, time.Now().Format("2006-01-02 15:04:05"))
	case <-time.After(10 * time.Second):
		fmt.Printf("失败创建：%s 的liverpc client，竟然从discovery拉取节点超时了：%s\n", c.AppID, time.Now().Format("2006-01-02 15:04:05"))
	}
	go cli.disproc(event)
	return cli
}

func (c *Client) disproc(event <-chan struct{}) {
	for {
		_, ok := <-event
		if !ok {
			return
		}
		c.disFetch()
	}
}

func (c *Client) disFetch() {
	ins, ok := c.dis.Fetch()
	if !ok {
		return
	}
	insZone, ok := ins[env.Zone]
	if !ok {
		return
	}
	addrs := make(GroupAddrs)
	for _, svr := range insZone {
		group, ok := svr.Metadata["group"]
		if !ok {
			group = c.defaultGroup
		}
		for _, addr := range svr.Addrs {
			u, err := url.Parse(addr)
			if err == nil && u.Scheme == _scheme {
				addrs[group] = append(addrs[group], u.Host)
			}
		}
	}
	if len(addrs) > 0 {
		c.addrs.Store(addrs)
	}
}

// pickConn pick conn by addrs
func (c *Client) pickConn(ctx context.Context, addrs []string) (*ClientConn, error) {
	var (
		lastErr error
	)
	if len(addrs) == 0 {
		lastErr = errors.New("addrs empty")
	} else {
		for i := 0; i < _dialRetries; i++ {
			idx := atomic.AddInt64(&c.addrsIdx, 1)
			addr := addrs[int(idx)%len(addrs)]
			cc, err := Dial(ctx, "tcp", addr, time.Duration(c.conf.Timeout), time.Duration(c.conf.ConnTimeout))
			if err != nil {
				lastErr = errors.Wrapf(err, "Dial %s error", addr)
				continue
			}
			return cc, nil
		}
	}
	if lastErr != nil {
		return nil, errors.WithMessage(errNoClient, lastErr.Error())
	}
	return nil, errors.WithStack(errNoClient)
}

// fetchAddrs fetch addrs by different strategies
// source_group first, come from request header if exists, currently only CallRaw supports source_group
// then env group, come from os.env
// since no invalid group found, return error
func (c *Client) fetchAddrs(request interface{}) (addrs []string, err error) {
	var (
		args        *Args
		groupAddrs  GroupAddrs
		ok          bool
		sourceGroup string
		groups      []string
	)
	defer func() {
		if err != nil {
			err = errors.WithMessage(errGroupInvalid, err.Error())
		}
	}()
	// try parse request header and fetch source group
	if args, ok = request.(*Args); ok && args.Header != nil {
		sourceGroup = args.Header.SourceGroup
		if sourceGroup != "" {
			groups = append(groups, sourceGroup)
		}
	}
	if c.defaultGroup != sourceGroup {
		groups = append(groups, c.defaultGroup)
	}
	if groupAddrs, ok = c.addrs.Load().(GroupAddrs); !ok {
		err = errors.New("addrs load error")
		return
	}
	if len(groupAddrs) == 0 {
		err = errors.New("group addrs empty")
		return
	}
	for _, group := range groups {
		if addrs, ok = groupAddrs[group]; ok {
			break
		}
	}
	if len(addrs) == 0 {
		err = errors.Errorf("addrs empty source(%s), default(%s)", sourceGroup, c.defaultGroup)
		return
	}
	return
}

// Call call the service method, waits for it to complete, and returns its error status.
// client: {service}
// serviceMethod: {version}|{controller.method}
// httpURL: /room/v1/Room/room_init
// httpURL: /{service}/{version}/{controller}/{method}
func (c *Client) Call(ctx context.Context, version int, serviceMethod string, in proto.Message, out proto.Message) (err error) {
	var (
		cc    *ClientConn
		addrs []string
	)
	defer func() {
		if cc != nil {
			cc.Close()
		}
	}() // for now it is non-persistent connection
	addrs, err = c.fetchAddrs(in)
	if err != nil {
		return
	}
	cc, err = c.pickConn(ctx, addrs)
	if err != nil {
		return
	}
	return cc.Call(ctx, version, serviceMethod, in, out)
}

// CallRaw call the service method, waits for it to complete, and returns reply its error status.
// this is can be use without protobuf
// client: {service}
// serviceMethod: {version}|{controller.method}
// httpURL: /room/v1/Room/room_init
// httpURL: /{service}/{version}/{controller}/{method}
func (c *Client) CallRaw(ctx context.Context, version int, serviceMethod string, in *Args) (out *Reply, err error) {
	var (
		cc    *ClientConn
		addrs []string
	)
	defer func() {
		if cc != nil {
			cc.Close()
		}
	}() // for now it is non-persistent connection
	addrs, err = c.fetchAddrs(in)
	if err != nil {
		return
	}
	cc, err = c.pickConn(ctx, addrs)
	if err != nil {
		return
	}
	return cc.CallRaw(ctx, version, serviceMethod, in)
}

//Close handle client exit
func (c *Client) Close() {
	if c.dis != nil {
		c.dis.Close()
	}
}

// getGroup get current group
func getGroup() (g string) {
	g = os.Getenv("group")
	if g == "" {
		g = _defaultGroup
	}
	return
}
