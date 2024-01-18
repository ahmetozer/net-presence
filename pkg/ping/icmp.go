package ping

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func Icmp(iface *net.Interface, ip *net.IP) error {

	addr, err := net.ResolveIPAddr("ip", ip.String())
	if err != nil {
		return err
	}
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer c.Close()

	b, err := (&icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte{150, 145, 154, 154, 157},
		},
	}).Marshal(nil)

	if err != nil {
		return err
	}

	// Send it
	n, err := c.WriteTo(b, addr)
	if err != nil {
		return err
	} else if n != len(b) {
		return fmt.Errorf("got %v; want %v", n, len(b))
	}
	return nil
}
