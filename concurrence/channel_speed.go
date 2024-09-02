package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

// 加速文件合并

// 三个生产者
const PROUDUC_NUM = 3

// 数据存放
var buffer = make(chan string, 100)

// 生产者使用信号
var pc_sync = make(chan struct{}, PROUDUC_NUM)

// 全部消费完信号
var all_over = make(chan struct{})

// 一个生产者，负责读取一个文件并且写入buffer
func produce(fileName string) {
	fin, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fin.Close()
	reader := bufio.NewReader(fin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF { //io.EOF 是 Go 语言中表示文件或输入流结束的错误值。当使用函数如 Read 或 ReadString 读取到文件或输入流的末尾时，会返回 io.EOF
				if len(line) > 0 {
					buffer <- (line + "\n")
				}
				break
			} else {
				fmt.Println(err)
			}
		} else {
			buffer <- line
		}
	}
	<-pc_sync
}

func consume(fileName string) {
	fout, err := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fout.Close()

	writer := bufio.NewWriter(fout)

	for {
		if len(buffer) == 0 { // 1.生产者还没生产完消息就被消费完了 2.生产者已经全部生产完成并且已经全部被消费完成
			if len(pc_sync) == 0 {
				// 2
				break
			}
			// 1 消费太快 睡一下等一下生产者生产
			time.Sleep(10 * time.Millisecond)
		} else {
			line := <-buffer
			writer.WriteString(line)
		}
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println(err)
		return
	}
	all_over <- struct{}{}
}

func main7() {
	for i := 0; i < PROUDUC_NUM; i++ {
		pc_sync <- struct{}{}
	}

	go produce("data/1.txt")
	go produce("data/2.txt")
	go produce("data/3.txt")
	go consume("data/big.txt")

	<-all_over
}
