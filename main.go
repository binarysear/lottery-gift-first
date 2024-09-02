package main

import (
	"context"
	"gift/database"
	"gift/handler"
	"gift/mq"
	"gift/util"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"
)

var (
	// 创建一个通道用于控制协程的生命周期
	stopChan = make(chan struct{})
)

func Init() {
	util.InitLog("log")                                  // 初始化日志
	database.InitGiftInventory()                         // 把商品库存从Mysql读取到redis
	if err := database.ClearLotteryOrder(); err != nil { // 清空订单表
		panic("clear LotteryOrders failed")
	} else {
		// 方便测试，清空订单表
		util.LogRus.Info("clear LotteryOrders success")
	}

	mq.InitRocketmqManager()
	util.OrderIdGenerator = util.NewWorkGenerator(5, 5)

	go handler.OrderTimeTask(stopChan)
	go handler.RollBackInventoryTimeTask(stopChan)
}

func main() {

	Init()

	gin.SetMode(gin.ReleaseMode)   // GIN线上发布模式 默认是debug
	gin.DefaultWriter = io.Discard // 禁止Gin的输出

	router := gin.Default()

	router.Static("/js", "views/js") //在url是访问目录/js相当于访问文件系统中的views/js目录
	router.Static("/img", "views/img")
	router.StaticFile("/favicon.ico", "views/img/dqq.png") //在url中访问文件/favicon.ico，相当于访问文件系统中的views/img/dqq.png文件
	router.LoadHTMLFiles("views/lottery.html")             //使用这些.html文件时就不需要加路径了

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "lottery.html", nil)
	})
	router.GET("/gifts", handler.GetAllGifts)     // 获取所有奖品信息
	router.GET("/lucky", handler.Lottery)         // // 点击抽奖按钮
	go http.ListenAndServe("127.0.0.1:1234", nil) //访问http://127.0.0.1:1234/debug/pprof查看goroutine异常

	//router.Run("0.0.0.0:5678")
	// 优雅启动停止
	srv := &http.Server{
		Addr:    "0.0.0.0:5679",
		Handler: router,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	close(stopChan)
	mq.MQManager.ShutdownMq()
	log.Println("Server exiting")

}
