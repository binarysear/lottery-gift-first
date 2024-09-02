package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gift/database"
	"gift/database/redis"
	"gift/util"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
	"strconv"
)

var (
	MQManager *RocketmqManager
)

type RocketmqManager struct {
	Producers            map[string]rocketmq.Producer
	TransactionProducers map[string]rocketmq.TransactionProducer
	Consumers            map[string]rocketmq.PushConsumer
}

func InitRocketmqManager() {
	MQManager = &RocketmqManager{
		Producers:            make(map[string]rocketmq.Producer),
		TransactionProducers: make(map[string]rocketmq.TransactionProducer),
		Consumers:            make(map[string]rocketmq.PushConsumer),
	}
	MQManager.InitCommonMQProducer()
	MQManager.InitTransactionMQProducer()
	MQManager.InitCommonMQConsumer()
	MQManager.InitTransactionMQConsumer()
	MQManager.InitCommonConfirmOrderMQConsumer()
}

// InitCommonMQProducer 初始化普通消息生产者
func (r *RocketmqManager) InitCommonMQProducer() {
	viper := util.CreateConfig("rocketmq")
	brokers := viper.GetString("brokers")
	producerCommonMQ, err = rocketmq.NewProducer(producer.WithNameServer([]string{brokers}))
	if brokers == "" || producerCommonMQ == nil {
		util.LogRus.Errorf("connect to mq failed %v", err)
		mqErr := errors.New("InitCommonMQProducer failed")
		panic(mqErr)
	}
	if err = producerCommonMQ.Start(); err != nil {
		mqErr := errors.New("InitCommonMQProducer Start failed")
		panic(mqErr)
	}
	r.Producers["commonMQMessageProducer"] = producerCommonMQ
}

// SendCommonMQMessages 发送普通消息
func (r *RocketmqManager) SendCommonMQMessages(topic string, body []byte) error {
	msg := primitive.NewMessage(topic, body)
	_, err := r.Producers["commonMQMessageProducer"].SendSync(context.Background(), msg)
	if err != nil {
		return err
	}
	return nil
}

// InitTransactionMQProducer 初始化事务消息生产者
func (r *RocketmqManager) InitTransactionMQProducer() {
	viper := util.CreateConfig("rocketmq")
	brokers := viper.GetString("brokers")
	groupName := viper.GetString("producerGroupName")
	OrderProducerMQ, CreateMqErr = rocketmq.NewTransactionProducer(
		NewOrderEntity(),
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{brokers})),
		producer.WithRetry(2),
		producer.WithGroupName(groupName), // 生产者组
	)
	if brokers == "" || OrderProducerMQ == nil {
		util.LogRus.Panicf("connect to mq failed %v", CreateMqErr)
		mqErr := errors.New("InitTransactionMQProducer failed")
		panic(mqErr)
	}
	if err := OrderProducerMQ.Start(); err != nil {
		mqErr := errors.New("InitTransactionMQProducer failed")
		panic(mqErr)
	}
	r.TransactionProducers["transactionMQMessageProducer"] = OrderProducerMQ
}

// SendTransactionMQMessage 发送事务消息
func (r *RocketmqManager) SendTransactionMQMessage(order database.LotteryOrder) error {
	// 1. 封装消息
	data, err := json.Marshal(order)
	if err != nil {
		util.LogRus.Errorf("序列化订单失败 error: %v", err)
		return errors.New("序列化订单失败")
	}
	viper := util.CreateConfig("rocketmq")
	topic := viper.GetString("transactionTopic")
	msg := &primitive.Message{
		Topic: topic,
		Body:  data,
	}

	// 2. 发送事务消息
	res, err := r.TransactionProducers["transactionMQMessageProducer"].SendMessageInTransaction(context.Background(), msg)
	if err != nil {
		//return errors.New("发送事务消息失败")
		return err
	}
	util.LogRus.Info("mq SendMessageInTransaction 成功", zap.Any("res", res))

	return nil
}

// InitCommonMQConsumer 初始化mq普通消息消费者
func (r *RocketmqManager) InitCommonMQConsumer() {
	util.InitLog("log") // 初始化日志
	viper := util.CreateConfig("rocketmq")
	brokers := viper.GetString("brokers")
	groupName := viper.GetString("consumerGroupName")
	topic := viper.GetString("rollbackTopic")
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{brokers}),
		consumer.WithGroupName(groupName), // 多个实例
	)
	if brokers == "" || topic == "" || c == nil {
		util.LogRus.Panicf("connect to mq failed %v", err)
		mqErr := errors.New("InitCommonMQConsumer failed")
		panic(mqErr)
	}

	err = c.Subscribe(topic, consumer.MessageSelector{}, r.RollbackInventoryMessageHandle)
	if err != nil {
		fmt.Println("读取消息失败")
	}
	if err = c.Start(); err != nil {
		mqErr := errors.New("InitCommonMQConsumer start failed")
		panic(mqErr)
	}
	r.Consumers["commonMQMessageConsumer"] = c
}

