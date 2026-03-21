package network

import (
	"sync"
)

// Hub управляет активными сессиями.
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewHub создаёт новый хаб.
func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*Session),
	}
}

// Register регистрирует сессию.
func (h *Hub) Register(id string, session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[id] = session
}

// Unregister удаляет сессию.
func (h *Hub) Unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, id)
}

// Send отправляет данные конкретной сессии.
func (h *Hub) Send(id string, data []byte) error {
	h.mu.RLock()
	session, ok := h.sessions[id]
	h.mu.RUnlock()
	if !ok {
		return ErrSessionNotFound
	}
	_, err := session.Write(data)
	return err
}

// Broadcast рассылает данные всем сессиям.
func (h *Hub) Broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, session := range h.sessions {
		session.Write(data) // игнорируем ошибки
	}
}

// Close закрывает все сессии.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, session := range h.sessions {
		session.Close()
	}
	h.sessions = make(map[string]*Session)
}

// ErrSessionNotFound ошибка, когда сессия не найдена.
var ErrSessionNotFound = &HubError{"session not found"}

// HubError представляет ошибку хаба.
type HubError struct {
	msg string
}

func (e *HubError) Error() string {
	return e.msg
}