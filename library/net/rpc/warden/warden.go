package warden

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	nmd "go-common/library/net/metadata"
	"go-common/library/net/netutil/breaker"
	"go-common/library/net/rpc/warden/status"
	"go-common/library/net/trace"
	"go-common/library/stat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	_family = "grpc"
	_noUser = "no_user"
)

var (
	statsClient = stat.RPCClient
	statsServer = stat.RPCServer
)

// handle return a new unary server interceptor for OpenTracing\Logging\LinkTimeout.
func (s *Server) handle() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, args *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			cancel func()
			user   string
			addr   string
			color  string
			remote string
		)
		start := time.Now()
		s.mutex.RLock()
		conf := s.conf
		s.mutex.RUnlock()
		// get derived timeout from grpc context,
		// compare with the warden configured,
		// and use the minimum one
		timeout := time.Duration(conf.Timeout)
		if dl, ok := ctx.Deadline(); ok {
			ctimeout := time.Until(dl)
			if ctimeout-time.Millisecond*20 > 0 {
				ctimeout = ctimeout - time.Millisecond*20
			}
			if timeout > ctimeout {
				timeout = ctimeout
			}
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()

		// get grpc metadata(trace & remote_ip & color)
		var t trace.Trace
		if gmd, ok := metadata.FromIncomingContext(ctx); ok {
			t, _ = trace.Extract(trace.GRPCFormat, gmd)
			if strs, ok := gmd[nmd.Color]; ok {
				color = strs[0]
			}
			if strs, ok := gmd[nmd.RemoteIP]; ok {
				remote = strs[0]
			}
		}
		if t == nil {
			t = trace.New(args.FullMethod)
		}
		// evil hack
		t.Visit(func(k, v string) {
			if k == trace.KeyTraceCaller {
				user = v
			}
		})

		if pr, ok := peer.FromContext(ctx); ok {
			addr = pr.Addr.String()
			t.SetTag(trace.String(trace.TagAddress, addr))
		}
		defer t.Finish(&err)

		// use common meta data context instead of grpc context
		ctx = nmd.NewContext(ctx, nmd.MD{
			nmd.Trace:    t,
			nmd.Color:    color,
			nmd.RemoteIP: remote,
		})

		resp, err = handler(ctx, req)
		ec := ecode.Cause(err)

		// monitor & logging
		dt := time.Since(start)
		if user == "" {
			user = _noUser
		}
		logging(ctx, user, addr, args.FullMethod, req, err, ec.Code(), dt.Seconds())
		statsServer.Timing(user, int64(dt/time.Millisecond), args.FullMethod)
		statsServer.Incr(user, args.FullMethod, ec.Error())
		err = status.ToStatus(ec)
		return
	}
}

// handle returns a new unary client interceptor for OpenTracing\Logging\LinkTimeout.
func (c *Client) handle() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		var (
			ok     bool
			cmd    nmd.MD
			t      trace.Trace
			gmd    metadata.MD
			conf   *ClientConfig
			cancel context.CancelFunc
			addr   string
			p      peer.Peer
		)
		ec := ecode.OK
		start := time.Now()
		// apm tracing
		if t, ok = trace.FromContext(ctx); ok {
			t = t.Fork(_family, method)
			defer t.Finish(&err)
		}
		// init outgoing metadata to carry trace and color info
		if gmd, ok = metadata.FromOutgoingContext(ctx); !ok {
			gmd = metadata.MD{}
			ctx = metadata.NewOutgoingContext(ctx, gmd)
		}
		trace.Inject(t, trace.GRPCFormat, gmd)
		c.mutex.RLock()
		if conf, ok = c.conf.Method[method]; !ok {
			conf = c.conf
		}
		c.mutex.RUnlock()
		brk := c.breaker.Get(method)
		if err = brk.Allow(); err != nil {
			statsClient.Incr(method, "breaker")
			return
		}
		defer onBreaker(brk, &err)
		_, ctx, cancel = conf.Timeout.Shrink(ctx)
		defer cancel()
		// meta color
		if cmd, ok = nmd.FromContext(ctx); ok {
			var color, ip string
			if color, ok = cmd[nmd.Color].(string); ok {
				gmd[nmd.Color] = []string{color}
			}
			if ip, ok = cmd[nmd.RemoteIP].(string); ok {
				gmd[nmd.RemoteIP] = []string{ip}
			}
		} else {
			cmd = nmd.MD{}
			ctx = nmd.NewContext(ctx, cmd)
		}
		opts = append(opts, grpc.Peer(&p))
		if err = invoker(ctx, method, req, reply, cc, opts...); err != nil {
			ec = status.ToEcode(err)
		}
		if p.Addr != nil {
			addr = p.Addr.String()
		}
		if t != nil {
			t.SetTag(trace.String(trace.TagAddress, addr), trace.String(trace.TagComment, ""))
		}
		// monitor & logging
		dt := time.Since(start)
		logging(ctx, "", addr, method, req, err, ec.Code(), dt.Seconds())
		statsClient.Timing(method, int64(dt/time.Millisecond))
		statsClient.Incr(method, ec.Error())
		if err != nil {
			err = ec
		}
		return
	}
}

func logging(ctx context.Context, user, addr, method string, args interface{}, err error, ret int, ts float64) {
	var (
		errmsg   string
		errstack string
	)
	lf := log.Infov
	if err != nil {
		lf = log.Errorv
		if ret > 0 {
			lf = log.Warnv
		}
		errmsg = err.Error()
		errstack = fmt.Sprintf("%+v", err)
	}

	lf(ctx,
		log.KV("user", user),
		log.KV("ip", addr),
		log.KV("path", method),
		log.KV("ret", ret),
		log.KV("args", args.(fmt.Stringer).String()),
		log.KV("stack", errstack),
		log.KV("error", errmsg),
		log.KV("ts", ts),
	)
}

func onBreaker(breaker breaker.Breaker, err *error) {
	if err != nil && *err != nil {
		if ecode.ServerErr.Equal(*err) || ecode.ServiceUnavailable.Equal(*err) || ecode.Deadline.Equal(*err) || ecode.LimitExceed.Equal(*err) {
			breaker.MarkFailed()
			return
		}
	}
	breaker.MarkSuccess()
}
