package model

import (
	"github.com/vearne/chat/consts"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/utils"
)

type CommonCmd struct {
	Cmd string `json:"cmd"`
}

type BaseRespCmd struct {
	Cmd  string `json:"cmd"`
	Code int32  `json:"code"`
}

type CmdCreateAccountReq struct {
	Cmd      string `json:"cmd"`
	NickName string `json:"nickName"`
}

func NewCmdCreateAccountReq() *CmdCreateAccountReq {
	var cmd CmdCreateAccountReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdCreateAccount)
	return &cmd
}

type CmdCreateAccountResp struct {
	BaseRespCmd
	NickName  string `json:"nickName"`
	AccountId uint64 `json:"accountId"`
	Token     string `json:"token"`
}

func NewCmdCreateAccountResp() *CmdCreateAccountResp {
	var cmd CmdCreateAccountResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdCreateAccount)
	return &cmd
}

type CmdMatchReq struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
}

func NewCmdMatchReq() *CmdMatchReq {
	var cmd CmdMatchReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdMatch)
	return &cmd
}

type CmdMatchResp struct {
	BaseRespCmd
	PartnerId   uint64 `json:"partnerId,omitempty"`
	PartnerName string `json:"partnerName,omitempty"`
	SessionId   uint64 `json:"sessionId,omitempty"`
}

func NewCmdMatchResp() *CmdMatchResp {
	var cmd CmdMatchResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdMatch)
	return &cmd
}

type CmdDialogueReq struct {
	Cmd       string `json:"cmd"`
	RequestId string `json:"requestId"`
	SenderId  uint64 `json:"senderId"`
	SessionId uint64 `json:"sessionId"`
	Content   string `json:"content"`
}

func NewCmdDialogueReq() *CmdDialogueReq {
	var cmd CmdDialogueReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdDialogue)
	return &cmd
}

type CmdDialogueResp struct {
	BaseRespCmd
	RequestId string `json:"requestId"`
	MsgId     uint64 `json:"msgId"`
}

func NewCmdDialogueResp() *CmdDialogueResp {
	var cmd CmdDialogueResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdDialogue)
	return &cmd
}

type CmdPushDialogueReq struct {
	Cmd       string `json:"cmd"`
	MsgId     uint64 `json:"msgId"`
	SenderId  uint64 `json:"senderId"`
	SessionId uint64 `json:"sessionId"`
	Content   string `json:"content"`
}

func NewCmdPushDialogueReq() *CmdPushDialogueReq {
	var cmd CmdPushDialogueReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdPushDialogue)
	return &cmd
}

type CmdPushDialogueResp struct {
	BaseRespCmd
}

func NewCmdPushDialogueResp() *CmdPushDialogueResp {
	var cmd CmdPushDialogueResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdPushDialogue)
	return &cmd
}

type CmdPushSignalReq struct {
	Cmd        string      `json:"cmd"`
	SenderId   uint64      `json:"senderId"`
	SessionId  uint64      `json:"sessionId"`
	ReceiverId uint64      `json:"receiverId"`
	SignalType string      `json:"signalType"`
	Data       interface{} `json:"data"`
}

func NewCmdPushSignalReq() *CmdPushSignalReq {
	var cmd CmdPushSignalReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdPushSignal)
	return &cmd
}

type CmdPushSignalResp struct {
	BaseRespCmd
}

func NewCmdPushSignalResp() *CmdPushSignalResp {
	var cmd CmdPushSignalResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdPushSignal)
	return &cmd
}

// 由broker发出
/*
	{
	"cmd": "PING_REQ",
	"accountId": 12000
	}
*/
type CmdPingReq struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
}

func NewCmdPingReq() *CmdPingReq {
	var cmd CmdPingReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdPing)
	return &cmd
}

// 由Client发出
/*
	{
		"cmd": "PING_RESP"
		"accountId": 12000
	}
*/
type CmdPingResp struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
}

func NewCmdPingResp() *CmdPingResp {
	var cmd CmdPingResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdPing)
	return &cmd
}

type CmdViewedAckReq struct {
	Cmd       string `json:"cmd"`
	SessionId uint64 `json:"sessionId"`
	AccountId uint64 `json:"accountId"`
	MsgId     uint64 `json:"MsgId"`
}

func NewCmdViewedAckReq() *CmdViewedAckReq {
	var cmd CmdViewedAckReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdViewedAck)
	return &cmd
}

type CmdViewedAckResp struct {
	BaseRespCmd
}

func NewCmdViewedAckResp() *CmdViewedAckResp {
	var cmd CmdViewedAckResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdViewedAck)
	return &cmd
}

type CmdPushViewedAckReq struct {
	Cmd       string `json:"cmd"`
	SessionId uint64 `json:"sessionId"`
	AccountId uint64 `json:"accountId"`
	MsgId     uint64 `json:"msgId"`
}

func NewCmdPushViewedAckReq() *CmdPushViewedAckReq {
	var cmd CmdPushViewedAckReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdPushViewedAck)
	return &cmd
}

type CmdReConnectReq struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
	Token     string `json:"token"`
}

func NewCmdReConnectReq() *CmdReConnectReq {
	var cmd CmdReConnectReq
	cmd.Cmd = utils.AssembleCmdReq(consts.CmdReConnect)
	return &cmd
}

type CmdReConnectResp struct {
	BaseRespCmd
}

func NewCmdReConnectResp() *CmdReConnectResp {
	var cmd CmdReConnectResp
	cmd.Cmd = utils.AssembleCmdResp(consts.CmdReConnect)
	return &cmd
}

type BrokerInfo struct {
	Addr   string
	Client pb.BrokerClient
}
