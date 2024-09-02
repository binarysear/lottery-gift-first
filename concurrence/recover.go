package main

import "fmt"

// recover只能放在匿名函数(func(){}())里面执行并且panic必须在匿名函数之后，并且只能捕获当前协程的panic
func main2() {
	defer func() {
		err := recover() // 如果没有panic，则recover()返回nil
		fmt.Println(err)
	}()
	go panic("故意panic") // panic必须发生在注册recover之后，并且跟recover同协程内才能被recover捕获
	fmt.Println("调用结束")
}
