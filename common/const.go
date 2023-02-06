package common

const (
	WorkerIDBits     = uint64(12) // 10 bit 工作机器ID中的 5bit workerID
	DataCenterIDBits = uint64(3)  // 10 bit 工作机器ID中的 5bit dataCenterID
	SequenceBits     = uint64(13)

	MaxWorkerID     = int64(-1) ^ (int64(-1) << WorkerIDBits) //节点ID的最大值 用于防止溢出
	MaxDataCenterID = int64(-1) ^ (int64(-1) << DataCenterIDBits)
	MaxSequence     = int64(-1) ^ (int64(-1) << SequenceBits)

	TimeLeft = uint8(25) // timeLeft = workerIDBits + sequenceBits // 时间戳向左偏移量
	DataLeft = uint8(16) // dataLeft = dataCenterIDBits + sequenceBits
	WorkLeft = uint8(12) // workLeft = sequenceBits // 节点IDx向左偏移量
	// 2020-05-20 08:0:00 +0800 CST
	Twepoch = int64(15809923200000) // 常量时间戳(毫秒) 13
)

const (
	WorkIdPathPrefix = "/IDMaker/Id-"
	WorkIdPath       = "/IDMaker"
)

const (
//MaxRetryTimes = MaxWorkerID / 2
)

type ServerType int

const (
	ServerTypeDefault = iota
	ServerTypeZk
)
