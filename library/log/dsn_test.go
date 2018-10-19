package log

import (
	"flag"
	"os"
	"testing"
	"time"

	xtime "go-common/library/time"

	"github.com/stretchr/testify/assert"
)

func TestUnixParseDSN(t *testing.T) {
	dsn := "unix:///data/log/lancer.sock?chan=1024&timeout=5s"
	flag.Parse()
	flag.Set("log.stdout", "true")
	as := parseDSN(dsn)
	assert.Equal(t, "unix", as.Proto)
	assert.Equal(t, "/data/log/lancer.sock", as.Addr)
	assert.Equal(t, 1024, as.Chan)
	assert.Equal(t, xtime.Duration(5*time.Second), as.Timeout)
}

func TestUDPParseDSN(t *testing.T) {
	dsn := "udp://localhost:8080?chan=1024"
	flag.Parse()
	flag.Set("log.stdout", "true")
	as := parseDSN(dsn)
	assert.Equal(t, "udp", as.Proto)
	assert.Equal(t, "localhost:8080", as.Addr)
}

func TestUnixFlagFormatDSN(t *testing.T) {
	flag.Parse()
	flag.Set("log.stdout", "true")
	flag.Set("log.dir", "/data/log/test")
	flag.Set("log.agent", "unixgram:///var/run/lancer/collector.sock?timeout=100ms&chan=1024")
	flag.Set("log.v", "100")
	flag.Set("log.module", "xxx=22, 333=444")
	flag.Set("log.filter", "key1, key2")

	Init(nil)
	assert.Equal(t, true, c.Stdout)
	assert.Equal(t, "/data/log/test", c.Dir)
	assert.NotNil(t, c.Agent)
	assert.Equal(t, "unixgram", c.Agent.Proto)
	assert.Equal(t, "/var/run/lancer/collector.sock", c.Agent.Addr)
	assert.Equal(t, 1024, c.Agent.Chan)
	assert.Equal(t, xtime.Duration(100*time.Millisecond), c.Agent.Timeout)
	assert.Equal(t, int32(100), c.V)
	assert.Equal(t, 2, len(c.Module))
	assert.Equal(t, 2, len(c.Filter))
}

func TestParseDSNSuccess(t *testing.T) {
	var dsn = "udp://172.16.0.204:514/?timeout=100ms&chan=1024"
	ac := parseDSN(dsn)
	assert.NotNil(t, ac)
	assert.Equal(t, "udp", ac.Proto)
	assert.Equal(t, "172.16.0.204:514", ac.Addr)
	assert.Equal(t, xtime.Duration(100*time.Millisecond), ac.Timeout)
	assert.Equal(t, 1024, ac.Chan)
}

func TestParseDSNPanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	var dsn = "udp:172.16.0.204:514/?timeout=100&chan=1024ms"
	ac := parseDSN(dsn)
	assert.NotNil(t, ac)
}

func TestUDPFlagFormatDSN(t *testing.T) {
	flag.Parse()
	flag.Set("log.stdout", "true")
	flag.Set("log.dir", "/data/log/test")
	flag.Set("log.agent", "udp://172.16.0.204:514/?timeout=100ms&chan=1024")
	flag.Set("log.v", "100")
	flag.Set("log.module", "xxx=22, 333=444")
	flag.Set("log.filter", "key1, key2")

	Init(nil)
	assert.Equal(t, true, c.Stdout)
	assert.Equal(t, "/data/log/test", c.Dir)
	assert.NotNil(t, c.Agent)
	assert.Equal(t, "udp", c.Agent.Proto)
	assert.Equal(t, "172.16.0.204:514", c.Agent.Addr)
	assert.Equal(t, 1024, c.Agent.Chan)
	assert.Equal(t, xtime.Duration(100*time.Millisecond), c.Agent.Timeout)
	assert.Equal(t, int32(100), c.V)
	assert.Equal(t, 2, len(c.Module))
	assert.Equal(t, 2, len(c.Filter))
}

func TestEnvFormatDSN(t *testing.T) {

	t.Run("module: env set, flag no", func(t *testing.T) {
		var (
			tenv = "LOG_MODULE"
			def  = "x1=1,z2=2"
			val  = &_module
		)
		err := os.Setenv(tenv, def)
		assert.Nil(t, err)
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		addFlag(fs)
		assert.Equal(t, 2, len(*val))
	})

	t.Run("module: env no, flag set", func(t *testing.T) {
		var (
			tflag = "log.module"
			def   = "x1=1,z2=2,z3=2"
			val   = &_module
		)
		flag.Parse()
		flag.Set(tflag, def)
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		addFlag(fs)
		assert.Equal(t, 3, len(*val))
	})
}

func TestDSNLog(t *testing.T) {
	flag.Parse()
	flag.Set("deploy.env", "uat")
	flag.Set("log.stdout", "true")
	flag.Set("log.dir", "/data/log/test")
	flag.Set("log.agent", "udp://172.16.0.204:514?timeout=100ms&chan=1024")
	flag.Set("log.v", "100")
	flag.Set("log.module", "xxx=22, 333=444")
	flag.Set("log.filter", "key1, key2")

	Init(nil)
	Info("dsn log.")
}
