package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func main15() {
	for {
		stopWork()
	}
}

func stopWork() {
	const Max = 10000
	const P = 100

	wg := sync.WaitGroup{}
	wg.Add(P)
	dataCh := make(chan int, 100)
	stopCh := make(chan struct{})

	for i := 0; i < P; i++ {
		go func() {
			defer wg.Done()
			for {
				data := rand.Intn(Max)
				select {
				case dataCh <- data:
				case <-stopCh:
					return
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case val := <-dataCh:
				if val == Max-1 {
					close(stopCh)
					return
				}
			}
		}
	}()

	wg.Wait()

	fmt.Println("stop")
}
