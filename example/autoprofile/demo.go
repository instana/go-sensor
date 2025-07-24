// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	instana "github.com/instana/go-sensor"
)

func useCPU(duration int, usage int) {
	for j := 0; j < duration; j++ {
		go func() {
			for i := 0; i < usage*80000; i++ {
				_ = strconv.Itoa(i)
			}
		}()

		time.Sleep(1 * time.Second)
	}
}

func simulateCPUUsage() {
	// sumulate CPU usage anomaly - every 45 minutes
	cpuAnomalyTicker := time.NewTicker(45 * time.Minute)
	go func() {
		for range cpuAnomalyTicker.C {
			// for 60 seconds produce generate 50% CPU usage
			useCPU(60, 50)
		}
	}()

	// generate constant ~10% CPU usage
	useCPU(math.MaxInt64, 10)
}

func leakMemory(duration int, size int) {
	mem := make([]string, 0)

	for j := 0; j < duration; j++ {
		go func() {
			for i := 0; i < size; i++ {
				mem = append(mem, strconv.Itoa(i))
			}
		}()

		time.Sleep(1 * time.Second)
	}
}

func simulateMemoryLeak() {
	// simulate memory leak - constantly
	constantTicker := time.NewTicker(2 * 3600 * time.Second)
	go func() {
		for range constantTicker.C {
			leakMemory(2*3600, 1000)
		}
	}()

	go leakMemory(2*3600, 1000)
}

func simulateChannelWait() {
	for {
		done := make(chan bool)

		go func() {
			wait := make(chan bool)

			go func() {
				time.Sleep(500 * time.Millisecond)

				wait <- true
			}()

			<-wait

			done <- true
		}()

		<-done

		time.Sleep(1000 * time.Millisecond)
	}
}

func simulateNetworkWait() {
	// start HTTP server
	go func() {
		http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			done := make(chan bool)

			go func() {
				time.Sleep(time.Duration(200+rand.Intn(5)) * time.Millisecond)
				done <- true
			}()
			<-done

			fmt.Fprintf(w, "OK")
		})

		if err := http.ListenAndServe(":5002", nil); err != nil {
			log.Fatal(err)
			return
		}
	}()

	requestTicker := time.NewTicker(500 * time.Millisecond)
	for range requestTicker.C {
		res, err := http.Get("http://localhost:5002/test")
		if err == nil {
			res.Body.Close()
		}
	}
}

func simulateSyscallWait() {
	for {
		done := make(chan bool)

		go func() {
			_, err := exec.Command("sleep", "1").Output()
			if err != nil {
				fmt.Println(err)
			}

			done <- true
		}()

		time.Sleep(1 * time.Second)

		<-done
	}
}

func simulateLockWait() {
	for {
		done := make(chan bool)

		lock := &sync.RWMutex{}
		lock.Lock()

		go func() {
			lock.RLock()
			defer lock.RUnlock()

			done <- true
		}()

		go func() {
			time.Sleep(200 * time.Millisecond)
			lock.Unlock()
		}()

		<-done

		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	instana.InitCollector(&instana.Options{
		Service:           "autoprofile-demo",
		LogLevel:          instana.Debug,
		EnableAutoProfile: true,
		Tracer:            instana.DefaultTracerOptions(),
	})

	go simulateCPUUsage()
	go simulateMemoryLeak()
	go simulateChannelWait()
	go simulateNetworkWait()
	go simulateSyscallWait()
	go simulateLockWait()

	select {}
}
