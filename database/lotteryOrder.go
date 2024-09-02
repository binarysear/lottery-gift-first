package database

import (
	"errors"
	selfMysql "gift/database/mysql"
	"gift/util"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

const TableNameLotteryOrder = "lottery_order"

// LotteryOrder mapped from table <lottery_order>
type LotteryOrder struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UserID     int64     `gorm:"column:user_id;not null;comment:用户id" json:"user_id"`            // 用户id
	ActivityID int64     `gorm:"column:activity_id;not null;comment:活动id" json:"activity_id"`    // 活动id
	GiftID     int64     `gorm:"column:gift_id;not null;comment:奖项id" json:"gift_id"`            // 奖项id
	OrderID    int64     `gorm:"column:order_id;not null;comment:订单id" json:"order_id"`          // 订单id
	State      int32     `gorm:"column:state;comment:状态 (0:用户不可见,1:用户可见,2:用户已经领取)" json:"state"` // 状态 (0:用户不可见,1:用户可见,2:用户已经领取)
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	Creator    string    `gorm:"column:creator" json:"creator"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
	Updater    string    `gorm:"column:updater" json:"updater"`
}

// TableName LotteryOrder's table name
func (*LotteryOrder) TableName() string {
	return TableNameLotteryOrder
}

func CreateLotteryOrder(order *LotteryOrder) error {
	db := selfMysql.GetGiftDBConnection()
	if err := db.Create(&order).Error; err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case 1062: // MySQL中表示重复条目的代码
				// 处理重复条目
				util.LogRus.Errorf("duplicate create order failed: %s", err)
				return errors.New("duplicate")
			default:
				// 处理其他错误
				util.LogRus.Errorf("create order failed: %s", err)
				return err
			}
		} else {
			// 处理非MySQL错误或未知错误
			util.LogRus.Errorf("create order failed: %s", err)
			return err
		}
	}
	util.LogRus.Debugf("create order id %d", order.OrderID)
	return nil
}

func ClearLotteryOrder() error {
	db := selfMysql.GetGiftDBConnection()
	var orders []LotteryOrder
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
		return db.Where("id>0").Delete(LotteryOrder{}).Error
	}
}

// UpdateOrderStateByOrderId 更新订单状态
func UpdateOrderStateByOrderId(orderId, state int64) error {
	db := selfMysql.GetGiftDBConnection()
	return db.Model(&LotteryOrder{}).Where("order_id=?", orderId).Update("state", state).Error
}
