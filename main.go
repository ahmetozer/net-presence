package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type macInfo struct {
	lastSeen int64
	lastIP   net.IP
}

var linfo, ldebug struct {
	Printf func(format string, v ...any)
}

func main() {

	var (
		macTTL        int64
		arpTTL        int64
		icmpTTL       int64
		err           error
		interfaceName = "eth0"
		mqttServer    = "tcp://127.0.0.1:1883"
	)

	var a func(format string, v ...any) = func(format string, v ...any) {}
	linfo.Printf = a
	ldebug.Printf = a

	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG", "debug":
		ldebug.Printf, linfo.Printf = log.Printf, log.Printf
	case "ERR", "err":
	default:
		linfo.Printf = log.Printf
	}

	macTTL, err = getEnvTTL("MAC_TTL", 30)
	errFatalf(err, "MAC_TTL")
	arpTTL, err = getEnvTTL("ARP_TTL", 15)
	errFatalf(err, "ARP_TTL")
	icmpTTL, err = getEnvTTL("PING_TTL", 12)
	errFatalf(err, "PING_TTL")

	if s := os.Getenv("INTERFACE"); s != "" {
		interfaceName = s
	}
	ldebug.Printf("interface name %s", interfaceName)

	if s := os.Getenv("MQTT"); s != "" {
		mqttServer = s
	}
	ldebug.Printf("mqtt server %s", mqttServer)

	client := mqtt.NewClient((&mqtt.ClientOptions{
		OnConnect:        connectHandler,
		OnConnectionLost: connectLostHandler,
		AutoReconnect:    true,
		ClientID:         "mac_presence_" + strconv.Itoa(os.Getpid()),
		ConnectTimeout:   time.Second * 2,
	}).AddBroker(mqttServer))

	ldebug.Printf("mqtt client %#s", client)

	if token := client.Connect(); token.Wait() {
		errFatalf(token.Error(), "token")
	}

	mac_presence := make(map[string]*macInfo)

	if iface, err := net.InterfaceByName(interfaceName); errFatalf(err) {
		go func() {
			var tdiff int64
			for {
				for mac, mInfo := range mac_presence {
					ldebug.Printf("mac %s, last seen %d, last ip %s", mac, mInfo.lastSeen, mInfo.lastIP)
					tdiff = time.Now().Unix() - mInfo.lastSeen
					macSrcMqtt := strings.Replace(mac, ":", "", -1)
					if tdiff > macTTL && mInfo.lastSeen != 0 {
						errPrintf(client.Publish("presence/"+macSrcMqtt+"/state", 0, false, "not_home").Error(), "mqtt", "not_home")
						linfo.Printf("not_home src %v %v\n", mac, mInfo)
						mac_presence[mac] = &macInfo{0, mac_presence[mac].lastIP}
					}
					if tdiff > icmpTTL && mInfo.lastSeen != 0 {
						ldebug.Printf("mac %s, last seen %d ago, icmp request %s", mac, tdiff, mInfo.lastIP)
						errPrintf(sendIcmp(iface, &mInfo.lastIP), "icmp", "control")
					}

					if tdiff > arpTTL && mInfo.lastSeen != 0 {
						ldebug.Printf("mac %s, last seen %d ago, arp request %s", mac, tdiff, mInfo.lastIP)
						errPrintf(sendArp(iface, &mInfo.lastIP), "arp", "control")
					}
				}
				time.Sleep(time.Second * 1)
			}

		}()
	}

	handle, err := pcap.OpenLive(interfaceName, 262144, true, pcap.BlockForever)
	errFatalf(err)
	errFatalf(handle.SetBPFFilter("inbound and (udp port 53 or udp port 5353 or icmp or multicast or broadcast)"))

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	var (
		layer              gopacket.Layer
		packet             gopacket.Packet
		macSrc, macSrcMqtt string
	)
	for packet = range packetSource.Packets() {

		macSrc = packet.LinkLayer().LinkFlow().Src().String()
		macSrcMqtt = strings.Replace(macSrc, ":", "", -1)

		if _, ok := mac_presence[macSrc]; !ok {
			//do something here
			mac_presence[macSrc] = &macInfo{lastSeen: 0}
			errPrintf(client.Publish("homeassistant/device_tracker/"+macSrcMqtt+"/config", 0, false, `{"state_topic": "presence/`+macSrcMqtt+`/state", "name": "`+macSrcMqtt+`", "payload_home": "home", "payload_not_home": "not_home"}`).Error(), "mqtt", "newdevice")
			linfo.Printf("new device %v\n", macSrc)
		}

		layer = packet.Layer(layers.LayerTypeIPv4)
		if layer != nil {
			mac_presence[macSrc].lastIP = layer.(*layers.IPv4).SrcIP
			ldebug.Printf("%s detected from ip package", macSrc)
		} else {
			layer = packet.Layer(layers.LayerTypeARP)
			if layer != nil {
				mac_presence[macSrc].lastIP = layer.(*layers.ARP).SourceProtAddress
				ldebug.Printf("%s detected from arp package", macSrc)
			}
		}

		if mac_presence[macSrc].lastSeen == 0 {
			errPrintf(client.Publish("presence/"+macSrcMqtt+"/state", 0, false, "home").Error(), "mqtt", "home")
			linfo.Printf("home %v\n", macSrc)
		}

		mac_presence[macSrc].lastSeen = time.Now().Unix()

	}

}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	linfo.Printf("MQTT Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	errPrintf(err, "mqtt", "connection_lost")
}

func sendArp(iface *net.Interface, ip *net.IP) error {
	var addr *net.IPNet
	if addrs, err := iface.Addrs(); err != nil {
		return err
	} else {
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					addr = &net.IPNet{
						IP:   ip4,
						Mask: ipnet.Mask[len(ipnet.Mask)-4:],
					}
					break
				}
			}
		}
	}
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(addr.IP),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(*ip),
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	gopacket.SerializeLayers(buf, opts, &eth, &arp)
	handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()
	if err := handle.WritePacketData(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func sendIcmp(iface *net.Interface, ip *net.IP) error {

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

func getEnvTTL(envName string, defaultTTL int64) (int64, error) {
	s := os.Getenv(envName)
	if s == "" {
		return defaultTTL, nil
	}

	ttl, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return ttl, nil
}

func errFatalf(err error, s ...string) bool {
	if err != nil {
		if len(s) > 0 {
			log.Fatalf(strings.Join(s, " ")+" %v", err)
		} else {
			log.Fatalf("%v", err)
		}
		return false
	}
	return true
}

func errPrintf(err error, s ...string) bool {

	if err != nil {
		if len(s) > 0 {
			log.Printf(strings.Join(s, " ")+" %v", err)
		} else {
			log.Printf("%v", err)
		}
		return false
	}
	return true
}
