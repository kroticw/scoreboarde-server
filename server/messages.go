package server

import (
	"scoreboarde-server/Play"
	"sync"
)

type Message struct {
	Time       int64        `json:"time"`
	CommandOne Play.Command `json:"command_one"`
	CommandTwo Play.Command `json:"command_two"`
	Period     Play.Period  `json:"period"`
}

type AtomicMessageHistory struct {
	mu      sync.Mutex
	history []Message
}

func InitAtomicMessageHistory() *AtomicMessageHistory {
	return &AtomicMessageHistory{
		history: make([]Message, 0),
	}
}

func (a *AtomicMessageHistory) Len() int {
	return len(a.history)
}

func (a *AtomicMessageHistory) Push(x interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = append(a.history, x.(Message))
}

func (a *AtomicMessageHistory) Pop() Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.history) == 0 {
		return Message{}
	}
	n := len(a.history)
	x := a.history[n-1]
	a.history = a.history[0 : n-1]
	return x
}

func (a *AtomicMessageHistory) PopFirst() Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.history) == 0 {
		return Message{}
	}
	x := a.history[0]
	a.history = a.history[1:]
	return x
}

func (a *AtomicMessageHistory) GetLast() *Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.history) == 0 {
		return &Message{}
	}
	return &a.history[len(a.history)-1]
}

func (a *AtomicMessageHistory) Get(index int) *Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.history) == index {
		return &Message{}
	}
	return &a.history[index]
}
