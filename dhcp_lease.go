package nudhcp

import (
	"fmt"
	"time"
)

type dhcplease struct {
	mac    string
	expire time.Time
}

func (lease *dhcplease) Expired() bool {
	return lease.expire.Before(time.Now())
}

func (lease *dhcplease) String() string {
	return fmt.Sprintf("<mac:%v, time left:%v>", lease.mac, lease.expire.String())
}

type MacReservation struct {
	Mac string
	Ip  string
}
