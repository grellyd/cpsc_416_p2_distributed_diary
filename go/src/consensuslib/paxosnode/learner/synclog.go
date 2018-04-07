package learner

// Code taken from https://medium.com/@deckarep/the-new-kid-in-town-gos-sync-map-de24a6bf7c2c

import (
	"sync"
)

type SyncLog struct {
	sync.RWMutex
	internal map[uint64]*MessageAccepted
}

func NewSyncLog() *SyncLog {
	return &SyncLog{
		internal: make(map[uint64]*MessageAccepted, 0),
	}
}

func (rm *SyncLog) Load(key uint64) (value *MessageAccepted, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

func (rm *SyncLog) Delete(key uint64) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *SyncLog) Store(key uint64, value *MessageAccepted) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}
