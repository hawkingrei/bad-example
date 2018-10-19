package warden

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"go-common/library/conf/dsn"
	"go-common/library/log"
	xtime "go-common/library/time"

	//this package is for json format response
	_ "go-common/library/net/rpc/warden/encoding/json"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

var (
	_grpcDSN        string
	_defaultSerConf = &ServerConfig{
		Network:           "tcp",
		Addr:              "0.0.0.0:9000",
		Timeout:           xtime.Duration(time.Second),
		IdleTimeout:       xtime.Duration(time.Second * 60),
		MaxLifeTime:       xtime.Duration(time.Hour * 2),
		ForceCloseWait:    xtime.Duration(time.Second * 20),
		KeepAliveInterval: xtime.Duration(time.Second * 60),
		KeepAliveTimeout:  xtime.Duration(time.Second * 20),
	}
	_abortIndex int8 = math.MaxInt8 / 2
)

// ServerConfig is rpc server conf.
type ServerConfig struct {
	// Network is grpc listen network,default value is tcp
	Network string `dsn:"network"`
	// Addr is grpc listen addr,default value is 0.0.0.0:9000
	Addr string `dsn:"address"`
	// Timeout is context timeout for per rpc call.
	Timeout xtime.Duration `dsn:"query.timeout"`
	// IdleTimeout is a duration for the amount of time after which an idle connection would be closed by sending a GoAway.
	// Idleness duration is defined since the most recent time the number of outstanding RPCs became zero or the connection establishment.
	IdleTimeout xtime.Duration `dsn:"query.idleTimeout"`
	// MaxLifeTime is a duration for the maximum amount of time a connection may exist before it will be closed by sending a GoAway.
	// A random jitter of +/-10% will be added to MaxConnectionAge to spread out connection storms.
	MaxLifeTime xtime.Duration `dsn:"query.maxLife"`
	// ForceCloseWait is an additive period after MaxLifeTime after which the connection will be forcibly closed.
	ForceCloseWait xtime.Duration `dsn:"query.closeWait"`
	// KeepAliveInterval is after a duration of this time if the server doesn't see any activity it pings the client to see if the transport is still alive.
	KeepAliveInterval xtime.Duration `dsn:"query.keepaliveInterval"`
	// KeepAliveTimeout  is After having pinged for keepalive check, the server waits for a duration of Timeout and if no activity is seen even after that
	// the connection is closed.
	KeepAliveTimeout xtime.Duration `dsn:"query.keepaliveTimeout"`
}

// Server is the framework's server side instance, it contains the GrpcServer, interceptor and interceptors.
// Create an instance of Server, by using NewServer().
type Server struct {
	conf  *ServerConfig
	mutex sync.RWMutex

	server   *grpc.Server
	handlers []grpc.UnaryServerInterceptor
}

func init() {
	addFlag(flag.CommandLine)
}

func addFlag(fs *flag.FlagSet) {
	v := os.Getenv("GRPC")
	if v == "" {
		v = "tcp://0.0.0.0:9000/?timeout=1s&idle_timeout=60s"
	}
	fs.StringVar(&_grpcDSN, "grpc", v, "listen grpc dsn, or use GRPC env variable.")
}

func parseDSN(rawdsn string) *ServerConfig {
	conf := new(ServerConfig)
	d, err := dsn.Parse(rawdsn)
	if err != nil {
		panic(errors.WithMessage(err, fmt.Sprintf("warden: invalid dsn: %s", rawdsn)))
	}
	if _, err = d.Bind(conf); err != nil {
		panic(errors.WithMessage(err, fmt.Sprintf("warden: invalid dsn: %s", rawdsn)))
	}
	return conf
}

// NewServer returns a new blank Server instance with a default server interceptor.
func NewServer(conf *ServerConfig, opt ...grpc.ServerOption) (s *Server) {
	if conf == nil {
		if !flag.Parsed() {
			fmt.Fprint(os.Stderr, "[warden] please call flag.Parse() before Init warden server, some configure may not effect\n")
		}
		conf = parseDSN(_grpcDSN)
	} else {
		fmt.Fprintf(os.Stderr, "[warden] config is Deprecated, argument will be ignored. please use -grpc flag or GRPC env to configure warden server.\n")
	}
	s = new(Server)
	if err := s.SetConfig(conf); err != nil {
		panic(errors.Errorf("warden: set config failed!err: %s", err.Error()))
	}
	keepParam := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(s.conf.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(s.conf.ForceCloseWait),
		Time:             time.Duration(s.conf.KeepAliveInterval),
		Timeout:          time.Duration(s.conf.KeepAliveTimeout),
		MaxConnectionAge: time.Duration(s.conf.MaxLifeTime),
	})
	opt = append(opt, keepParam, grpc.UnaryInterceptor(s.interceptor))
	s.server = grpc.NewServer(opt...)
	s.Use(s.recovery(), s.handle(), s.stats(), s.validate())
	return
}

