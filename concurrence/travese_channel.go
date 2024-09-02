package main

import (
	"fmt"
	"sync"
	"time"
)

var ch = make(chan int, 3)
var twg = sync.WaitGroup{}

func addCh() {
	defer func() {
		twg.Done()
	}()
	for i := 0; i < 10; i++ {
		ch <- i
	}
	time.Sleep(1 * time.Second)
	for i := 0; i < 10; i++ {
		ch <- i
	}
	close(ch) // 管道用完记得关掉，防止死锁 fatal error: all goroutines are asleep - deadlock!
}

func traveseCh() {
	defer func() {
		twg.Done()
	}()
	//for els := range ch { // 此处如果管道没有关闭会一直等待 容易出现死锁
	//	fmt.Println(els)
	//}

	for {
		els, ok := <-ch
		if ok {
			fmt.Println(els)
		} else {
			break
		}
	}

	fmt.Println("bye bye")
}

func main4() {
	twg.Add(2)
	go addCh()
	go traveseCh()

	twg.Wait()
}
