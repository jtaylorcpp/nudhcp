package nudhcp

import (
	"testing"
)

const config string = `
servers:
  - interface: eth0
    serverAddress: 192.168.2.2
    subnet: 192.168.2.0/24
    gateway: 192.168.2.1
    dnsServers: 8.8.8.8
    leaseDuration: 1m
    ipReservations:
      - mac: thisisamac
        ip: 192.168.2.3
`

func TestConfigFileParser(t *testing.T) {
	testConfig := parseConfigFile(config)
	t.Log("Parsed config: ",testConfig)
}


func TestDHCPServerFromConfig(t *testing.T) {
	testConfig := parseConfigFile(config)
	newServer := serverFromConfig(testConfig.Servers[0])
	t.Log("New sever: ",newServer)
}
