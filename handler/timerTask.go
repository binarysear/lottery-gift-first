package handler

import (
	"context"
	"encoding/json"
	"gift/database"
	myredis "gift/database/redis"
	"gift/mq"
	"gift/util"
	"time"
)

// OrderTimeTask 定时任务，每隔5分钟检查一次，找出那些需要更改状态的订单，发送mq消息 之所以不在此更改是因为更改订单影响查找速度
func OrderTimeTask(stopChan chan struct{}) {
	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			// 1.先查询redis中那些成功扣减了库存的订单
			hashMap, err := myredis.GetRedisClient().HGetAll(context.Background(), "commit").Result()
			if err != nil {
				util.LogRus.Errorf("redis hashmap find error: %v", err)
			}

			// 2.遍历hashmap,更新订单状态为1表示确认订单
			for orderId, _ := range hashMap {
				// 2.1 发送mq消息同志消费者更改订单状态 发送失败可以重复发送，消费者做幂等性处理就行
				data, _ := json.Marshal(orderId)
				//if err := mq.SendCommonMQMessage("lottery_order_confirm_topic", data); err != nil {
				//	util.LogRus.Errorf("send confirm order common mq message error: %v", err)
				//}
				if err := mq.MQManager.SendCommonMQMessages("lottery_order_confirm_topic", data); err != nil {
					util.LogRus.Errorf("send confirm order common mq message error: %v", err)
				} else {
					// 发送成功就把其从redis中删除
					myredis.GetRedisClient().HDel(context.Background(), "commit", orderId)
				}
			}
		case <-stopChan:
			util.LogRus.Info("OrderTimeTask stop")
			return
		}
	}
}

// RollBackInventoryTimeTask 定时任务，每隔5分钟检查一次，找出那些因为订单超时但是还未回滚库存的订单
// 遍历redis中hashmap的toCommit,写lua脚本
// 先判断orderId在不在commit中，如果在表示用户抽奖成功了，不需要后续处理
// 不在就设置状态为rollback,然后回滚库存
func RollBackInventoryTimeTask(stopChan chan struct{}) {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			// 1.先查询redis中那些成功扣减了库存的订单
			hashMap, err := myredis.GetRedisClient().HGetAll(context.Background(), "toCommit").Result()
			if err != nil {
				util.LogRus.Errorf("redis hashmap find error: %v", err)
			}

			for orderId, _ := range hashMap {
				if err := database.CheckInventoryRollBack(myredis.GetRedisClient(), orderId); err != nil {
					util.LogRus.Errorf("CheckInventoryRollBack orderId:%v  error: %v", orderId, err)
				}
			}
		case <-stopChan:
			util.LogRus.Info("OrderTimeTask stop")
			return
		}
	}
}
