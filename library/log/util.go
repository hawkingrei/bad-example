package log

import (
	"context"
	"fmt"
	"runtime"
	"strconv"

	"go-common/library/conf/env"
	"go-common/library/net/metadata"
	"go-common/library/net/trace"
)

func addExtraField(ctx context.Context, fields map[string]interface{}) {
	if t, ok := trace.FromContext(ctx); ok {
		if s, ok := t.(fmt.Stringer); ok {
			fields[_tid] = s.String()
		} else {
			fields[_tid] = fmt.Sprintf("%s", t)
		}
	}
	if caller := metadata.String(ctx, metadata.Caller); caller != "" {
		fields[_caller] = caller
	}
	fields[_deplyEnv] = env.DeployEnv
	fields[_zone] = env.Zone
	fields[_appID] = c.Family
	fields[_instanceID] = c.Host
	if metadata.Bool(ctx, metadata.Mirror) {
		fields[_mirror] = true
	}
}

// funcName get func name.
func funcName() (name string) {
	if pc, _, lineNo, ok := runtime.Caller(5); ok {
		name = runtime.FuncForPC(pc).Name() + ":" + strconv.FormatInt(int64(lineNo), 10)
	}
	return
}
