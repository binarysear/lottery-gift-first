package main

import (
	"sync"
	"time"
)

// 接口调用限制
var qpsCh = make(chan struct{}, 100)

func handler() {
	qpsCh <- struct{}{} // 调用接口时向管道发送消息
	defer func() {
		<-qpsCh // 调用结束时取出消息
	}()
	time.Sleep(3 * time.Second)
}

func main5() {
	const P = 1000
	twg := sync.WaitGroup{}
	twg.Add(P)
	for i := 0; i < P; i++ {
		go func() {
			defer twg.Done()
			handler()
		}()
	}

	twg.Wait()
}
