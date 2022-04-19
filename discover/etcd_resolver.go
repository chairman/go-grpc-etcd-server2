package discover

import (
	"context"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"log"
	"sync"
)

type etcdResolver struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cc         resolver.ClientConn
	etcdClient *clientv3.Client
	scheme     string
	ipPool     sync.Map
}

func (e *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	log.Println("etcd resolver resolve now")
}

func (e *etcdResolver) Close() {
	log.Println("etcd resolver close")
	e.cancel()
}

func (e *etcdResolver) watcher() {

	watchChan := e.etcdClient.Watch(context.Background(), "/"+e.scheme, clientv3.WithPrefix())

	for {
		select {
		case val := <-watchChan:
			for _, event := range val.Events {
				switch event.Type {
				case 0: // 0 是有数据增加
					e.store(event.Kv.Key, event.Kv.Value)
					log.Println("put:", string(event.Kv.Key))
					e.updateState()
				case 1: // 1是有数据减少
					log.Println("del:", string(event.Kv.Key))
					e.del(event.Kv.Key)
					e.updateState()
				}
			}
		case <-e.ctx.Done():
			return
		}

	}
}

func (e *etcdResolver) store(k, v []byte) {
	e.ipPool.Store(string(k), string(v))
}

func (s *etcdResolver) del(key []byte) {
	s.ipPool.Delete(string(key))
}

func (e *etcdResolver) updateState() {
	var addrList resolver.State
	e.ipPool.Range(func(k, v interface{}) bool {
		tA, ok := v.(string)
		if !ok {
			return false
		}
		log.Printf("conn.UpdateState key[%v];val[%v]\n", k, v)
		addrList.Addresses = append(addrList.Addresses, resolver.Address{Addr: tA})
		return true
	})

	e.cc.UpdateState(addrList)
}
