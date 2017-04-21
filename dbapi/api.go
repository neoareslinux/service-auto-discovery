package dbapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

// Watch events
const (
	WatchServiceEventAdd = iota
	WatchServiceEventDel
	WatchServiceEventErr
)

// API cc
type API interface {
	GetObj(key string, retValue interface{}) error
	SetObj(key string, value interface{}) error
	DelObj(key string) error
	ListDir(key string) ([]string, error)
	RegisterService(serviceInfo ServiceInfo) error
	GetService(name string) ([]ServiceInfo, error)
}

// ServiceInfo cc
type ServiceInfo struct {
	ServiceName string
	Keyname     string
	TTL         int
	HostAddr    string
	HostPort    int
	Hostname    string
}

// EtcdServiceState cc
type EtcdServiceState struct {
	ServiceName string
	Keyname     string
	Version     string
	TTL         time.Duration
	HostAddr    string
	Port        int
	Hostname    string
	stopCh      chan bool
}

// EtcdAPI cc
type EtcdAPI struct {
	client client.Client
	//kapi   client.KeysAPI
}

// WatchServiceEvent xx
type WatchServiceEvent struct {
	EventType   uint
	ServiceInfo ServiceInfo
}

// NewEtcdAPI cc
func NewEtcdAPI() EtcdAPI {
	endPointURL := []string{"http://127.0.0.1:2379"}
	cfg := client.Config{
		Endpoints: endPointURL,
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		// handle error
		fmt.Println("Fali to create new etcd client")
	}
	etcdClient := EtcdAPI{client: c}
	return etcdClient
}

//RegisterService cc
func (e *EtcdAPI) RegisterService(srvInfo ServiceInfo) {
	keyName := "/nodes/" + srvInfo.HostAddr + ":" + strconv.Itoa(srvInfo.HostPort)
	jsonVal, err := json.Marshal(srvInfo)
	ttl := time.Duration(srvInfo.TTL) * time.Second
	if err != nil {
		fmt.Println("ERROR to prepare json data to register agent")
	}
	etcdState := EtcdServiceState{
		ServiceName: srvInfo.ServiceName,
		Keyname:     keyName,
		Version:     "1.0",
		TTL:         ttl,
		HostAddr:    srvInfo.HostAddr,
		Port:        srvInfo.HostPort,
		Hostname:    srvInfo.Hostname,
		stopCh:      make(chan bool, 1),
	}
	fmt.Println("Start to refresh service loop")
	e.refreshService(&etcdState, string(jsonVal[:]))
}

func (e *EtcdAPI) refreshService(etcdInfo *EtcdServiceState, keyVal string) {
	fmt.Println("start refresh service....")
	kapi := client.NewKeysAPI(e.client)
	_, err := kapi.Set(context.Background(), etcdInfo.Keyname, keyVal, &client.SetOptions{TTL: etcdInfo.TTL})
	if err != nil {
		fmt.Println("Fail to set ", etcdInfo.Keyname)
	}
	// loop forever
	for {
		select {
		case <-time.After(etcdInfo.TTL / 3):
			_, err := kapi.Set(context.Background(), etcdInfo.Keyname, keyVal, &client.SetOptions{TTL: etcdInfo.TTL})
			if err != nil {
				fmt.Println("Fail to set ", etcdInfo.Keyname)
			}
		case <-etcdInfo.stopCh:
			fmt.Println("stop updating ", etcdInfo.Keyname)
			return
		}
	}
}

// DiscoveryService cc
func (e *EtcdAPI) DiscoveryService(nodeChan chan WatchServiceEvent, stopCh chan bool) {
	keyname := "/nodes"
	kapi := client.NewKeysAPI(e.client)
	watchCtx, watchCancel := context.WithCancel(context.Background())
	watchCh := make(chan *client.Response, 1)
	go func() {
		watchIndex := e.InitStateDriver(keyname, watchCh)
		watcher := kapi.Watcher(keyname, &client.WatcherOptions{AfterIndex: watchIndex, Recursive: true})
		if watcher == nil {
			fmt.Println("Fail to watch", keyname)
		}
		for {
			etcdRsp, err := watcher.Next(watchCtx)
			if err != nil {
				fmt.Println("err to watch next")
			}
			watchCh <- etcdRsp
		}
	}()

	go func() {
		var serMap = make(map[string]ServiceInfo)
		for {
			select {
			case watchResp := <-watchCh:
				var srvInfo ServiceInfo
				srvKey := strings.TrimPrefix(watchResp.Node.Key, "/nodes")
				if _, ok := serMap[srvKey]; !ok && watchResp.Action == "set" {
					err := json.Unmarshal([]byte(watchResp.Node.Value), &srvInfo)
					if err != nil {
						fmt.Println("Can not dumps json", err)
						nodeChan <- WatchServiceEvent{
							EventType:   WatchServiceEventErr,
							ServiceInfo: srvInfo,
						}
					}
					fmt.Println("Add key into serMap", srvKey, srvInfo)
					serMap[srvKey] = srvInfo
					nodeChan <- WatchServiceEvent{
						EventType:   WatchServiceEventAdd,
						ServiceInfo: srvInfo,
					}
				} else if (watchResp.Action == "delete") || (watchResp.Action == "expire") {
					err := json.Unmarshal([]byte(watchResp.PrevNode.Value), &srvInfo)
					if err != nil {
						fmt.Println("Can not dumps json", err)
					}
					delete(serMap, srvKey)
					nodeChan <- WatchServiceEvent{
						EventType:   WatchServiceEventDel,
						ServiceInfo: srvInfo,
					}
				}
			case stopReq := <-stopCh:
				if stopReq {
					watchCancel()
					return
				}
			}
		}
	}()
}

// InitStateDriver will scan all agents that exist before start to watch it
func (e *EtcdAPI) InitStateDriver(key string, watchCh chan *client.Response) uint64 {
	kapi := client.NewKeysAPI(e.client)
	resp, err := kapi.Get(context.Background(), key, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		// should retry here
		fmt.Println("fail to  scan all agents", err)
	}
	for _, node := range resp.Node.Nodes {
		fmt.Println("uyyyyyyy", node)
		watchCh <- resp
	}

	fmt.Println("Finish init state driver")
	return resp.Index
}
