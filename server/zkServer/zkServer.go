package zkServer

import (
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/spf13/cast"

	"github.com/lypee/snowFlake/base"
	"github.com/lypee/snowFlake/common"
	"github.com/lypee/snowFlake/pool"
	"github.com/lypee/snowFlake/utils"

	"github.com/samuel/go-zookeeper/zk"
)

type connOpt struct {
	addr           string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	sessionTimeout time.Duration
	servers        []string
}

func DefaultOpt() *connOpt {
	return &connOpt{
		readTimeout:    3 * time.Second,
		writeTimeout:   3 * time.Second,
		sessionTimeout: 3 * time.Second,
		servers:        []string{},
	}
}

func WithAddr(addr string) ConnOptFunc {
	return func(opt *connOpt) {
		opt.addr = addr
	}
}

func WithServers(servers []string) ConnOptFunc {
	return func(opt *connOpt) {
		opt.servers = servers
	}
}

func WithReadTimeout(d time.Duration) ConnOptFunc {
	return func(opt *connOpt) {
		opt.readTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) ConnOptFunc {
	return func(opt *connOpt) {
		opt.writeTimeout = d
	}
}

func WithSessionTimeout(d time.Duration) ConnOptFunc {
	return func(opt *connOpt) {
		opt.sessionTimeout = d
	}
}

type ConnOptFunc func(opt *connOpt)

type ZkServer struct {
	lock  sync.RWMutex
	errCh chan error
	opt   *connOpt
}

func NewZkServer(errCh chan error, opt *connOpt) *ZkServer {
	return &ZkServer{
		opt:   opt,
		errCh: errCh,
	}
}

// GetWorkerIdWithPool -
func (srv *ZkServer) GetWorkerIdWithPool() (id int, err error) {
	srv.lock.Lock()
	defer func() {
		srv.lock.Unlock()
	}()
	conn := pool.ClientConnPool.Get()
	if c, ok := conn.(*zk.Conn); ok {
		var exist bool
		timeNow := time.Now()
		for i := 0; i < int(common.MaxWorkerID/2); i++ {
			base.InfoF("retry: %d times", i)
			workId := utils.RandomNum(0, int(common.MaxWorkerID))
			path := common.WorkIdPathPrefix + strconv.Itoa(workId)
			exist, _, err = c.Exists(path)
			if err != nil {
				base.ErrorF("path %v not exist", path)
				return 0, err
			}
			if exist {
				base.InfoF("path [%+v] exist", path)
				continue
			}
			_, err = c.Create(path, utils.Int64ToBytes(timeNow.Unix()), 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				base.ErrorF("set path: %v fail", path)
				return 0, err
			}

			base.InfoF("set path: %v success", path)
			return workId, nil
		}
	}
	pool.ClientConnPool.Put(conn)
	return 0, common.ConnErr.WithTrueErr(err)
}

// GetWorkerId -
func (srv *ZkServer) GetWorkerId() (id int, err error) {
	if len(srv.opt.servers) < 1 {
		return 0, common.ServersErr
	}
	srv.lock.Lock()
	defer srv.lock.Unlock()

	servers := srv.opt.servers
	sessionTimeout := srv.opt.sessionTimeout

	c, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		base.ErrorF("zk.Connect-err:[%+v]", err)
		return 0, common.StartConnErr.WithTrueErr(err)
	}
	var exist bool
	for i := 0; i < int(common.MaxWorkerID/2); i++ {
		base.InfoF("retry: %d times", i)
		workId := utils.RandomNum(0, int(common.MaxWorkerID))
		path := common.WorkIdPathPrefix + strconv.Itoa(workId)
		// check path valid
		if valid, err := srv.validatePath(path, false); !valid || err != nil {
			base.ErrorF("validatePath-fail:[%+v]", path, err)
		}
		exist, _, err = c.Exists(path) // todo handle father node if not exist
		if err != nil {
			base.ErrorF("c.Exist-err:[%+v]", err, path)
			return 0, err
		}
		if exist {
			// try create father node

			base.InfoF("path %v exist", path)
			continue
		}
		//str, err := c.CreateProtectedEphemeralSequential(path, []byte{}, zk.WorldACL(zk.PermAll))
		resPath, err := c.Create(path, []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
		if err != nil {
			base.ErrorF("set path: %v fail", path)
			return 0, err
		}
		base.InfoF("set path: [%+v] success", resPath, path, workId)
		return workId, nil
	}

	return 0, common.ConnErr
}

func (srv *ZkServer) RemoveAllNode(basePath string) (bool, error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	conn := pool.ClientConnPool.Get()
	if c, ok := conn.(*zk.Conn); ok {
		cds, _, err := c.Children(basePath)
		if err != nil {
			return false, common.OpErr.WithTrueErr(err)
		}
		delNums := 0
		var path string
		for i := 0; i < len(cds); i++ {
			path = utils.SpliceString(basePath, "/", cds[i])
			base.InfoF("completePath: %s", path)
			err = c.Delete(path, -1)
			if err != nil {
				base.InfoF("c.Delete-err:[%+v]", err)
				continue
			}
			delNums++
		}
		base.InfoF("delete.Nums:[%d]", delNums)
	}
	return false, common.ConnErr
}

// RemoveNode 删除单个节点
func (srv *ZkServer) RemoveNode(basePath, nodePath string) (success bool, err error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	conn := pool.ClientConnPool.Get()
	if c, ok := conn.(*zk.Conn); ok {
		path := utils.SpliceString(basePath, nodePath)
		base.InfoF("completePath: %s", path)
		err = c.Delete(path, -1)
		if err != nil {
			base.ErrorF("c.Delete-err:[%+v]", err, path)
			return false, err
		}
		base.InfoF("c.Delete-success:[%+v]", path)
		return true, nil
	}

	return false, common.ConnErr
}

func (srv *ZkServer) Shutdown() {
	srv.RemoveAllNode(common.WorkIdPath)
	close(srv.errCh)
}

// todo watch and delete node

// Watch -
//func (srv *ZkServer) Watch() {
//	servers := srv.opt.servers
//	sessionTimeout := srv.opt.sessionTimeout
//
//	c, _, err := zk.Connect(servers, sessionTimeout)
//	if err != nil {
//		base.ErrorF("zk.Connect-err:[%+v]", err)
//		return
//	}
//
//}

// 创建父节点
func (srv *ZkServer) createFatherNode(path string) (success bool, err error) {
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return false, common.PathLengthErr
	}

	var tmpPath string
	servers := srv.opt.servers
	sessionTimeout := srv.opt.sessionTimeout

	c, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		base.ErrorF("zk.Connect-err:[%+v]", err)
		return false, common.StartConnErr.WithTrueErr(err)
	}

	for i := 0; i < len(paths)-1; i++ {
		if paths[i] == "" {
			continue
		}
		tmpPath = utils.SpliceString(tmpPath, "/", paths[i])
		_, err = c.Create(tmpPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			base.InfoF("create-err:[%+v] ", err, tmpPath)
			continue
		}
		base.InfoF("createPath-success:[%+v]", tmpPath)

	}
	base.InfoF("createFatherNode-success:[%+v]", path)

	return true, nil
}

