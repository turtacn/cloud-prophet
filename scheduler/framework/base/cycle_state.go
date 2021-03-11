//
package base

import (
	"errors"
	"sync"
)

const (
	NotFound = "not found"
)

type StateData interface {
	Clone() StateData
}

type StateKey string

type CycleState struct {
	mx      sync.RWMutex
	storage map[StateKey]StateData
}

func NewCycleState() *CycleState {
	return &CycleState{
		storage: make(map[StateKey]StateData),
	}
}

func (c *CycleState) Clone() *CycleState {
	if c == nil {
		return nil
	}
	copy := NewCycleState()
	for k, v := range c.storage {
		copy.Write(k, v.Clone())
	}
	return copy
}

func (c *CycleState) Read(key StateKey) (StateData, error) {
	if v, ok := c.storage[key]; ok {
		return v, nil
	}
	return nil, errors.New(NotFound)
}

func (c *CycleState) Write(key StateKey, val StateData) {
	c.storage[key] = val
}

func (c *CycleState) Delete(key StateKey) {
	delete(c.storage, key)
}

func (c *CycleState) Lock() {
	c.mx.Lock()
}

func (c *CycleState) Unlock() {
	c.mx.Unlock()
}

func (c *CycleState) RLock() {
	c.mx.RLock()
}

func (c *CycleState) RUnlock() {
	c.mx.RUnlock()
}
