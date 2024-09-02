package database

import (
	"context"
	"errors"
	myredis "gift/database/redis"
	"gift/util"
	"github.com/dtm-labs/dtmcli/logger"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	prefix = "{lottery-gift-stock}-" // 所有redis key设置统一的前缀，方便后续按照前缀遍历key
)

// InitGiftInventory
// 从Mysql中读出所有奖品的初始库存,存入Redis，如果同时又跟多用户来参与抽奖活动，不能交发去Mysql里减库存，mysql扛不住这么高的并发量，redis可以抗住
func InitGiftInventory() {
	giftCh := make(chan *Gift, 100)
	go GetAllGiftsV2(giftCh)

	client := myredis.GetRedisClient()
	for {
		gift, ok := <-giftCh
		if !ok { // channel已经消费完，并且channel已经关闭
			break
		}

		// 剔除掉库存为0的商品
		if gift.Count <= 0 {
			continue
		}

		err := client.Set(context.Background(), prefix+strconv.Itoa(gift.Id), gift.Count, 0).Err() // 不设置过期时间
		if err != nil {
			util.LogRus.Panicf("set gift %d:%s count to %d failed: %s", gift.Id, gift.Name, gift.Count, err)
		}
	}

}

// GetAllGiftInventory
// 从redis获取所有奖品剩余的库存
func GetAllGiftInventory() []*Gift {
	client := myredis.GetRedisClient()
	keys, err := client.Keys(context.Background(), prefix+"*").Result() // 根据前缀获取所有奖品的key
	if err != nil {
		util.LogRus.Errorf("iterate all keys by prefix %s failed: %s", prefix, err)
		return nil
	}

	gifts := make([]*Gift, 0, len(keys))
	for _, key := range keys {
		giftId, err := strconv.Atoi(key[len(prefix):])
		if err != nil {
			util.LogRus.Errorf("invalid redis key %s", key)
		}
		count, err := client.Get(context.Background(), key).Int()
		if err != nil {
			util.LogRus.Errorf("invalid gift inventory %s", client.Get(context.Background(), key).String())
		}
		gifts = append(gifts, &Gift{Id: giftId, Count: count})
	}

	return gifts
}

// RedisCheckAdjustAmount 扣减库存
// 先查找有没有对应的商品，有就接着判断商品库存是否足够，足够就扣减库存，否则返回错误
// 如果此订单号是第一次扣减成功，就插入一条数据到redis的hashmap中的toCommit中，key为orderId,value为商品id表示该用户抽中了商品
// 如果此订单号不是第一次扣减，就返回重复错误，防止重复扣减
func RedisCheckAdjustAmount(rd *redis.Client, orderId, key string, amount int) error {
	v, err := rd.Eval(rd.Context(), ` -- RedisCheckAdjustAmount
local v = redis.call('GET', KEYS[1])
local e1 = redis.call('HGET','toCommit', KEYS[2])

if v == false or v + ARGV[1] < 0 then
	return 'FAILURE'
end

if e1 ~= false then
	return 'DUPLICATE'
end

redis.call('HSET','toCommit', KEYS[2], KEYS[1])

redis.call('INCRBY', KEYS[1], ARGV[1])
`, []string{key, orderId}, amount).Result()
	logger.Debugf("lua return v: %v err: %v", v, err)
	if err == redis.Nil {
		err = nil
	}
	if err == nil && v == "DUPLICATE" {
		return errors.New("DUPLICATE Reduce Inventory")
	}
	if err == nil && v == "FAILURE" {
		err = errors.New("FAILURE")
	}
	return err
}

// RedisCheckBack 用于rocketmq事务消息做库存回查，检查库存是否扣减成功
// 主要就是看orderId有没有找到对应的值，没有找到表示本地事务扣减库存失败，设置orderId对应的value为rollback
// 如果是rollback表示已经回滚过了不需要操作
func RedisCheckBack(rd *redis.Client, orderId string, barrierExpire int) error {
	v, err := rd.Eval(rd.Context(), ` -- RedisQueryPrepared
local v = redis.call('HGET','toCommit', KEYS[1])
if v == false then
	redis.call('HSET','toCommit', KEYS[1], "rollback")
	v = 'rollback'
end
if v == 'rollback' then
	return 'FAILURE'
end
`, []string{orderId}, barrierExpire).Result()
	logger.Debugf("lua return v: %v err: %v", v, err)
	if err == redis.Nil {
		err = nil
	}
	if err == nil && v == "FAILURE" {
		err = errors.New("FAILURE")
	}
	return err
}

