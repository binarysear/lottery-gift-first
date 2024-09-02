package main

import (
	"fmt"
	"gift/util"
)

func main() {
	viper := util.CreateConfig("rocketmq")
	brokers := viper.GetString("brokers")
	fmt.Println(brokers)
}
