package main

import (
	"fmt"
	"flag"
	"runtime/debug"
	"sync"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

const (
	bufferSize = 4096 * 10
	windowSize = 200000
)

type (
	message []byte
	buffer  [windowSize]message
)

var worstLatency time.Duration
var bestLatency time.Duration
var averageLatency time.Duration

var bufPool = sync.Pool{
	New: func() interface{} {
		return make(message, bufferSize)
	},
}

func idleThread(b *buffer, cycles, period, buffers int, latencies *[]time.Duration, wg *sync.WaitGroup) {
	sleepPeriod := (time.Duration)(period) * time.Millisecond
	bestLatency = time.Minute

	defer wg.Done()
	bar := pb.StartNew(cycles)

	for i:= 0; i < cycles; i++ {
		bar.Increment()

		for j := 0; j < buffers; j++ {
			m := bufPool.Get().(message)
			for i := range m {
				m[i] = byte(j)
			}

			if (*b)[j%windowSize] != nil {
				bufPool.Put((*b)[j%windowSize])
			}
			(*b)[j%windowSize] = m
		}

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

		for j := 0; j < buffers; j++ {
			bufPool.Put((*b)[j])
			(*b)[j%windowSize] = nil
		}
	}

	bar.FinishPrint("Done")
}

func main() {
	var wg sync.WaitGroup
	var b buffer
	var gcStats debug.GCStats
	var latencies []time.Duration

	cycles := flag.Int("cycles", 500, "number of sleeping cycles")
	sleepPeriod := flag.Int("period", 100, "Sleeping period (in milliseconds)")
	numBuffer := flag.Int("buffers", 10, "Number of allocated buffer per cycle")
	debugPause := flag.Bool("debug", false, "Dump all GC pauses and latencies")
	flag.Parse()

	fmt.Printf("%d cycles, %dms sleep period:\n", *cycles, *sleepPeriod)

	wg.Add(1)

	go idleThread(&b, *cycles, *sleepPeriod, *numBuffer, &latencies, &wg)

	wg.Wait()

	fmt.Printf("Latency: [Avg %vµs, Best %v, Worst %v]\n", (averageLatency.Nanoseconds()/(int64)(*cycles))/1000, bestLatency, worstLatency)

	if *debugPause {
		fmt.Printf("Latencies: %v\n", latencies)
	}

	debug.ReadGCStats(&gcStats)

	fmt.Printf("\nGC Stats:\n")

	fmt.Printf("\tNumber of GC runs %v\n", gcStats.NumGC)
	fmt.Printf("\tTotal GC pause time %v\n", gcStats.PauseTotal)

	if *debugPause {
		fmt.Printf("\tGC pauses: %v\n", gcStats.Pause)
	}
}
