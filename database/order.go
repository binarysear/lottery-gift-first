package database

import (
	"errors"
	"gift/database/mysql"
	"gift/util"
	"gorm.io/gorm"
)

type Order struct {
	Id     int `gorm:"column:id"`
	GiftId int `gorm:"column:gift_id"`
	UserId int `gorm:"column:user_id"`
}

func (*Order) TableName() string {
	return "orders"
}

func ClearOrder() error {
	db := mysql.GetGiftDBConnection()
	var orders []Order
	err := db.Where("id>0").Find(&orders).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if len(orders) == 0 {
		return nil
	} else {
		return db.Where("id>0").Delete(Order{}).Error
	}
}

// CreateOrder
// 创建订单
func CreateOrder(order *Order) int {
	db := mysql.GetGiftDBConnection()
	if err := db.Create(&order).Error; err != nil {
		util.LogRus.Errorf("create order failed: %s", err)
		return 0
	} else {
		util.LogRus.Debugf("create order id %d", order.Id)
		return order.Id
	}
}
