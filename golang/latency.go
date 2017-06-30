package main

import (
	"fmt"
	"flag"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"gopkg.in/cheggaaa/pb.v2"
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

func storeLatencies(path string, latencies []time.Duration) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	for _, l := range latencies {
		latency := []byte(fmt.Sprintf("%f\n", (float32)(l)/1000))
		_, err = f.Write(latency)
		if err != nil {
			return err
		}
	}

	return nil
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return make(message, bufferSize)
	},
}

func newMessage(pool bool) message {
	if pool {
		return bufPool.Get().(message)
	}

	return make(message, bufferSize)
}

func idleThread(b *buffer, cycles, period, buffers int, usePool bool, latenciesFile string, wg *sync.WaitGroup) {
	var templ pb.ProgressBarTemplate = `{{string . "Cycles"}}{{counters . }} {{bar . }} {{percent . }}{{string . ""}}`
	var latenciesArray []time.Duration
	sleepPeriod := (time.Duration)(period) * time.Millisecond
	bestLatency = time.Minute

	defer wg.Done()

	bar := pb.StartNew(cycles)
	bar = bar.SetTemplate(templ) 

	for i:= 0; i < cycles; i++ {
		bar.Increment()

		for j := 0; j < buffers; j++ {
			m := newMessage(usePool)
			for i := range m {
				m[i] = byte(j)
			}

			if (*b)[j%windowSize] != nil && usePool {
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

		latenciesArray = append(latenciesArray, latency)

		if usePool {
			for j := 0; j < buffers; j++ {
				bufPool.Put((*b)[j])
				(*b)[j%windowSize] = nil
			}
		}
	}

	bar.Finish()

	if latenciesFile != "" {
		storeLatencies(latenciesFile, latenciesArray)
	}
}

func main() {
	var wg sync.WaitGroup
	var b buffer
	var gcStats debug.GCStats

	cycles := flag.Int("cycles", 500, "number of sleeping cycles")
	sleepPeriod := flag.Int("period", 100, "Sleeping period (in milliseconds)")
	numBuffer := flag.Int("buffers", 10, "Number of allocated buffer per cycle")
	usePool := flag.Bool("no-pool", false, "Do not use golang pools for memory allocations")
	debugPause := flag.Bool("debug", false, "Dump all GC pauses and latencies")
	latenciesFile := flag.String("file", "", "File to store all latencies in microseconds")
	flag.Parse()

	fmt.Printf("%d cycles - %dms sleep period - %d buffers allocated per cycle - Golang pools %v\n",
		*cycles, *sleepPeriod, *numBuffer, !(*usePool))

	wg.Add(1)

	go idleThread(&b, *cycles, *sleepPeriod, *numBuffer, !(*usePool), *latenciesFile, &wg)

	wg.Wait()

	fmt.Printf("Latency: [Avg %vµs, Best %v, Worst %v]\n", (averageLatency.Nanoseconds()/(int64)(*cycles))/1000, bestLatency, worstLatency)

	debug.ReadGCStats(&gcStats)

	fmt.Printf("\nGC Stats:\n")

	fmt.Printf("\tNumber of GC runs %v\n", gcStats.NumGC)
	fmt.Printf("\tTotal GC pause time %v\n", gcStats.PauseTotal)

	if *debugPause {
		fmt.Printf("\tGC pauses: %v\n", gcStats.Pause)
	}
}
