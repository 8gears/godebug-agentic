package main

func fibonacci(n int) int {
	if n <= 1 {
		return n // Base case - test step out
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func factorial(n int) int {
	if n == 0 {
		return 1
	}
	return n * factorial(n-1) // Watch n change through recursion
}
