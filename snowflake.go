package main

import (
	"errors"
	"github.com/spf13/cast"
	base "lpynnng/engineering/snowFlake/base"
	"lpynnng/engineering/snowFlake/common"
	"lpynnng/engineering/snowFlake/config"
	"lpynnng/engineering/snowFlake/server/zkServer"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Worker struct {
	ZkSrv *zkServer.ZkServer

	mu           sync.Mutex
	LastStamp    int64 // 记录上一次ID的时间戳
	WorkerID     int64 // 该节点的ID
	DataCenterID int64 // 该节点的 数据中心ID
	Sequence     int64 // 当前毫秒已经生成的ID序列号(从0 开始累加) 1毫秒内最多生成4096个ID

}

var (
	SfWorker *Worker
	wg       sync.WaitGroup
)

func init() {
	config.InitConfig("conf", "/conf.yaml")

	errCh := make(chan error, 3)

	// zk-workerId
	zkSrv := zkServer.NewZkServer(errCh)
	workId, _ := zkSrv.GetWorkerId()

	// initialization
	SfWorker = newWorker(int64(workId), 1, zkSrv)

	// start-monitor
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go SfWorker.Monitor(errCh, c)
}

// newWorker 分布式情况下 通过外部配置文件或其他方式为个worker分配独立的id
// eg: 静态配置文件、zk发号、redis发号
func newWorker(workerID, dataCenterID int64, zkSrv *zkServer.ZkServer) *Worker {
	if workerID > common.MaxWorkerID || workerID < 0 {
		workerID = common.MaxWorkerID
	}
	if dataCenterID > common.MaxDataCenterID || dataCenterID < 0 {
		dataCenterID = common.MaxDataCenterID
	}
	//log.Println(fmt.Sprintf("newWorker,workerId:[%d] ,dataCenterId:[%d]", workerID, dataCenterID))
	return &Worker{
		WorkerID:     workerID,
		LastStamp:    0,
		Sequence:     0,
		DataCenterID: dataCenterID,
		ZkSrv:        zkSrv,
	}
}

func (w *Worker) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func (w *Worker) NextID() (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.nextID()
}

func (w *Worker) nextID() (uint64, error) {
	timeStamp := w.getMilliSeconds()
	if timeStamp < w.LastStamp {
		return 0, errors.New("time is moving backwards,waiting until")
	}

	if w.LastStamp == timeStamp {
		w.Sequence = (w.Sequence + 1) & common.MaxSequence
		if w.Sequence == 0 {
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMilliSeconds()
			}
		}
	} else {
		w.Sequence = 0
	}

	w.LastStamp = timeStamp
	id := ((timeStamp - common.Twepoch) << common.TimeLeft) |
		(w.DataCenterID << common.DataLeft) |
		(w.WorkerID << common.WorkLeft) | w.Sequence

	return uint64(id), nil
}

func (w *Worker) Monitor(errCh chan error, sigCh chan os.Signal) {
	for {
		select {
		case err := <-errCh:
			base.WarningF("%v", err)
		case s := <-sigCh:
			base.InfoF("receive signal %v", s)
			//app.GetApplication().Close()
			w.ZkSrv.RemoveNode(common.WorkIdPathPrefix, cast.ToString(w.WorkerID))
			os.Exit(0)
		}
	}
}
