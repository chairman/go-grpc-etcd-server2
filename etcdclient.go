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

	// PUT
	ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	key := "/ns/service"
	value := "127.0.0.1:8000"
	_, err = client.Put(ctx, key, value)
	if err != nil {
		log.Printf("etcd put error,%v\n", err)
		return
	}

	// GET
	getResponse, err := client.Get(ctx, key)
	if err != nil {
		log.Printf("etcd GET error,%v\n", err)
		return
	}

	for _,kv := range getResponse.Kvs {
		fmt.Printf("%s=%s\n",kv.Key,kv.Value)
	}

	// /ns/service=127.0.0.1:8000
}

