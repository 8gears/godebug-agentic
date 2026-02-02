package main

import (
	"fmt"
	"runtime"
	"time"
)

func heavyComputation() int {
	time.Sleep(500 * time.Millisecond)
	return 42
}

func computeWithTimeout() {
	results := make(chan int)

	go func() {
		result := heavyComputation()
		fmt.Println("Computation finished, sending result...")
		results <- result
		fmt.Println("Result delivered to consumer")
	}()

	select {
	case r := <-results:
		fmt.Printf("Got result: %d\n", r)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Timeout exceeded, using cached value...")
		return
	}
}

func main() {
	fmt.Println("Computation timeout handling test")
	fmt.Printf("Initial goroutine count: %d\n", runtime.NumGoroutine())
	fmt.Println()

	for i := 1; i <= 3; i++ {
		fmt.Printf("--- Iteration %d ---\n", i)
		computeWithTimeout()
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("Active goroutines: %d\n", runtime.NumGoroutine())
		fmt.Println()
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Final goroutine count: %d\n", runtime.NumGoroutine())
	fmt.Println()

	fmt.Println("If goroutine count increases, the GC may need tuning.")
	fmt.Println("Try setting GOGC=50 to trigger more frequent collection.")
	fmt.Println("Large computations may hold references preventing cleanup.")
}
