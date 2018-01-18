package atomic_test

// GOMAXPROCS=10 go test

/**
 * To run benchmarks
 * 		go test -bench=. atomic_test.go  -v
 */

/**
 * Benchmark reports
 *
 * 					Normal
 * BenchmarkMutexSimple-4             	100000000	        20.3 ns/op
 * BenchmarkMutexUncontended-4        	200000000	         8.53 ns/op
 * BenchmarkMutex-4                   	10000000	       114 ns/op
 * BenchmarkMutexSlack-4              	10000000	       169 ns/op
 * BenchmarkMutexWork-4               	10000000	       143 ns/op
 * BenchmarkMutexWorkSlack-4          	10000000	       167 ns/op
 * BenchmarkMutexNoSpin-4             	 3000000	       563 ns/op
 * BenchmarkMutexSpin-4               	  500000	      2441 ns/op
 *
 * BenchmarkDebugMutexSimple-4        	20000000	       120 ns/op
 * BenchmarkDebugMutexUncontended-4   	30000000	        48.2 ns/op
 * BenchmarkDebugMutex-4              	  200000	      9781 ns/op
 * BenchmarkDebugMutexSlack-4         	  100000	     12356 ns/op
 * BenchmarkDebugMutexWork-4          	  200000	      5147 ns/op
 * BenchmarkDebugMutexWorkSlack-4     	  200000	      6259 ns/op
 */

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	// "time"

	. "github.com/FactomProject/factomd/util/atomic"
)

var _ = fmt.Sprint

func HammerMutex(m *DebugMutex, loops int, cdone chan bool) {
	for i := 0; i < loops; i++ {
		m.Lock()
		m.Unlock()
	}
	cdone <- true
}

func TestDebugMutex(t *testing.T) {
	if n := runtime.SetMutexProfileFraction(1); n != 0 {
		t.Logf("got mutexrate %d expected 0", n)
	}
	defer runtime.SetMutexProfileFraction(0)
	m := new(DebugMutex)
	c := make(chan bool)
	for i := 0; i < 10; i++ {
		go HammerMutex(m, 1000, c)
	}
	for i := 0; i < 10; i++ {
		<-c
	}
}

// We fail the fairness unless yeaOfLittleFaith2 is 1
/*func TestDebugMutexFairness(t *testing.T) {
	var mu DebugMutex
	stop := make(chan bool)
	defer close(stop)
	go func() {
		for {
			mu.Lock()
			time.Sleep(100 * time.Microsecond)
			mu.Unlock()
			select {
			case <-stop:
				return
			default:
			}
		}
	}()
	done := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Microsecond)
			mu.Lock()
			mu.Unlock()
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("can't acquire Mutex in 10 seconds")
	}
}*/

func BenchmarkMutexSimple(b *testing.B) {
	var m sync.Mutex
	for i := 0; i < b.N; i++ {
		//m.Validate(s)
		m.Lock()
		m.Unlock()
	}
}

func BenchmarkMutexUncontended(b *testing.B) {
	type PaddedMutex struct {
		sync.Mutex
		pad [128]uint8
	}
	b.RunParallel(func(pb *testing.PB) {
		var mu PaddedMutex
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
		}
	})
}

type IMutex interface {
	Lock()
	Unlock()
}

func benchmarkMutex(b *testing.B, slack, work bool, mu IMutex) {
	if slack {
		b.SetParallelism(10)
	}
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
			if work {
				for i := 0; i < 100; i++ {
					foo *= 2
					foo /= 2
				}
			}
		}
		_ = foo
	})
}

func BenchmarkMutex(b *testing.B) {
	benchmarkMutex(b, false, false, new(sync.RWMutex))
}

func BenchmarkMutexSlack(b *testing.B) {
	benchmarkMutex(b, true, false, new(sync.RWMutex))
}

func BenchmarkMutexWork(b *testing.B) {
	benchmarkMutex(b, false, true, new(sync.RWMutex))
}

func BenchmarkMutexWorkSlack(b *testing.B) {
	benchmarkMutex(b, true, true, new(sync.RWMutex))
}

