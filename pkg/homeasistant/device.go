package homeasistant

import "net"

type DeviceInfo struct {
	LastSeen int64
	IP       net.IP
	Mac      string
}

type Device map[string]*DeviceInfo

var Devices Device

func init() {
	Devices = make(Device)
}
