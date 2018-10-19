package permit

import (
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
)

const (
	_verifyURI     = "/api/session/verify"
	_permissionURI = "/x/admin/manager/permission"
	_sessIDKey     = "_AJSESSIONID"
	_sessUIDKey    = "uid"
	_sessUnKey     = "username"
	// CtxPermissions will be set into ctx
	CtxPermissions = "permissions"
)

// permissions .
type permissions struct {
	UID   int64    `json:"uid"`
	Perms []string `json:"perms"`
}

// Permit is an auth middleware.
type Permit struct {
	verifyURI       string
	permissionURI   string
	dashboardCaller string
	dsClient        *bm.Client // dashboard client
	maClient        *bm.Client // manager-admin client

	sm *SessionManager // user Session
}

//Verify only export Verify function because of less configure
type Verify interface {
	Verify() bm.HandlerFunc
}

// New new an auth service.
func New(c *Config) *Permit {
	a := &Permit{
		dashboardCaller: c.DashboardCaller,
		verifyURI:       c.DashboardHost + _verifyURI,
		permissionURI:   c.ManagerHost + _permissionURI,
		dsClient:        bm.NewClient(c.DsHTTPClient),
		maClient:        bm.NewClient(c.MaHTTPClient),
		sm:              newSessionManager(c.Session),
	}
	return a
}

// New2 .
func New2(c *Config2) *Permit {
	if c.DashboardCaller == "" {
		c.DashboardCaller = _defaultDashboardCaller
	}
	if c.CookieName == "" {
		c.CookieName = _defaultCookieName
	}
	_defaultSession.CookieName = c.CookieName
	_defaultSession.Memcache = c.Memcache
	return &Permit{
		dashboardCaller: c.DashboardCaller,
		verifyURI:       _defaultVerifyURL,
		permissionURI:   _defaultPermitURL,
		dsClient:        bm.NewClient(_defaultDsHTTPClient),
		maClient:        bm.NewClient(_defaultMaHTTPClient),
		sm:              newSessionManager(_defaultSession),
	}
}

// NewVerify new a verify service.
func NewVerify(c *Config) Verify {
	a := &Permit{
		verifyURI:       c.DashboardHost + _verifyURI,
		dsClient:        bm.NewClient(c.DsHTTPClient),
		dashboardCaller: c.DashboardCaller,
		sm:              newSessionManager(c.Session),
	}
	return a
}

// NewVerify2 .
func NewVerify2(c *Config2) Verify {
	if c.DashboardCaller == "" {
		c.DashboardCaller = _defaultDashboardCaller
	}
	if c.CookieName == "" {
		c.CookieName = _defaultCookieName
	}
	_defaultSession.CookieName = c.CookieName
	_defaultSession.Memcache = c.Memcache
	return &Permit{
		dashboardCaller: c.DashboardCaller,
		verifyURI:       _defaultVerifyURL,
		permissionURI:   _defaultPermitURL,
		dsClient:        bm.NewClient(_defaultDsHTTPClient),
		sm:              newSessionManager(_defaultSession),
	}
}

// Verify return bm HandlerFunc which check whether the user has logged in.
func (a *Permit) Verify() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		si, err := a.login(ctx)
		if err != nil {
			ctx.JSON(nil, ecode.Unauthorized)
			ctx.Abort()
			return
		}
		a.sm.SessionRelease(ctx, si)
	}
}

// Permit return bm HandlerFunc which check whether the user has logged in and has the access permission of the location.
// If `permit` is empty,it will allow any logged in request.
func (a *Permit) Permit(permit string) bm.HandlerFunc {
	return func(ctx *bm.Context) {
		var (
			si    *Session
			uid   int64
			perms []string
			err   error
		)
		si, err = a.login(ctx)
		if err != nil {
			ctx.JSON(nil, ecode.Unauthorized)
			ctx.Abort()
			return
		}
		defer a.sm.SessionRelease(ctx, si)
		uid, perms, err = a.permissions(ctx, si.Get(_sessUnKey).(string))
		if err == nil {
			si.Set(_sessUIDKey, uid)
			ctx.Set(_sessUIDKey, uid)
		}
		if len(perms) > 0 {
			ctx.Set(CtxPermissions, perms)
		}
		if !a.permit(permit, perms) {
			ctx.JSON(nil, ecode.AccessDenied)
			ctx.Abort()
			return
		}
	}
}

// login check whether the user has logged in.
func (a *Permit) login(ctx *bm.Context) (si *Session, err error) {
	si = a.sm.SessionStart(ctx)
	if si.Get(_sessUnKey) == nil {
		var username string
		if username, err = a.verify(ctx); err != nil {
			return
		}
		si.Set(_sessUnKey, username)
	}
	ctx.Set(_sessUnKey, si.Get(_sessUnKey))
	return
}

func (a *Permit) verify(ctx *bm.Context) (username string, err error) {
	var (
		sid string
		r   = ctx.Request
	)
	session, err := r.Cookie(_sessIDKey)
	if err == nil {
		sid = session.Value
	}
	if sid == "" {
		err = ecode.Unauthorized
		return
	}
	username, err = a.verifyDashboard(ctx, sid)
	return
}

// permit check whether user has the access permission of the location.
func (a *Permit) permit(permit string, permissions []string) bool {
	if permit == "" {
		return true
	}
	for _, p := range permissions {
		if p == permit {
			// access the permit
			return true
		}
	}
	return false
}

// verifyDashboard check whether the user is valid from Dashboard.
func (a *Permit) verifyDashboard(ctx *bm.Context, sid string) (username string, err error) {
	params := url.Values{}
	params.Set("session_id", sid)
	params.Set("encrypt", "md5")
	params.Set("caller", a.dashboardCaller)
	var res struct {
		Code     int    `json:"code"`
		UserName string `json:"username"`
	}
	if err = a.dsClient.Get(ctx, a.verifyURI, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		log.Error("dashboard get verify Session url(%s) error(%v)", a.verifyURI+"?"+params.Encode(), err)
		return
	}
	if ecode.Int(res.Code) != ecode.OK {
		log.Error("dashboard get verify Session url(%s) error(%v)", a.verifyURI+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
		return
	}
	username = res.UserName
	return
}

// permissions get user's permisssions from manager-admin.
func (a *Permit) permissions(ctx *bm.Context, username string) (uid int64, perms []string, err error) {
	params := url.Values{}
	params.Set(_sessUnKey, username)
	var res struct {
		Code int         `json:"code"`
		Data permissions `json:"data"`
	}
	if err = a.maClient.Get(ctx, a.permissionURI, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		log.Error("dashboard get permissions url(%s) error(%v)", a.permissionURI+"?"+params.Encode(), err)
		return
	}
	if ecode.Int(res.Code) != ecode.OK {
		log.Error("dashboard get permissions url(%s) error(%v)", a.permissionURI+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
		return
	}
	perms = res.Data.Perms
	uid = res.Data.UID
	return
}
