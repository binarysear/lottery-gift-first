package test

import (
	"fmt"
	"gift/database"
	"gift/util"
	"testing"
)

func init() {
	util.InitLog("log") // 初始化日志，因为初始化mysql的时候使用了log
}

func TestInitGiftInventory(t *testing.T) {
	database.InitGiftInventory()
	gifts := database.GetAllGiftInventory()
	if len(gifts) == 0 {
		t.Fail()
	}
	for _, gift := range gifts {
		fmt.Println(gift)
	}
}
