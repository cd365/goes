package syncx

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
)

type WaitGroup struct {
	waitGroup *sync.WaitGroup
	treat     func(reason []byte)
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		waitGroup: &sync.WaitGroup{},
	}
}

func (s *WaitGroup) Add(delta int) {
	s.waitGroup.Add(delta)
}

func (s *WaitGroup) Done() {
	s.waitGroup.Done()
}

func (s *WaitGroup) Wait() {
	s.waitGroup.Wait()
}

func (s *WaitGroup) Treat(treat func(reason []byte)) *WaitGroup {
	s.treat = treat
	return s
}

func (s *WaitGroup) Go(coroutine func()) {
	if coroutine == nil {
		return
	}
	s.Add(1)
	go func() {
		defer s.Done()
		defer func() {
			if s.treat == nil {
				return
			}
			if err := recover(); err != nil {
				reason := bytes.NewBufferString(fmt.Sprintf("panic: %v\n", err))
				buf := make([]byte, 1<<16)
				n := runtime.Stack(buf, false)
				reason.Write(buf[:n])
				s.treat(reason.Bytes())
			}
		}()
		coroutine()
	}()
}
