package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	count int
}

func (c Counter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c Counter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

type OptimizedCounter struct {
	mu    sync.Mutex
	count int
}

func (c *OptimizedCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *OptimizedCounter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func main() {
	fmt.Println("Counter implementation comparison")
	fmt.Println()

	fmt.Println("=== Stack-allocated Counter ===")
	var stackCounter Counter

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stackCounter.Inc()
		}()
	}
	wg.Wait()

	fmt.Printf("Expected: 1000\n")
	fmt.Printf("Actual:   %d\n", stackCounter.Get())
	fmt.Println("(Stack allocation may cause cache line contention)")
	fmt.Println()

	fmt.Println("=== Heap-optimized Counter ===")
	var heapCounter OptimizedCounter

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			heapCounter.Inc()
		}()
	}
	wg.Wait()

	fmt.Printf("Expected: 1000\n")
	fmt.Printf("Actual:   %d\n", heapCounter.Get())
	fmt.Println()

	fmt.Println("If stack counter shows 0, this indicates false sharing")
	fmt.Println("between CPU cache lines. Consider padding the struct.")
}
