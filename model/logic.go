package model

import (
	pb "github.com/vearne/chat/proto"
	"sync"
)

// melody.Session keys 是一个Map，多线程处理有并发问题
type BrokerHub struct {
	sync.RWMutex
	brokerMap map[string]pb.BrokerClient
}

func NewBrokerHub() *BrokerHub {
	h := BrokerHub{}
	h.brokerMap = make(map[string]pb.BrokerClient, 10)
	return &h
}

func (h *BrokerHub) GetBroker(brokerAddr string) (pb.BrokerClient, bool) {
	h.RLock()
	defer h.RUnlock()
	client, ok := h.brokerMap[brokerAddr]
	if ok {
		return client, true
	} else {
		return nil, false
	}
}

func (h *BrokerHub) SetBroker(brokerAddr string, client pb.BrokerClient) {
	h.Lock()
	defer h.Unlock()
	h.brokerMap[brokerAddr] = client
}

func (h *BrokerHub) Size() int {
	h.RLock()
	defer h.RUnlock()
	return len(h.brokerMap)
}
