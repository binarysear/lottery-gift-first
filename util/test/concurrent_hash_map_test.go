package test

import (
	"gift/util"
	"math/rand"
	"sync"
	"testing"
)

var chMap = util.NewConcurrentHashMap[int64](8, 1000)
var sMap = sync.Map{}

func readConMap() {
	for i := 0; i < 1000; i++ {
		key := rand.Int63()
		chMap.Get(key)
	}
}

func writeConMap() {
	for i := 0; i < 1000; i++ {
		key := rand.Int63()
		chMap.Set(key, 1)
	}
}

func readSMap() {
	for i := 0; i < 1000; i++ {
		key := rand.Int63()
		sMap.Load(key)
	}
}

func writeSMap() {
	for i := 0; i < 1000; i++ {
		key := rand.Int63()
		sMap.Store(key, 1)
	}
}

func BenchmarkConMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		const P = 300
		wg.Add(2 * P)
		for j := 0; j < P; j++ {
			go func() {
				defer wg.Done()
				for k := 0; k < 10; k++ {
					readConMap()
				}
			}()
		}
		for j := 0; j < P; j++ {
			go func() {
				defer wg.Done()
				for k := 0; k < 10; k++ {
					writeConMap()
				}
			}()
		}

		wg.Wait()
	}
}

func BenchmarkSMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		const P = 300
		wg.Add(2 * P)
		for j := 0; j < P; j++ {
			go func() {
				defer wg.Done()
				for k := 0; k < 10; k++ {
					readSMap()
				}
			}()
		}
		for j := 0; j < P; j++ {
			go func() {
				defer wg.Done()
				for k := 0; k < 10; k++ {
					writeSMap()
				}
			}()
		}

		wg.Wait()
	}
}

// go test ./util/test -bench=Map -run=^$ -count=1 -benchmem -benchtime=3s
// goos: windows
//goarch: amd64
//pkg: gift/util/test
//cpu: AMD Ryzen 7 6800H with Radeon Graphics

//BenchmarkConMap-16             7         522154943 ns/op(测试函数平均耗时)        632557229 B/op(分配的内存)  18081422 allocs/op(内存分配次数)
//BenchmarkSMap-16               2        4045394400 ns/op        503018624 B/op  12055655 allocs/op
//PASS
//

// go test: 这是运行Go测试的命令。
//./util/test: 这指定要测试的包或目录。在这种情况下，它是util/test包。
//-bench=Map: 这个标志指定要运行的性能测试函数的名称。在这种情况下，它是Map函数。
//-run=^$: 这个标志指定要匹配测试名称的正则表达式模式。在这种情况下，它匹配所有测试。
//-count=1: 这个标志指定要运行测试的次数。在这种情况下，它设置为1。
//-benchmem: 这个标志启用了性能测试的内存分配分析。
//-benchtime=3s: 这个标志设置了每个性能测试的最大持续时间。在这种情况下，它设置为3秒。
