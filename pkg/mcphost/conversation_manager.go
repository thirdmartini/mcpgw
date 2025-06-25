package mcphost

import (
	"sync"

	"github.com/thirdmartini/mcpgw/pkg/history"
)

type ConversationManager struct {
	lock         sync.RWMutex
	systemPrompt string
	conversation map[string]*Conversation
}

func (s *ConversationManager) newConversation(id string) *Conversation {
	conversation := &Conversation{
		Id:       id,
		Messages: []history.HistoryMessage{},
		Window:   16,
	}

	if s.systemPrompt != "" {
		conversation.Append(history.HistoryMessage{
			Role: "system",
			Content: []history.ContentBlock{
				{
					Type: "text",
					Text: s.systemPrompt,
				},
			},
		})
	}
	return conversation
}

func (s *ConversationManager) GetConversation(id string) *Conversation {
	if id == "" {
		return s.newConversation("")
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if conversation, ok := s.conversation[id]; ok {
		return conversation
	}

	return s.newConversation(id)
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

func NewConversationManager(systemPrompt string) *ConversationManager {
	return &ConversationManager{
		systemPrompt: systemPrompt,
		conversation: make(map[string]*Conversation),
	}
}
