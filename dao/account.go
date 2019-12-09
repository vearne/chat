package dao

import (
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
	"time"
)

func GetAccount(accountId uint64) *model.Account {
	var account model.Account
	resource.MySQLClient.Where("id = ?", accountId).First(&account)
	return &account
}

func GetSession(sessionId uint64) *model.Session {
	var session model.Session
	resource.MySQLClient.Where("id = ?", sessionId).First(&session)
	return &session
}

func CreateOutMsg(msgType pb.MsgTypeEnum, senderId, sessionId uint64, content string) *model.OutBox {
	outMsg := model.OutBox{SenderId: senderId, SessionId: sessionId}
	outMsg.Status = consts.OutBoxStatusNormal
	outMsg.MsgType = int(msgType)
	outMsg.Content = content
	outMsg.CreatedAt = time.Now()
	outMsg.ModifiedAt = outMsg.CreatedAt

	resource.MySQLClient.Create(&outMsg)
	return &outMsg
}

func CreateInMsg(senderId, msgId, receiverId uint64) *model.InBox {
	inMsg := model.InBox{}
	inMsg.SenderId = senderId
	inMsg.MsgId = msgId
	inMsg.ReceiverId = receiverId
	resource.MySQLClient.Create(&inMsg)
	return &inMsg
}
