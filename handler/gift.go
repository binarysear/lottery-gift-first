package handler

import (
	"context"
	"encoding/json"
	"errors"
	"gift/database"
	myredis "gift/database/redis"
	"gift/mq"
	"gift/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// GetAllGifts 获取所有商品信息
func GetAllGifts(c *gin.Context) {
	gifts := database.GetAllGiftsV1()
	if len(gifts) == 0 {
		c.JSON(http.StatusInternalServerError, nil)
	} else {
		// 抹掉敏感信息
		for _, gift := range gifts {
			gift.Count = 0
		}
		c.JSON(http.StatusOK, gifts)
	}
}

// Lottery 抽奖接口
func Lottery(c *gin.Context) {
	ids, counts, err := CheckInventory()
	if err != nil {
		c.String(http.StatusOK, strconv.Itoa(0)) // 0表示所有商品已经抽完
		return
	}

	// 3.调用抽奖算法获取抽中的商品index 此index即为商品在ids和counts中的索引
	index := util.Lottory(counts)
	giftId := ids[index]

	// 4.尝试创建订单
	// 4.1创建订单号 使用雪花算法
	orderId := util.OrderIdGenerator.GeneratorID()

	// 4.2 创建订单信息
	order := database.LotteryOrder{
		OrderID:    orderId,
		UserID:     1,
		GiftID:     int64(giftId),
		State:      0, // 正常创建给0,待确认后续抽奖流程成功后置为1
		ActivityID: 1,
		Creator:    "hua",
		Updater:    "hua",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 5.发送mq事务消息
	//err = mq.Create(order)
	err = mq.MQManager.SendTransactionMQMessage(order)

	if err != nil {
		util.LogRus.Warnf("用户：%d 尝试生成订单消息失败,error: %v", order.UserID, err)
		c.String(http.StatusOK, strconv.Itoa(database.EMPRY_GIFT))
		return
	}

	// 6.需要接着判断订单是否生成成功才能返回给用户确认消息
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	queryTick := time.NewTicker(1 * time.Millisecond)
	defer queryTick.Stop()
	stockKey := "{lottery-gift-stock}-" + strconv.Itoa(int(order.GiftID))
	for {
		select {
		case <-ctx.Done():
			// 超时就需要回滚库存 发送回滚消息
			util.LogRus.Warnf("用户：%d 尝试扣减库存超时,error: %v", order.UserID, err)
			data, _ := json.Marshal(order)
			//if err := mq.SendCommonMQMessage("lottery_order_rollback_topic", data); err != nil {
			//	util.LogRus.Errorf("user:%d send mq message of order:%d rollback failed,redis need to rollback inventory", order.UserID, order.OrderID)
			//}
			if err := mq.MQManager.SendCommonMQMessages("lottery_order_rollback_topic", data); err != nil {
				util.LogRus.Errorf("user:%d send mq message of order:%d rollback failed,redis need to rollback inventory", order.UserID, order.OrderID)
			}
			// 超时或者错误就返回谢谢参与
			c.String(http.StatusOK, strconv.Itoa(database.EMPRY_GIFT))
			return

		case <-queryTick.C:
			// 间隔25毫秒查询一次redis,查询到扣减库存成功就设置扣除成功的状态(commit) 后续创建订单需要检查这个状态
			state := database.RedisQueryInventory(myredis.GetRedisClient(), strconv.Itoa(int(orderId)), stockKey)

			switch state {
			case 0:
				util.LogRus.Infof("**************用户：%d 扣减库存成功,giftId: %v,orderId: %v", order.UserID, order.GiftID, order.OrderID)
				c.String(http.StatusOK, strconv.Itoa(giftId)) // 扣减库存成功，返回给前端抽中的奖品id
				return
			case 1:
				util.LogRus.Infof("**************用户：%d 库存已经回滚,giftId: %v,orderId: %v", order.UserID, order.GiftID, order.OrderID)
				c.String(http.StatusOK, strconv.Itoa(database.EMPRY_GIFT))
				return
			case 2:
				util.LogRus.Infof("**************用户：%d 扣减库存失败,giftId: %v,orderId: %v", order.UserID, order.GiftID, order.OrderID)
				c.String(http.StatusOK, strconv.Itoa(database.EMPRY_GIFT))
				return
			case 3, 4:
				util.LogRus.Infof("**************用户：%d 正在尝试扣减库存,giftId: %v,orderId: %v", order.UserID, order.GiftID, order.OrderID)
				continue
			}
		}
	}
}

// CheckInventory 检查商品库存
func CheckInventory() (ids []int, counts []float64, err error) {
	// 1.从redis获取现在商品的库存
	gifts := database.GetAllGiftInventory()

	// 2.记录商品信息 (ids商品id切片 counts商品库存切片-库存对应抽中概率)
	ids = make([]int, 0, len(gifts))
	counts = make([]float64, 0, len(gifts))
	for _, gift := range gifts {
		if gift.Count > 0 {
			ids = append(ids, gift.Id)
			counts = append(counts, float64(gift.Count))
		}
	}

	// 2.1判断商品库存是否还有库存
	if len(ids) == 0 {
		err = errors.New("inventory is zero")
		return
	}

	return
}
