package snowFlake

import (
	"context"
	"log"
	"testing"
	"time"
)

func BenchmarkSnowflake(b *testing.B) {
	// 64bit = 8b
	ctx := context.Background()

	sonCtx, cancel := context.WithCancel(ctx)
	nums := 10000
	length := 10000
	ch := make(chan uint64, length+1)
	defer close(ch)
	errCh := make(chan error, 3)
	sf, _ := NewSfWorker(errCh)

	go countMap(sonCtx, ch)

	wg.Add(nums)
	b.ResetTimer()
	startTime := time.Now().Unix()
	for i := 0; i < nums; i++ {
		go func() {
			defer wg.Done()
			id, _ := sf.NextID()
			log.Println(id)
			ch <- id
		}()
	}
	wg.Wait()
	time.Sleep(time.Second)
	cancel()
	time.Sleep(time.Second)
	log.Println("timeDiff: ", time.Now().Unix()-startTime, " s")
	return
}

func countMap(ctx context.Context, ch chan uint64) {
	syncMap := make(map[uint64]int, 1)
	dupNum := 0
	unDupNum := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("dupNum: ", dupNum)
			log.Println("unDupNum", unDupNum)
			log.Println("done")
			return
		case data, _ := <-ch:
			if _, ok := syncMap[data]; ok {
				dupNum++
			} else {
				unDupNum++
			}
			syncMap[data] = 1
		default:

		}
	}
}

func TestNewSfWorker(t *testing.T) {
	errCh := make(chan error, 3)
	sf, _ := NewSfWorker(errCh)
	sf.NextID()
}
