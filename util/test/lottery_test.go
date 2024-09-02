package test

import (
	"fmt"
	"gift/util"
	"math/rand"
	"slices"
	"testing"
)

func TestBinarySearch(t *testing.T) {
	L := 100
	for j := 0; j < 30; j++ {
		arr := make([]float64, 0, L)
		for i := 0; i < L; i++ {
			arr = append(arr, rand.Float64())
		}

		slices.Sort(arr)
		var target float64
		// 先测试两个边界
		target = arr[0] - 1.0
		if util.BinarySearch(arr, target) != 0 {
			t.Fail()
		}
		target = arr[len(arr)-1] + 1.0
		if util.BinarySearch(arr, target) != len(arr) {
			t.Fail()
		}

		target = arr[0]
		if util.BinarySearch(arr, target) != 0 {
			t.Fail()
		}

		// 测试每个点和中间点
		for i := 0; i < len(arr)-1; i++ {
			target = (arr[i] + arr[i+1]) / 2
			if util.BinarySearch(arr, target) != i+1 {
				t.Fail()
			}
			target = arr[i+1]
			if util.BinarySearch(arr, target) != i+1 {
				t.Fail()
			}
		}

	}
}

func TestLottory(t *testing.T) {
	probs := []float64{5, 2, 4}
	countMap := make(map[int]float64, len(probs))
	for i := 0; i < len(probs); i++ {
		countMap[i] = 0
	}
	for i := 0; i < 100; i++ {
		index := util.Lottory(probs)
		countMap[index] += 1
	}
	fmt.Println(countMap[0] / probs[0])
	fmt.Println(countMap[1] / probs[1])
	fmt.Println(countMap[2] / probs[2])
}
