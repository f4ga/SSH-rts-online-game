package events

import (
	"sync"
	"sync/atomic"

	"ssh-arena-app/internal"
)

// eventBus реализует internal.EventBus.
type eventBus struct {
	mu            sync.RWMutex
	subscribers   map[string][]*subscriberEntry
	syncPublish   bool // если true, обработчики вызываются синхронно
}

type subscriberEntry struct {
	id       uint64
	listener internal.EventListener
}

var nextSubscriberID uint64

// NewEventBus создаёт новую шину событий с синхронной публикацией по умолчанию.
func NewEventBus() internal.EventBus {
	return &eventBus{
		subscribers: make(map[string][]*subscriberEntry),
		syncPublish: true,
	}
}

// NewAsyncEventBus создаёт шину событий, где обработчики вызываются асинхронно.
func NewAsyncEventBus() internal.EventBus {
	return &eventBus{
		subscribers: make(map[string][]*subscriberEntry),
		syncPublish: false,
	}
}

// Publish публикует событие всем подписчикам соответствующего типа.
// Если syncPublish == true, обработчики вызываются синхронно в порядке подписки.
// Иначе каждый обработчик запускается в отдельной горутине, но ожидается завершение всех.
func (b *eventBus) Publish(event internal.Event) {
	b.mu.RLock()
	entries, ok := b.subscribers[event.Type]
	if !ok {
		b.mu.RUnlock()
		return
	}
	// Создаём копию списка обработчиков, чтобы избежать гонок при изменении подписчиков.
	handlers := make([]internal.EventListener, len(entries))
	for i, entry := range entries {
		handlers[i] = entry.listener
	}
	b.mu.RUnlock()

	if b.syncPublish {
		for _, h := range handlers {
			h(event)
		}
	} else {
		var wg sync.WaitGroup
		for _, h := range handlers {
			wg.Add(1)
			go func(handler internal.EventListener) {
				defer wg.Done()
				handler(event)
			}(h)
		}
		wg.Wait()
	}
}

// Subscribe добавляет подписчика на события определённого типа.
// Возвращает идентификатор подписки, который можно использовать для UnsubscribeByID.
// Реализует internal.EventBus.Subscribe.
func (b *eventBus) Subscribe(eventType string, listener internal.EventListener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := atomic.AddUint64(&nextSubscriberID, 1)
	entry := &subscriberEntry{
		id:       id,
		listener: listener,
	}
	b.subscribers[eventType] = append(b.subscribers[eventType], entry)
}

// Unsubscribe удаляет подписчика по ссылке на функцию.
// Реализует internal.EventBus.Unsubscribe.
func (b *eventBus) Unsubscribe(eventType string, listener internal.EventListener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entries, ok := b.subscribers[eventType]
	if !ok {
		return
	}
	for i, entry := range entries {
		// Сравниваем указатели на функции (как в оригинале).
		if &entry.listener == &listener {
			b.subscribers[eventType] = append(entries[:i], entries[i+1:]...)
			if len(b.subscribers[eventType]) == 0 {
				delete(b.subscribers, eventType)
			}
			return
		}
	}
}

// SubscribeWithID добавляет подписчика и возвращает ID подписки.
// Это расширение, не входящее в стандартный интерфейс.
func (b *eventBus) SubscribeWithID(eventType string, listener internal.EventListener) uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := atomic.AddUint64(&nextSubscriberID, 1)
	entry := &subscriberEntry{
		id:       id,
		listener: listener,
	}
	b.subscribers[eventType] = append(b.subscribers[eventType], entry)
	return id
}

// UnsubscribeByID удаляет подписчика по идентификатору подписки.
// Это расширение, не входящее в стандартный интерфейс.
func (b *eventBus) UnsubscribeByID(eventType string, id uint64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	entries, ok := b.subscribers[eventType]
	if !ok {
		return false
	}
	for i, entry := range entries {
		if entry.id == id {
			b.subscribers[eventType] = append(entries[:i], entries[i+1:]...)
			if len(b.subscribers[eventType]) == 0 {
				delete(b.subscribers, eventType)
			}
			return true
		}
	}
	return false
}