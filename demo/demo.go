package main

import (
	"flag"
	"github.com/swfrench/lamport-go"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Run the Lamport distributed lock demo for n communicating goroutines
func demo(n int) {
	// create input channel for each goroutine
	chs := make([]chan lamport.Message, n)
	for p := range chs {
		chs[p] = make(chan lamport.Message, 512)
	}

	// initialize the waitgroup
	var group sync.WaitGroup
	group.Add(n)

	// initialize the shared test var
	var tvar int32

	// spawn goroutine "workers"
	for p, _ := range chs {
		go func(myProc int, ptvar *int32) {
			// initialize the distributed lock
			lock := lamport.Start(myProc, chs)

			// acquire
			lock.Acquire()
			log.Println(myProc, "Acquired lock")

			// lock is acquired - set the test var to my proc id
			atomic.StoreInt32(ptvar, int32(myProc))

			// sleep for a bit
			time.Sleep(100 * time.Millisecond)

			// sample the test var
			tval := atomic.LoadInt32(ptvar)

			// release
			lock.Release()
			log.Println(myProc, "Released lock")

			// check the sampled test var
			if tval != int32(myProc) {
				log.Fatal("Error: test var =", tval, "!= myProc")
			} else {
				log.Println(" OK: mutating test var is still", tval)
			}

			// sync
			group.Done()
		}(p, &tvar)
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
	demo(*n)
}
