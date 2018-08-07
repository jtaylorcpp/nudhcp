package nudhcp

import (
	"github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"

	"log"
	"net"
	"time"
)

/*
TODO:

add mechanism to periodically release leases/send disconnect message

add mechanism to reserve mac address leases
*/

// Find maximum address using net.IPNet {IP, Mask}
// and subtract 1 so 192.168.1.0/24 doesnt return
// 192.168.1.255 but 192.168.1.254 (reserved broadcast addr)
func getStopAddr(netIP *net.IPNet) net.IP {
	stop := make([]byte, len(netIP.IP))
	for idx, _ := range netIP.IP {
		stop[idx] = netIP.IP[idx] + (byte(255) - netIP.Mask[idx])
	}
	stop[3] -= 1
	return stop
}

// Bumps address returned by net.ParseCIDR in case
// something like 192.168.1.0/24 is parsed where
// start IP would be 192.168.1.0 (reserved)
func getStartAddr(netIP *net.IPNet) net.IP {
	if netIP.IP[3] == 0 {
		return net.IP{netIP.IP[0], netIP.IP[1], netIP.IP[2], netIP.IP[3] + 1}
	}
	return netIP.IP
}

type DHCPServer struct {
	iface         string
	serverAddr    net.IP
	startAddr     net.IP
	stopAddr      net.IP
	gatewayAddr   net.IP
	dnsAddr       net.IP
	netmask       *net.IPNet
	leaseRange    int
	leaseDuration time.Duration
	leases        map[string]*dhcplease
	options       dhcp4.Options
}

// DHCP Server Constructor
// Need:
//  - interface 	-> eth0
//  - server addr 	-> 192.168.1.2
//  - subnet		-> 192.168.1.0/24
//	- dns			-> 8.8.8.8
//	- gateway		-> 192.169.1.1 (if not default)
//
// Can Derive:
//  - netmask		-> from subnet string
//  - begin addr	-> from subnet string
//  - end addr		-> subnet stirng
//	- gateway		-> from subnet (192.168.1.1)
//
// Parses out what it needs to, reserves addresses as necessary, and returns server

func NewDHCPServer(iface, serverAddr, subnet, gateway, dns string,
	duration time.Duration) *DHCPServer {
	server := &DHCPServer{}
	log.Println("Configuring DHCP Server...")
	log.Printf("  using interface %v\n", iface)
	server.iface = iface
	log.Printf("  using server address %v\n", serverAddr)
	server.serverAddr = net.ParseIP(serverAddr)
	log.Printf("  serving subnet %v\n", subnet)
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Fatalf("subnet %v is not valid...exiting\n", subnet)
		panic(err)
	}
	server.netmask = ipNet
	server.stopAddr = getStopAddr(ipNet)
	server.startAddr = getStartAddr(ipNet)
	server.leaseRange = dhcp4.IPRange(server.startAddr, server.stopAddr)
	log.Printf("  start/stop addresses are %v/%v total range is %v\n",
		server.startAddr, server.stopAddr, server.leaseRange)
	log.Printf("  using gateway address %v\n", gateway)
	server.gatewayAddr = net.ParseIP(gateway)
	log.Printf("  using dns address %v\n", dns)
	server.dnsAddr = net.ParseIP(dns)
	server.leases = make(map[string]*dhcplease)
	server.ReserveStaticLeases(server.serverAddr, server.gatewayAddr, server.dnsAddr)
	server.PrintReservedLeases()
	server.leaseDuration = duration
	server.options = GetDefaultDHCPOptions(net.IP(server.netmask.Mask),
		server.gatewayAddr, server.dnsAddr)

	return server
}

// Reserve IP addresses in block by IP address
func (ds *DHCPServer) ReserveStaticLeases(ips ...net.IP) {
	for _, ip := range ips {
		if dhcp4.IPInRange(ds.startAddr, ds.stopAddr, ip) {
			// set lease for 10 years
			// which will be shocking if this ever runs longer without failure
			ds.leases[ip.String()] = &dhcplease{mac: "00:00:00:00:00",
				expire: time.Now().Add(time.Hour * 24 * 365 * 10)}
		}
	}
}

// Reserve IP address in block by MAC/IP address pair
//
// Allows an address to be reserved for a specific computer but to be
// assigned by DHCP
func (ds *DHCPServer) ReserverDynamicLeases(reservations ...MacReservation) {}

func (ds *DHCPServer) expireLeases() {
	for addr, currentLease := range ds.leases {
		if currentLease.Expired() {
			delete(ds.leases, addr)
		}
	}
}

func (ds *DHCPServer) findNextFree() net.IP {
	log.Println("DHCP SERVER: finding next unassigned in subnet...")
	log.Printf("DHCP SERVER: current subnet is: %v %v\n", ds.netmask.IP.String(),
		ds.netmask.Mask.String())
	log.Printf("DHCP SERVER: total lease range is %v\n", ds.leaseRange)
	for i := 0; i < ds.leaseRange; i++ {
		possibleAddr := dhcp4.IPAdd(ds.startAddr, i).String()
		log.Printf("DHCP SERVER: evaluating lease number %v %v...", i, possibleAddr)
		if _, leased := ds.leases[possibleAddr]; !leased {
			return dhcp4.IPAdd(ds.startAddr, i)
		}
		log.Printf("DHCP SERVER: lease %v %v already reserved...", i, possibleAddr)
	}
	log.Println("DHCP SERVER: no available leases ... returning 0.0.0.0")
	return net.IPv4zero
}

