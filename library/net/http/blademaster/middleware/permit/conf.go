package permit

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/conf"
	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"
)

var (
	_defaultDashboardCaller = "manager-go"
	_defaultCookieName      = "mng-go"
	_defaultPermitURL       = "http://manager.hhhhhh.co/x/admin/manager/permission"
	_defaultVerifyURL       = "http://dashboard-mng.hhhhhh.co/api/session/verify"
	_defaultSession         = &SessionConfig{
		SessionIDLength: 32,
		CookieLifeTime:  2592000,
		Domain:          ".hhhhhh.co",
	}
	_defaultDsHTTPClient = &bm.ClientConfig{
		App: &conf.App{
			Key:    "manager-go",
			Secret: "949bbb2dd3178252638c2407578bc7ad",
		},
		Dial:    xtime.Duration(time.Millisecond * 100),
		Timeout: xtime.Duration(time.Millisecond * 300),
	}
	_defaultMaHTTPClient = &bm.ClientConfig{
		App: &conf.App{
			Key:    "f6433799dbd88751",
			Secret: "36f8ddb1806207fe07013ab6a77a3935",
		},
		Dial:    xtime.Duration(time.Millisecond * 100),
		Timeout: xtime.Duration(time.Millisecond * 300),
	}
)

// Config identify config.
type Config struct {
	DsHTTPClient    *bm.ClientConfig // dashboard client config. appkey can not reuse.
	MaHTTPClient    *bm.ClientConfig // manager-admin client config
	Session         *SessionConfig
	ManagerHost     string
	DashboardHost   string
	DashboardCaller string
}

// Config2 .
type Config2 struct {
	DashboardCaller string
	CookieName      string
	Memcache        *memcache.Config
}
