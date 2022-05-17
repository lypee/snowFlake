package snowFlake

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/lypee/snowFlake/base"
	"github.com/lypee/snowFlake/common"
	"github.com/lypee/snowFlake/server/zkServer"

	"github.com/spf13/cast"
)

type SfWorker struct {
	zkSrv *zkServer.ZkServer

	mu           sync.Mutex
	lastStamp    int64 // 记录上一次ID的时间戳
	workerID     int64 // 该节点的ID
	dataCenterID int64 // 该节点的 数据中心ID
	sequence     int64 // 当前毫秒已经生成的ID序列号(从0 开始累加) 1毫秒内最多生成4096个ID

}

var (
	wg sync.WaitGroup
)

func NewSfWorker(ofs ...zkServer.ConnOptFunc) *SfWorker {
	//config.InitConfig("conf", "/conf.yaml")

	opt := zkServer.DefaultOpt()
	for _, op := range ofs {
		op(opt)
	}
	errCh := make(chan error, 3)
	zkSrv := zkServer.NewZkServer(errCh, opt)
	workId, _ := zkSrv.GetWorkerId()
	// initialization
	sfWorker := newWorker(int64(workId), 1, zkSrv)

	// start-monitor
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	return sfWorker
	//go SfWorker.monitor(errCh, c)
}

// newWorker 分布式情况下 通过外部配置文件或其他方式为个worker分配独立的id
// eg: 静态配置文件、zk发号、redis发号
func newWorker(workerID, dataCenterID int64, zkSrv *zkServer.ZkServer) *SfWorker {
	if workerID > common.MaxWorkerID || workerID < 0 {
		workerID = common.MaxWorkerID
	}
	if dataCenterID > common.MaxDataCenterID || dataCenterID < 0 {
		dataCenterID = common.MaxDataCenterID
	}
	//log.Println(fmt.Sprintf("newWorker,workerId:[%d] ,dataCenterId:[%d]", workerID, dataCenterID))
	return &SfWorker{
		workerID:     workerID,
		lastStamp:    0,
		sequence:     0,
		dataCenterID: dataCenterID,
		zkSrv:        zkSrv,
	}
}

func (w *SfWorker) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func (w *SfWorker) NextID() (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.nextID()
}

func (w *SfWorker) nextID() (uint64, error) {
	timeStamp := w.getMilliSeconds()
	if timeStamp < w.lastStamp {
		return 0, errors.New("time is moving backwards,waiting until")
	}

	if w.lastStamp == timeStamp {
		w.sequence = (w.sequence + 1) & common.MaxSequence
		if w.sequence == 0 {
			for timeStamp <= w.lastStamp {
				timeStamp = w.getMilliSeconds()
			}
		}
	} else {
		w.sequence = 0
	}

	w.lastStamp = timeStamp
	id := ((timeStamp - common.Twepoch) << common.TimeLeft) |
		(w.dataCenterID << common.DataLeft) |
		(w.workerID << common.WorkLeft) | w.sequence

	return uint64(id), nil
}

// monitor -
func (w *SfWorker) monitor(errCh chan error, sigCh chan os.Signal) {
	for {
		select {
		case err := <-errCh:
			base.WarningF("%v", err)
		case s := <-sigCh:
			base.InfoF("receive signal %v", s)
			//app.GetApplication().Close()
			w.zkSrv.RemoveNode(common.WorkIdPathPrefix, cast.ToString(w.workerID))
			os.Exit(0)
		}
	}
}
