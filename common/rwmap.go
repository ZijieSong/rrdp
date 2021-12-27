package common

import "sync"

type Rwmap struct {
	Lock  sync.RWMutex
	Store map[interface{}]interface{}
}

func NewRwmap() *Rwmap {
	return &Rwmap{
		Lock:  sync.RWMutex{},
		Store: make(map[interface{}]interface{}),
	}
}

func (m *Rwmap) Get(key interface{}) interface{} {
	m.Lock.RLock()
	defer m.Lock.RUnlock()
	return m.Store[key]
}

func (m *Rwmap) Put(key interface{}, val interface{}) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	m.Store[key] = val
}

func (m *Rwmap) Delete(key interface{}) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	delete(m.Store, key)
}
