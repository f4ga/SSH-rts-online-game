package events

import (
	"sync"

	"ssh-arena-app/internal"
)

// eventBus реализует internal.EventBus.
type eventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]internal.EventListener
}

// NewEventBus создаёт новую шину событий.
func NewEventBus() internal.EventBus {
	return &eventBus{
		subscribers: make(map[string][]internal.EventListener),
	}
}

// Publish публикует событие всем подписчикам соответствующего типа.
func (b *eventBus) Publish(event internal.Event) {
	b.mu.RLock()
	handlers, ok := b.subscribers[event.Type]
	b.mu.RUnlock()
	if !ok {
		return
	}
	// Вызываем обработчики асинхронно, чтобы не блокировать издателя.
	for _, h := range handlers {
		go h(event)
	}
}

// Subscribe добавляет подписчика на события определённого типа.
func (b *eventBus) Subscribe(eventType string, listener internal.EventListener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], listener)
}

// Unsubscribe удаляет подписчика.
func (b *eventBus) Unsubscribe(eventType string, listener internal.EventListener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	handlers := b.subscribers[eventType]
	for i, h := range handlers {
		if &h == &listener { // сравнение указателей на функции
			b.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}