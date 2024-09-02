package test

import (
	"fmt"
	"gift/util"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestSnowflake(t *testing.T) {
	generator := util.NewWorkGenerator(5, 5)
	ch := make(chan int64, 10000)
	defer close(ch)
	count := 10000
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			id := generator.GeneratorID()
			ch <- id
		}()
	}

	wg.Wait()
	m := make(map[int64]int)
	for i := 0; i < count; i++ {
		id := <-ch
		_, ok := m[id]
		if ok {
			t.Error("id重复")
			return
		}
		m[id] = i
	}
	fmt.Println("ALL", len(m))
}
