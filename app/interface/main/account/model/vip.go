package model

import (
	vipmol "go-common/app/service/main/vip/model"
	"go-common/library/time"
)

// vip tips.
const (
	PanelPosition int8 = iota + 1
	PgcPosition
)

// VIPInfo vip info.
type VIPInfo struct {
	Mid       int64  `json:"mid"`
	Type      int8   `json:"vipType"`
	DueDate   int64  `json:"vipDueDate"`
	DueMsec   int64  `json:"vipSurplusMsec"`
	DueRemark string `json:"dueRemark"`
	Status    int8   `json:"accessStatus"`
	VipStatus int8   `json:"vipStatus"`
}

// TipsReq tips request.
type TipsReq struct {
	Version  int64  `form:"build"`
	Platform string `form:"platform" validate:"required"`
	Position int8   `form:"position" default:"1"`
}

//CodeInfoReq code info request
type CodeInfoReq struct {
	Appkey    string    `form:"appkey" validate:"required"`
	Sign      string    `form:"sign"`
	Ts        time.Time `form:"ts"`
	StartTime time.Time `form:"start_time" validate:"required"`
	EndTime   time.Time `form:"end_time" validate:"required"`
	Cursor    int64     `form:"cursor"`
}

// VipPanelRes .
type VipPanelRes struct {
	Device    string `form:"device"`
	MobiApp   string `form:"mobi_app"`
	Platform  string `form:"platform" default:"pc"`
	SortTP    int8   `form:"sort_type"`
	PanelType string `form:"panel_type" default:"normal"`
	Month     int32  `form:"month"`
	SubType   int32  `form:"order_type"`
}

// ArgVipCoupon req.
type ArgVipCoupon struct {
	ID       int64  `form:"id" validate:"required,min=1,gte=1"`
	Platform string `form:"platform" default:"pc"`
}

// ArgVipCancelPay req.
type ArgVipCancelPay struct {
	CouponToken string `form:"coupon_token" validate:"required"`
}

// GetVipPanelPlat .
func (vr *VipPanelRes) GetVipPanelPlat() int8 {
	switch {
	case vr.PanelType == vipmol.PanelTypeFriend:
		return vipmol.PlatVipPriceConfigFriendsGift
	case vr.MobiApp == "iphone":
		return vipmol.PlatVipPriceConfigIOS
	case vr.MobiApp == "ipad" && vr.Device == "pad":
		return vipmol.PlatVipPriceConfigIPADHD
	default:
		return vipmol.PlatVipPriceConfigOther
	}
}

// coupon cancel explain
const (
	CouponCancelExplain = "解锁成功,请重新选择劵信息"
)

// const for vip
const (
	MobiAppIphone = iota + 1
	MobiAppIpad
	MobiAppPC
	MobiAppANDROID
)

//MobiAppByName .
var MobiAppByName = map[string]int{
	"iphone":  MobiAppIphone,
	"ipad":    MobiAppIpad,
	"pc":      MobiAppPC,
	"android": MobiAppANDROID,
}

// MobiAppPlat .
func MobiAppPlat(mobiApp string) (p int) {
	p = MobiAppByName[mobiApp]
	if p == 0 {
		// def pc.
		p = MobiAppPC
	}
	return
}

// ArgVipPanel arg panel.
type ArgVipPanel struct {
	Device    string `form:"device"`
	Build     int64  `form:"build"`
	MobiApp   string `form:"mobi_app"`
	Platform  string `form:"platform" default:"pc"`
	SortTP    int8   `form:"sort_type"`
	PanelType string `form:"panel_type" default:"normal"`
	Mid       int64
	IP        string
}

// VipPanelResp vip panel resp.
type VipPanelResp struct {
	Vps        []*vipmol.VipPanelInfo          `json:"price_list"`
	CodeSwitch int8                            `json:"code_switch"`
	GiveSwitch int8                            `json:"give_switch"`
	Privileges map[int8]*vipmol.PrivilegesResp `json:"privileges,omitempty"`
	TipInfo    *vipmol.TipsResp                `json:"tip_info,omitempty"`
	UserInfo   *vipmol.VipPanelExplain         `json:"user_info,omitempty"`
}

// ManagerResp manager resp.
type ManagerResp struct {
	JointlyInfo []*vipmol.JointlyResp `json:"jointly_info"`
}
