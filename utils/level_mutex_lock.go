package utils

import (
	"sync"
	"sync/atomic"
)

type LockCounter struct {
	mu      sync.Mutex
	counter int32
}

type DBKeyLock struct {
	locks sync.Map
}

func (tl *DBKeyLock) Lock(key interface{}) {
	wrapper, _ := tl.locks.LoadOrStore(key, &LockCounter{})
	atomic.AddInt32(&wrapper.(*LockCounter).counter, 1)
	wrapper.(*LockCounter).mu.Lock()
}

func (tl *DBKeyLock) Unlock(key interface{}) {
	wrapper, ok := tl.locks.Load(key)
	if !ok {
		return // 键不存在，无需解锁
	}

	w := wrapper.(*LockCounter)
	w.mu.Unlock()
	if atomic.AddInt32(&w.counter, -1) == 0 {
		tl.locks.Delete(key)
	}
}
