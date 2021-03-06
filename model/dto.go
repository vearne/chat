package model

type CommonCmd struct {
	Cmd string `json:"cmd"`
}

type CmdCreateAccountReq struct {
	Cmd      string `json:"cmd"`
	NickName string `json:"nickName"`
}

type CmdCreateAccountResp struct {
	Cmd       string `json:"cmd"`
	NickName  string `json:"nickName"`
	AccountId uint64 `json:"accountId"`
}

type CmdMatchReq struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
}

type CmdMatchResp struct {
	Cmd         string `json:"cmd"`
	PartnerId   uint64 `json:"partnerId,omitempty"`
	PartnerName string `json:"partnerName,omitempty"`
	SessionId   uint64 `json:"sessionId,omitempty"`
	Code        int32  `json:"code"`
}

type CmdDialogueReq struct {
	Cmd       string `json:"cmd"`
	RequestId string `json:"requestId"`
	SenderId  uint64 `json:"senderId"`
	SessionId uint64 `json:"sessionId"`
	Content   string `json:"content"`
}

type CmdDialogueResp struct {
	Cmd       string `json:"cmd"`
	RequestId string `json:"requestId"`
	MsgId     uint64 `json:"msgId"`
	Code      int32  `json:"code"`
}

type CmdPushDialogueReq struct {
	Cmd       string `json:"cmd"`
	MsgId     uint64 `json:"msgId"`
	SenderId  uint64 `json:"senderId"`
	SessionId uint64 `json:"sessionId"`
	Content   string `json:"content"`
}

type CmdPushDialogueResp struct {
	Cmd  string `json:"cmd"`
	Code int32  `json:"code"`
}

type CmdPushSignalReq struct {
	Cmd        string      `json:"cmd"`
	SenderId   uint64      `json:"senderId"`
	SessionId  uint64      `json:"sessionId"`
	ReceiverId uint64      `json:"receiverId"`
	SignalType string      `json:"signalType"`
	Data       interface{} `json:"data"`
}

type CmdPushSignalResp struct {
	Cmd  string `json:"cmd"`
	Code int32  `json:"code"`
}

// 由broker发出
/*
	{
	"cmd": "PING",
	"accountId": 12000
	}
*/
type CmdPingReq struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
}

// 由Client发出
/*
	{
		"cmd": "PONG"
		"accountId": 12000
	}
*/
type CmdPingResp struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
}

type CmdViewedAckReq struct {
	Cmd       string `json:"cmd"`
	SessionId uint64 `json:"sessionId"`
	AccountId uint64 `json:"accountId"`
	MsgId     uint64 `json:"MsgId"`
}

type CmdViewedAckResp struct {
	Cmd  string `json:"cmd"`
	Code int32  `json:"code"`
}

type CmdPushViewedAckReq struct {
	Cmd       string `json:"cmd"`
	SessionId uint64 `json:"sessionId"`
	AccountId uint64 `json:"accountId"`
	MsgId     uint64 `json:"msgId"`
}
