package main

import (
	"fmt"
	"time"

	"github.com/jtaylorcpp/nudhcp"
)

func example_main() {
	fmt.Println("Test!")
	// interface, server address, subnet to address, gateway, dns, lease time
	dhcps := nudhcp.NewDHCPServer("eth1", "192.168.2.2", "192.168.2.0/24",
		"192.168.2.1", "8.8.8.8", time.Hour*3)
	fmt.Printf("%v\n", dhcps)
	dhcps.Start()
}
