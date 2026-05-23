package session

import (
	"sync"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Store struct {
	sessions map[string][]Message
	mu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string][]Message),
	}
}

func (s *Store) AddMessage(sessionID, role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages := s.sessions[sessionID]
	messages = append(messages, Message{Role: role, Content: content})

	if len(messages) > 6 {
		messages = messages[len(messages)-6:]
	}

	s.sessions[sessionID] = messages
}

func (s *Store) GetMessages(sessionID string) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}

	result := make([]Message, len(messages))
	copy(result, messages)
	return result
}

func (s *Store) Clear(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}