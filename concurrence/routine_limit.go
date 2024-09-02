package main

import (
	"fmt"
	"runtime"
	"time"
)

func main3() {
	limiter := NewGoroutineLimiter(100)
	work := func() {
		time.Sleep(10 * time.Second)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			// 每一秒统计一下当前goroutine数量
			<-ticker.C
			fmt.Printf("当前goroutine数量为: %d\n", runtime.NumGoroutine())
		}
	}()

	for i := 0; i < 10000; i++ {
		limiter.Run(work) // 通过Run创建协程
	}

	time.Sleep(11 * time.Second)
}

type GoroutineLimiter struct {
	limit int           // 限制goroutine的数量,运行的协程不会超过这个数量
	ch    chan struct{} // 通过channel限制
}

func NewGoroutineLimiter(limit int) *GoroutineLimiter {
	return &GoroutineLimiter{
		limit: limit,
		ch:    make(chan struct{}, limit),
	}
}

func (g *GoroutineLimiter) Run(f func()) {
	g.ch <- struct{}{} // 子协程开启时向管道发送一个信号
	go func() {
		f()
		<-g.ch // 子协程退出时从管道中取出一个信号
	}()
}
