package ping

import (
	"net"
	"os"
	"strconv"
	"time"

	b "github.com/ahmetozer/basiclog"
	ha "github.com/ahmetozer/net-presence/pkg/homeasistant"
)

func Existence(interfaceName string) {

	var (
		presenceTTL      int64
		presenceInterval time.Duration
		arpTTL           int64
		icmpTTL          int64
		tdiff            int64
		err              error
	)

	presenceTTL, err = getEnvTTL("PRESENCE_TTL", 30)
	b.ErrNil(b.Fatal, err)

	p, err := getEnvTTL("PRESENCE_INTERVAL", 30)
	b.ErrNil(b.Fatal, err)
	presenceInterval = time.Duration(p)

	arpTTL, err = getEnvTTL("ARP_TTL", 15)
	b.ErrNil(b.Fatal, err)
	icmpTTL, err = getEnvTTL("PING_TTL", 12)
	b.ErrNil(b.Fatal, err)

	iface, err := net.InterfaceByName(interfaceName)
	b.ErrNil(b.Fatal, err)

	for {
		for deviceId, DeviceInfo := range ha.Devices {
			b.Debug("mac %s, last seen %d, last ip %s", DeviceInfo.Mac, DeviceInfo.LastSeen, DeviceInfo.LastIP)
			tdiff = time.Now().Unix() - DeviceInfo.LastSeen
			if tdiff > presenceTTL && DeviceInfo.LastSeen != 0 {
				b.ErrNil(b.Error, ha.ClientNotHome(DeviceInfo))
				b.Info("not_home src %v %v\n", DeviceInfo.Mac, DeviceInfo)
				ha.Devices[deviceId].LastSeen = 0
			}
			if tdiff > icmpTTL && DeviceInfo.LastSeen != 0 {
				b.Debug("mac %s, last seen %d ago, icmp request %s", DeviceInfo.Mac, tdiff, DeviceInfo.LastIP)
				b.ErrNil(b.Error, Icmp(iface, &DeviceInfo.LastIP))
			}

			if tdiff > arpTTL && DeviceInfo.LastSeen != 0 {
				b.Debug("mac %s, last seen %d ago, arp request %s", DeviceInfo.Mac, tdiff, DeviceInfo.LastIP)
				b.ErrNil(b.Error, Arp(iface, &DeviceInfo.LastIP))
			}

			// if tdiff > 86400 { // if the client is not present 1 day delete from memory
			// 	delete(ha.Devices, deviceId)
			// }
		}
		time.Sleep(time.Second * presenceInterval)
	}

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