// validatePath check path is valid
func (srv *ZkServer) validatePath(path string, isSequential bool) (bool, error) {
	if path == "" {
		return false, common.InvalidPathErr
	}

	if path[0] != '/' {
		return false, common.InvalidPathErr
	}

	n := len(path)
	if n == 1 {
		// path is just the root
		return false, nil
	}

	if !isSequential && path[n-1] == '/' {
		return false, common.InvalidPathErr
	}

	// Start at rune 1 since we already know that the first character is
	// a '/'.
	for i, w := 1, 0; i < n; i += w {
		r, width := utf8.DecodeRuneInString(path[i:])
		switch {
		case r == '\u0000':
			return false, common.InvalidPathErr
		case r == '/':
			last, _ := utf8.DecodeLastRuneInString(path[:i])
			if last == '/' {
				return false, common.InvalidPathErr
			}
		case r == '.':
			last, lastWidth := utf8.DecodeLastRuneInString(path[:i])

			// Check for double dot
			if last == '.' {
				last, _ = utf8.DecodeLastRuneInString(path[:i-lastWidth])
			}

			if last == '/' {
				if i+1 == n {
					return false, common.InvalidPathErr
				}

				next, _ := utf8.DecodeRuneInString(path[i+w:])
				if next == '/' {
					return false, common.InvalidPathErr
				}
			}
		case r >= '\u0000' && r <= '\u001f',
			r >= '\u007f' && r <= '\u009f',
			r >= '\uf000' && r <= '\uf8ff',
			r >= '\ufff0' && r < '\uffff':
			return false, common.InvalidPathErr
		}
		w = width
	}
	return true, nil
}

func (srv *ZkServer) genTrueWorkerIdByNodeName(nodeName string) (int, error) {
	strs := strings.Split(nodeName, "-")
	if len(strs) < 1 {
		return 0, common.NodeNameErr
	}
	return cast.ToInt(strs[len(strs)-1]), nil
}
