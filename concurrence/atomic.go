package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var n int32 = 0

func inc() {
	//n++  // n++会被分成三部分执行 a=n;b=n+1;n=b 并发时容易出问题
	atomic.AddInt32(&n, 1)
}

func main9() {
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			inc()
		}()
	}

	wg.Wait()
	fmt.Println(n)
}
