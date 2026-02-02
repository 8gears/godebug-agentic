// Package main demonstrates a race condition with WaitGroup.Add().
//
// BUG: Calling wg.Add(1) inside the goroutine races with wg.Wait().
// The Wait() call might execute before all Add() calls complete,
// causing it to return early while goroutines are still running.
//
// The correct pattern is to call Add() BEFORE starting the goroutine.
//
// DEBUGGER TEST:
// - Run with race detector: go run -race .
// - The race detector will catch the Add/Wait race
// - Program may complete before all work is done
// - Observe inconsistent "completed" counts
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func buggyWaitGroup() {
	fmt.Println("=== Buggy WaitGroup (Add inside goroutine) ===")

	var wg sync.WaitGroup
	var completed int32

	for i := 0; i < 10; i++ {
		go func(id int) {
			wg.Add(1) // BUG: races with Wait()!
			defer wg.Done()

			// Simulate work
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			fmt.Printf("Worker %d finished\n", id)
		}(i)
	}

	// BUG: Wait() might return before all Add() calls happen
	wg.Wait()

	fmt.Printf("Wait() returned. Completed: %d/10\n", atomic.LoadInt32(&completed))
	if atomic.LoadInt32(&completed) < 10 {
		fmt.Println("BUG: Some workers didn't finish before Wait() returned!")
	}
	fmt.Println()
}

func correctWaitGroup() {
	fmt.Println("=== Correct WaitGroup (Add before goroutine) ===")

	var wg sync.WaitGroup
	var completed int32

	for i := 0; i < 10; i++ {
		wg.Add(1) // CORRECT: Add before starting goroutine
		go func(id int) {
			defer wg.Done()

			// Simulate work
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			fmt.Printf("Worker %d finished\n", id)
		}(i)
	}

	wg.Wait()

	fmt.Printf("Wait() returned. Completed: %d/10\n", atomic.LoadInt32(&completed))
	fmt.Println()
}

func anotherBuggyPattern() {
	fmt.Println("=== Another Bug: Add(n) then spawn less than n ===")

	var wg sync.WaitGroup
	wg.Add(5) // Expect 5 goroutines

	for i := 0; i < 3; i++ { // BUG: Only spawn 3!
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Worker %d done\n", id)
		}(i)
	}

	// This will hang forever - waiting for 2 Done() calls that never come
	fmt.Println("Waiting... (this will hang because Add(5) but only 3 Done() calls)")

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Completed")
	case <-time.After(200 * time.Millisecond):
		fmt.Println("Timeout! WaitGroup counter never reached zero")
	}
	fmt.Println()
}

func main() {
	fmt.Println("Starting WaitGroup race demo...")
	fmt.Println()

	buggyWaitGroup()

	// Small delay to let any stragglers finish
	time.Sleep(50 * time.Millisecond)
	fmt.Println("---")
	fmt.Println()

	correctWaitGroup()

	fmt.Println("---")
	fmt.Println()

	anotherBuggyPattern()

	fmt.Println("Run with: go run -race . to detect the race condition")
	fmt.Println("The race detector will report: 'race on sync.WaitGroup'")
}
