package main

import (
	"flag"
	"github.com/swfrench/lamport-go"
	"log"
	"sync"
	"time"
)

// Run the Lamport distributed lock demo for n communicating goroutines
func Demo(n int) {
	// create input channel for each goroutine
	chs := make([]chan lamport.Message, n)
	for p := range chs {
		// ensure channel buffer is of at least size n (simultaneous
		// acquisition attempts will not block)
		chs[p] = make(chan lamport.Message, n)
	}

	// init the waitgroup
	var group sync.WaitGroup
	group.Add(n)

	// spawn goroutines
	for p, _ := range chs {
		go func(myProc int) {
			// initialize the distributed lock
			lock := lamport.Start(myProc, chs)

			// acquire
			lock.Acquire()
			log.Println(myProc, "Have lock")

			// release
			lock.Release()
			log.Println(myProc, "Released lock")

			time.Sleep(2 * time.Second)

			// sync
			group.Done()
		}(p)
	}

	// wait on the team
	group.Wait()
}

func main() {
	// get number of processes (goroutines in the demo)
	var n = flag.Int("n", 2, "number of processes")
	flag.Parse()

	// check n for sensible values
	if *n < 2 {
		log.Fatal("Error: nonsense number of processes ", *n)
	}

	// run the demo
	Demo(*n)
}
