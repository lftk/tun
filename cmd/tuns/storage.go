package main

import (
	"encoding/json"
	"errors"
	"sync"
)

type storage interface {
	Load(string) (*proxy, error)
	Store(*proxy) error
	Update(*proxy) error
	Delete(string) error
}

type fileStorage struct {
	sync.RWMutex
	proxies map[string]*proxy
}

func openFileStorage(path string) (s storage, err error) {
	s = &fileStorage{
		proxies: make(map[string]*proxy),
	}
	return
}

func (s *fileStorage) flush() (err error) {
	b, err := json.Marshal(s.proxies)
	if err != nil {
		return
	}
	_ = b
	return
}

func (s *fileStorage) Load(name string) (p *proxy, err error) {
	s.RLock()
	defer s.RUnlock()

	p, ok := s.proxies[name]
	if !ok {
		err = errors.New("")
	}
	return
}

func (s *fileStorage) Store(p *proxy) (err error) {
	s.Lock()
	s.proxies[p.Name] = p
	s.Unlock()
	return
}

func (s *fileStorage) Update(p *proxy) (err error) {
	s.Lock()
	s.proxies[p.Name] = p
	s.Unlock()
	return
}

func (s *fileStorage) Delete(name string) (err error) {
	s.Lock()
	delete(s.proxies, name)
	s.Unlock()
	return
}
