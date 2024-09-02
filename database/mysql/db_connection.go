package mysql

import (
	"fmt"
	"gift/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	ormlog "gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

var (
	gift_mysql      *gorm.DB
	gift_mysql_once sync.Once
	giftLog         ormlog.Interface
)

func init() {
	giftLog = ormlog.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		ormlog.Config{
			SlowThreshold: 100 * time.Millisecond, // 慢sql阈值 超过此值会被记录
			LogLevel:      ormlog.Silent,          // 不记录任何日志
			Colorful:      false,                  // 彩色输出
		},
	)
}

func createMysqlDB(dbname, host, user, pass string, port int) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)
	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: giftLog, PrepareStmt: true})
	if err != nil {
		util.LogRus.Panicf("connect to mysql use dsn %s failed: %s", dsn, err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(20)
	util.LogRus.Infof("connect to mysql db %s", dbname)
	return db
}

// 获取gorm.DB
func GetGiftDBConnection() *gorm.DB {
	gift_mysql_once.Do(func() {
		if gift_mysql == nil { // 数据库没有初始化
			// 初始化数据库
			viper := util.CreateConfig("mysql")
			dbName := viper.GetString("mysql" + ".db")
			host := viper.GetString("mysql" + ".host")
			port := viper.GetInt("mysql" + ".port")
			user := viper.GetString("mysql" + ".user")
			pass := viper.GetString("mysql" + ".pass")
			gift_mysql = createMysqlDB(dbName, host, user, pass, port)
		}
	})
	return gift_mysql
}
