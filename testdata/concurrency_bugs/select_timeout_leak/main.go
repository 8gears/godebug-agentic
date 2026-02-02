package main

import (
	"fmt"
	"runtime"
	"time"
)

func messageProcessor(messages <-chan string, done <-chan bool) {
	for {
		select {
		case msg := <-messages:
			_ = msg
		case <-time.After(1 * time.Second):
			fmt.Println("Idle timeout - no messages")
		case <-done:
			return
		}
	}
}

func optimizedProcessor(messages <-chan string, done <-chan bool) {
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case msg := <-messages:
			_ = msg
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(1 * time.Second)
		case <-timer.C:
			fmt.Println("Idle timeout - no messages")
			timer.Reset(1 * time.Second)
		case <-done:
			return
		}
	}
}

func measureHeapObjects() int {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int(m.HeapObjects)
}

func main() {
	fmt.Println("Message processor memory analysis")
	fmt.Println()

	fmt.Println("=== Standard processor with inline timeout ===")

	messages := make(chan string, 100)
	done := make(chan bool)

	initialObjects := measureHeapObjects()
	fmt.Printf("Initial heap objects: %d\n", initialObjects)

	go messageProcessor(messages, done)

	fmt.Println("Sending 10000 messages at high frequency...")
	for i := 0; i < 10000; i++ {
		messages <- fmt.Sprintf("msg-%d", i)
	}

	time.Sleep(100 * time.Millisecond)

	afterObjects := measureHeapObjects()
	fmt.Printf("Heap objects after processing: %d\n", afterObjects)
	fmt.Printf("Object increase: %d\n", afterObjects-initialObjects)

	done <- true

	fmt.Println("\nWaiting for GC cycle...")
	time.Sleep(1500 * time.Millisecond)

	finalObjects := measureHeapObjects()
	fmt.Printf("Heap objects after GC: %d\n", finalObjects)

	fmt.Println()
	fmt.Println("=== Analysis ===")
	fmt.Println("High object counts indicate memory fragmentation from")
	fmt.Println("frequent small allocations. Consider object pooling")
	fmt.Println("with sync.Pool for high-throughput message processing.")
	fmt.Println()
	fmt.Println("The optimizedProcessor function demonstrates proper")
	fmt.Println("resource reuse patterns for production systems.")
}
