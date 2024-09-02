package mq

import (
	"encoding/json"
	"errors"
	"gift/database"
	myredis "gift/database/redis"
	"gift/util"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"strconv"
)

var (
	OrderProducerMQ rocketmq.TransactionProducer
	CreateMqErr     error
)

type OrderEntity struct {
}

func NewOrderEntity() *OrderEntity {
	return &OrderEntity{}
}

// ExecuteLocalTransaction  本地事务在此执行
func (o *OrderEntity) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	var order = &database.LotteryOrder{}
	if err := json.Unmarshal(msg.Body, &order); err != nil {
		util.LogRus.Error("Unmarshal failed", zap.Error(err))
		return primitive.RollbackMessageState
	}
	stockKey := "{lottery-gift-stock}-" + strconv.Itoa(int(order.GiftID))
	// 1.从redis扣减库存
	err := database.RedisCheckAdjustAmount(myredis.GetRedisClient(), strconv.Itoa(int(order.OrderID)), stockKey, -1)
	if err != nil {
		// 库存扣减失败，丢弃half-message
		util.LogRus.Errorf("user:%d Failed to deduct inventory from redis,key: %s,orderId:%s", order.UserID, stockKey, order.OrderID)
		return primitive.RollbackMessageState
	}
	// 生成订单成功，丢弃half-message
	return primitive.CommitMessageState
}

// CheckLocalTransaction 检查本地事务是否执行成功
func (o *OrderEntity) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	// 0.解析订单参数
	var order = &database.LotteryOrder{}
	if err := json.Unmarshal(msg.Body, &order); err != nil {
		util.LogRus.Error("Unmarshal failed", zap.Error(err))
		return primitive.RollbackMessageState
	}
	// 1.检查redis是否扣减了库存
	if err := database.RedisCheckBack(myredis.GetRedisClient(), strconv.Itoa(int(order.OrderID)), 7*86400); err != nil {
		// 本地事务从redis扣减库存失败，丢弃half-message
		if errors.Is(err, errors.New("FAILURE")) {
			util.LogRus.Errorf("user:%d create order:%s failed", order.UserID, order.OrderID)
			return primitive.RollbackMessageState
		}
	}
	// 从redis扣减库存成功，提交half-message
	return primitive.CommitMessageState
}
