package storage

/*
	cache storage
	1. local cache(lru)
	2. redis cache
	3. database
*/

import (
	lru "github.com/hashicorp/golang-lru"
)

type StWithCache struct {
	store Storage
	cache *lru.Cache
}

func NewCache(store Storage, cacheSize int) (Storage, error) {
	r := &StWithCache{}
	r.cache, _ = lru.New(cacheSize)
	r.store = store
	return r, nil
}

func (s *StWithCache) Get(key string) (string, error) {
	// local cache get
	c, exist := s.cache.Get(key)
	if exist {
		return c.(string), nil
	}
	// redis cache get
	v, err := s.store.Get(key)
	if v != "" {
		// grab data from redis and write into local cache
		s.cache.Add(key, v)
	}
	return v, err
}

func (s *StWithCache) Set(key string, value string) error {
	old, exist := s.cache.Get(key)
	s.cache.Add(key, value)
	err := s.store.Set(key, value)
	if err != nil && exist {
		s.cache.Add(key, old)
	}
	return err
}

func (s *StWithCache) Delete(key string) error {
	old, exist := s.cache.Get(key)

	s.cache.Remove(key)
	err := s.store.Delete(key)

	if err != nil && exist {
		s.cache.Add(key, old)
	}
	return err
}

func (s *StWithCache) Close() error {
	return s.store.Close()
}

