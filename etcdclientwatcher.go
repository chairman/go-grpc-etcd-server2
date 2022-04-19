package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"log"
	"strconv"
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

	// 监听变化
	go watcher(client,key)

	value := "127.0.0.1:800"
	ctx := context.Background()

	// 每隔2秒重新PUT一次
	for i:=0 ;i<100;i++ {
		time.Sleep(2*time.Second)
		_, err := client.Put(ctx, key, value+strconv.Itoa(i))
		if err != nil {
			log.Printf("put error %v",err)
		}
	}

}


func watcher(client *clientv3.Client,key string) {

	// 监听这个chan
	watchChan := client.Watch(context.Background(), key)

	for watchResponse := range watchChan {

		for _, event := range watchResponse.Events {
			fmt.Printf("Type:%s,Key:%s,Value:%s\n",event.Type,event.Kv.Key,event.Kv.Value)
			/**
			Type:PUT,Key:/ns/service,Value:127.0.0.1:8000
			Type:PUT,Key:/ns/service,Value:127.0.0.1:8001
			Type:PUT,Key:/ns/service,Value:127.0.0.1:8002
			Type:PUT,Key:/ns/service,Value:127.0.0.1:8003
			...
			 */
		}
	}
}

