package main

import (
	"os"

	b "github.com/ahmetozer/basiclog"

	ha "github.com/ahmetozer/net-presence/pkg/homeasistant"
	"github.com/ahmetozer/net-presence/pkg/ping"
	"github.com/ahmetozer/net-presence/pkg/presence"
)

func main() {

	b.Init()

	ha.Connect()

	interfaceName := "eth0"

	if s := os.Getenv("INTERFACE"); s != "" {
		interfaceName = s
	}
	b.Debug("interface is %s", interfaceName)

	go ping.Existence(interfaceName)
	presence.Watch(interfaceName)

}
