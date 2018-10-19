package ecode

// 直播　号段　　1000000 - 1999999
var (
	// wallet 1000000 - 1001999
	CoinNotEnough = New(1000000)
	PayFailed     = New(1000001)

	// live-test 100 2000 - 100 2999
	RoomNotFound = New(1002001)
)
