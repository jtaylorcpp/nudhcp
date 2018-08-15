 package nudhcp

import (
	"net"
	"time"
	"testing"
)

func  TestGetStopAddr(t *testing.T) {
	startAddress1 := &net.IPNet{
				IP: []byte{192,168,1,0},
				Mask: []byte{255,255,255,0},
			}
	t.Logf("Test Address 1: %v\n",startAddress1)
	stopAddress1 := getStopAddr(startAddress1)
	if "192.168.1.254" != stopAddress1.String() {
		t.Errorf("Stop address %v should be 192.168.1.254\n",stopAddress1)
	}

	startAddress2 := &net.IPNet {
				IP: []byte{10,0,0,0},
				Mask: []byte{255,255,0,0},
			}
	t.Logf("Test Address 2: %v\n", startAddress2)
	stopAddress2 := getStopAddr(startAddress2)
	if "10.0.255.254" != stopAddress2.String() {
		t.Errorf("Stop address %v should be 10.0.255.254\n", stopAddress2)
	}

}

func TestGetStartAddress(t *testing.T) {
	startAddress1 := &net.IPNet{
                                IP: []byte{192,168,1,0},
                                Mask: []byte{255,255,255,0},
                        }
        t.Logf("Test Address 1: %v\n",startAddress1)
	firstAddress1 := getStartAddr(startAddress1)
	if "192.168.1.1" != firstAddress1.String() {
		t.Errorf("First address %v should be 192.168.1.1\n",firstAddress1)
	}
	startAddress2 := &net.IPNet {
                                IP: []byte{10,0,0,0},
                                Mask: []byte{255,255,0,0},
                        }
        t.Logf("Test Address 2: %v\n", startAddress2)
	firstAddress2 := getStartAddr(startAddress2)
	if "10.0.0.1" != firstAddress2.String() {
		 t.Errorf("First address %v should be 10.0.0.1\n",firstAddress2)
	}
}

func TestDHCPServer(t *testing.T) {
	ds := NewDHCPServer("testETH","10.0.0.2","10.0.0.0/16","10.0.0.1","4.4.4.4", 1 * time.Second)

	if ds.iface != "testETH" {
		t.Errorf("Eth %v should be \"testETH\"\n",ds.iface)
	}

	if ds.serverAddr.String() != "10.0.0.2" {
		t.Errorf("Server IP %v should be 10.0.0.2\n", ds.serverAddr)
	}

	if net.IP(ds.netmask.Mask).String() != "255.255.0.0" {
		t.Errorf("Netmask %v should be 255.255.0.0\n", ds.netmask.Mask.String())
	}

	if ds.startAddr.String() != "10.0.0.1" {
		t.Errorf("Start address %v should be 10.0.0.1\n", ds.startAddr)
	}

	if ds.stopAddr.String() != "10.0.255.254" {
		t.Errorf("Stop address %v should be 10.0.255.254\n", ds.stopAddr)
	}

	if ds.gatewayAddr.String() != "10.0.0.1" {
		t.Errorf("Gateway address %v should be 10.0.0.1\n",ds.gatewayAddr)
	}

	if ds.dnsAddr.String() != "4.4.4.4" {
		t.Errorf("DNS address %v should be 4.4.4.4\n",ds.dnsAddr)
	}

	testIP1 := net.ParseIP("10.0.0.2")
	ds.ReserveStaticLeases(testIP1)
	if _, ok := ds.leases["10.0.0.2"]; !ok {
		t.Error("Unable to reserve address 10.0.0.2")
	}

	for ip := 3; ip < 128; ip += 1 {
		nextIp := ds.findNextFree()
		expectedIp := net.IP([]byte{10,0,0,byte(ip)})
		if nextIp.String() != expectedIp.String() {
			t.Errorf("Assigned IP %v should have been %v\n", nextIp, expectedIp)
		}
		ds.ReserveStaticLeases(nextIp)
	}

	if _,assigned := ds.findMacAssigned("will fail"); assigned {
		t.Error("Invalid MAC address \"will fail\"")
	}
}
