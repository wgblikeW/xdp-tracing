package main

import (
	"context"
	"time"

	"github.com/p1nant0m/xdp-tracing/handler/utils"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	PUT    mvccpb.Event_EventType = 0
	DELETE mvccpb.Event_EventType = 1
)

var localCache map[string]string = make(map[string]string)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.99999999",
	})

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		logrus.Fatal("error occurs when create new etcdv3 client err:%v", err.Error())
	}

	logrus.Info(utils.FontSet("[Strategy Deployer]") + " Service Start successfully!")
	watchCh := cli.Watch(context.Background(), "node", clientv3.WithPrefix())
	for watchResp := range watchCh {
		switch watchResp.Events[0].Type {
		case PUT:
			logrus.WithFields(logrus.Fields{
				"key":     string(watchResp.Events[0].Kv.Key),
				"value":   string(watchResp.Events[0].Kv.Value),
				"revison": watchResp.Header.Revision,
			}).Info(utils.FontSet("[Etcd Watcher]") + " This is a Put Event")

			localCache[string(watchResp.Events[0].Kv.Key)] = string(watchResp.Events[0].Kv.Value)
			logrus.Infof(utils.FontSet("[Etcd Watcher]")+" Host %v (%v) has been added to localCache",
				string(watchResp.Events[0].Kv.Key), string(watchResp.Events[0].Kv.Value))
		case DELETE:
			logrus.WithFields(logrus.Fields{
				"key":     string(watchResp.Events[0].Kv.Key),
				"value":   string(watchResp.Events[0].Kv.Value),
				"revison": watchResp.Header.Revision,
			}).Info(utils.FontSet("[Etcd Watcher]") + " This is a Delete Event")

			logrus.Infof(utils.FontSet("[Etcd Watcher]")+" Host %v (%v) has been removed from localCache",
				string(watchResp.Events[0].Kv.Key), localCache[string(watchResp.Events[0].Kv.Key)])
			delete(localCache, string(watchResp.Events[0].Kv.Key))
		}
	}
}
