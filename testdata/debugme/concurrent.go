package main

import (
	"fmt"
	"sync"
)

func runWorkerPool(workers int, tasks []string) {
	var wg sync.WaitGroup
	taskChan := make(chan string, len(tasks))

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(i, taskChan, &wg) // Debug goroutines
	}

	// Send tasks
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	wg.Wait()
}

func worker(id int, tasks <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		processTask(id, task) // Breakpoint in goroutine
	}
}

func processTask(workerID int, task string) {
	fmt.Printf("Worker %d processing: %s\n", workerID, task)
}
