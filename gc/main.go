package main

import (
	"fmt"
	"flag"
	"runtime/debug"
//	"sort"
	"sync"
	"time"
)

const (
	bufferSize = 4096
	windowSize = 200000
)

type (
	message []byte
	buffer  [windowSize]message
)

var worstLatency time.Duration
var bestLatency time.Duration
var averageLatency time.Duration

func memoryThrasher(b *buffer) {
	for i := 0; ; i++ {
		m := make(message, bufferSize)
		for i := range m {
			m[i] = byte(i)
		}

		(*b)[i%windowSize] = m
	}
}

func idleThread(cycles, period int, latencies *[]time.Duration, wg *sync.WaitGroup) {
	sleepPeriod := (time.Duration)(period) * time.Millisecond

	defer wg.Done()
	for i:= 0; i < cycles; i++ {
		start := time.Now()
		time.Sleep(sleepPeriod)

		wakeup := time.Now()
		perfectWakeup := start.Add(sleepPeriod)
		latency := wakeup.Sub(perfectWakeup)

		averageLatency += latency

		if latency > worstLatency {
			worstLatency = latency
		}

		if latency < bestLatency {
			bestLatency = latency
		}

		*latencies = append(*latencies, latency)
	}
}

func main() {
	var wg sync.WaitGroup	
	var b buffer
	var gcStats debug.GCStats
	var latencies []time.Duration

	cycles := flag.Int("cycles", 500, "number of sleeping cycles")
	sleepPeriod := flag.Int("sleep", 100, "Sleeping period (in milliseconds)")
	flag.Parse()

	fmt.Printf("%d cycles, %dms sleep period:\n", *cycles, *sleepPeriod)

	wg.Add(1)

	go memoryThrasher(&b)
	go idleThread(*cycles, *sleepPeriod, &latencies, &wg)
	
	wg.Wait()

	fmt.Printf("\tWorst latency: %v\n", worstLatency)
	fmt.Printf("\tBest latency: %v\n", bestLatency)
	fmt.Printf("\tAverage latency: %vµs\n", (averageLatency.Nanoseconds()/(int64)(*cycles))/1000)
	fmt.Printf("\tLatencies: %v\n", latencies)

	debug.ReadGCStats(&gcStats)

	fmt.Printf("\nGC Stats:\n")
	fmt.Printf("\tLast GC run %v\n", gcStats.LastGC)
	fmt.Printf("\tNumber of GC runs %v\n", gcStats.NumGC)
	fmt.Printf("\tTotal GC pause time %v\n", gcStats.PauseTotal)

//	sort.Sort(gcStats.Pause)
	fmt.Printf("\tSorted pauses: %v\n", gcStats.Pause)
}
