package database

import (
	"errors"
	"gift/database/mysql"
	"gift/util"
	"gorm.io/gorm"
)

const EMPRY_GIFT = 1 // 空奖品 ("谢谢参与")的ID

type Gift struct {
	Id      int    `gorm:"column:id;primaryKey"`
	Name    string `gorm:"column:name"`
	Price   int    `gorm:"column:price"`
	Picture string `gorm:"column:picture"`
	Count   int    `gorm:"column:count"`
}

func (*Gift) TableName() string {
	return "inventory"
}

var (
	_all_gift_field = util.GetGormFields(Gift{})
)

// GetAllGiftsV1 正常获取全部商品信息
func GetAllGiftsV1() []*Gift {
	db := mysql.GetGiftDBConnection()
	var gifts []*Gift
	err := db.Select(_all_gift_field).Find(&gifts).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			util.LogRus.Errorf("read table %s failed: %s", (&Gift{}).TableName(), err)
		}
	}
	return gifts
}

// GetAllGiftsV2 防止查询的数据量太大，分次查询，把每次的数据放到channel，然后再从channel中慢慢消费
// 千万级以上大表遍历方案
func GetAllGiftsV2(giftCh chan<- *Gift) {
	size := 500
	maxId := 0
	db := mysql.GetGiftDBConnection()
	var gifts []*Gift
	for {
		err := db.Select(_all_gift_field).Where("id>?", maxId).Limit(size).Find(&gifts).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				util.LogRus.Errorf("read table %s failed: %s", (&Gift{}).TableName(), err)
			}
			break
		}

		if len(gifts) == 0 {
			break
		}

		for _, gift := range gifts {
			if gift.Id > maxId {
				maxId = gift.Id
			}
			giftCh <- gift
		}
	}
	close(giftCh)
}
