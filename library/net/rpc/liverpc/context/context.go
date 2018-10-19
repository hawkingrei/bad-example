package context

import (
	"context"
	"time"

	"go-common/library/net/rpc/liverpc"
)

// WithHeader returns new context with header
func WithHeader(ctx context.Context, header *liverpc.Header) (ret context.Context) {
	ret = context.WithValue(ctx, liverpc.KeyHeader, header)
	return
}

// WithTimeout set timeout to rpc request
// Notice this is nothing related to to built-in context.WithTimeout
func WithTimeout(ctx context.Context, time time.Duration) (ret context.Context) {
	ret = context.WithValue(ctx, liverpc.KeyTimeout, time)
	return
}

// AddHeaderToBm add header to blademaster context
// returns the created header
//
// Deprecated
// the client will create header automatically
// 	so no just remove the call if you use it
func AddHeaderToBm(ctx context.Context) (header *liverpc.Header) {
	return
}

// HeaderFromContext returns the header in ctx, if not exist, return nil
func HeaderFromContext(ctx context.Context) (hdr *liverpc.Header) {
	hdr, _ = ctx.Value(liverpc.KeyHeader).(*liverpc.Header)
	return
}

// TimeoutFromContext returns the time in ctx, if not exist, return 0
func TimeoutFromContext(ctx context.Context) time.Duration {
	t, _ := ctx.Value(liverpc.KeyTimeout).(time.Duration)
	return t
}
