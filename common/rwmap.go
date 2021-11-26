package common

import "sync"

type Rwmap struct {
	lock  sync.RWMutex
	store map[interface{}]interface{}
}

func NewRwmap() *Rwmap {
	return &Rwmap{
		lock:  sync.RWMutex{},
		store: make(map[interface{}]interface{}),
	}
}

func (m *Rwmap) Get(key interface{}) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.store[key]
}

func (m *Rwmap) Put(key interface{}, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.store[key] = val
}

func (m *Rwmap) Delete(key interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.store, key)
}
