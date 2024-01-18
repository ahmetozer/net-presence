package presence

import (
	"net"
	"time"

	b "github.com/ahmetozer/basiclog"
	ha "github.com/ahmetozer/net-presence/pkg/homeasistant"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Packets are proccessed linear, to prevent spikes at GC, global variables is used
var (
	ip         net.IP
	layer      gopacket.Layer
	packet     gopacket.Packet
	macSrc, id string
)

func Watch(interfaceName string) {

	handle, err := pcap.OpenLive(interfaceName, 262144, true, pcap.BlockForever)
	b.ErrNil(b.Fatal, err)
	b.ErrNil(b.Fatal, handle.SetBPFFilter("inbound and (udp port 53 or udp port 123 or icmp or multicast or broadcast)"))

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet = range packetSource.Packets() {

		macSrc = packet.LinkLayer().LinkFlow().Src().String()

		getIp()

		if ip.Equal(net.IP{}) { // Empty
			b.Debug("ip is not determined for %s", macSrc)
			continue // to proccess next packet
		}

		id = macSrc + ip.String()
		if _, ok := ha.Devices[id]; !ok {
			//do something here
			ha.Devices[id] = &ha.DeviceInfo{LastSeen: 0, LastIP: ip, Mac: macSrc}
			b.ErrNil(b.Error, ha.ClientNew(ha.Devices[id]))
			b.Info("new device %v\n", macSrc)
		}

		if ha.Devices[id].LastSeen == 0 {
			b.ErrNil(b.Error, ha.ClientAtHome(ha.Devices[id]))
			b.Info("home %v\n", macSrc)
		}

		ha.Devices[id].LastSeen = time.Now().Unix()

	}

}

func getIp() {
	layer = packet.Layer(layers.LayerTypeIPv4)
	if layer != nil {
		ip = layer.(*layers.IPv4).SrcIP
		b.Debug("%s detected from ip package", macSrc)
	} else {
		layer = packet.Layer(layers.LayerTypeARP)
		if layer == nil {
			ip = net.IP{}
			return
		}

		ip = layer.(*layers.ARP).SourceProtAddress
		b.Debug("%s detected from arp package", macSrc)

	}
}
