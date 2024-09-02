package main

import (
	"context"
	"fmt"
	"time"
)

const (
	WorkUseTime = 500 * time.Millisecond
	Timeout     = 100 * time.Millisecond
)

// 模拟一个长时间工作的接口
func WorkLongTime() int {
	time.Sleep(WorkUseTime)
	return 888
}

// 模拟一个处理接口
func Handle1() {
	workCh := make(chan int, 1)
	dealCh := make(chan struct{}, 1)
	go func() {
		n := WorkLongTime()
		workCh <- n
	}()

	go func() {
		time.Sleep(Timeout)
		dealCh <- struct{}{}
	}()

	select {
	case <-workCh:
		fmt.Println("work over")
	case <-dealCh:
		fmt.Println("work timeout")
	}
}

func Handle2() {
	workCh := make(chan int, 1)

	go func() {
		n := WorkLongTime()
		workCh <- n
	}()

	select {
	case <-workCh:
		fmt.Println("work over")
	case <-time.After(Timeout):
		fmt.Println("work timeout")
	}
}

func Handle3() {
	workCh := make(chan int, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		n := WorkLongTime()
		workCh <- n
	}()
	go func() {
		time.Sleep(Timeout)
		cancel()
	}()

	select {
	case <-workCh:
		fmt.Println("work over")
	case <-ctx.Done():
		fmt.Println("work timeout")
	}
}

func Handle4() {
	workCh := make(chan int, 1)
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	go func() {
		n := WorkLongTime()
		workCh <- n
	}()

	select {
	case <-workCh:
		fmt.Println("work over")
	case <-ctx.Done():
		fmt.Println("work timeout")
	}
}

func main11() {
	Handle1()
	Handle2()
	Handle3()
	Handle4()
}
