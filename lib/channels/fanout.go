package channels

import (
	"fmt"
	"slices"
	"sync"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type FanoutChannel[T any] struct {
	name           string
	size           int
	publishTimeout time.Duration
	listeners      []chan T
	mu             sync.Mutex
	closed         bool
	logger         wailsLogger.Logger
}

func NewFanoutChannel[T any](name string, size int, publishTimeout time.Duration, logger wailsLogger.Logger) *FanoutChannel[T] {
	return &FanoutChannel[T]{name: name, size: size, publishTimeout: publishTimeout, logger: logger}
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

	fan.logger.Trace(fmt.Sprintf("[FanoutChannel:%s] Publishing to %d listeners", fan.name, len(fan.listeners)))
	for i, c := range fan.listeners {
		select {
		case c <- value:
			fan.logger.Trace(fmt.Sprintf("[FanoutChannel:%s] Sent to listener %d successfully", fan.name, i))
		case <-time.After(fan.publishTimeout):
			fan.logger.Trace(fmt.Sprintf("[FanoutChannel:%s] Timeout sending to listener %d", fan.name, i))
		}
	}
	fan.logger.Trace(fmt.Sprintf("[FanoutChannel:%s] Publish complete", fan.name))
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
