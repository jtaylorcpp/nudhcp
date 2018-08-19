package nudhcp

import (
	"log"
)


// Provides the shared context for all of the DHCP servers ran in this instance
type DHCPManager struct {
	servers map[string]*DHCPServer
}

func (dm *DHCPManager) StartDHCPServers() {
	log.Printf("Starting DHCP Servers...")
	statusChan := make(chan error, 1)
	defer close(statusChan)
	for  _, server := range dm.servers {
		go server.Start(statusChan)
	}

	for statusMsg := range statusChan {
		log.Fatal(statusMsg)
	}
}
