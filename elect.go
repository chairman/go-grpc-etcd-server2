package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"time"
)

const prefix = "/root/elect"
const prop = "local"

func main() {
	chans := [3]chan bool{make(chan bool, 1), make(chan bool, 1), make(chan bool, 1)}
	leader := make(chan int, 10)

	go func() {
		sec := 0
		for {
			<-time.After(1 * time.Second)
			sec++
			fmt.Printf("========第%d秒========\n", sec)
		}
	}()

	go job(0, chans[0], leader)
	go job(1, chans[1], leader)
	go job(2, chans[2], leader)
	for {
		time.Sleep(5 * time.Second)
		nowLeader := <-leader
		chans[nowLeader] <- true
		time.Sleep(1 * time.Second)
		go job(nowLeader, chans[nowLeader], leader)
	}
}

func job(workerId int, quitC chan bool, leader chan int) {
	endpoints := []string{"127.0.0.1:2379"}
	quitA := make(chan int, 1)
	quitB := make(chan int, 1)
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	leaderFlag := false
	go campaign(cli, prefix, prop, workerId, leader, &leaderFlag, quitA)
	go doCrontab(workerId, &leaderFlag, quitB)

	<-quitC
	fmt.Printf("worker:%d 退出\n", workerId)
	quitA <- 1
	quitB <- 1
}

func campaign(cli *clientv3.Client, election string, prop string, workerId int, leader chan int, leaderFlag *bool, quitA chan int) {
	quit := false
	timer := time.NewTimer(1 * time.Second)
	for range timer.C {
		if quit {
			break
		}

		s, err := concurrency.NewSession(cli, concurrency.WithTTL(5))
		if err != nil {
			fmt.Println(err)
			continue
		}
		e := concurrency.NewElection(s, election)
		ctx := context.TODO()

		fmt.Printf("worker:%d 参加选举\n", workerId)
		if err = e.Campaign(ctx, prop); err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("worker:%d 选举：成功,Key:%s \n", workerId, e.Key())
		*leaderFlag = true
		leader <- workerId

		select {
		case <-s.Done():
			*leaderFlag = false
			fmt.Println("选举：超时")

		case <-quitA:
			s.Close()
			quit = true
		}
	}
	fmt.Printf("worker:%d 退出选举\n", workerId)
}

func doCrontab(workerId int, leaderFlag *bool, quitB chan int) {
	var cronCnt int
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	quit := false
	for !quit {
		select {
		case <-ticker.C:
			cronCnt++
			if *leaderFlag == true {
				fmt.Printf("worker:%d 每1s执行定时任务: %d\n", workerId, cronCnt)
			} else {
				fmt.Printf("worker:%d Follow\n", workerId)
			}
		case <-quitB:
			quit = true
		}
	}
	fmt.Printf("worker:%d 结束定时任务\n", workerId)
}
