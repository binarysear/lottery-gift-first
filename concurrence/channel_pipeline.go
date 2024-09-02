package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

// 一共分为上中下三个
// 上游负责读取数据
// 中游负责处理数据
// 下游负责从中游读取并写入数据

// 上游
var readWg = sync.WaitGroup{}
var textCh = make(chan string, 100)

// 中游
var dealWg = sync.WaitGroup{}
var dataCh = make(chan int, 100)

// 下游
var writeFinishCh = make(chan struct{})

func readFile(fileName string) {
	defer readWg.Done()
	fin, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fin.Close()

	reader := bufio.NewReader(fin)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				return
			}
		} else {
			textCh <- string(line)
		}
	}
}

func dealData() {
	defer dealWg.Done()
	for {
		str, ok := <-textCh
		if !ok {
			break
		} else {
			sum := calculate(str)
			dataCh <- sum
		}
	}
}

func calculate(str string) int {
	sum := 0
	for _, s := range str {
		sum += int(s)
	}
	return sum
}

func writerFile(fileName string) {
	fout, err := os.OpenFile(fileName, os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fout.Close()

	writer := bufio.NewWriter(fout)

	for {
		data, ok := <-dataCh
		if !ok {
			break
		} else {
			writer.WriteString(strconv.Itoa(data))
			writer.WriteString("\n")
		}
	}

	err = writer.Flush()
	writeFinishCh <- struct{}{}

}

func main8() {
	// 第一阶段 io密集型，并行执行提高速度
	readWg.Add(2)
	go readFile("data/rsa_private_key.pem")
	go readFile("data/rsa_public_key.pem")

	// 第二阶段 cpu密集型，多分配几个内核线程
	dealWg.Add(4)
	for i := 0; i < 4; i++ {
		go dealData()
	}

	// 第三阶段 汇总，写一个文件
	go writerFile("data/digit.txt")

	// 第一阶段结束后，关闭管道textCh
	readWg.Wait()
	close(textCh)

	// 第二阶段结束后，关闭管道dataCh
	dealWg.Wait()
	close(dataCh)

	// 第三阶段结束后，往writeFinishCh里面取数据
	<-writeFinishCh
}
