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
	users    map[string]string
	profiles map[string]string
	mu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string][]Message),
		users:    make(map[string]string),
		profiles: make(map[string]string),
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

func (s *Store) GetAllSessions() map[string][]Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]Message)
	for k, v := range s.sessions {
		result[k] = v
	}
	return result
}

func (s *Store) SetSessionUser(sessionID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[sessionID] = userID
}

func (s *Store) GetSessionUser(sessionID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[sessionID]
}

func (s *Store) SetSessionProfile(sessionID, profileID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles[sessionID] = profileID
}

func (s *Store) GetSessionProfile(sessionID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profiles[sessionID]
}

func (s *Store) ClearSessionUser(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, sessionID)
	delete(s.profiles, sessionID)
}