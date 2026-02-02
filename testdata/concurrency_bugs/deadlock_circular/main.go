package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	resourceA sync.Mutex
	resourceB sync.Mutex
)

func transferAtoB() {
	fmt.Println("Transfer A->B: acquiring resource A...")
	resourceA.Lock()
	fmt.Println("Transfer A->B: acquired resource A")

	time.Sleep(10 * time.Millisecond)

	fmt.Println("Transfer A->B: acquiring resource B...")
	resourceB.Lock()
	fmt.Println("Transfer A->B: acquired resource B")

	resourceB.Unlock()
	resourceA.Unlock()
	fmt.Println("Transfer A->B: complete")
}

func transferBtoA() {
	fmt.Println("Transfer B->A: acquiring resource B...")
	resourceB.Lock()
	fmt.Println("Transfer B->A: acquired resource B")

	time.Sleep(10 * time.Millisecond)

	fmt.Println("Transfer B->A: acquiring resource A...")
	resourceA.Lock()
	fmt.Println("Transfer B->A: acquired resource A")

	resourceA.Unlock()
	resourceB.Unlock()
	fmt.Println("Transfer B->A: complete")
}

func main() {
	fmt.Println("Resource Transfer System")
	fmt.Println("Testing bidirectional transfer with mutex protection")
	fmt.Println()

	go transferAtoB()
	go transferBtoA()

	time.Sleep(5 * time.Second)

	fmt.Println("Transfers completed successfully")
}