// RollbackInventoryMessageHandle 接收rocketmq的事务消息，回滚库存
func (r *RocketmqManager) RollbackInventoryMessageHandle(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		var data database.LotteryOrder
		err := json.Unmarshal(msgs[i].Body, &data)
		if err != nil {
			util.LogRus.Errorf("json.Unmarshal RollbackMsg failed", zap.Error(err))
			continue
		}
		util.LogRus.Infof("read common mq message -> %v", data)
		stockKey := "{lottery-gift-stock}-" + strconv.Itoa(int(data.GiftID))
		// 将库存回滚
		rdb := redis.GetRedisClient()
		orderId := strconv.Itoa(int(data.OrderID))
		err = database.RedisRollBackInventory(rdb, orderId, stockKey)
		if err != nil {
			// 如果库存回滚失败，则返回重试
			if !errors.Is(err, errors.New("Order already set to rollback.")) {
				// 判断是否重复回滚。不是重复回滚，消费失败重新尝试
				return consumer.ConsumeRetryLater, nil
			}
		}
		util.LogRus.Infof("redis: %v rollback inventory,count 1", stockKey)
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}

// InitTransactionMQConsumer 初始化mq事务消费者
func (r *RocketmqManager) InitTransactionMQConsumer() {
	util.InitLog("log") // 初始化日志
	viper := util.CreateConfig("rocketmq")
	brokers := viper.GetString("brokers")
	groupName := viper.GetString("transactionConsumerGroupName")
	topic := viper.GetString("createOrderTopic")
	util.LogRus.Infof("mq brokers %s,groupName %s,topic: %s", brokers, groupName, topic)
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{brokers}),
		consumer.WithGroupName(groupName), // 多个实例
	)
	util.LogRus.Infof("mq c %v", c)
	if brokers == "" || topic == "" || c == nil {
		util.LogRus.Panicf("connect to mq failed %v", err)
		mqErr := errors.New("InitTransactionMQConsumer failed")
		panic(mqErr)
	}

	err = c.Subscribe(topic, consumer.MessageSelector{}, r.TransactionMessagesHandle)
	if err != nil {
		fmt.Println("读取消息失败")
	}
	if err := c.Start(); err != nil {
		mqErr := errors.New("InitTransactionMQConsumer start failed")
		panic(mqErr)
	}
	r.Consumers["transactionMQMessageConsumer"] = c
}

// TransactionMessagesHandle 接收rocketmq的事务消息，生成订单
func (r *RocketmqManager) TransactionMessagesHandle(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		var order database.LotteryOrder
		err := json.Unmarshal(msgs[i].Body, &order)
		if err != nil {
			util.LogRus.Errorf("json.Unmarshal RollbackMsg failed", zap.Error(err))
			continue
		}
		util.LogRus.Infof("read transaction mq message -> %v", order)
		if err := database.CreateLotteryOrder(&order); err == nil {
			// 插入状态0订单成功
			return consumer.ConsumeSuccess, nil
		} else {
			// 如果订单插入失败，判断是否是重复订单
			if errors.Is(err, errors.New("duplicate")) {
				// 如果是重复状态订单错误，说明已经创建过了
				return consumer.ConsumeSuccess, nil
			}
			// 其他原因创建失败，需要重试
			util.LogRus.Errorf("create order:%v failed,error: %v", order, err)
			return consumer.ConsumeRetryLater, nil
		}
	}
	return consumer.ConsumeSuccess, nil
}

// InitCommonConfirmOrderMQConsumer 初始化mq确认订单普通消息消费者
// 检查配置是否有效
func (r *RocketmqManager) InitCommonConfirmOrderMQConsumer() {
	util.InitLog("log") // 初始化日志
	viper := util.CreateConfig("rocketmq")
	brokers := viper.GetString("brokers")
	commonGroupName := viper.GetString("commonGroupName")
	topic := viper.GetString("orderConfirmTopic")
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{brokers}),
		consumer.WithGroupName(commonGroupName), // 多个实例
	)
	if brokers == "" || commonGroupName == "" || topic == "" || c == nil {
		util.LogRus.Panicf("connect to mq failed %v", err)
		mqErr := errors.New("InitTransactionMQConsumer failed")
		panic(mqErr)
	}

	err = c.Subscribe(topic, consumer.MessageSelector{}, r.ConfirmOrderMessagesHandle)
	if err != nil {
		fmt.Println("读取消息失败")
	}
	if err := c.Start(); err != nil {
		mqErr := errors.New("InitTransactionMQConsumer start failed")
		panic(mqErr)
	}
	r.Consumers["commonConfirmOrderMQMessageConsumer"] = c
}

// ConfirmOrderMessagesHandle 接收rocketmq的事务消息，确认订单状态
func (r *RocketmqManager) ConfirmOrderMessagesHandle(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		var orderIdStr string
		err := json.Unmarshal(msgs[i].Body, &orderIdStr)
		if err != nil {
			util.LogRus.Errorf("json.Unmarshal RollbackMsg failed", zap.Error(err))
			continue
		}
		util.LogRus.Infof("read confirm order common mq message -> %v", orderIdStr)
		orderId, _ := strconv.Atoi(orderIdStr)
		err = database.UpdateOrderStateByOrderId(int64(orderId), 1)
		if err != nil {
			// 如果更新订单状态失败，则返回重试
			// 判断是否重复回滚。不是重复回滚，消费失败重新尝试
			return consumer.ConsumeRetryLater, nil
		}
		util.LogRus.Infof("update order:%v success", orderIdStr)
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}

// ShutdownMq 关闭mq
func (r *RocketmqManager) ShutdownMq() {
	for _, prod := range r.Producers {
		prod.Shutdown()
	}
	for _, tp := range r.TransactionProducers {
		tp.Shutdown()
	}
	for _, cons := range r.Consumers {
		cons.Shutdown()
	}
}
