# fwp

[![GoDoc](https://godoc.org/github.com/CAFxX/fwp?status.svg)](https://godoc.org/github.com/CAFxX/fwp)

`fwp` (fast worker pool) is a simple, very fast bounded worker pool with an unlimited work queue.

When the worker pool is idle it consumes no memory (or goroutines).

## Usage

```go
// A worker pool with up to 1000 workers.
p := fwp.WorkerPool{Max: 1000}

p.Go(func() {
    // ...
})
p.Go(func() {
    // ...
})
// ...
```

If you need to wait for completion:

```go
p := fwp.WorkerPool{Max: 1000}
var wg sync.WaitGroup

wg.Add(1)
p.Go(func() {
    // ...
    wg.Done()
})

wg.Wait()
```

It is possible to submit tasks from inside other tasks:

```go
p := fwp.WorkerPool{Max: 1000}

p.Go(func() {
    p.Go(func() {
        // ...
    })
    // ...
})
```

If tasks depend on each other it is recommended, to prevent deadlocks
that may be caused by `Max` tasks becoming blocked at the same time,
to resubmit tasks (instead of blocking) in case a task is executed
before its dependencies are ready:

```go
p := fwp.WorkerPool{Max: 1000}

var fn func()
fn = func() {
    if some_precondition_is_not_yet_met {
        p.Go(fn)
        return
    }
    // ...
}
p.Go(fn)
```

## Performance

`fwp` is pretty fast. Indeed it is faster than any other workerpool tested, and for high volumes of short tasks it can even be faster than spawning goroutines without a semaphore:

```
name                       time/op
FastWorkerPool-6            242ns ± 4%
GammazeroWorkerPool-6      1.42µs ± 1%
AlittoPond-6                405ns ± 8%
Panjf2000Ants-6            1.17µs ± 1%
GoroutineCond-6            1.04µs ± 3%
GoroutineCondPre-6          907ns ± 2%
GoroutineChannelSema-6      491ns ± 3%
GoroutineChannelSemaPre-6   735ns ± 3%
GoroutineXSyncSemaphore-6  1.66µs ± 5%
Goroutine-6                 266ns ±18%
```

The performance is due to three factors:

- Goroutines are reused to process multiple tasks (this minimizes
  allocation of new goroutines as well as stack growths).
- The length of critical sections is kept as short as possible
  (this minimizes contention on the mutex that guards the internals
  of the worker pool).
- The internal behavior of the pool is adaptive to the workload,
  with 2 different regimes selected automatically based on the
  number and duration of tasks submitted.

## License

[MIT](LICENSE)
