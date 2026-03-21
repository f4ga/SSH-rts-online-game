package events

import "ssh-arena-app/internal"

// EventBus возвращает реализацию шины событий.
func EventBus() internal.EventBus {
	return NewEventBus()
}