func (ds *DHCPServer) findMacAssigned(mac string) (net.IP, bool) {
	for addr, lease := range ds.leases {
		if mac == lease.mac {
			return net.ParseIP(addr), true
		}
	}
	return nil, false
}

func (ds *DHCPServer) PrintReservedLeases() {
	log.Println("Current Leases:")
	for addr, lease := range ds.leases {
		log.Printf("\t%v <> %v\n", addr, lease.String())
	}
}
func (ds *DHCPServer) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, opts dhcp4.Options) dhcp4.Packet {
	log.Println("DHCP Packet recieved...")
	switch msgType {
	case dhcp4.Discover:
		log.Println("DHCP DISCOVER recieved...")
		resAddr, reserved := ds.findMacAssigned(p.CHAddr().String())
		log.Printf("DHCP DISCOVER: %v lookup -> (%v, %v)\n", p.CHAddr(), resAddr, reserved)
		if reserved {
			log.Printf("DHCP DISCOVER: %v already assigned %v\n", p.CHAddr(), resAddr)
			return dhcp4.ReplyPacket(p, dhcp4.Offer, ds.serverAddr, resAddr,
				ds.leaseDuration,
				ds.options.SelectOrderOrAll(opts[dhcp4.OptionParameterRequestList]))
		}
		addr := ds.findNextFree()
		log.Printf("DHCP DISCOVER: %v possible new address -> %v\n", p.CHAddr(), addr)
		if !addr.Equal(net.IPv4zero) {
			log.Printf("DHCP DISCOVER: %v assigned %v\n", p.CHAddr(), addr)
			return dhcp4.ReplyPacket(p, dhcp4.Offer, ds.serverAddr, addr,
				ds.leaseDuration,
				ds.options.SelectOrderOrAll(opts[dhcp4.OptionParameterRequestList]))
		}
		log.Printf("DHCP DISCOVER: nothing returned for %v\n", p)
	case dhcp4.Request:
		log.Println("DHCP REQUEST recieved...")
		if server, ok := opts[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(ds.serverAddr) {
			log.Printf("DHCP REQUEST: packer <%v> not for this server...\n", p)
			return nil //wrong dhcp server DO NOT RESPOND
		}
		requestedIP := net.IP(opts[dhcp4.OptionRequestedIPAddress])
		if requestedIP == nil {
			requestedIP = net.IP(p.CIAddr())
		}

		if len(requestedIP) == 4 && !requestedIP.Equal(net.IPv4zero) {
			// only ipv4 and no 0.0.0.0
			log.Printf("DHCP REQUEST: %v requesting %v...\n", p.CHAddr(), requestedIP)
			if dhcp4.IPInRange(ds.startAddr, ds.stopAddr, requestedIP) {
				log.Printf("DHCP REQUEST: %v is in ip range of server...\n", requestedIP)
				if lease, leased := ds.leases[requestedIP.String()]; !leased || lease.mac == p.CHAddr().String() {
					ds.leases[requestedIP.String()] = &dhcplease{
						mac:    p.CHAddr().String(),
						expire: time.Now().Add(ds.leaseDuration),
					}
					log.Printf("DHCP REQUEST: new lease assigned %v...\n",
						ds.leases[requestedIP.String()])
					return dhcp4.ReplyPacket(p, dhcp4.ACK, ds.serverAddr,
						requestedIP, ds.leaseDuration,
						ds.options.SelectOrderOrAll(opts[dhcp4.OptionParameterRequestList]))
				}
			}
		}
		log.Printf("DHCP REQUEST: sending NAK for some reason...\n")
		return dhcp4.ReplyPacket(p, dhcp4.NAK, ds.serverAddr, nil, 0, nil)

	case dhcp4.Release, dhcp4.Decline:
		log.Printf("DHCP RELEASE/DECLINE recieved...")
		for addr, currentLease := range ds.leases {
			if currentLease.mac == p.CHAddr().String() {
				delete(ds.leases, addr)
				break
			}
		}
	}
	return nil
}

func (ds *DHCPServer) Start() {
	if ds.iface != "" {
		log.Printf("DHCP SERVER: starting and bound to %v...\n", ds.iface)
		//udpConn, err := conn.NewUDP4FilterListener(ds.iface, ":67")
		udpConn, err := conn.NewUDP4BoundListener(ds.iface, ":67")

		if err != nil {
			panic(err)
		}
		log.Fatal(dhcp4.Serve(udpConn, ds))
	} else {
		log.Printf("DHCP SERVER: starting...\n")
		log.Fatal(dhcp4.ListenAndServe(ds))
	}
}
