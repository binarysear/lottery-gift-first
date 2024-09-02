package util

import "math/rand"

func Lottory(probs []float64) int {
	if len(probs) == 0 {
		return -1
	}
	acc := make([]float64, 0, len(probs))
	sum := 0.0
	for _, prob := range probs {
		sum += prob
		acc = append(acc, sum)
	}

	// 从[0,sum)随机取一个数
	r := rand.Float64() * sum
	index := BinarySearch(acc, r)
	return index
}

// BinarySearch 二分查找有序排序切片中的目标所处的区间索引 例如 arr = [1，5，7，9]  target=2 即落在(1,5]之间，index为5
func BinarySearch(arr []float64, target float64) int {
	if len(arr) == 0 {
		return -1
	}
	begin, end := 0, len(arr)-1
	middle := 0

	for {
		// 优先进行边界判断 即arr范围之外
		// 左边边界
		if arr[begin] >= target {
			return begin
		}
		// 右边
		if arr[end] < target {
			return end + 1
		}

		// 如果begin 和 end只相差1需要特殊判断 不在边界两边 就在边界里面，当此条件满足时target就在begin和end中间
		// 无此判断会导致循环时出现begin一致等于middle死循环
		// 因为跟传统的二分不一样，当arr[middle]!=target时 begin和end不是直接+1或者-1 而是直接等于middle
		// 不能直接+1或者-1是因为现在要判断的是区间范围 例如 [1,4,5,7] target = 4.1 如果直接+1会直接跳过这个区间范围
		if begin == end-1 {
			return end
		}
		middle = (begin + end) / 2
		if arr[middle] < target {
			begin = middle
		} else if arr[middle] > target {
			end = middle
		} else {
			return middle
		}

	}
}
