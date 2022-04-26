package zkServer

import (
	"log"
	"lpynnng/engineering/snowFlake/common"
	"lpynnng/engineering/snowFlake/config"
	"lpynnng/engineering/snowFlake/utils"
	"sync"
	"testing"
	"time"
)

func init() {
	config.InitConfig("../../conf", "/conf.yaml")
	errCh := make(chan error, 3)
	zkSrv = NewZkServer(errCh)
}

var (
	zkSrv *ZkServer
)

func TestZkServer_GetWorkIdWithPool(t *testing.T) {
	nums := 500
	ids := make([]int , 0 , nums)
	rwLock := sync.RWMutex{}
	wg := &sync.WaitGroup{}
	wg.Add(nums)
	for i := 0 ; i < nums ; i++{
		go func() {
			defer wg.Done()
			id, err := zkSrv.GetWorkerIdWithPool()
			if err != nil {
				log.Println("err: ", err.Error())
			}
			rwLock.Lock()
			defer rwLock.Unlock()
			ids = append(ids , id)
		}()
	}
	wg.Wait()
	time.Sleep(10 * time.Second)
	log.Println("ids : ", ids )
}

func TestZkServer_GetWorkId(t *testing.T) {
	nums := 1000
	ids := make([]int , 0 , nums)
	rwLock := sync.RWMutex{}
	wg := &sync.WaitGroup{}
	wg.Add(nums)
	for i := 0 ; i < nums ; i++{
		go func() {
			defer wg.Done()
			id, err := zkSrv.GetWorkerId()
			if err != nil {
				log.Println("err: ", err.Error())
			}
			rwLock.Lock()
			defer rwLock.Unlock()
			ids = append(ids , id)
		}()
	}
	wg.Wait()
	time.Sleep(10 * time.Second)
	log.Println("ids : ", ids )
}

func TestZkServer_Test(t *testing.T) {
	mmap := make(map[int]struct{}, 0)
	log.Println(int(common.MaxWorkerID))
	for i := 0; i < 100000; i++ {
		workId := utils.RandomNum(0, int(common.MaxWorkerID))
		mmap[workId] = struct{}{}
	}

	log.Println(len(mmap))
}

func TestZkServer_RemoveAllNode(t *testing.T) {
	zkSrv.RemoveAllNode(common.WorkIdPath)
}
