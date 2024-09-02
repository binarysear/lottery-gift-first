package main

import "fmt"

//func main() {
//	ch := make(chan Gift, 100)
//
//	go func() {
//		for i := 0; i < 10000; i++ {
//			ch <- Gift{}
//		}
//		close(ch)
//		fmt.Println("写入完成")
//	}()
//
//	for {
//		gift, ok := <-ch
//		if !ok {
//			break // ok==false表示 channel为空 并且channel已经关闭
//		}
//		_ = gift // 将gift写入redis
//	}
//	time.Sleep(3 * time.Second)
//}

type Gift struct{}

func main6() {
	ch := make(chan int, 3) // channel有缓存，close(channel)之后能够继续读取channel， 如果channel有数据就返回数据，没有数据就返回管道类型初始默认值
	// 如果没有缓存 直接读取已经close的channel会deadlock
	//ch <- 1
	//fmt.Println(<-ch)
	//ch <- 2
	//fmt.Println(<-ch)
	//ch <- 3
	//fmt.Println(<-ch)
	close(ch)

	fmt.Println(<-ch)
	fmt.Println(<-ch)
}
