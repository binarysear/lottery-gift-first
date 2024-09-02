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

func TestGetAllGiftV1(t *testing.T) {
	gifts := database.GetAllGiftsV1()
	if len(gifts) == 0 {
		t.Fail()
	} else {
		for _, gift := range gifts {
			fmt.Printf("%+v\n", *gift)
		}
	}
}

func TestGetAllGiftsV2(t *testing.T) {
	giftCh := make(chan *database.Gift, 500)
	go database.GetAllGiftsV2(giftCh)
	for {
		gift, ok := <-giftCh
		if !ok {
			break
		}
		fmt.Println(gift)
	}

}
