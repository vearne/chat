package biz

import (
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/model"
	"time"
)

func CreateSession(partner1, partner2 uint64) (*model.Session, error) {
	var session model.Session
	var err error
	session.Status = consts.SessionStatusInUse
	session.CreatedAt = time.Now()
	session.ModifiedAt = session.CreatedAt
	err = resource.MySQLClient.Create(&session).Error
	if err != nil {
		return nil, err
	}
	// 2. 创建会话中的对象 session-account
	s1 := model.SessionAccount{SessionId: session.ID, AccountId: partner1}
	err = resource.MySQLClient.Create(&s1).Error
	if err != nil {
		return nil, err
	}
	s2 := model.SessionAccount{SessionId: session.ID, AccountId: partner2}
	err = resource.MySQLClient.Create(&s2).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}
