package test

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const url = "http://localhost:5679/lucky"
const P = 400 //模拟p个用户，在疯狂地点击“抽奖”

func TestLottery2(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(P)

	hitMap := make(map[string]int, 10)
	giftCh := make(chan string, 10000)
	counterCh := make(chan struct{})

	go func() {
		for {
			giftId, ok := <-giftCh
			if !ok {
				break
			}
			if cnt, exists := hitMap[giftId]; exists {
				hitMap[giftId] = cnt + 1
			} else {
				hitMap[giftId] = 1
			}
		}
		counterCh <- struct{}{}
	}()

	var totalCall int64    // 记录接口总的调用次数
	var totalUseTime int64 // 记录接口总的调用时间
	begin := time.Now()
	for i := 0; i < P; i++ {
		go func() {
			defer wg.Done()
			for {
				t1 := time.Now()
				resp, err := http.Get(url)
				atomic.AddInt64(&totalCall, 1)
				atomic.AddInt64(&totalUseTime, time.Since(t1).Milliseconds())
				if err != nil {
					fmt.Println(err)
					break
				}
				bs, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					break
				}
				resp.Body.Close()
				giftId := string(bs)
				if giftId == "0" { // 奖品已经抽完
					break
				}
				giftCh <- giftId
			}
		}()
	}
	wg.Wait()
	close(giftCh)
	<-counterCh // 等待hitMap准备好

	totalTime := int64(time.Since(begin).Seconds())
	if totalTime > 0 && totalCall > 0 {
		qps := totalCall / totalTime        // 每秒钟处理请求的数量
		avgTime := totalUseTime / totalCall // 每个接口平均调用花费时间
		fmt.Printf("QPS %d, avg time %dms\n", qps, avgTime)

		total := 0
		for giftId, count := range hitMap {
			total += count
			fmt.Printf("%s\t%d\n", giftId, count)
		}
		fmt.Printf("共计%d件商品\n", total)
	}
}

// go test -v .\handler\test\ -run=^TestLottery1$ -count=1
// go test -v ./handler/test/ -run=^TestLottery2$ -count=1
