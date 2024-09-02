package main

import (
	"fmt"
	"sync"
)

var mp = make(map[int]int, 1000)
var lock sync.RWMutex
var smp = sync.Map{}

func readMap() {
	for i := 0; i < 1000; i++ {
		lock.RLock()
		_ = mp[5]
		lock.RUnlock()
		smp.Load(10)
	}
}

func writeMap() {
	for i := 0; i < 1000; i++ {
		lock.Lock()
		mp[10] = 10
		lock.Unlock()
		smp.Store(10, 10)
	}
}

func mapInc(smp *sync.Map, key int) {
	lock.Lock()
	defer lock.Unlock()
	if value, ok := smp.Load(key); ok {
		smp.Store(key, value.(int)+1)
	} else {
		smp.Store(key, 1)
	}
}

func main10() {
	// fatal error: concurrent map read and map write
	// 直接并发读取和写map和会报错误
	//go readMap()
	//go writeMap()
	//time.Sleep(1 * time.Second)

	wg := sync.WaitGroup{}
	const P = 1000
	const key = 8

	wg.Add(P)
	for i := 0; i < P; i++ {
		go func() {
			defer wg.Done()
			mapInc(&smp, key)
		}()
	}
	wg.Wait()

	fmt.Println(smp.Load(key))
}
