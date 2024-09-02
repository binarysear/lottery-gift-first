package main

import (
	"fmt"
	"reflect"
	"time"
)

func main1() {
	ch := make(chan struct{})
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("子协程结束")
		ch <- struct{}{}
	}()
	<-ch

	testEmptyStruct()
}

type A struct{}
type B struct{}

// 空结构体统一使用一个地址，不占内存
func testEmptyStruct() {
	a := A{}
	b := B{}
	fmt.Printf("%p  %p\n", &a, &b)
	typeA := reflect.TypeOf(a)
	typeB := reflect.TypeOf(b)
	fmt.Printf("%d %d", typeA.Size(), typeB.Size())
	// 0x10b82c0  0x10b82c0
	// 0 0
}
