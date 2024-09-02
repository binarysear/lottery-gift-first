package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main3() {
	g := gen.NewGenerator(gen.Config{
		//  设置输出路径
		OutPath: "./query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // 选择生成模式
	})
	//  建立数据库连接
	gormdb, _ := gorm.Open(mysql.Open("root:root@(127.0.0.1:3306)/gift?charset=utf8mb4&parseTime=True&loc=Local"))
	g.UseDB(gormdb) // 选择数据库连接

	// 为结构模型生成基本类型安全的 DAO API。用户的以下约定

	g.ApplyBasic(
		g.GenerateModel("lottery_order"),
	)
	// 生成代码
	g.Execute()
}
