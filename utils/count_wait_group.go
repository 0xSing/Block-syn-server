package utils

import (
	"sync"
	"sync/atomic"
)

type FinWaitGroup struct {
	wg      sync.WaitGroup
	counter int32
}

func NewFinWaitGroup() *FinWaitGroup {
	return &FinWaitGroup{
		counter: 0,
	}
}

func (f *FinWaitGroup) Add(delta int) {
	atomic.AddInt32(&f.counter, int32(delta))
	f.wg.Add(delta)
}

func (f *FinWaitGroup) Done() {
	atomic.AddInt32(&f.counter, -1)
	f.wg.Done()
}

func (f *FinWaitGroup) Wait() {
	f.wg.Wait()
}

func (f *FinWaitGroup) GetCounter() int32 {
	return f.counter
}
