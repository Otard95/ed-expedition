package channels

import (
	"slices"
	"sync"
)

type FanoutChannel[T any] struct {
	size      int
	listeners []chan T
	mu        sync.Mutex
	closed    bool
}

func NewFanoutChannel[T any](size int) *FanoutChannel[T] {
	return &FanoutChannel[T]{size: size}
}

func (fan *FanoutChannel[T]) Subscribe() chan T {
	fan.mu.Lock()
	defer fan.mu.Unlock()
	if fan.closed {
		panic("Tried to call Subscribe on a closed FanoutChannel")
	}

	c := make(chan T, fan.size)
	fan.listeners = append(fan.listeners, c)
	return c
}

func (fan *FanoutChannel[T]) Unsubscribe(c chan T) {
	fan.mu.Lock()
	defer fan.mu.Unlock()
	if fan.closed {
		panic("Tried to call Unsubscribe on a closed FanoutChannel")
	}

	i := slices.Index(fan.listeners, c)
	if i == -1 {
		return
	}

	close(fan.listeners[i])
	fan.listeners[i] = fan.listeners[len(fan.listeners)-1]
	fan.listeners = fan.listeners[:len(fan.listeners)-1]
}

func (fan *FanoutChannel[T]) Publish(value T) {
	fan.mu.Lock()
	defer fan.mu.Unlock()
	if fan.closed {
		panic("Tried to call Publish on a closed FanoutChannel")
	}

	for _, c := range fan.listeners {
		select {
		case c <- value:
		default:
		}
	}
}

func (fan *FanoutChannel[T]) Close() {
	fan.mu.Lock()
	defer fan.mu.Unlock()
	if fan.closed {
		panic("Tried to call Close on a closed FanoutChannel")
	}

	for _, c := range fan.listeners {
		close(c)
	}

	fan.closed = true
	fan.listeners = nil
}
