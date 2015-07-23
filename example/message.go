package main

import (
	"sync"
	"time"
)

type Message struct {
	Id     int64  `json:"id"`
	Msg    string `json:"message"`
	Time   int64  `json:"time"`
	Remark string `json:"remark,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////

type MessageManager struct {
	msgs map[int64]*Message

	counter int64
	lock    sync.RWMutex
}

func NewMessageManager() *MessageManager {
	m := new(MessageManager)
	m.msgs = make(map[int64]*Message)
	return m
}

func (m *MessageManager) Add(s string, remark string) *Message {
	m.lock.Lock()

	m.counter += 1
	msg := new(Message)
	msg.Id = m.counter
	msg.Msg = s
	msg.Time = time.Now().Unix()
	msg.Remark = remark
	m.msgs[msg.Id] = msg

	m.lock.Unlock()

	return msg
}

func (m *MessageManager) Get(id int64) *Message {
	m.lock.RLock()
	msg, _ := m.msgs[id]
	m.lock.RUnlock()
	return msg
}

func (m *MessageManager) List() []*Message {
	m.lock.RLock()

	msgs := make([]*Message, 0, len(m.msgs))
	for _, msg := range m.msgs {
		msgs = append(msgs, msg)
	}

	m.lock.RUnlock()

	return msgs
}
