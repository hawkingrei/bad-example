package model

import (
	"fmt"
	"math"
	"strconv"

	col "go-common/app/service/main/coupon/model"
	"go-common/library/time"
)

// vip_price_config suit_type
const (
	AllUser int8 = iota
	OldVIP
	NewVIP
	OldSubVIP
	NewSubVIP
	OldPackVIP
	NewPackVIP
)

// order type
const (
	NoRenew int8 = iota
	OtherRenew
	IOSRenew
)

// order type by month for vip_user_discount_history table
const (
	OneMonthSub int8 = iota + 1
	ThreeMonthSub
	OneYearSub
)

// const month
const (
	OneMonth   = int8(1)
	ThreeMonth = int8(3)
	OneYear    = int8(12)
)

// const vip_price_config beforeSuitType
const (
	All int8 = iota
	VIP
	Sub
	Pack
)

// const panel month sort
const (
	PanelMonthDESC int8 = iota
	PanelMonthASC
)

// const PanelType
const (
	PanelTypeNormal = "normal"
	PanelTypeFriend = "friend"
)

const (
	// PlatVipPriceConfigOther 其他平台
	PlatVipPriceConfigOther int8 = iota + 1
	// PlatVipPriceConfigIOS IOS平台
	PlatVipPriceConfigIOS
	// PlatVipPriceConfigIPADHD ipad hd平台
	PlatVipPriceConfigIPADHD
	// PlatVipPriceConfigFriendsGift 好友赠送
	PlatVipPriceConfigFriendsGift
)

// const select
const (
	PanelNotSelected int32 = iota
	PanelSelected
)

// VipPriceConfig price config.
type VipPriceConfig struct {
	ID          int64     `json:"id"`
	Plat        int8      `json:"platform"`
	PdName      string    `json:"product_name"`
	PdID        string    `json:"product_id"`
	SuitType    int8      `json:"suit_type"`
	TopSuitType int8      `json:"-"`
	Month       int16     `json:"month"`
	SubType     int8      `json:"sub_type"`
	OPrice      float64   `json:"original_price"`
	DPrice      float64   `json:"discount_price"`
	Selected    int8      `json:"selected"`
	Remark      string    `json:"remark"`
	Status      int8      `json:"status"`
	Forever     bool      `json:"-"`
	Operator    string    `json:"operator"`
	OpID        int64     `json:"oper_id"`
	CTime       time.Time `json:"ctime"`
	MTime       time.Time `json:"mtime"`
}

// VipPirceResp vip pirce resp.
type VipPirceResp struct {
	Vps         []*VipPanelInfo               `json:"price_list"`
	CouponInfo  *col.CouponAllowancePanelInfo `json:"coupon_info"`
	CouponSwith int8                          `json:"coupon_switch"`
	CodeSwitch  int8                          `json:"code_switch"`
	GiveSwitch  int8                          `json:"give_switch"`
	ExistCoupon int8                          `json:"exist_coupon"`
	Privileges  *PrivilegesResp               `json:"privileges"`
}

// VipPirceResp5 vip pirce resp.
type VipPirceResp5 struct {
	Vps         []*VipPanelInfo               `json:"price_list"`
	CouponInfo  *col.CouponAllowancePanelInfo `json:"coupon_info"`
	CouponSwith int8                          `json:"coupon_switch"`
	CodeSwitch  int8                          `json:"code_switch"`
	GiveSwitch  int8                          `json:"give_switch"`
	Privileges  map[int8]*PrivilegesResp      `json:"privileges"`
}

// VipDPriceConfig price discount config.
type VipDPriceConfig struct {
	ID       int64     `json:"id"`
	PdID     string    `json:"product_id"`
	DPrice   float64   `json:"discount_price"`
	STime    time.Time `json:"stime"`
	ETime    time.Time `json:"etime"`
	Remark   string    `json:"remark"`
	Operator string    `json:"operator"`
	OpID     int64     `json:"oper_id"`
	CTime    time.Time `json:"ctime"`
	MTime    time.Time `json:"mtime"`
}

// DoTopSuitType .
func (vpc *VipPriceConfig) DoTopSuitType() {
	switch vpc.SuitType {
	case OldPackVIP, NewPackVIP:
		vpc.TopSuitType = Pack
	case OldSubVIP, NewSubVIP:
		vpc.TopSuitType = Sub
	case OldVIP, NewVIP:
		vpc.TopSuitType = VIP
	case AllUser:
		vpc.TopSuitType = All
	}
}

// DoCheckRealPrice ,
func (vpc *VipPriceConfig) DoCheckRealPrice(mvp map[int64]*VipDPriceConfig) {
	if vp, ok := mvp[vpc.ID]; ok {
		vpc.PdID = vp.PdID
		vpc.DPrice = vp.DPrice
		vpc.Remark = vp.Remark
	}
	if vpc.DPrice == 0 {
		vpc.DPrice = vpc.OPrice
	}
}

// DoSubMonthKey .
func (vpc *VipPriceConfig) DoSubMonthKey() string {
	return fmt.Sprintf("%d%d", vpc.Month, vpc.SubType)
}

// FormatRate .
func (vpc *VipPriceConfig) FormatRate() string {
	if vpc.DPrice == 0 {
		return ""
	}
	if vpc.DPrice/vpc.OPrice == 1 {
		return ""
	}
	return strconv.FormatFloat(math.Floor((vpc.DPrice/vpc.OPrice)*100)/10, 'f', -1, 64) + "折"
}

// DoPayOrderTypeKey .
func (po *PayOrder) DoPayOrderTypeKey() string {
	if po.OrderType == IOSRenew {
		po.OrderType = OtherRenew
	}
	return fmt.Sprintf("%d%d", po.BuyMonths, po.OrderType)
}

// IsSub .
func (po *PayOrder) IsSub() bool {
	return po.OrderType == OtherRenew || po.OrderType == IOSRenew
}

// VipPirce vip pirce.
type VipPirce struct {
	Panel  *VipPanelInfo            `json:"pirce_info"`
	Coupon *col.CouponAllowanceInfo `json:"coupon_info"`
}

// GetVipPanelPlat .
func (vr *ArgPanel) GetVipPanelPlat() int8 {
	switch {
	case vr.PanelType == PanelTypeFriend:
		return PlatVipPriceConfigFriendsGift
	case vr.MobiApp == "iphone":
		return PlatVipPriceConfigIOS
	case vr.MobiApp == "ipad" && vr.Device == "pad":
		return PlatVipPriceConfigIPADHD
	default:
		return PlatVipPriceConfigOther
	}
}

// VipPanelExplain vip panel explain.
type VipPanelExplain struct {
	BackgroundURL string `json:"background_url"`
	Explain       string `json:"user_explain"`
}