// SetConfig hot reloads server config
func (s *Server) SetConfig(conf *ServerConfig) (err error) {
	if conf == nil {
		conf = _defaultSerConf
	}
	if conf.Timeout <= 0 {
		conf.Timeout = xtime.Duration(time.Second)
	}
	if conf.IdleTimeout <= 0 {
		conf.IdleTimeout = xtime.Duration(time.Second * 60)
	}
	if conf.MaxLifeTime <= 0 {
		conf.MaxLifeTime = xtime.Duration(time.Hour * 2)
	}
	if conf.ForceCloseWait <= 0 {
		conf.ForceCloseWait = xtime.Duration(time.Second * 20)
	}
	if conf.KeepAliveInterval <= 0 {
		conf.KeepAliveInterval = xtime.Duration(time.Second * 60)
	}
	if conf.KeepAliveTimeout <= 0 {
		conf.KeepAliveTimeout = xtime.Duration(time.Second * 20)
	}
	if conf.Addr == "" {
		conf.Addr = "0.0.0.0:9000"
	}
	if conf.Network == "" {
		conf.Network = "tcp"
	}
	s.mutex.Lock()
	s.conf = conf
	s.mutex.Unlock()
	return nil
}

// interceptor is a single interceptor out of a chain of many interceptors.
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryServer(one, two, three) will execute one before two before three, and three
// will see context changes of one and two.
func (s *Server) interceptor(ctx context.Context, req interface{}, args *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var (
		i     int
		chain grpc.UnaryHandler
	)

	n := len(s.handlers)
	if n == 0 {
		return handler(ctx, req)
	}

	chain = func(ic context.Context, ir interface{}) (interface{}, error) {
		if i == n-1 {
			return handler(ic, ir)
		}
		i++
		return s.handlers[i](ic, ir, args, chain)
	}

	return s.handlers[0](ctx, req, args, chain)
}

// Server return the grpc server for registering service.
func (s *Server) Server() *grpc.Server {
	return s.server
}

// Use attachs a global inteceptor to the server.
// For example, this is the right place for a rate limiter or error management inteceptor.
func (s *Server) Use(handlers ...grpc.UnaryServerInterceptor) *Server {
	finalSize := len(s.handlers) + len(handlers)
	if finalSize >= int(_abortIndex) {
		panic("warden: server use too many handlers")
	}
	mergedHandlers := make([]grpc.UnaryServerInterceptor, finalSize)
	copy(mergedHandlers, s.handlers)
	copy(mergedHandlers[len(s.handlers):], handlers)
	s.handlers = mergedHandlers
	return s
}

// Run create a tcp listener and start goroutine for serving each incoming request.
// Run will return a non-nil error unless Stop or GracefulStop is called.
func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		err = errors.WithStack(err)
		log.Error("failed to listen: %v", err)
		return err
	}
	reflection.Register(s.server)
	return s.Serve(lis)
}

// RunUnix create a unix listener and start goroutine for serving each incoming request.
// RunUnix will return a non-nil error unless Stop or GracefulStop is called.
func (s *Server) RunUnix(file string) error {
	lis, err := net.Listen("unix", file)
	if err != nil {
		err = errors.WithStack(err)
		log.Error("failed to listen: %v", err)
		return err
	}
	reflection.Register(s.server)
	return s.Serve(lis)
}

// Start create a new goroutine run server with configured listen addr
// will panic if any error happend
// return server itself
func (s *Server) Start() (*Server, error) {
	lis, err := net.Listen(s.conf.Network, s.conf.Addr)
	if err != nil {
		return nil, err
	}
	reflection.Register(s.server)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return s, nil
}

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each.
// Serve will return a non-nil error unless Stop or GracefulStop is called.
func (s *Server) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

// Shutdown stops the server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished or the context deadline is reached.
func (s *Server) Shutdown(ctx context.Context) (err error) {
	ch := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(ch)
	}()
	select {
	case <-ctx.Done():
		s.server.Stop()
		err = ctx.Err()
	case <-ch:
	}
	return
}
