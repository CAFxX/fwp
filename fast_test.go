package fwp_test

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/CAFxX/fwp"
	"github.com/alitto/pond"
	"github.com/alphadose/itogami"
	"github.com/gammazero/workerpool"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/sync/semaphore"
)

func TestWorkerPool(t *testing.T) {
	c := make([]bool, 1<<20)
	s := fwp.WorkerPool{Max: runtime.GOMAXPROCS(0)}
	var wg sync.WaitGroup

	var fn func(int)
	fn = func(n int) {
		if n >= len(c) {
			return
		}
		if c[n] {
			t.Fatalf("duplicated call: %d", n)
		}
		c[n] = true
		wg.Add(2)
		s.Go(func() {
			fn(n * 2)
			wg.Done()
		})
		s.Go(func() {
			fn(n*2 + 1)
			wg.Done()
		})
	}
	fn(1)
	wg.Wait()

	for i, e := range c {
		if !e && i > 0 {
			t.Fatalf("missing call: %d", i)
		}
	}
}

func BenchmarkFastWorkerPool(b *testing.B) {
	s := fwp.WorkerPool{Max: runtime.GOMAXPROCS(0)}
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			// ...
			wg.Done()
		}
		for pb.Next() {
			wg.Add(1)
			s.Go(fn)
		}
		wg.Wait()
	})
}

func BenchmarkGammazeroWorkerPool(b *testing.B) {
	wp := workerpool.New(runtime.GOMAXPROCS(0))
	b.RunParallel(func(pb *testing.PB) {
		fn := func() {
			// ...
		}
		for pb.Next() {
			wp.Submit(fn)
		}
	})
	wp.StopWait()
}

func BenchmarkAlittoPond(b *testing.B) {
	p := pond.New(runtime.GOMAXPROCS(0), 1<<16)
	b.RunParallel(func(pb *testing.PB) {
		fn := func() {
			// ...
		}
		for pb.Next() {
			p.Submit(fn)
		}
	})
	p.StopAndWait()
}

func BenchmarkPanjf2000Ants(b *testing.B) {
	p, err := ants.NewPool(runtime.GOMAXPROCS(0))
	if err != nil {
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			// ...
			wg.Done()
		}
		for pb.Next() {
			wg.Add(1)
			p.Submit(fn)
		}
		wg.Wait()
	})
}

func BenchmarkAlphadoseItogami(b *testing.B) {
	b.Skip("extremely buggy/unreliable")
	p := itogami.NewPool(uint64(runtime.GOMAXPROCS(0)))
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			// ...
			wg.Done()
		}
		for pb.Next() {
			wg.Add(1)
			p.Submit(fn)
		}
		wg.Wait()
	})
}

func BenchmarkGoroutineCond(b *testing.B) {
	var m sync.Mutex
	c := sync.Cond{L: &m}
	n := runtime.GOMAXPROCS(0)
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			m.Lock()
			for n == 0 {
				c.Wait()
			}
			n--
			m.Unlock()

			// ...

			m.Lock()
			n++
			m.Unlock()
			c.Signal()

			wg.Done()
		}
		for pb.Next() {
			wg.Add(1)
			go fn()
		}
		wg.Wait()
	})
}

func BenchmarkGoroutineCondPre(b *testing.B) {
	var m sync.Mutex
	c := sync.Cond{L: &m}
	n := runtime.GOMAXPROCS(0)
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			// ...

			m.Lock()
			n++
			m.Unlock()
			c.Signal()

			wg.Done()
		}
		for pb.Next() {
			m.Lock()
			for n == 0 {
				c.Wait()
			}
			n--
			m.Unlock()

			wg.Add(1)
			go fn()
		}
		wg.Wait()
	})
}

func BenchmarkGoroutineChannelSema(b *testing.B) {
	ch := make(chan struct{}, runtime.GOMAXPROCS(0))
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			ch <- struct{}{}

			// ...

			<-ch
			wg.Done()
		}
		for pb.Next() {
			wg.Add(1)
			go fn()
		}
		wg.Wait()
	})
}

func BenchmarkGoroutineChannelSemaPre(b *testing.B) {
	ch := make(chan struct{}, runtime.GOMAXPROCS(0))
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			// ...

			<-ch
			wg.Done()
		}
		for pb.Next() {
			ch <- struct{}{}
			wg.Add(1)
			go fn()
		}
		wg.Wait()
	})
}

func BenchmarkGoroutineXSyncSemaphore(b *testing.B) {
	s := semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
	b.RunParallel(func(pb *testing.PB) {
		fn := func() {
			s.Acquire(context.Background(), 1)

			// ...

			s.Release(1)
		}
		for pb.Next() {
			go fn()
		}
	})
	s.Acquire(context.Background(), int64(runtime.GOMAXPROCS(0)))
}

func BenchmarkGoroutine(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		fn := func() {
			// ...
			wg.Done()
		}
		for pb.Next() {
			wg.Add(1)
			go fn()
		}
		wg.Wait()
	})
}
