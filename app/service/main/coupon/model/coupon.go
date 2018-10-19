package model

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/time"
)

// CouponChangeLog coupon change log.
type CouponChangeLog struct {
	ID          int64     `json:"-"`
	CouponToken string    `json:"coupon_token"`
	Mid         int64     `json:"mid"`
	State       int8      `json:"state"`
	Ctime       time.Time `json:"ctime"`
	Mtime       time.Time `json:"mtime"`
}

// CouponPageResp coupon page.
type CouponPageResp struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Time  int64  `json:"time"`
	RefID int64  `json:"ref_id"`
	Tips  string `json:"tips"`
	Count int64  `json:"count"`
}

// CouponOrder coupon order info.
type CouponOrder struct {
	ID           int64     `json:"id"`
	OrderNo      string    `json:"order_no"`
	Mid          int64     `json:"mid"`
	Count        int64     `json:"count"`
	State        int8      `json:"state"`
	CouponType   int8      `json:"coupon_type"`
	ThirdTradeNo string    `json:"third_trade_no"`
	Remark       string    `json:"remark"`
	Tips         string    `json:"tips"`
	UseVer       int64     `json:"use_ver"`
	Ver          int64     `json:"ver"`
	Ctime        time.Time `json:"ctime"`
	Mtime        time.Time `json:"mtime"`
}

// CouponOrderLog coupon order log.
type CouponOrderLog struct {
	ID      int64     `json:"id"`
	OrderNo string    `json:"order_no"`
	Mid     int64     `json:"mid"`
	State   int8      `json:"state"`
	Ctime   time.Time `json:"ctime"`
	Mtime   time.Time `json:"mtime"`
}

// CouponBalanceChangeLog coupon balance change log.
type CouponBalanceChangeLog struct {
	ID            int64     `json:"id"`
	OrderNo       string    `json:"order_no"`
	Mid           int64     `json:"mid"`
	BatchToken    string    `json:"batch_token"`
	Balance       int64     `json:"balance"`
	ChangeBalance int64     `json:"change_balance"`
	ChangeType    int8      `json:"change_type"`
	Ctime         time.Time `json:"ctime"`
	Mtime         time.Time `json:"mtime"`
}

// CouponCartoonPageResp coupon cartoon page.
type CouponCartoonPageResp struct {
	Count       int64             `json:"count"`
	CouponCount int64             `json:"coupon_count"`
	List        []*CouponPageResp `json:"list"`
}

// CouponBatchInfo coupon batch info.
type CouponBatchInfo struct {
	ID            int64     `json:"id"`
	AppID         int64     `json:"app_id"`
	Name          string    `json:"name"`
	BatchToken    string    `json:"batch_token"`
	MaxCount      int64     `json:"max_count"`
	CurrentCount  int64     `json:"current_count"`
	LimitCount    int64     `json:"limit_count"`
	StartTime     int64     `json:"start_time"`
	ExpireTime    int64     `json:"expire_time"`
	ExpireDay     int64     `json:"expire_day"`
	Ver           int64     `json:"ver"`
	Ctime         time.Time `json:"ctime"`
	Mtime         time.Time `json:"mtime"`
	FullAmount    float64   `json:"full_amount"`
	Amount        float64   `json:"amount"`
	State         int8      `json:"state"`
	CouponType    int8      `json:"coupon_type"`
	PlatformLimit string    `json:"platform_limit"`
}

// CouponAllowancePanelInfo allowance coupon panel info.
type CouponAllowancePanelInfo struct {
	CouponToken         string  `json:"coupon_token"`
	Amount              float64 `json:"coupon_amount"`
	State               int32   `json:"state"`
	FullLimitExplain    string  `json:"full_limit_explain"`
	ScopeExplain        string  `json:"scope_explain"`
	FullAmount          float64 `json:"full_amount"`
	CouponDiscountPrice float64 `json:"coupon_discount_price"`
	StartTime           int64   `json:"start_time"`
	ExpireTime          int64   `json:"expire_time"`
	Selected            int8    `json:"selected"`
	DisablesExplains    string  `json:"disables_explains"`
	OrderNO             string  `json:"order_no"`
	Name                string  `json:"name"`
	Usable              int8    `json:"usable"`
}

// CouponAllowanceChangeLog coupon allowance change log.
type CouponAllowanceChangeLog struct {
	ID          int64     `json:"-"`
	CouponToken string    `json:"coupon_token"`
	OrderNO     string    `json:"order_no"`
	Mid         int64     `json:"mid"`
	State       int8      `json:"state"`
	ChangeType  int8      `json:"change_type"`
	Ctime       time.Time `json:"ctime"`
	Mtime       time.Time `json:"mtime"`
}

//CouponReceiveLog receive log.
type CouponReceiveLog struct {
	ID          int64  `json:"id"`
	Appkey      string `json:"appkey"`
	OrderNo     string `json:"order_no"`
	Mid         int64  `json:"mid"`
	CouponToken string `json:"coupon_token"`
	CouponType  int8   `json:"coupon_type"`
}

//CouponAllowancePanelResp def.
type CouponAllowancePanelResp struct {
	Usables  []*CouponAllowancePanelInfo `json:"usables"`
	Disables []*CouponAllowancePanelInfo `json:"disables"`
	Using    []*CouponAllowancePanelInfo `json:"using"`
}

// ScopeExplainFmt get scope explain fmt.
func (c *CouponAllowancePanelInfo) ScopeExplainFmt(pstr string) {
	var (
		ps    []string
		plats string
		err   error
		p     int
	)
	if len(pstr) == 0 {
		c.ScopeExplain = ScopeNoLimit
		return
	}
	ps = strings.Split(pstr, ",")
	for _, v := range ps {
		if p, err = strconv.Atoi(v); err != nil {
			continue
		}
		plats += PlatformByCode[p] + ","
	}
	if len(plats) > 0 {
		plats = plats[:len(plats)-1]
		plats = fmt.Sprintf(ScopeFmt, plats)
	}
	c.ScopeExplain = plats
}
