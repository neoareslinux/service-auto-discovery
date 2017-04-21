package main

import "fmt"

func main() {
	// c := Controller{ClusterMode: "testmode"}
	var c Controller
	c = NewController()
	//c.Init()
	c.DiscoveryService()
	fmt.Println("finish")
}
