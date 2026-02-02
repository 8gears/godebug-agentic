// Package main demonstrates a classic deadlock from circular lock ordering.
//
// BUG: Two goroutines acquire two mutexes in opposite order:
// - Goroutine A: lock1 -> lock2
// - Goroutine B: lock2 -> lock1
// This creates a circular wait condition where each goroutine holds one lock
// and waits for the other, causing a deadlock.
//
// DEBUGGER TEST:
// - Run program: it will hang
// - Use delve to inspect goroutine stack traces
// - Observe both goroutines waiting on Lock() calls
// - Notice they hold locks the other needs
package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	lock1 sync.Mutex
	lock2 sync.Mutex
)

func goroutineA() {
	fmt.Println("Goroutine A: acquiring lock1...")
	lock1.Lock()
	fmt.Println("Goroutine A: acquired lock1")

	// Small delay to ensure goroutine B acquires lock2
	time.Sleep(10 * time.Millisecond)

	fmt.Println("Goroutine A: acquiring lock2...")
	lock2.Lock() // DEADLOCK: waiting for lock2, which B holds
	fmt.Println("Goroutine A: acquired lock2")

	lock2.Unlock()
	lock1.Unlock()
	fmt.Println("Goroutine A: done")
}

func goroutineB() {
	fmt.Println("Goroutine B: acquiring lock2...")
	lock2.Lock()
	fmt.Println("Goroutine B: acquired lock2")

	// Small delay to ensure goroutine A acquires lock1
	time.Sleep(10 * time.Millisecond)

	fmt.Println("Goroutine B: acquiring lock1...")
	lock1.Lock() // DEADLOCK: waiting for lock1, which A holds
	fmt.Println("Goroutine B: acquired lock1")

	lock1.Unlock()
	lock2.Unlock()
	fmt.Println("Goroutine B: done")
}

func main() {
	fmt.Println("Starting deadlock demo...")
	fmt.Println("Two goroutines will acquire locks in opposite order")
	fmt.Println()

	go goroutineA()
	go goroutineB()

	// Wait long enough to observe the deadlock
	time.Sleep(5 * time.Second)

	// This line should never be reached due to deadlock
	fmt.Println("Program completed (you should not see this)")
}
