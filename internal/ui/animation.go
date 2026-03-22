package ui

import (
	"fmt"
	"sync"
	"time"
)

// Animation представляет анимационный эффект.
type Animation struct {
	ID        string
	X, Y      int          // позиция в мировых координатах
	Symbol    rune         // отображаемый символ
	Color     ColorRGB     // цвет
	StartTime time.Time    // время начала
	Duration  time.Duration // длительность
}

// AnimationManager управляет активными анимациями.
type AnimationManager struct {
	mu         sync.RWMutex
	animations map[string]*Animation
}

// NewAnimationManager создаёт новый менеджер анимаций.
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		animations: make(map[string]*Animation),
	}
}

// Add добавляет анимацию.
func (am *AnimationManager) Add(anim *Animation) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.animations[anim.ID] = anim
}

// Remove удаляет анимацию.
func (am *AnimationManager) Remove(id string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.animations, id)
}

// Update удаляет устаревшие анимации.
func (am *AnimationManager) Update(now time.Time) {
	am.mu.Lock()
	defer am.mu.Unlock()
	for id, anim := range am.animations {
		if now.Sub(anim.StartTime) > anim.Duration {
			delete(am.animations, id)
		}
	}
}

// GetActive возвращает все активные анимации.
func (am *AnimationManager) GetActive() []*Animation {
	am.mu.RLock()
	defer am.mu.RUnlock()
	var result []*Animation
	for _, anim := range am.animations {
		result = append(result, anim)
	}
	return result
}

// CreateExplosion создаёт анимацию взрыва в заданной точке.
func CreateExplosion(x, y int) *Animation {
	return &Animation{
		ID:        fmt.Sprintf("explosion-%d-%d", x, y),
		X:         x,
		Y:         y,
		Symbol:    '*',
		Color:     ColorUnit, // красный
		StartTime: time.Now(),
		Duration:  200 * time.Millisecond,
	}
}

// CreateMoveTrail создаёт след движения.
func CreateMoveTrail(x, y int) *Animation {
	return &Animation{
		ID:        fmt.Sprintf("trail-%d-%d", x, y),
		X:         x,
		Y:         y,
		Symbol:    '.',
		Color:     ColorRGB{0, 255, 255}, // голубой
		StartTime: time.Now(),
		Duration:  100 * time.Millisecond,
	}
}