//go:build performance_test

package bank

import (
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadOperationsPerformance(t *testing.T) {
	bankPerformance := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()

		bank := New(1000)
		b.ResetTimer()

		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				bank.TotalAmount()
			}
		})
	})

	emulatorPerformance := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()

		mx := make([]sync.RWMutex, 1000)
		b.ResetTimer()

		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {

				for i := range 1000 {
					mx[i].RLock()
				}

				v := 0
				for range 1000 {
					v++
				}

				for i := range 1000 {
					mx[i].RUnlock()
				}
			}
		})
	})

	require.LessOrEqual(t, float64(bankPerformance.NsPerOp())/float64(emulatorPerformance.NsPerOp()), 1.5)
}

func TestTransferFineGrainedLocks(t *testing.T) {
	bankPerformance := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()

		bank := New(1000)
		for i := range 1000 {
			bank.Deposit(i, DefaultMaxAmount)
		}
		b.ResetTimer()

		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				bank.Transfer(rand.N[int](1000), rand.N[int](1000), 1)
			}
		})
	})

	emulatorPerformance := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()

		mutexes := make([]sync.Mutex, 1000)

		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mx1 := rand.N[int](1000)
				mx2 := rand.N[int](1000)

				if mx1 == mx2 {
					continue
				}

				if mx1 > mx2 {
					mx1, mx2 = mx2, mx1
				}

				mutexes[mx1].Lock()
				mutexes[mx2].Lock()

				var v int
				for range 200 {
					v++
				}

				mutexes[mx1].Unlock()
				mutexes[mx2].Unlock()
			}
		})
	})

	require.LessOrEqual(t, float64(bankPerformance.NsPerOp())/float64(emulatorPerformance.NsPerOp()), 1.)
}

func TestConsolidateFineGrainedLocks(t *testing.T) {
	bankPerformance := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()

		bank := New(1000)
		for i := range 1000 {
			bank.Deposit(i, DefaultMaxAmount)
		}
		b.ResetTimer()

		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				src := []int{rand.N[int](1000)}
				bank.Consolidate(src, rand.N[int](1000))
			}
		})
	})

	emulatorPerformance := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()

		mutexes := make([]sync.Mutex, 1000)

		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mx1 := rand.N[int](1000)
				mx2 := rand.N[int](1000)

				if mx1 == mx2 {
					continue
				}

				if mx1 > mx2 {
					mx1, mx2 = mx2, mx1
				}

				mutexes[mx1].Lock()
				mutexes[mx2].Lock()

				var v int
				for range 200 {
					v++
				}

				mutexes[mx1].Unlock()
				mutexes[mx2].Unlock()
			}
		})
	})

	require.LessOrEqual(t, float64(bankPerformance.NsPerOp())/float64(emulatorPerformance.NsPerOp()), 2.)
}
