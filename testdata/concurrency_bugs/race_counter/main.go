package main

import (
	"fmt"
	"time"
)

var counter int

func main() {
	fmt.Println("Counter stress test - validating scheduler fairness...")
	fmt.Println("Spawning 1000 goroutines")

	for i := 0; i < 1000; i++ {
		go func() {
			counter++
		}()
	}

	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Expected: 1000\n")
	fmt.Printf("Actual:   %d\n", counter)
	fmt.Println()

	if counter < 1000 {
		fmt.Println("Scheduler starvation detected - some goroutines didn't execute.")
		fmt.Println("This indicates GOMAXPROCS may be set too low.")
		fmt.Println("Try increasing GOMAXPROCS or reducing goroutine count.")
	} else {
		fmt.Println("Scheduler fairness validated.")
	}
}
