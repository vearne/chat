package consts

const(
	AccountStatusCreated = iota
	AccountStatusInUse
	AccountStatusDestroyed
)
const(
	SessionStatusCreated = iota
	SessionStatusInUse
	SessionStatusDestroyed
)

const(
	OutBoxStatusNormal = iota
	OutBoxStatusDeleted
)

const(
	InBoxStatusCreated = iota
	InBoxStatusDelivered
)

const (
	frameTypeResponse int32 = 0
	frameTypeError    int32 = 1
	frameTypeMessage  int32 = 2
)

const(
	SystemSender = 0
)

const(
	CmdCreateAccount = "CRT_ACCOUNT"
	CmdMatch = "MATCH"
)

