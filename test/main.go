package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lehaisonmath6/etcdconfig"
)

func main() {
	etcdconfig.InitModels(nil)
	go func() {
		epChan := make(chan *etcdconfig.Endpoint)
		go etcdconfig.WatchChangeService("/test/kvstorage", epChan)
		for ep := range epChan {
			fmt.Println("event change endpoint", ep)
		}
	}()
	time.Sleep(2 * time.Second)
	err := etcdconfig.SetEndpoint(&etcdconfig.Endpoint{
		SID:    "/test/kvstorage",
		Schema: "thrift_binary",
		Host:   "10.60.68.102",
		Port:   "1203",
	})
	if err != nil {
		log.Fatalln(err)
	}
	ep, err := etcdconfig.GetEndpoint("/test/kvstorage", "thrift_binary")
	if ep != nil {
		log.Println(ep)
	}
	etcdconfig.SetEndpoint(&etcdconfig.Endpoint{
		SID:    "/test/kvstorage",
		Schema: "thrift_compact",
		Host:   "10.60.68.102",
		Port:   "1204",
	})
	allEp, _ := etcdconfig.GetAllEndpoint("/test/kvstorage")
	fmt.Println(allEp)
	waitKey := make(chan bool)
	<-waitKey
}
