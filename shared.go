package main

import (
	"sync"
)

type storage struct {
	lock *sync.RWMutex
	Data map[string]any
}

var SharedData *storage

func init() {
	sync.OnceFunc(func() {
		SharedData = &storage{
			lock: &sync.RWMutex{},
			Data: make(map[string]any),
		}
	})()
}

func (s *storage) Set(key string, value any) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Data[key] = value

}

func (s *storage) Get(key string) any {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value := s.Data[key]
	return value
}

func (s *storage) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.Data, key)
}

func (s *storage) Foreach(callback func(key string, value any)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for key, value := range s.Data {
		callback(key, value)
	}
}

func (s *storage) Transfer(callback func(key string, value any)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for key, value := range s.Data {
		callback(key, value)
		delete(s.Data, key)
	}
}

func (s *storage) Exists(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, exists := s.Data[key]
	return exists
}

func (s *storage) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.Data)
}

func (s *storage) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Data = make(map[string]any)
}
