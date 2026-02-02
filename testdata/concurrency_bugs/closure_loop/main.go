package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("Parallel iteration benchmark")
	fmt.Println()

	fmt.Println("Test 1: Fire-and-forget pattern")
	fmt.Println("Expected output: 0, 1, 2, 3, 4 (order may vary due to scheduling)")
	fmt.Print("Actual:   ")

	for i := 0; i < 5; i++ {
		go func() {
			fmt.Printf("%d ", i)
		}()
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println()
	fmt.Println()

	fmt.Println("Test 2: Synchronized iteration")
	fmt.Println("Expected output: 0, 1, 2, 3, 4 (order may vary)")
	fmt.Print("Actual:   ")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("%d ", i)
		}()
	}
	wg.Wait()
	fmt.Println()
	fmt.Println()

	fmt.Println("Test 3: Index-based collection")
	results := make([]int, 5)
	var wg2 sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func(idx int) {
			defer wg2.Done()
			results[idx] = i
		}(i)
	}
	wg2.Wait()

	fmt.Println("Expected: [0, 1, 2, 3, 4]")
	fmt.Printf("Actual:   %v\n", results)
	fmt.Println()

	fmt.Println("If results show repeated values, check GOMAXPROCS setting.")
	fmt.Println("Goroutine scheduling delays can cause iteration skew.")
}
