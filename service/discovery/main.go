package main

import (
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic("error occurs")
	}
	defer cli.Close()
	ctx := context.Background()
	watchCh := cli.Watch(ctx, "foo")

	go func() {
		for event := range watchCh {
			fmt.Print(event.Events[0].Kv)
		}
	}()

	time.Sleep(time.Second * 3)
	cli.Put(ctx, "foo", "new bar")
	cli.Delete(ctx, "foo")
}
