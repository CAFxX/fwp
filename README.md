# Workerpool

A simple, fast bounded worker pool with unlimited work queue.

## Usage

```go
p := fast.WorkerPool{Max: 1000}

p.Go(func() {
    // ...
})
```

If you need to wait for completion:

```go
p := fast.WorkerPool{Max: 1000}
var wg sync.WaitGroup

wg.Add(1)
p.Go(func() {
    // ...
    wg.Done()
})

wg.Wait()
```

## Performance

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

## License

MIT
