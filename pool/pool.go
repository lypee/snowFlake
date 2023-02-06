package pool

import (
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
)

var (
	duration time.Duration
)

func init() {
	duration = viper.GetDuration("Zookeeper.SessionTimeout")
	if duration < time.Duration(2) {
		duration = time.Duration(2)
	}
}

var ClientConnPool = sync.Pool{
	New: func() interface{} {
		servers := viper.GetStringSlice("Zookeeper.Servers")
		sessionTimeout := time.Second * duration
		conn, _, err := zk.Connect(servers, sessionTimeout)
		if err != nil {
			return nil
		}
		return conn
	},
}
