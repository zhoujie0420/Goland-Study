package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func worker(ch <-chan bool) { // 单向通道 只能接收 ch chan <- bool 只能发送
	defer wg.Done()
	for {
		select {
		case <-ch:
			break
		default:
			fmt.Println("worker")
			time.Sleep(1 * time.Second)
		}

	}
}

func main() {
	var exitChan = make(chan bool, 1)

	wg.Add(1)
	go worker(exitChan)
	wg.Wait()
	fmt.Println("main")
}
