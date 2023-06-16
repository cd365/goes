package container

import (
	"errors"
	"hash/fnv"
	"sync"
)

type mapNode struct {
	rwm *sync.RWMutex
	val map[string]interface{}
}

func (s *mapNode) Read(fn func(m map[string]interface{})) {
	if fn == nil {
		return
	}
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	fn(s.val)
}

func (s *mapNode) Write(fn func(m map[string]interface{})) {
	if fn == nil {
		return
	}
	s.rwm.Lock()
	defer s.rwm.Unlock()
	fn(s.val)
}

func (s *mapNode) Put(key string, value interface{}) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.val[key] = value
}

func (s *mapNode) Get(key string) interface{} {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	if val, ok := s.val[key]; ok {
		return val
	}
	return nil
}

func (s *mapNode) Delete(key string) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	delete(s.val, key)
}

func newMapNode() *mapNode {
	return &mapNode{
		rwm: &sync.RWMutex{},
		val: make(map[string]interface{}),
	}
}

type SliceMap struct {
	length int
	nodes  []*mapNode
}

func (s *SliceMap) Length() int {
	return s.length
}

func (s *SliceMap) Index(key string, length int) (int, error) {
	h := fnv.New32a()
	n, err := h.Write([]byte(key))
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errors.New("the value of `key` is empty")
	}
	return int(h.Sum32() % uint32(length)), nil
}

func (s *SliceMap) Put(key string, value interface{}) error {
	i, err := s.Index(key, s.length)
	if err != nil {
		return err
	}
	s.nodes[i].Put(key, value)
	return nil
}

func (s *SliceMap) Get(key string) (interface{}, error) {
	i, err := s.Index(key, s.length)
	if err != nil {
		return nil, err
	}
	return s.nodes[i].Get(key), nil
}

func (s *SliceMap) Delete(key string) error {
	i, err := s.Index(key, s.length)
	if err != nil {
		return err
	}
	s.nodes[i].Delete(key)
	return nil
}

func (s *SliceMap) AllNodeLoopRead(fn func(i int, m map[string]interface{})) {
	for i := 0; i < s.length; i++ {
		s.nodes[i].Read(func(m map[string]interface{}) { fn(i, m) })
	}
}

func (s *SliceMap) AllNodeLoopWrite(fn func(i int, m map[string]interface{})) {
	for i := 0; i < s.length; i++ {
		s.nodes[i].Write(func(m map[string]interface{}) { fn(i, m) })
	}
}

func NewSliceMap(length int) *SliceMap {
	nodes := make([]*mapNode, length)
	for i := 0; i < length; i++ {
		nodes[i] = newMapNode()
	}
	return &SliceMap{
		length: length,
		nodes:  nodes,
	}
}
