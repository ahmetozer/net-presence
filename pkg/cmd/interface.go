package cmd

import (
	"os"

	b "github.com/ahmetozer/basiclog"
)

func InterfaceName() string {
	interfaceName := "eth0"

	if s := os.Getenv("INTERFACE"); s != "" {
		interfaceName = s
	}
	b.Debug("interface is %s", interfaceName)

	return interfaceName
}
