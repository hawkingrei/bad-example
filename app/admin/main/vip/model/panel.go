package model

import "go-common/library/time"

// VipPriceConfigPlat vip价格面版平台 1. 其他平台 2.IOS平台 3.IOS的HD平台
type VipPriceConfigPlat int8

const (
	// PlatVipPriceConfigOther 其他平台
	PlatVipPriceConfigOther VipPriceConfigPlat = iota + 1
	// PlatVipPriceConfigIOS IOS平台
	PlatVipPriceConfigIOS
	// PlatVipPriceConfigIOSHD IOS的HD平台
	PlatVipPriceConfigIOSHD
	// PlatVipPriceConfigFriendsGift 好友赠送
	PlatVipPriceConfigFriendsGift
)

// VipPriceConfigStatus vip价格面版配置状态 0. 有效 1. 失效 2.待生效
type VipPriceConfigStatus int8

const (
	// VipPriceConfigStatusON 有效
	VipPriceConfigStatusON VipPriceConfigStatus = iota
	// VipPriceConfigStatusOFF 失效
	VipPriceConfigStatusOFF
	// VipPriceConfigStatusFuture 待生效
	VipPriceConfigStatusFuture
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

// const .
const (
	DefualtZeroTimeFromDB = 0
	TimeFormatDay         = "2006-01-02 15:04:05"
	DefulatTimeFromDB     = "1970-01-01 08:00:00"
)

// VipPriceConfig price config.
type VipPriceConfig struct {
	ID       int64                `json:"id"`
	Plat     VipPriceConfigPlat   `json:"platform"`
	PdName   string               `json:"product_name"`
	PdID     string               `json:"product_id"`
	SuitType int8                 `json:"suit_type"`
	Month    int16                `json:"month"`
	SubType  int8                 `json:"sub_type"`
	OPrice   float64              `json:"original_price"`
	NPrice   float64              `json:"now_price"`
	Selected int8                 `json:"selected"`
	Remark   string               `json:"remark"`
	Status   VipPriceConfigStatus `json:"status"`
	Operator string               `json:"operator"`
	OpID     int64                `json:"oper_id"`
	CTime    time.Time            `json:"ctime"`
	MTime    time.Time            `json:"mtime"`
}

// VipDPriceConfig price discount config.
type VipDPriceConfig struct {
	DisID    int64     `json:"discount_id"`
	ID       int64     `json:"vpc_id"`
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

// ArgAddOrUpVipPrice .
type ArgAddOrUpVipPrice struct {
	ID       int64              `form:"id"`
	Plat     VipPriceConfigPlat `form:"platform" validate:"required"`
	PdName   string             `form:"product_name" validate:"required"`
	PdID     string             `form:"product_id"`
	Month    int16              `form:"month" validate:"required"`
	SubType  int8               `form:"sub_type"`
	SuitType int8               `form:"suit_type"`
	OPrice   float64            `form:"original_price" validate:"required"`
	Remark   string             `form:"remark"`
	Operator string             `form:"operator" validate:"required"`
	OpID     int64              `form:"oper_id" validate:"required"`
}

// ArgAddOrUpVipDPrice .
type ArgAddOrUpVipDPrice struct {
	DisID    int64     `form:"discount_id"`
	ID       int64     `form:"vpc_id" validate:"required"`
	PdID     string    `form:"product_id"`
	DPrice   float64   `form:"discount_price"`
	STime    time.Time `form:"stime" validate:"required"`
	ETime    time.Time `form:"etime"`
	Remark   string    `form:"remark"`
	Operator string    `form:"operator" validate:"required"`
	OpID     int64     `form:"oper_id" validate:"required"`
}

// CheckProductID .
func (vpc *VipPriceConfig) CheckProductID(arg *ArgAddOrUpVipDPrice) bool {
	return (vpc.Plat == PlatVipPriceConfigIOS || vpc.Plat == PlatVipPriceConfigIOSHD) && arg.PdID == ""
}

// ExistPlat .
func (aavp *ArgAddOrUpVipPrice) ExistPlat() bool {
	return aavp.Plat == PlatVipPriceConfigOther || aavp.Plat == PlatVipPriceConfigIOS || aavp.Plat == PlatVipPriceConfigIOSHD || aavp.Plat == PlatVipPriceConfigFriendsGift
}

// CheckProductID .
func (aavp *ArgAddOrUpVipPrice) CheckProductID() bool {
	return (aavp.Plat == PlatVipPriceConfigIOS || aavp.Plat == PlatVipPriceConfigIOSHD) && aavp.PdID == ""
}

// ArgVipPriceID .
type ArgVipPriceID struct {
	ID int64 `form:"id" validate:"required"`
}

// ArgVipDPriceID .
type ArgVipDPriceID struct {
	DisID int64 `form:"discount_id" validate:"required"`
}

// ArgVipPrice .
type ArgVipPrice struct {
	Plat     VipPriceConfigPlat `form:"platform" default:"-1"`
	Month    int16              `form:"month" default:"-1"`
	SubType  int8               `form:"sub_type" default:"-1"`
	SuitType int8               `form:"suit_type" default:"-1"`
}
