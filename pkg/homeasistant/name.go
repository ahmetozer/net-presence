package homeasistant

import (
	"context"
	"net"
	"os"
	"time"
)

var (
	dnsServer  string
	clientName func(id string, ha *DeviceInfo) string
	resolver   = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second,
			}
			return d.DialContext(ctx, network, dnsServer)
		},
	}
)

func init() {
	clientName = nameById
}

func nameById(id string, device *DeviceInfo) string {
	return "presence_" + id
}

func nameByDNS(id string, device *DeviceInfo) string {

	result, err := resolver.LookupAddr(context.Background(), device.IP.String())
	if err == nil && len(result) != 0 {
		if result[0][len(result[0])-1] == 46 {
			return result[0][:len(result[0])-1]
		}
		return result[0]
	}

	return nameById(id, device)
}

func checkDnsVar() {
	if dnsServer = os.Getenv("DNS_SERVER"); dnsServer != "" {
		clientName = nameByDNS
	}
}
