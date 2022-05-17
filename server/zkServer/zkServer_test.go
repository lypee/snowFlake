package zkServer

import (
	"github.com/lypee/snowFlake/config"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/lypee/snowFlake/base"
	"github.com/samuel/go-zookeeper/zk"

	"github.com/lypee/snowFlake/common"
	"github.com/lypee/snowFlake/utils"
)

func init() {
	config.InitConfig("../.. /conf", "/conf.yaml")
	errCh := make(chan error, 3)
	opt := DefaultOpt()
	opt.servers = []string{"43.138.36.75:2181"}
	zkSrv = NewZkServer(errCh, opt)
}

var (
	zkSrv *ZkServer
)

func TestZkServer_GetWorkIdWithPool(t *testing.T) {
	nums := 500
	ids := make([]int, 0, nums)
	rwLock := sync.RWMutex{}
	wg := &sync.WaitGroup{}
	wg.Add(nums)
	for i := 0; i < nums; i++ {
		go func() {
			defer wg.Done()
			id, err := zkSrv.GetWorkerIdWithPool()
			if err != nil {
				log.Println("err: ", err.Error())
			}
			rwLock.Lock()
			defer rwLock.Unlock()
			ids = append(ids, id)
		}()
	}
	wg.Wait()
	time.Sleep(10 * time.Second)
	log.Println("ids : ", ids)
}

func TestZkServer_GetWorkId(t *testing.T) {
	nums := 100
	ids := make([]int, 0, nums)
	rwLock := sync.RWMutex{}
	wg := &sync.WaitGroup{}
	wg.Add(nums)
	for i := 0; i < nums; i++ {
		go func() {
			defer wg.Done()
			id, err := zkSrv.GetWorkerId()
			if err != nil {
				log.Println("err: ", err.Error())
			}
			rwLock.Lock()
			defer rwLock.Unlock()
			ids = append(ids, id)
		}()
	}
	wg.Wait()
	time.Sleep(10 * time.Second)
	log.Println("ids : ", ids)
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

func TestZkServer_CreateProtectedEphemeralSequential(t *testing.T) {
	c, _, err := zk.Connect([]string{"43.138.36.75:2181"}, time.Minute)
	if err != nil {
		base.ErrorF("err:[%+v]", err)
	}
	str1, err := c.Create(utils.SpliceString(common.WorkIdPathPrefix, strconv.Itoa(123)), []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		base.InfoF("err1:[%+v]", err)
	}
	str2, err := c.Create(utils.SpliceString(common.WorkIdPathPrefix, strconv.Itoa(123)), []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		base.InfoF("err2:[%+v]", err)
	}
	str3, err := c.Create(utils.SpliceString(common.WorkIdPathPrefix, strconv.Itoa(124)), []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		base.InfoF("err3:[%+v]", err)
	}
	str4, err := c.Create(utils.SpliceString(common.WorkIdPathPrefix, strconv.Itoa(123)), []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		base.InfoF("err4:[%+v]", err)
	}
	log.Println(str1)
	log.Println(str2)
	log.Println(str3)
	log.Println(str4)
}

func TestZkServer_RemoveAllNode(t *testing.T) {
	zkSrv.RemoveAllNode(common.WorkIdPath)
}

func TestZkServer_validatePath(t *testing.T) {
	log.Println(zkSrv.validatePath(common.WorkIdPathPrefix, false))
}


func TestZkServer_createFatherNode(t *testing.T) {
	//str := "\base\"
	//log.Println(strings.Split(common.WorkIdPathPrefix , "/"))
	str := "/IDMfr/222/123/das/fa/dasa"
	zkSrv.createFatherNode(str)
}