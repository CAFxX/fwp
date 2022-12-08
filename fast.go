package fwp

// WorkerPool is a worker pool with bounded workers (up to Max tasks
// can be running concurrently) and unbounded queue (no limit to the
// number of tasks waiting for a worker to become available). This is
// similar to how the native `go` construct operates, but with an
// (optional) bound to the number of tasks executed concurrently.
type WorkerPool struct {
	// Max is the maximum number of tasks being executed concurrently.
	// A value of 0 means no limit, i.e. the WorkerPool behaves exactly
	// like the native `go` construct.
	Max int
	m   mutex
	n   int
	q   cbuf
}

// Go submits a task for asynchronous execution by the worker
// pool. It is similar to the native `go` construct.
//
// The task will be processed by one the pool workers as
// soon as one becomes available. WorkerPool (similarly to Go
// goroutines) provides no guarantees about the order in which
// tasks are executed (if you need such guarantees, use external
// synchronization mechanisms, but taking care to not cause
// deadlocks; you can do this by resubmitting a task to be
// executed later in case some resources can not be acquired).
//
// To wait for one or more tasks to complete use an explicit
// synchronization mechanism such as channels, sync.WaitGroup,
// or similar.
func (s *WorkerPool) Go(fn func()) {
	if s.Max <= 0 {
		go fn()
		return
	}
	s.m.Lock()
	if s.n >= s.Max {
		s.q.put(fn)
		s.m.Unlock()
		return
	}
	s.n++
	s.m.Unlock()
	go s.worker(fn)
}

func (s *WorkerPool) worker(fn func()) {
	for {
		fn()

		s.m.Lock()
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
			s.m.Unlock()
			return
		}
		s.m.Unlock()
	}
}

// Stats returns statistics about the worker pool.
func (s *WorkerPool) Stats() Stats {
	s.m.Lock()
	r := Stats{
		Running: s.n,
		Queued:  s.q.len(),
	}
	s.m.Unlock()
	return r
}

// Stats contains statistics about the worker pool.
type Stats struct {
	// Running is the number of tasks currently being run.
	// It is never greater than the number of Max workers.
	Running int
	// Queued is the number of tasks currently queued, waiting
	// for a worker to become available for processing.
	// This number is only bound by the amount of memory available.
	//
	// The total number of tasks in the worker pool is therefore
	// Queued+Running.
	Queued int
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

func (c *cbuf) len() int {
	if c.w >= c.r {
		return c.w - c.r
	}
	return len(c.e) - c.r + c.w
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
