package main

import (
	"go-grpc-etcd-server/com/discover"
	"go-grpc-etcd-server/com/lb"
	pb "go-grpc-etcd-server/com/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	//"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

func main() {
	// 注册自定义ETCD解析器
	etcdResolverBuilder := discover.NewEtcdResolverBuilder()
	resolver.Register(etcdResolverBuilder)

	// 使用自带的DNS解析器和负载均衡实现方式
	conn, err := grpc.Dial("etcd:///", grpc.WithBalancerName(lb.Name), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	c := pb.NewGreeterClient(conn)
	// 初始化上下文，设置请求超时时间为1秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 延迟关闭请求会话
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "world"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	// 打印服务的返回的消息
	log.Printf("Greeting: %s", r.Message)
	//for i := 0; i < 200; i++ {
	//	time.Sleep(2 * time.Second)
	//	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "world"})
	//	if err != nil {
	//		log.Fatalf("could not greet: %v", err)
	//	}
	//	// 打印服务的返回的消息
	//	log.Printf("Greeting: %s", r.Message)
	//}

}
