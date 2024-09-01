package sync_primitives

import (
	"sync/atomic"
)

type Barrier struct {
	remains uint32
	once    uint32
	done    chan struct{}
}

func NewBarrier(remains uint32) *Barrier {
	return &Barrier{
		remains: remains,
		done:    make(chan struct{}),
	}
}

// Join block goroutine until other N goroutines will be joined
func (b *Barrier) Join() {
	for {
		// if all goroutines joined - barrier is open
		remains := atomic.LoadUint32(&b.remains)
		if remains == 0 {
			break
		}

		// try acquire barrier position
		if atomic.CompareAndSwapUint32(&b.remains, remains, remains-1) {
			// if all goroutines joined, need open barrier
			if atomic.LoadUint32(&b.remains) == 0 && atomic.CompareAndSwapUint32(&b.once, 0, 1) {
				close(b.done)
			}
			// wait open
			<-b.done
		}
	}
}
