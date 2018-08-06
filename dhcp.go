package nudhcp

import (
	"github.com/krolaw/dhcp4"

	"log"
	"net"
)

func GetDefaultDHCPOptions(subnet net.IP, gateway net.IP, dns net.IP) dhcp4.Options {
	return dhcp4.Options{
		dhcp4.OptionSubnetMask:       []byte(subnet),
		dhcp4.OptionRouter:           []byte(gateway),
		dhcp4.OptionDomainNameServer: []byte(dns),
	}
}

type DHCPManager struct {
	servers map[string]*DHCPServer
}

func (dm *DHCPManager) StartDHCPServers() {
	log.Printf("Starting DHCP Servers...")
}
