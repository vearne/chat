package consts

const (
	AccountStatusCreated = iota
	AccountStatusInUse
	AccountStatusDestroyed
)
const (
	SessionStatusCreated = iota
	SessionStatusInUse
	SessionStatusDestroyed
)

const (
	OutBoxStatusNormal = iota
	OutBoxStatusDeleted
)

const (
	InBoxStatusCreated = iota
	InBoxStatusDelivered
)

const (
	SystemSender = 0
)

const (
	CmdCreateAccount = "CRT_ACCOUNT"
	CmdMatch         = "MATCH"
	CmdDialogue      = "DIALOGUE"
	CmdPushDialogue  = "PUSH_DIALOGUE"
	CmdPushSignal    = "PUSH_SIGNAL"
	CmdPing          = "PING"
	CmdViewedAck     = "VIEWED_ACK"
	CmdPushViewedAck = "PUSH_VIEWED_ACK"
	CmdReConnect     = "RECONNECT"
)
