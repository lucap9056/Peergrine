package genericchannels

import "sync"

type Channels[T any] struct {
	mux      *sync.RWMutex
	channels map[string]chan T
}

func New[T any]() Channels[T] {
	return Channels[T]{
		mux:      new(sync.RWMutex),
		channels: make(map[string]chan T),
	}
}

func (s *Channels[T]) Add(key string) chan T {
	ch := make(chan T)

	s.mux.Lock()
	s.channels[key] = ch
	s.mux.Unlock()

	return ch
}

func (s *Channels[T]) Get(key string) chan T {
	s.mux.RLock()
	ch, exists := s.channels[key]
	s.mux.RUnlock()

	if exists {
		return ch
	}
	return nil
}

func (s *Channels[T]) Del(key string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	ch, exists := s.channels[key]
	if exists {
		close(ch)
		delete(s.channels, key)
	}
}

func (s *Channels[T]) Close() {
	s.mux.Lock()
	defer s.mux.Unlock()

	for key, channel := range s.channels {
		close(channel)
		delete(s.channels, key)
	}
}
