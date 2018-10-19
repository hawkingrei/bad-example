package blademaster

import (
	"fmt"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
)

// Logger is logger  middleware
func Logger() HandlerFunc {
	const noUser = "no_user"
	return func(c *Context) {
		now := time.Now()
		ip := metadata.String(c, metadata.RemoteIP)
		req := c.Request
		path := req.URL.Path
		params := req.Form

		c.Next()

		mid, _ := c.Get("mid")
		err := c.Error
		cerr := ecode.Cause(err)
		dt := time.Since(now)
		caller := metadata.String(c, metadata.Caller)
		if caller == "" {
			caller = noUser
		}

		stats.Incr(caller, path[1:], strconv.FormatInt(int64(cerr.Code()), 10))
		stats.Timing(caller, int64(dt/time.Millisecond), path[1:])

		lf := log.Infov
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
			lf = log.Errorv
			if cerr.Code() > 0 {
				lf = log.Warnv
			}
		}
		lf(c,
			log.KV("method", req.Method),
			log.KV("mid", mid),
			log.KV("ip", ip),
			log.KV("user", caller),
			log.KV("path", path),
			log.KV("params", params.Encode()),
			log.KV("ret", cerr.Code()),
			log.KV("msg", cerr.Message()),
			log.KV("stack", fmt.Sprintf("%+v", err)),
			log.KV("err", errmsg),
			log.KV("ts", dt.Seconds()),
		)
	}
}
