package util

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/viper"
)

var (
	//ProjectRootPath = getoncurrentPath() + "/../"
	ProjectRootPath = getCurrentPath() + "/"
)

func getoncurrentPath() string {
	_, filename, _, _ := runtime.Caller(0) // 获取调用getoncurrentPath函数的位置 也就是当前的D:/code/gocode/gift/util/../
	return path.Dir(filename)
}

func getCurrentPath() string {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		panic(err)
	}
	return currentDir
}

func CreateConfig(file string) *viper.Viper {
	config := viper.New()
	config.AddConfigPath(ProjectRootPath + "config/")
	config.SetConfigName(file)
	config.SetConfigType("yaml")

	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("找不到配置文件:%s", ProjectRootPath+"config/"+file+".yaml"))
		} else {
			panic(fmt.Errorf("解析配置文件%s出错:%s", ProjectRootPath+"config/"+file+".yaml"))
		}
	}
	return config
}
