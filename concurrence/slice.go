package main

import (
	"fmt"
	"sync"
)

// 使用goroutine加速切片的遍历操作 为了避免并发冲突，把slice分成一片一片的，让每一个goroutine各司其职
func main14() {
	const LEN = 100
	const LOOP = 100
	const P = 10
	arr := make([]int, LEN)
	wg := sync.WaitGroup{}
	wg.Add(P)

	for i := 0; i < P; i++ {
		go func(i int) {
			defer wg.Done()
			fmt.Println(i)
			start := i * (LEN / P)
			end := start + (LEN / P)
			for j := 0; j < LOOP; j++ {
				for k := start; k < end; k++ {
					arr[k]++
				}
			}
		}(i)
	}
	wg.Wait()
	sum := 0
	for k := 0; k < LEN; k++ {
		sum += arr[k]
	}
	fmt.Println(sum)
}
