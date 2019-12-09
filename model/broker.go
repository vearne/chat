package model

import (
	"gopkg.in/olahol/melody.v1"
	"sync"
	"time"
)

// melody.Session keys 是一个Map，多线程处理有并发问题
type BizHub struct {
	sync.RWMutex
	clientMap map[uint64]*Client
}

func NewBizHub() *BizHub {
	h := BizHub{}
	h.clientMap = make(map[uint64]*Client, 10)
	return &h
}

func (h *BizHub) SetClient(nickName string, accountId uint64, session *melody.Session) {
	h.Lock()
	defer h.Unlock()
	h.clientMap[accountId] = NewClient(nickName, accountId, session)
}

func (h *BizHub) RemoveClient(accountId uint64) {
	h.Lock()
	defer h.Unlock()
	if _, ok := h.clientMap[accountId]; ok {
		delete(h.clientMap, accountId)
	}
}

func (h *BizHub) GetSession(accountId uint64) (*melody.Session, bool) {
	h.RLock()
	defer h.RUnlock()
	client, ok := h.clientMap[accountId]
	if ok {
		return client.Session, true
	} else {
		return nil, false
	}
}

func (h *BizHub) SetLastPong(accountId uint64, t time.Time) {
	h.Lock()
	defer h.Unlock()
	if client, ok := h.clientMap[accountId]; ok {
		client.LastPong = t
	}
}

func (h *BizHub) GetAllClient() []*Client {
	h.RLock()
	defer h.RUnlock()
	res := make([]*Client, len(h.clientMap))
	for _, v := range h.clientMap {
		res = append(res, v)
	}
	return res
}

type Client struct {
	Session *melody.Session `json:"-"`

	NickName  string    `json:"nickName"`
	AccountId uint64    `json:"accountId"`
	LastPong  time.Time `json:"lastPong"`
}

func NewClient(nickName string, accountId uint64, session *melody.Session) *Client {
	c := Client{NickName: nickName, AccountId: accountId, Session: session}
	c.LastPong = time.Now()
	return &c
}
