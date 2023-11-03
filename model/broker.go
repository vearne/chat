package model

import (
	"encoding/json"
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

	delete(h.clientMap, accountId)
}

func (h *BizHub) GetClient(accountId uint64) (*Client, bool) {
	h.RLock()
	defer h.RUnlock()
	client, ok := h.clientMap[accountId]
	return client, ok
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
	res := make([]*Client, 0, len(h.clientMap))
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

func (s *Client) Write(obj any) error {
	bt, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return s.Session.Write(bt)
}

type SessionWrapper struct {
	Session *melody.Session `json:"-"`
}

func NewSessionWrapper(s *melody.Session) *SessionWrapper {
	return &SessionWrapper{Session: s}
}

func (s *SessionWrapper) Write(obj any) error {
	bt, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return s.Session.Write(bt)
}
