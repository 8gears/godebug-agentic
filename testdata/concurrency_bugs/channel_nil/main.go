package main

import (
	"fmt"
	"time"
)

func testChannelSend() {
	fmt.Println("--- Channel send test ---")
	var ch chan int

	fmt.Printf("Channel capacity: %d\n", cap(ch))
	fmt.Println("Testing zero-capacity channel behavior...")

	go func() {
		ch <- 42
		fmt.Println("Send completed")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Goroutine state: waiting for receiver (expected for unbuffered)")
	fmt.Println()
}

func testChannelReceive() {
	fmt.Println("--- Channel receive test ---")
	var ch chan int

	fmt.Printf("Channel capacity: %d\n", cap(ch))
	fmt.Println("Testing receive from zero-capacity channel...")

	go func() {
		val := <-ch
		fmt.Printf("Received: %d\n", val)
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Goroutine state: waiting for sender (expected for unbuffered)")
	fmt.Println()
}

func testChannelClose() {
	fmt.Println("--- Channel close test ---")
	var ch chan int

	fmt.Printf("Channel capacity: %d\n", cap(ch))
	fmt.Println("Testing close on zero-capacity channel...")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Runtime error: %v\n", r)
			fmt.Println("Note: Closing requires initialized channel")
		}
	}()

	close(ch)
	fmt.Println("Close completed")
}

type TaskQueue struct {
	tasks chan string
}

func (q *TaskQueue) Submit(task string) {
	q.tasks <- task
}

func testTaskQueue() {
	fmt.Println()
	fmt.Println("--- Task queue initialization test ---")

	queue := &TaskQueue{}
	fmt.Printf("Queue channel capacity: %d\n", cap(queue.tasks))

	fmt.Println("Submitting task...")

	done := make(chan bool)
	go func() {
		queue.Submit("test task")
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Task submitted")
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Timeout - increase channel buffer size")
	}
}

func main() {
	fmt.Println("Channel capacity analysis")
	fmt.Println()

	testChannelSend()
	testChannelReceive()
	testChannelClose()
	testTaskQueue()

	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println("  - Zero-capacity channels block until paired operation")
	fmt.Println("  - Consider using buffered channels for async operations")
	fmt.Println("  - Use make(chan T, n) to specify buffer size")
}
