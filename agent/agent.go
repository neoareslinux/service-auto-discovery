package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"servicedis/dbapi"
)

// CommonAgent interface for agents
type CommonAgent interface {
	RegisterAgent() error
}

// Agent represent an service discovery agent deamon
type Agent struct {
	HostAddr    string
	HostPort    int
	stateDriver dbapi.EtcdAPI
}

// NewAgent init an agent instance
func NewAgent(opts CliOpts) Agent {
	agent := Agent{HostAddr: opts.HostAddr, HostPort: opts.Port, stateDriver: dbapi.NewEtcdAPI()}
	return agent
}

// RegisterAgent report agent state to server
func (ag *Agent) RegisterAgent() {
	fmt.Println("start registering agent service")
	agentInfo := dbapi.ServiceInfo{
		ServiceName: "agents",
		Keyname:     "agenttest",
		TTL:         20,
		HostAddr:    ag.HostAddr,
		HostPort:    ag.HostPort,
		Hostname:    "local",
	}
	ag.stateDriver.RegisterService(agentInfo)
}

// CliOpts ..
type CliOpts struct {
	HostAddr string
	Port     int
}

func main() {
	var opts CliOpts
	var flagSet *flag.FlagSet
	flagSet = flag.NewFlagSet("agent", flag.ExitOnError)
	flagSet.StringVar(&opts.HostAddr,
		"hostaddr",
		"127.0.0.1",
		"Ip address of agent")
	flagSet.IntVar(&opts.Port,
		"port",
		9906,
		"Port of agent")
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to parse command. Error: %s", err)
	}
	agent := NewAgent(opts)
	agent.RegisterAgent()
}
