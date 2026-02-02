// Package main demonstrates a timer/memory leak from time.After in a loop.
//
// BUG: Using time.After() inside a select loop creates a new timer each iteration.
// These timers are not garbage collected until they fire, even if the select
// chooses a different case. In a busy loop, this causes unbounded timer growth.
//
// DEBUGGER TEST:
// - Run program and observe memory/timer growth
// - Use delve to inspect runtime timer count
// - Memory usage grows even though work is being processed
package main

import (
	"fmt"
	"runtime"
	"time"
)

func leakyLoop(messages <-chan string, done <-chan bool) {
	for {
		select {
		case msg := <-messages:
			// Process message quickly
			_ = msg
		case <-time.After(1 * time.Second):
			// BUG: Creates a NEW timer each iteration!
			// If messages arrive frequently, these timers accumulate
			// and won't be GC'd until they fire (1 second later)
			fmt.Println("Timeout - no messages")
		case <-done:
			return
		}
	}
}

func correctLoop(messages <-chan string, done <-chan bool) {
	// CORRECT: Create timer once, reset it each iteration
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case msg := <-messages:
			_ = msg
			// Reset timer after receiving message
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(1 * time.Second)
		case <-timer.C:
			fmt.Println("Timeout - no messages")
			timer.Reset(1 * time.Second)
		case <-done:
			return
		}
	}
}

func countTimers() int {
	// Force GC to get accurate count of non-collectable timers
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// HeapObjects gives us a rough proxy for leaked resources
	return int(m.HeapObjects)
}

func main() {
	fmt.Println("Starting time.After leak demo...")
	fmt.Println()

	// Demonstrate the leaky pattern
	fmt.Println("=== Leaky Pattern: time.After in loop ===")

	messages := make(chan string, 100)
	done := make(chan bool)

	// Measure initial state
	initialObjects := countTimers()
	fmt.Printf("Initial heap objects: %d\n", initialObjects)

	// Start the leaky loop
	go leakyLoop(messages, done)

	// Send many messages rapidly - each iteration creates a new timer
	fmt.Println("Sending 10000 rapid messages...")
	for i := 0; i < 10000; i++ {
		messages <- fmt.Sprintf("msg-%d", i)
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	afterObjects := countTimers()
	fmt.Printf("Heap objects after rapid messages: %d\n", afterObjects)
	fmt.Printf("Increase: %d objects\n", afterObjects-initialObjects)

	// Stop the leaky loop
	done <- true

	// Wait for timers to fire and get collected
	fmt.Println("\nWaiting 1.5 seconds for leaked timers to fire...")
	time.Sleep(1500 * time.Millisecond)

	finalObjects := countTimers()
	fmt.Printf("Heap objects after timers fired: %d\n", finalObjects)

	fmt.Println()
	fmt.Println("=== Explanation ===")
	fmt.Println("Each iteration of the select loop with time.After() creates")
	fmt.Println("a new timer. If messages arrive faster than the timeout,")
	fmt.Println("timers accumulate because they're not GC'd until they fire.")
	fmt.Println()
	fmt.Println("In production with high message rates and long timeouts,")
	fmt.Println("this can consume significant memory.")
	fmt.Println()
	fmt.Println("FIX: Use time.NewTimer() once and Reset() it each iteration.")
	fmt.Println("See the correctLoop function in this file for the proper pattern.")
}