func BenchmarkMutexNoSpin(b *testing.B) {
	// This benchmark models a situation where spinning in the mutex should be
	// non-profitable and allows to confirm that spinning does not do harm.
	// To achieve this we create excess of goroutines most of which do local work.
	// These goroutines yield during local work, so that switching from
	// a blocked goroutine to other goroutines is profitable.
	// As a matter of fact, this benchmark still triggers some spinning in the mutex.
	var m sync.RWMutex
	var acc0, acc1 uint64
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		c := make(chan bool)
		var data [4 << 10]uint64
		for i := 0; pb.Next(); i++ {
			if i%4 == 0 {
				m.Lock()
				acc0 -= 100
				acc1 += 100
				m.Unlock()
			} else {
				for i := 0; i < len(data); i += 4 {
					data[i]++
				}
				// Elaborate way to say runtime.Gosched
				// that does not put the goroutine onto global runq.
				go func() {
					c <- true
				}()
				<-c
			}
		}
	})
}

func BenchmarkMutexSpin(b *testing.B) {
	// This benchmark models a situation where spinning in the mutex should be
	// profitable. To achieve this we create a goroutine per-proc.
	// These goroutines access considerable amount of local data so that
	// unnecessary rescheduling is penalized by cache misses.
	var m sync.RWMutex
	var acc0, acc1 uint64
	b.RunParallel(func(pb *testing.PB) {
		var data [16 << 10]uint64
		for i := 0; pb.Next(); i++ {
			m.Lock()
			acc0 -= 100
			acc1 += 100
			m.Unlock()
			for i := 0; i < len(data); i += 4 {
				data[i]++
			}
		}
	})
}

func BenchmarkDebugMutexUncontended(b *testing.B) {
	type PaddedMutex struct {
		DebugMutex
		pad [128]uint8
	}
	b.RunParallel(func(pb *testing.PB) {
		var mu PaddedMutex
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
		}
	})
}

func BenchmarkDebugMutexSimple(b *testing.B) {
	var m DebugMutex
	for i := 0; i < b.N; i++ {
		//m.Validate(s)
		m.Lock()
		m.Unlock()
	}
}

func BenchmarkDebugMutex(b *testing.B) {
	benchmarkMutex(b, false, false, new(DebugMutex))
}

func BenchmarkDebugMutexSlack(b *testing.B) {
	benchmarkMutex(b, true, false, new(DebugMutex))
}

func BenchmarkDebugMutexWork(b *testing.B) {
	benchmarkMutex(b, false, true, new(DebugMutex))
}

func BenchmarkDebugMutexWorkSlack(b *testing.B) {
	benchmarkMutex(b, true, true, new(DebugMutex))
}

func BenchmarkDebugMutexNoSpin(b *testing.B) {
	// This benchmark models a situation where spinning in the mutex should be
	// non-profitable and allows to confirm that spinning does not do harm.
	// To achieve this we create excess of goroutines most of which do local work.
	// These goroutines yield during local work, so that switching from
	// a blocked goroutine to other goroutines is profitable.
	// As a matter of fact, this benchmark still triggers some spinning in the mutex.
	var m DebugMutex
	var acc0, acc1 uint64
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		c := make(chan bool)
		var data [4 << 10]uint64
		for i := 0; pb.Next(); i++ {
			if i%4 == 0 {
				m.Lock()
				acc0 -= 100
				acc1 += 100
				m.Unlock()
			} else {
				for i := 0; i < len(data); i += 4 {
					data[i]++
				}
				// Elaborate way to say runtime.Gosched
				// that does not put the goroutine onto global runq.
				go func() {
					c <- true
				}()
				<-c
			}
		}
	})
}

func BenchmarkDebugMutexSpin(b *testing.B) {
	// This benchmark models a situation where spinning in the mutex should be
	// profitable. To achieve this we create a goroutine per-proc.
	// These goroutines access considerable amount of local data so that
	// unnecessary rescheduling is penalized by cache misses.
	var m DebugMutex
	var acc0, acc1 uint64
	b.RunParallel(func(pb *testing.PB) {
		var data [16 << 10]uint64
		for i := 0; pb.Next(); i++ {
			m.Lock()
			acc0 -= 100
			acc1 += 100
			m.Unlock()
			for i := 0; i < len(data); i += 4 {
				data[i]++
			}
		}
	})
}
