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
		go etcdconfig.WatchChangeService("/trustkeys/cryptocurrencywallet/node/bsctest", epChan)
		for ep := range epChan {
			fmt.Println("event change endpoint", ep)
		}
	}()
	go func() {
		epChan := make(chan *etcdconfig.Endpoint)
		go etcdconfig.WatchChangeService("/trustkeys/cryptocurrencywallet/node/bsc", epChan)
		for ep := range epChan {
			fmt.Println("event change endpoint", ep)
		}
	}()
	fmt.Println("begin")
	time.Sleep(5 * time.Second)

	err := etcdconfig.SetEndpoint(&etcdconfig.Endpoint{
		SID:    "/trustkeys/cryptocurrencywallet/node/bsctest",
		Host:   "data-seed-prebsc-1-s2.binance.org",
		Port:   "8545",
		Schema: "https",
	})
	if err != nil {
		log.Println("[ERROR] set ep", err)
	}
	// ep, err := etcdconfig.GetEndpoint("/trustkeys/cryptocurrencywallet/node/bsctest", "https")
	// fmt.Println(ep, err)
	// if err != nil {
	// 	log.Println(err)
	// 	time.Sleep(5 * time.Second)
	// }
	// ep, err := etcdconfig.GetEndpoint("/test/kvstorage", "thrift_binary")
	// if ep != nil {
	// 	log.Println(ep)
	// }
	// etcdconfig.SetEndpoint(&etcdconfig.Endpoint{
	// 	SID:    "/test/kvstorage",
	// 	Schema: "thrift_compact",
	// 	Host:   "10.60.68.102",
	// 	Port:   "1204",
	// })
	// allEp, _ := etcdconfig.GetAllEndpoint("/test/kvstorage")
	// fmt.Println(allEp)
	waitKey := make(chan bool)
	<-waitKey
}
