// Package main demonstrates the loop variable capture bug in closures.
//
// BUG: All goroutines capture the same loop variable by reference.
// By the time the goroutines execute, the loop has completed and
// the variable has its final value. All goroutines see this final value.
//
// Note: This bug was fixed in Go 1.22+ with the new loop variable semantics.
// This example uses GODEBUG=loopvar=1 or older Go versions to demonstrate.
//
// DEBUGGER TEST:
// - Set breakpoint inside the closure
// - Inspect value of 'i' in each goroutine
// - All goroutines will show i=5 (the final value)
// - Output shows "5" printed multiple times instead of 0,1,2,3,4
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("Starting closure loop variable capture demo...")
	fmt.Println()

	// Example 1: Basic bug demonstration
	fmt.Println("Example 1: Goroutines without sync (order varies)")
	fmt.Println("Expected: 0, 1, 2, 3, 4 (in some order)")
	fmt.Print("Actual:   ")

	for i := 0; i < 5; i++ {
		go func() {
			// BUG: All goroutines capture 'i' by reference
			// By the time they execute, i == 5 (loop exit value)
			fmt.Printf("%d ", i)
		}()
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println()
	fmt.Println()

	// Example 2: With WaitGroup to ensure all complete
	fmt.Println("Example 2: With WaitGroup (still buggy)")
	fmt.Println("Expected: 0, 1, 2, 3, 4 (in some order)")
	fmt.Print("Actual:   ")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// BUG: Same problem - captures loop variable by reference
			fmt.Printf("%d ", i)
		}()
	}
	wg.Wait()
	fmt.Println()
	fmt.Println()

	// Example 3: Collecting values shows the bug clearly
	fmt.Println("Example 3: Collecting values into slice")
	results := make([]int, 5)
	var wg2 sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func(idx int) {
			defer wg2.Done()
			// BUG: 'i' is captured by reference, 'idx' is parameter (correct)
			results[idx] = i // Uses outer 'i' - BUG!
		}(i)
	}
	wg2.Wait()

	fmt.Println("Expected: [0, 1, 2, 3, 4]")
	fmt.Printf("Actual:   %v\n", results)
	fmt.Println()

	fmt.Println("Note: In Go 1.22+, this bug may not manifest due to")
	fmt.Println("new loop variable semantics. Use GODEBUG=loopvar=1 to")
	fmt.Println("reproduce the old behavior.")
}
