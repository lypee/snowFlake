分布式雪花算法
目前使用zk进行worker机器id分配


```
errCh := make(chan error, 3)
sf := NewSfWorker(errCh,zkServer.WithServers([]string{"your host"}))
sf.NextID()
``
