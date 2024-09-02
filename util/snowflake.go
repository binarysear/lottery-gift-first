package util

import (
	"sync"
	"time"
)

const (
	WorkIDBits     = uint64(5)  // 工作机器id 5bit
	DataCenterBits = uint64(5)  // 工作机器数据中心 5bit
	SequenceBits   = uint64(12) // 序列号 12bit

	WorkIDBitsMax     = -1 ^ (-1 << WorkIDBits)     // 工作机器ID最大值
	DataCenterBitsMax = -1 ^ (-1 << DataCenterBits) // 工作机器数据中心最大值
	SequenceBitsMax   = -1 ^ (-1 << SequenceBits)   // 序列号最大值

	TimeShift           = int64(22) // 时间戳偏移量
	DataCenterBitsShift = int(17)   // 数据中心偏移量
	WorkIDBitsShift     = int(12)   // 工作机器id偏移量

	Twepech   = int64(1692704479) // 当前时间的时间戳 毫秒
	MaxLength = 2000              // 最大允许的时间回拨的毫秒数
)

var SequenceHistory = make([]int64, MaxLength) // 缓存近2s内的每毫秒最大序列号

type Work struct {
	mu         sync.RWMutex
	LastStamp  int64 // 上次生成id的时间
	Sequence   int64 // 序列号
	WorkID     int64 // 工作机器id
	DataCenter int64 // 数据中心id
}

var (
	OrderIdGenerator *Work
)

func NewWorkGenerator(workID, dataCenter int64) *Work {
	if workID < 0 || workID > WorkIDBitsMax {
		panic("机器id太大")
	}
	if dataCenter < 0 || dataCenter > DataCenterBitsMax {
		panic("数据中心id太大")
	}
	return &Work{
		LastStamp:  0,
		Sequence:   0,
		WorkID:     workID,
		DataCenter: dataCenter,
	}
}

// getMillisSecond 获取当前毫秒数
func (w *Work) getMillisSecond() int64 {
	return time.Now().UnixMilli()
}

// GeneratorID 生成ID
func (w *Work) GeneratorID() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	// 1.获取当前时间
	timeStamp := w.getMillisSecond()
	// 索引记录当前毫秒数下最大序列号所在数组的位置
	index := int(timeStamp % MaxLength)

	// 2.出现时间回拨
	if timeStamp < w.LastStamp {
		// 判断是否在缓存时间范围内
		if w.LastStamp-timeStamp > MaxLength {
			panic("时间会退范围太大，超过2000ms缓存")
		}

		w.Sequence = 0
		// 在缓存范围内
		for {
			// 拿到之前毫秒数生成的最大序列号
			preSequence := SequenceHistory[index]
			// 判断之前的序列号是否已经到达最大值
			w.Sequence = SequenceBitsMax & (preSequence + 1)
			// 如果到达最大值，重新获取毫秒数
			if w.Sequence == 0 {
				timeStamp = w.getMillisSecond()
				index = int(timeStamp % MaxLength)
			} else {
				// 没有到达最大值，更新缓存，生成id
				SequenceHistory[index] = w.Sequence
				id := ((timeStamp - Twepech) << TimeShift) |
					(w.DataCenter << DataCenterBitsShift) |
					(w.WorkID << WorkIDBitsShift) |
					w.Sequence
				return int64(id)
			}
			// 结束循环条件
			if timeStamp >= w.LastStamp {
				break
			}
		}
	}

	// 3.没有出现时间回拨
	// 如果时间相等
	if timeStamp == w.LastStamp {
		// 判断生成的序列号是否超出范围
		w.Sequence = (w.Sequence + 1) & SequenceBitsMax
		// 如果超出范围
		if w.Sequence == 0 {
			// 当前毫秒生成序列号的个数已满，重新获取下一毫秒
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMillisSecond()
				index = int(timeStamp % MaxLength)
			}
		}
	} else {
		// 时间大于上次生成id的时间，重置序列号sequence
		w.Sequence = 0
	}

	// 刷新缓存
	SequenceHistory[index] = w.Sequence
	w.LastStamp = timeStamp

	id := ((timeStamp - Twepech) << TimeShift) |
		(w.DataCenter << DataCenterBitsShift) |
		(w.WorkID << WorkIDBitsShift) |
		w.Sequence
	return int64(id)
}
