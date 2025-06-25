package mcphost

import (
	"sync"

	"github.com/thirdmartini/mcpgw/pkg/history"
)

type ConversationManager struct {
	lock         sync.RWMutex
	conversation map[string]*Conversation
}

func (s *ConversationManager) GetConversation(id string) *Conversation {
	if id == "" {
		return &Conversation{
			Id:       "",
			Messages: []history.HistoryMessage{},
			Window:   16,
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if session, ok := s.conversation[id]; ok {
		return session
	}

	return &Conversation{
		Id:       id,
		Messages: []history.HistoryMessage{},
		Window:   16,
	}
}

func (s *ConversationManager) PutConversation(conversation *Conversation) {
	if conversation.Id == "" {
		return
	}

	conversation.Prune()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.conversation[conversation.Id] = conversation
}

func NewConversationManager() *ConversationManager {
	return &ConversationManager{
		conversation: make(map[string]*Conversation),
	}
}
