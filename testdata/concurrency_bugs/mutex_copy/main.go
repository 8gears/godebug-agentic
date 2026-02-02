// Package main demonstrates the mutex copy bug from value receivers.
//
// BUG: When a method has a value receiver, the struct is copied on each call.
// This copies the mutex too, creating independent mutexes that don't protect
// the shared state. Each method call operates on its own copy.
//
// DEBUGGER TEST:
// - Run with race detector: go run -race .
// - Set breakpoint in Inc() method
// - Inspect address of c.mu across different calls - they differ!
// - The original struct's mutex is never locked
package main

import (
	"fmt"
	"sync"
)

// Counter has a mutex for protecting the count field
type Counter struct {
	mu    sync.Mutex
	count int
}

// Inc has a VALUE receiver - BUG!
// Each call copies the entire Counter struct, including the mutex.
// The copy's mutex is locked, not the original's.
func (c Counter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
	// The incremented c.count is discarded when this method returns
	// because we modified a copy, not the original
}

// Get also has a value receiver for consistency in demonstrating the bug
func (c Counter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// CorrectCounter shows the proper way with pointer receivers
type CorrectCounter struct {
	mu    sync.Mutex
	count int
}

func (c *CorrectCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *CorrectCounter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func main() {
	fmt.Println("Starting mutex copy bug demo...")
	fmt.Println()

	// Demonstrate the buggy counter
	fmt.Println("=== Buggy Counter (value receiver) ===")
	var buggy Counter

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buggy.Inc() // Each call copies the struct!
		}()
	}
	wg.Wait()

	fmt.Printf("Expected: 1000\n")
	fmt.Printf("Actual:   %d\n", buggy.Get())
	fmt.Println("(Value is always 0 because Inc() modifies copies)")
	fmt.Println()

	// Demonstrate the correct counter
	fmt.Println("=== Correct Counter (pointer receiver) ===")
	var correct CorrectCounter

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			correct.Inc()
		}()
	}
	wg.Wait()

	fmt.Printf("Expected: 1000\n")
	fmt.Printf("Actual:   %d\n", correct.Get())
	fmt.Println()

	fmt.Println("Run with: go run -race . to see race detector warnings")
	fmt.Println("The race detector catches the mutex copy issue.")
}
