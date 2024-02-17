package main

import (
	_ "embed"

	b "github.com/ahmetozer/basiclog"

	cmd "github.com/ahmetozer/net-presence/pkg/cmd"
	ha "github.com/ahmetozer/net-presence/pkg/homeasistant"
	"github.com/ahmetozer/net-presence/pkg/ping"

	"github.com/ahmetozer/net-presence/pkg/presence"
)

//go:embed Readme.md

var Readme string

func main() {

	cmd.Readme = &Readme
	cmd.Flag()

	b.Init()

	ha.Connect()

	interfaceName := cmd.InterfaceName()
	go ping.Existence(interfaceName)
	presence.Watch(interfaceName)

}
