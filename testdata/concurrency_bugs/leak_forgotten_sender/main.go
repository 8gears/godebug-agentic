// Package main demonstrates a goroutine leak from an abandoned channel sender.
//
// BUG: When a timeout fires before the slow computation completes, the sender
// goroutine remains blocked forever on the unbuffered channel send.
// The receiver has moved on, so nobody will ever receive the value.
// The goroutine and its resources are leaked.
//
// DEBUGGER TEST:
// - Run program and observe timeout message
// - Use delve to list goroutines after timeout
// - Notice leaked goroutine blocked on channel send (results <- ...)
// - The goroutine count never goes back to 1
package main

import (
	"fmt"
	"runtime"
	"time"
)

func slowComputation() int {
	// Simulate slow work that takes longer than the timeout
	time.Sleep(500 * time.Millisecond)
	return 42
}

func leakyOperation() {
	results := make(chan int) // BUG: unbuffered channel

	go func() {
		result := slowComputation()
		fmt.Println("Computation finished, trying to send result...")
		results <- result // BUG: blocks forever if timeout fires first
		fmt.Println("Result sent (you won't see this after timeout)")
	}()

	select {
	case r := <-results:
		fmt.Printf("Got result: %d\n", r)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Timeout! Abandoning slow operation...")
		// BUG: The sender goroutine is now leaked!
		return
	}
}

func main() {
	fmt.Println("Starting goroutine leak demo...")
	fmt.Printf("Initial goroutine count: %d\n", runtime.NumGoroutine())
	fmt.Println()

	// Run the leaky operation multiple times
	for i := 1; i <= 3; i++ {
		fmt.Printf("--- Iteration %d ---\n", i)
		leakyOperation()
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("Goroutine count after iteration %d: %d\n", i, runtime.NumGoroutine())
		fmt.Println()
	}

	// Wait a bit and check goroutine count
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Goroutine count after waiting: %d\n", runtime.NumGoroutine())
	fmt.Println()

	fmt.Println("Notice: Goroutine count keeps increasing!")
	fmt.Println("Each timeout leaks one goroutine blocked on channel send.")
	fmt.Println()
	fmt.Println("Use delve to inspect the leaked goroutines:")
	fmt.Println("  dlv debug . -- then type 'goroutines' to list them")
}
