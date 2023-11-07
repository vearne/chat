package dao

import (
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"time"
)

func GetAccount(accountId uint64) (*model.Account, error) {
	var account model.Account
	err := resource.MySQLClient.Where("id = ?", accountId).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func GetSession(sessionId uint64) (*model.Session, error) {
	var session model.Session
	err := resource.MySQLClient.Where("id = ?", sessionId).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func GetSessionPartner(sessionId uint64, accountId uint64) (*model.SessionAccount, error) {
	var partner model.SessionAccount
	err := resource.MySQLClient.Where("session_id = ? and account_id != ?",
		sessionId, accountId).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

func CreateOutMsg(msgType pb.MsgTypeEnum, senderId, sessionId uint64, content string) (*model.OutBox, error) {
	outMsg := model.OutBox{SenderId: senderId, SessionId: sessionId}
	outMsg.Status = consts.OutBoxStatusNormal
	outMsg.MsgType = int(msgType)
	outMsg.Content = content
	outMsg.CreatedAt = time.Now()
	outMsg.ModifiedAt = outMsg.CreatedAt

	err := resource.MySQLClient.Create(&outMsg).Error
	if err != nil {
		return nil, err
	}
	return &outMsg, nil
}

func CreateInMsg(senderId, msgId, receiverId uint64) (*model.InBox, error) {
	inMsg := model.InBox{}
	inMsg.SenderId = senderId
	inMsg.MsgId = msgId
	inMsg.ReceiverId = receiverId

	err := resource.MySQLClient.Create(&inMsg).Error
	if err != nil {
		return nil, err
	}
	return &inMsg, nil
}
