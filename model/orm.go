package model

import "time"

type Account struct {
	ID         uint64    `gorm:"column:id" json:"id"`
	NickName   string    `gorm:"column:nickname" json:"nickname"`
	Status     int       `gorm:"column:status" json:"status"`
	Broker     string    `gorm:"column:broker" json:"broker"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"-"`
	ModifiedAt time.Time `gorm:"column:modified_at" json:"-"`
}

func (Account) TableName() string {
	return "account"
}

type Session struct {
	ID         uint64    `gorm:"column:id" json:"id"`
	Status     int       `gorm:"column:status" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"-"`
	ModifiedAt time.Time `gorm:"column:modified_at" json:"-"`
}

func (Session) TableName() string {
	return "session"
}

type SessionAccount struct {
	ID        uint64 `gorm:"column:id" json:"id"`
	SessionId uint64 `gorm:"column:session_id" json:"session_id"`
	AccountId uint64 `gorm:"column:account_id" json:"account_id"`
}

func (SessionAccount) TableName() string {
	return "session_account"
}

type OutBox struct {
	ID         uint64    `gorm:"column:id" json:"id"`
	SenderId   uint64    `gorm:"column:sender_id" json:"sender_id"`
	SessionId  uint64    `gorm:"column:session_id" json:"session_id"`
	Status     int       `gorm:"column:status" json:"status"`
	MsgType    int       `gorm:"column:msg_type" json:"msg_type"`
	Content    string    `gorm:"column:content" json:"content"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"-"`
	ModifiedAt time.Time `gorm:"column:modified_at" json:"-"`
}

func (OutBox) TableName() string {
	return "outbox"
}

type InBox struct {
	ID         uint64 `gorm:"column:id" json:"id"`
	SenderId   uint64 `gorm:"column:sender_id" json:"sender_id"`
	MsgId      uint64 `gorm:"column:msg_id" json:"msg_id"`
	ReceiverId uint64 `gorm:"column:receiver_id" json:"receiver_id"`
}

func (InBox) TableName() string {
	return "inbox"
}
