package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"time"
)

func rpc() int {
	time.Sleep(50 * time.Millisecond)
	return 888
}

func homeHandler(ctx *gin.Context) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	workDone := make(chan int)
	defer cancel()
	go func() {
		n := rpc()
		select {
		case workDone <- n:
			fmt.Println("over")
		default:
			fmt.Println("done")
			return
		}
		// dealCh <- n  // 这里会一直阻塞 导致goroutine泄露
	}()
	select {
	case n := <-workDone:
		ctx.String(http.StatusOK, strconv.Itoa(n))
	case <-timeoutCtx.Done():
		ctx.String(http.StatusInternalServerError, strconv.Itoa(0))
	}
}

func main12() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			fmt.Println(runtime.NumGoroutine())
		}
	}()

	go http.ListenAndServe("127.0.0.1:1234", nil) //访问http://127.0.0.1:1234/debug/pprof查看goroutine异常

	//gin.DefaultWriter = io.Discard
	r := gin.Default()
	r.GET("/", homeHandler)
	r.Run("127.0.0.1:5678")
}