// RedisRollBackInventory redis回滚库存
// 先查看orderId对应的value值是否是rollback，不是则设置成rollback，然后回滚库存
// 是rollback就代表已经回滚过了不需要操作
func RedisRollBackInventory(rd *redis.Client, orderId, key string) error {
	v, err := rd.Eval(rd.Context(), `-- Lua脚本：检查并设置orderId的值，检查stockKey的值，然后增加库存
local function checkSetOrderIdAndIncrStockIfNotRollback()
    -- 获取orderId的当前值
    local orderIdValue = redis.call('HGET','toCommit', KEYS[1])

    -- 检查orderId的值是否已经是"rollback"
    if orderIdValue ~= "rollback" then
        -- 如果不是"rollback"，则设置为"rollback"
        redis.call('HSET','toCommit', KEYS[1], "rollback")
		-- 回滚库存
		redis.call('INCRBY', KEYS[2], 1)
    else
        -- 如果orderId已经是"rollback"，则不进行任何操作
        return 'INFO: OrderId already set to rollback.'
    end
end

-- 调用Lua函数
return checkSetOrderIdAndIncrStockIfNotRollback()`, []string{orderId, key}).Result()
	if err == redis.Nil {
		err = nil
	}
	if err == nil && v == "INFO: OrderId already set to rollback." {
		err = errors.New("Order already set to rollback.")
	}
	return err
}

// RedisQueryInventory 检查redis是扣减成功
// 返回状态值0:成功，1:回滚，2:失败，3:未知
// 主要是orderId有没有在toCommit中找到对应的数据，找到就接着判断状态值value，如果value跟用户抽中的奖品号码stockKey一样代表扣减成功，如果value是rollback表示已经回滚
// 没有就表示可能还没扣减
func RedisQueryInventory(rd *redis.Client, orderId, stockKey string) int {
	v, err := rd.Eval(rd.Context(), `
-- Lua脚本：检查并设置orderId的值，检查stockKey的值，然后增加库存

local function RedisQueryInventory()
    -- 获取orderId的当前值
    local orderIdValue = redis.call('HGET', 'toCommit', KEYS[1])

    -- 检查orderIdValue是否为空
    if orderIdValue == nil then
        return 'unknown'
    end

    -- 检查orderIdValue是否为'rollback'
    if orderIdValue == 'rollback' then
        return 'rollback'
    end

    -- 检查orderIdValue是否与KEYS[2]相等
    if orderIdValue ~= KEYS[2] then
        return 'failure'
    else
        -- 如果相等，则代表成功扣减库存，移交信息到commit等待确认订单
        redis.call('HSET', 'commit', KEYS[1], "commit")
        redis.call('HDEL', 'toCommit', KEYS[1])
        return 'success'
    end
end

-- 调用Lua函数
return RedisQueryInventory()`, []string{orderId, stockKey}).Result()

	value := v.(string)
	if err == nil || err == redis.Nil {
		if value == "success" {
			return 0
		} else if value == "rollback" {
			return 1
		} else if value == "failure" {
			return 2
		} else if value == "unknown" {
			return 3
		}
	}
	// 其他错误
	return 4
}

// CheckInventoryRollBack 定时任务检查库存是否需要执行回滚
// 1.先判断orderId在toCommit中的状态，如果是rollback表示已经回滚过库存，如果是其他状态接着往下判断
// 2.判断orderId在不在commit中，如果在表示用户抽奖成功了，不需要后续处理
// 3.不在就设置状态为rollback,然后回滚库存
func CheckInventoryRollBack(rd *redis.Client, orderId string) error {
	v, err := rd.Eval(rd.Context(), `-- Lua脚本：检查并设置orderId的值，检查stockKey的值，然后增加库存
local function RedisQueryInventory()
    -- 获取orderId的当前值
    local orderIdValue = redis.call('HGET','toCommit', KEYS[1])

	-- 1.判断orderId在toCommit中的状态，如果是rollback表示已经回滚过库存，如果是空代表已经移交确认不需要处理，如果是其他状态接着往下判断
	if orderIdValue == "rollback" or orderIdValue == nil then 
	   --已经回滚或者确认直接返回
		return 'success'
	end

    -- 2.判断orderId在不在commit中，如果在表示用户抽奖成功了，不需要后续处理
	local orderIdCommitValue = redis.call('HGET','commit', KEYS[1])
    if orderIdCommitValue == nil then
		-- 3.不在就设置状态为rollback,然后回滚库存
		redis.call('HSET','toCommit', KEYS[1], "rollback")
		-- 回滚库存
		redis.call('INCRBY', orderIdValue, 1)
	else
		return 'success'
    end
end

-- 调用Lua函数
return RedisQueryInventory()`, []string{orderId}).Result()
	if err == redis.Nil && v == "success" {
		return nil
	}
	return err
}
