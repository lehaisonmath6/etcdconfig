package etcdconfig

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	etcdEndpoints = []string{"http://10.60.1.20:2379", "http://10.110.1.100:2379", "http://10.110.68.103:2379"}
	etcdClient    *clientv3.Client
)

func init() {
	fmt.Println("init etcdconfig")
	InitModels(etcdEndpoints)
}
func InitModels(aEtcdEndpoints []string) {
	var err error
	if len(aEtcdEndpoints) > 0 {
		etcdEndpoints = aEtcdEndpoints
	}
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err == context.DeadlineExceeded {
		log.Println("[ERROR] connect to etcd endpoint", etcdEndpoints, err)
		return
	}
}

func Close() {
	if etcdClient != nil {
		etcdClient.Close()
	}
}

type Endpoint struct {
	SID    string
	Host   string
	Port   string
	Schema string
}

// example with schema key /trustkeys/socialnetwork/postservice/thrift_binary:0.0.0.0:12010 value thrift_binary://0.0.0.0:12010
// example no schema  key /trustkeys/socialnetwork/status/postservice value 10.60.68.102:12010
// schema is thrift_compact | thrift_binary | ""
func GetEndpoint(sid string, schema string) (*Endpoint, error) {
	if etcdClient == nil {
		return nil, errors.New("etcd Client null")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	opts := clientv3.WithPrefix()

	resp, err := etcdClient.Get(ctx, sid, opts)
	cancel()
	if err != nil {
		return nil, err
	}
	for _, kv := range resp.Kvs {
		if strings.Contains(string(kv.Key), schema) {
			return parseEndpoint(sid, string(kv.Value), schema)
		}
	}
	return nil, errors.New("not found")
}
func DeleteEndpoint(sid string) error {
	if etcdClient == nil {
		return errors.New("etcd Client null")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	opts := clientv3.WithPrefix()

	_, err := etcdClient.Delete(ctx, sid, opts)
	cancel()
	if err != nil {
		return err
	}
	return nil
}
func SetEndpoint(ep *Endpoint) error {
	if etcdClient == nil {
		return errors.New("etcd Client null")
	}
	var key, value string
	if ep.Schema == "" {
		key = ep.SID
		value = ep.Host + ":" + ep.Port

	} else if ep.Schema == "https" {
		key = ep.SID + "/" + ep.Schema + ":" + ep.Host
		value = ep.Schema + "://" + ep.Host
	} else {
		key = ep.SID + "/" + ep.Schema + ":" + ep.Host + ":" + ep.Port
		value = ep.Schema + "://" + ep.Host + ":" + ep.Port
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	_, err := etcdClient.Put(ctx, key, value)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func GetAllEndpoint(sid string) ([]*Endpoint, error) {
	if etcdClient == nil {
		return nil, errors.New("etcd Client null")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	opts := clientv3.WithPrefix()

	resp, err := etcdClient.Get(ctx, sid, opts)
	cancel()
	if err != nil {
		return nil, err
	}
	var listEndpoints []*Endpoint
	for _, kv := range resp.Kvs {
		schema := getSchema(string(kv.Value))
		ep, err := parseEndpoint(sid, string(kv.Value), schema)
		if err != nil {
			continue
		}
		listEndpoints = append(listEndpoints, ep)
	}
	if len(listEndpoints) == 0 {
		return nil, errors.New("not found")
	}

	return listEndpoints, nil
}

func WatchChangeService(sid string, epChan chan *Endpoint) {
	if etcdClient == nil {
		return
	}
	if epChan == nil {
		epChan = make(chan *Endpoint, 10)
	}
	opts := clientv3.WithPrefix()

	watchChan := etcdClient.Watch(context.Background(), sid, opts)

	for wresp := range watchChan {
		for _, ev := range wresp.Events {
			if ev.Type == clientv3.EventTypePut {
				schema := getSchema(string(ev.Kv.Value))
				ep, err := parseEndpoint(sid, string(ev.Kv.Value), schema)
				if err != nil {
					log.Println("[ERROR] etcd watch parse endpoint sid", sid, err)
				}
				epChan <- ep
			}
		}
	}
}

func getSchema(value string) string {
	arr := strings.Split(value, "://")
	if len(arr) != 2 {
		return ""
	}
	return arr[0]
}

func parseEndpoint(sid, value, schema string) (*Endpoint, error) {
	address := ""
	if schema != "" {
		arr := strings.Split(value, "//")
		if len(arr) != 2 {
			return nil, errors.New("invalid endpoint value")
		}
		address = arr[1]
	}
	if schema == "https" {
		return &Endpoint{
			SID:    sid,
			Host:   address,
			Port:   "",
			Schema: schema,
		}, nil
	}
	arr := strings.Split(address, ":")
	if len(arr) != 2 {
		return nil, errors.New("invalid endpoint value")
	}
	return &Endpoint{
		SID:    sid,
		Host:   arr[0],
		Port:   arr[1],
		Schema: schema,
	}, nil
}
