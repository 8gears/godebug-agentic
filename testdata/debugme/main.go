package main

import "fmt"

func main() {
	// 1. Basic nested function calls
	result := outerFunc(10)

	// 2. Recursion test
	fib := fibonacci(10)

	// 3. Loop with state changes
	items := processItems([]int{1, 2, 3, 4, 5})

	// 4. Goroutines and channels
	runWorkerPool(3, []string{"task1", "task2", "task3"})

	// 5. Nested structs
	user := createUser("Alice", 30)
	updateUser(&user, "Bob")

	fmt.Println(result, fib, items, user)
}

func outerFunc(x int) int {
	y := middleFunc(x * 2)
	return y + 1
}

func middleFunc(x int) int {
	z := innerFunc(x + 5)
	return z * 2
}

func innerFunc(x int) int {
	return x * x // Breakpoint here to test call stack
}

func processItems(input []int) []int {
	result := make([]int, 0, len(input))
	for i, item := range input {
		processed := item * 2
		result = append(result, processed)
		_ = i // Use i to avoid unused variable warning
	}
	return result
}
