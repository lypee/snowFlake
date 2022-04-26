package pool

import (
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/viper"
)

var ClientConnPool = sync.Pool{
	New: func() interface{} {
		servers := viper.GetStringSlice("Zookeeper.Servers")
		sessionTimeout := time.Second * viper.GetDuration("Zookeeper.SessionTimeout")
		conn, _, err := zk.Connect(servers, sessionTimeout)
		if err != nil {
			return nil
		}
		return conn
	},
}
