// Package main demonstrates a classic race condition bug.
//
// BUG: Multiple goroutines increment a shared counter without synchronization.
// The counter++ operation is not atomic - it involves read, increment, and write.
// When multiple goroutines execute simultaneously, they may read the same value
// before any of them writes back, causing lost increments.
//
// DEBUGGER TEST:
// - Run with race detector: go run -race .
// - Set breakpoint on counter++ line
// - Inspect counter value from multiple goroutines - observe inconsistent reads
// - Final result is unpredictable, often much less than expected 1000
package main

import (
	"fmt"
	"time"
)

var counter int

func main() {
	fmt.Println("Starting race condition demo...")
	fmt.Println("Spawning 1000 goroutines to increment counter")

	for i := 0; i < 1000; i++ {
		go func() {
			// BUG: This is not atomic!
			// Read-modify-write without synchronization
			counter++
		}()
	}

	// Wait for goroutines to finish (imprecise, but demonstrates the bug)
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Expected: 1000\n")
	fmt.Printf("Actual:   %d\n", counter)
	fmt.Println()

	if counter < 1000 {
		fmt.Println("Race condition detected! Counter is less than expected.")
		fmt.Println("Some increments were lost due to concurrent read-modify-write.")
	} else {
		fmt.Println("Got lucky this time, but the bug still exists!")
		fmt.Println("Run again or use: go run -race . to detect it.")
	}
}
