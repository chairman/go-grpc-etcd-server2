package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func main() {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalln(err)
	}

	key := "/ns/service"
	value := "127.0.0.1:800"
	ctx := context.Background()

	// 获取一个租约 有效期为5秒
	leaseGrant, err := client.Grant(ctx, 5)
	if err != nil {
		log.Printf("grant error %v",err)
		return
	}

	// PUT 租约期限为5秒
	_, err = client.Put(ctx, key, value, clientv3.WithLease(leaseGrant.ID))
	if err != nil {
		log.Printf("put error %v",err)
		return
	}

	// 续租
	keepaliveResponseChan, err := client.KeepAlive(ctx, leaseGrant.ID)
	if err != nil {
		log.Printf("KeepAlive error %v",err)
		return
	}
	
	for {
		ka := <-keepaliveResponseChan
		fmt.Println("ttl:", ka.TTL)
		//ttl: 5
		//ttl: 5
		//ttl: 5
	}
}

