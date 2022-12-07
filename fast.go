package fwp

import (
	"runtime"
	"sync"

	rtm "github.com/0xmjk/go-tsx-rtm"
	"github.com/intel-go/cpuid"
)

type WorkerPool struct {
	Max int
	m   sync.Mutex
	n   int
	q   cbuf
}

func (s *WorkerPool) Go(fn func()) {
	if s.Max <= 0 {
		go fn()
		return
	}
	s.lock()
	if s.n >= s.Max {
		s.q.put(fn)
		s.unlock()
		return
	}
	s.n++
	s.unlock()
	go s.worker(fn)
}

func (s *WorkerPool) worker(fn func()) {
	for {
		fn()

		s.lock()
		var ok bool
		fn, ok = s.q.get()
		if !ok {
			s.n--
			if s.n == 0 {
				// We are the last running worker and we are shutting down
				// because the work queue is empty: reset the queue
				// (dropping also any queue storage) so that we do not keep
				// references to previously-enqueued functions alive
				// (preventing them from being GCed), and so that in general
				// we consume no memory while we are idle.
				s.q.reset()
			}
			s.unlock()
			return
		}
		s.unlock()
	}
}

type cbuf struct {
	e []func()
	r int
	w int
}

func (c *cbuf) put(v func()) {
	var w int
	var ne []func()
	if len(c.e) == 0 {
		const minLen = 4
		ne = make([]func(), minLen)
	} else if next(c.w+1, len(c.e)) == c.r {
		ne = make([]func(), len(c.e)*2)
		if c.r < c.w {
			w = copy(ne, c.e[c.r:c.w])
		} else if c.r > c.w {
			w = copy(ne, c.e[c.r:])
			w += copy(ne[w:], c.e[:c.w])
		}
	} else {
		c.e[c.w] = v
		c.w = next(c.w+1, len(c.e))
		return
	}
	ne[w] = v
	c.e, c.r, c.w = ne, 0, w+1
}

func (c *cbuf) get() (func(), bool) {
	if c.r == c.w {
		return nil, false
	}

	v := c.e[c.r]
	c.r = next(c.r+1, len(c.e))
	return v, true
}

func (c *cbuf) reset() {
	*c = cbuf{}
}

func next(n, m int) int {
	if n >= m {
		return n - m
	}
	return n
}

var tsx = cpuid.HasExtendedFeature(cpuid.RTM)

func (s *WorkerPool) lock() {
	if runtime.GOARCH == "amd64" && tsx {
		if rtm.TxBegin() == rtm.TxBeginStarted {
			return
		}
	}
	s.m.Lock()
}

func (s *WorkerPool) unlock() {
	if runtime.GOARCH == "amd64" && tsx && rtm.TxTest() == 1 {
		rtm.TxEnd()
		return
	}
	s.m.Unlock()
}
