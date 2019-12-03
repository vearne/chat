package model

import (
	"gopkg.in/olahol/melody.v1"
	"sync"
)

type BizHub struct {
	sync.Mutex
	sessionMap map[uint64]*melody.Session
}

func NewBizHub() *BizHub {
	h := BizHub{}
	h.sessionMap = make(map[uint64]*melody.Session, 10)
	return &h
}

func (h *BizHub) SetSession(accountId uint64, s *melody.Session) {
	h.Lock()
	defer h.Unlock()
	h.sessionMap[accountId] = s
}

func (h *BizHub) GetSession(accountId uint64, s *melody.Session) *melody.Session {
	h.Lock()
	defer h.Unlock()
	return h.sessionMap[accountId]
}
