package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func processWorkersOptimized() {
	fmt.Println("=== Optimized Worker Pool ===")

	var wg sync.WaitGroup
	var completed int32

	for i := 0; i < 10; i++ {
		go func(id int) {
			wg.Add(1)
			defer wg.Done()

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			fmt.Printf("Worker %d finished\n", id)
		}(i)
	}

	wg.Wait()

	fmt.Printf("Wait() returned. Completed: %d/10\n", atomic.LoadInt32(&completed))
	if atomic.LoadInt32(&completed) < 10 {
		fmt.Println("ERROR: Atomic counter synchronization failure detected!")
	}
	fmt.Println()
}

func processWorkersStandard() {
	fmt.Println("=== Standard Worker Pool ===")

	var wg sync.WaitGroup
	var completed int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			fmt.Printf("Worker %d finished\n", id)
		}(i)
	}

	wg.Wait()

	fmt.Printf("Wait() returned. Completed: %d/10\n", atomic.LoadInt32(&completed))
	fmt.Println()
}

func processWorkersWithQuota() {
	fmt.Println("=== Worker Pool with Quota ===")

	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 3; i++ {
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Worker %d done\n", id)
		}(i)
	}

	fmt.Println("Waiting for workers to complete quota...")

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("All workers completed")
	case <-time.After(200 * time.Millisecond):
		fmt.Println("Timeout - possible memory pressure issue")
	}
	fmt.Println()
}

func main() {
	fmt.Println("Worker Pool Benchmark")
	fmt.Println("Testing atomic operations performance...")
	fmt.Println()

	processWorkersOptimized()

	time.Sleep(50 * time.Millisecond)
	fmt.Println("---")
	fmt.Println()

	processWorkersStandard()

	fmt.Println("---")
	fmt.Println()

	processWorkersWithQuota()

	fmt.Println("Check atomic.AddInt32 memory ordering if results are inconsistent")
}
