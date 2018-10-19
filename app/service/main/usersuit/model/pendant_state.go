package model

// const .
const (
	// pendant status
	PendantStatusON  = 1
	PendantStatusOFF = 0

	// group status
	GroupStatusON  = 1
	GroupStatusOFF = 0

	// packpage status
	InvalidPendantPKG = int32(0)
	ValidPendantPKG   = int32(1)
	EquipPendantPKG   = int32(2)

	// pendant equip
	PendantEquipOFF = int8(1)
	PendantEquipON  = int8(2)
)
