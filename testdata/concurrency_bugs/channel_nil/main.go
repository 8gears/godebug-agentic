// Package main demonstrates bugs with nil channel operations.
//
// BUG: A nil channel has special behavior:
// - Sending to nil blocks forever
// - Receiving from nil blocks forever
// - Closing nil panics
//
// This often happens when channels are conditionally initialized
// or when struct fields are not properly set up.
//
// DEBUGGER TEST:
// - Run program and observe it hangs
// - Use delve to inspect the channel - it shows nil
// - Goroutines are blocked on channel operations that will never complete
package main

import (
	"fmt"
	"time"
)

func demonstrateSendToNil() {
	fmt.Println("--- Send to nil channel ---")
	var ch chan int // nil channel, not initialized

	fmt.Printf("Channel value: %v (nil: %t)\n", ch, ch == nil)
	fmt.Println("Attempting to send to nil channel...")
	fmt.Println("(This will block forever)")

	go func() {
		ch <- 42 // BUG: blocks forever on nil channel
		fmt.Println("Send completed (you won't see this)")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Goroutine is blocked - send to nil never completes")
	fmt.Println()
}

func demonstrateReceiveFromNil() {
	fmt.Println("--- Receive from nil channel ---")
	var ch chan int // nil channel

	fmt.Printf("Channel value: %v (nil: %t)\n", ch, ch == nil)
	fmt.Println("Attempting to receive from nil channel...")
	fmt.Println("(This will block forever)")

	go func() {
		val := <-ch // BUG: blocks forever on nil channel
		fmt.Printf("Received: %d (you won't see this)\n", val)
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Goroutine is blocked - receive from nil never completes")
	fmt.Println()
}

func demonstrateCloseNil() {
	fmt.Println("--- Close nil channel ---")
	var ch chan int // nil channel

	fmt.Printf("Channel value: %v (nil: %t)\n", ch, ch == nil)
	fmt.Println("Attempting to close nil channel...")
	fmt.Println("(This will panic)")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()

	close(ch) // BUG: panics on nil channel
	fmt.Println("Close completed (you won't see this)")
}

// RealWorldExample shows how nil channels can sneak into code
type Worker struct {
	tasks chan string // might be nil if not initialized
}

func (w *Worker) ProcessTask(task string) {
	// BUG: If w.tasks was never initialized, this blocks forever
	w.tasks <- task
}

func demonstrateRealWorld() {
	fmt.Println()
	fmt.Println("--- Real-world example: uninitialized struct field ---")

	worker := &Worker{} // tasks channel is nil!
	fmt.Printf("Worker tasks channel: %v (nil: %t)\n", worker.tasks, worker.tasks == nil)

	fmt.Println("Attempting to process task...")

	done := make(chan bool)
	go func() {
		worker.ProcessTask("important task") // BUG: blocks forever
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Task processed")
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Timeout! Worker is blocked on nil channel")
	}
}

func main() {
	fmt.Println("Starting nil channel operations demo...")
	fmt.Println()

	demonstrateSendToNil()
	demonstrateReceiveFromNil()
	demonstrateCloseNil()
	demonstrateRealWorld()

	fmt.Println()
	fmt.Println("Summary of nil channel behavior:")
	fmt.Println("  - send to nil:    blocks forever")
	fmt.Println("  - receive from nil: blocks forever")
	fmt.Println("  - close nil:      panics")
	fmt.Println()
	fmt.Println("Use delve to inspect blocked goroutines and see nil channel values.")
}
