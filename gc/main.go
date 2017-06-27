package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	bufferSize = 4096
	windowSize = 200000
	msgCount   = 1000000
)

type (
	message []byte
	buffer  [windowSize]message
)

var worst time.Duration
var sleepTime = 100 * time.Millisecond

func memoryThrasher(b *buffer, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < msgCount; i++ {
		m := make(message, bufferSize)
		for i := range m {
			m[i] = byte(i)
		}

		(*b)[i%windowSize] = m
	}
}

func idleThread() {
	for {
		start := time.Now()
		time.Sleep(sleepTime)

		wakeup := time.Now()
		perfectWakeup := start.Add(sleepTime)	
		latency := wakeup.Sub(perfectWakeup)

		if latency > worst {
			worst = latency
		}
	}
}

func main() {
	var wg sync.WaitGroup	
	var b buffer

	wg.Add(1)

	go memoryThrasher(&b, &wg)
	go idleThread()
	
	wg.Wait()

	fmt.Println("Worst latency: ", worst)
}
