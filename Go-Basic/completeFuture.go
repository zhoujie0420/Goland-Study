package main

import "sync"

func foo1(wg *sync.WaitGroup, result chan<- int) {
	defer wg.Done()
	println("foo1")
	result <- 1
}

func foo2(wg *sync.WaitGroup, result chan<- int) {
	defer wg.Done()
	println("foo2")
	result <- 2
}

func main() {
	var wg sync.WaitGroup
	result := make(chan int, 2)
	wg.Add(2)
	go foo1(&wg, result)
	go foo2(&wg, result)
	wg.Wait()
	close(result)
	for i := range result {
		println(i)
	}
	println("main")
}
