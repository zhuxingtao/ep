package ep

import (
	"github.com/google/uuid"
	"sync"
	"sync/atomic"
)

func StrUUID() string {
	u, _ := uuid.NewUUID()
	return u.String()
}

func SyncLen(m *sync.Map) int {
	if m == nil {
		return 0
	}
	n := 0
	m.Range(func(key, value interface{}) bool {
		n++
		return true
	})
	return n
}

type AtomicBool struct {
	v int32
}

func (b *AtomicBool) Val() bool {
	return atomic.LoadInt32(&b.v) > 0
}

func (b *AtomicBool) Set() {
	atomic.StoreInt32(&b.v, 1)
}

func (b *AtomicBool) Unset() {
	atomic.StoreInt32(&b.v, 0)
}
