package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ChopStick struct { // declare a ChopStick type struct (mutex)
	sync.Mutex
}

type Philo struct { // declare a Philosopher type struct
	leftCS       *ChopStick // L and R chopsticks
	rightCS      *ChopStick
	id           int      // id number
	eatCount     int      // count of times eaten
	stopEating   chan int // stop eating channel
	startEating  chan int // start eating channel
	requestToEat chan int // request to eat channel
}

var wg sync.WaitGroup // declare the global wait group "wg"

func main() {
	cSticks := make([]*ChopStick, 5) // make a slice of 5 chopsticks
	philos := make([]*Philo, 5)      // make a slice of 5 philosophers
	stopEating := make(chan int)     // make the required channels
	startEating := make(chan int)
	requestToEat := make(chan int)
	allDone := make(chan int)

	rand.Seed(time.Now().UTC().UnixNano()) // seed the random number generator

	for i := 0; i < 5; i++ { // populate chopstick slice with 5 new chopsticks
		cSticks[i] = new(ChopStick)
	}

	for i := 0; i < 5; i++ {
		// populate philosopher slice with 5 new philosophers
		philos[i] = &Philo{
			leftCS:       cSticks[i],
			rightCS:      cSticks[(i+1)%5],
			id:           i + 1,
			eatCount:     0,
			stopEating:   stopEating,
			startEating:  startEating,
			requestToEat: requestToEat,
		}
	}

	// host goroutine runs while philosophers eat
	go host(requestToEat, startEating, stopEating, allDone)

	wg.Add(5) // add 5 to the wait group (for the 5 philosphers)

	for _, p := range philos { // loop through the 5 philosophers,
		go p.eat() // run goroutine for each philosopher to eat
	}

	wg.Wait()    // wait for 5 philosophers to finish eating ("wg.Done" 5x)
	allDone <- 1 // signal the allDone channel
}

func host(requestToEat, startEating, stopEating, allDone chan int) {
	philosEating := 0 // count of philosophers currently eating

	for {
		select {

		case <-requestToEat: // case there's a signal on the request to eat channel,
			if philosEating < 2 { // if less than 2 philosophers currently eating,
				philosEating++   // add 1 to the number of philosophers eating
				startEating <- 1 // signal 1 (allow) on the startEating channel for this philosopher
			} else {
				startEating <- 0 // else, signal 0 (deny) on the startEating channel for this philosopher
			}

		case <-stopEating: // case there's a signal on the stop eating channel,
			philosEating-- // subtract 1 from the number of philosophers eating

		case <-allDone: // case there's a signal on the allDone channel, return
			return
		}
	}
}

func (p Philo) eat() { // philosopher eating function
	for {

		if p.eatCount == 3 { // if the philosopher has eaten 3 times,
			p.stopEating <- 1 // signal the stopEating channel
			wg.Done()         // signal done to the wait group, and return
			return
		}

		// else, signal the request to eat channel
		p.requestToEat <- 1

		// block and wait for a signal on the startEating channel
		if allowedToEat := <-p.startEating; allowedToEat == 1 { // if the signal is "1" (allow),
			p.leftCS.Lock() // lock both chopsticks
			p.rightCS.Lock()
			fmt.Printf("Starting to eat %d\n", p.id)
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(p.id*1000))) // eat for a random amount of time
			fmt.Printf("Finishing eating %d\n", p.id)
			p.eatCount++      // increase the eat counter +1
			p.stopEating <- 1 // signal the stopEating channel
			p.leftCS.Unlock() // release both chopsticks
			p.rightCS.Unlock()
		}
	}
}
