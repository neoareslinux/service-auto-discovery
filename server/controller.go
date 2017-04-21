package main

import (
	"fmt"
	"servicedis/dbapi"
)

// Controller comment
type Controller struct {
	clusterMode string
	stateDriver dbapi.EtcdAPI
	agentDb     map[string]*Node
}

// Node represent an anent demon
type Node struct {
	HostAddress string
	Port        uint16
}

// Init controller
//func (c *Controller) Init() {
//	serInfo := dbapi.ServiceInfo{}
//c.stateDriver.RegisterService(serInfo)
//}

// NewController will create an new Controller instance
func NewController() Controller {
	return Controller{
		clusterMode: "lone",
		stateDriver: dbapi.NewEtcdAPI(),
	}
}

// DiscoveryService  Loop monitor to discovery agent node
func (c *Controller) DiscoveryService() {
	stopCh := make(chan bool, 1)
	agentEvent := make(chan dbapi.WatchServiceEvent, 1)
	fmt.Println("Start to discovery service")
	go c.stateDriver.DiscoveryService(agentEvent, stopCh)
	for {
		agentEv := <-agentEvent
		if agentEv.EventType == dbapi.WatchServiceEventAdd {
			fmt.Println("Find an new agent", agentEv.ServiceInfo)
		} else if agentEv.EventType == dbapi.WatchServiceEventDel {
			fmt.Println("Delete an agent", agentEv.ServiceInfo)
		} else if agentEv.EventType == dbapi.WatchServiceEventErr {
			fmt.Println("Error occur when discovery agents")
		}
	}
}
