package model

// ArgAllowanceList req.
type ArgAllowanceList struct {
	State int8 `form:"state"`
}

// ArgCouponPage req .
type ArgCouponPage struct {
	State int8 `form:"state"`
	Pn    int  `form:"pn"`
	Ps    int  `form:"ps"`
}
