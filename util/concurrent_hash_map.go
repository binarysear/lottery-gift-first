package util

import (
	farmhash "github.com/leemcloughlin/gofarmhash"
	"sync"
	"unsafe"
)

// 自行实现支持并发读写的map
type ConcurrentHashMap[T comparable] struct {
	chMap []map[T]any    // 有多个小map组成的map数组
	locks []sync.RWMutex // 每个小map配一把锁
	seg   int            // 有多少个小map
	seed  uint32         // 生成hash值需要的参数
}

// seg: 有多少个小map cap:预估存放多少个key-value
func NewConcurrentHashMap[T comparable](seg, cap int) *ConcurrentHashMap[T] {
	chMap := make([]map[T]any, seg)
	locks := make([]sync.RWMutex, seg)
	for i := 0; i < seg; i++ {
		chMap[i] = make(map[T]any, cap/seg)
		locks[i] = sync.RWMutex{}
	}
	return &ConcurrentHashMap[T]{
		chMap: chMap,
		locks: locks,
		seed:  0,
		seg:   seg,
	}
}

// 指针转整形
// 先把key转成unsafe.Point指针类型 unsafe.Pointer()可以存储任意类型的指针类型
// 然后强转成int指针类型 最后在取值  不安全，但是比反射快
func Point2Int[T comparable](key *T) int {
	return *(*int)(unsafe.Pointer(key))
}

// 获取key对应哪个小map
func (c *ConcurrentHashMap[T]) getSegIndex(key T) int {
	hash := int(farmhash.Hash32WithSeed(IntToBytes(Point2Int(&key)), c.seed))
	return hash & (c.seg - 1) // % x 等于 & (x -1)
}

func (c *ConcurrentHashMap[T]) Set(key T, value any) {
	index := c.getSegIndex(key)
	c.locks[index].Lock()
	defer c.locks[index].Unlock()
	c.chMap[index][key] = value
}

func (c *ConcurrentHashMap[T]) Get(key T) (any, bool) {
	index := c.getSegIndex(key)
	c.locks[index].RLock()
	defer c.locks[index].RUnlock()
	value, exists := c.chMap[index][key]
	return value, exists
}
