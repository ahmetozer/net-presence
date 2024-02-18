package cmd

import (
	"flag"
	"fmt"
	"os"
)

var Readme *string
var BuildVersion string
var BuildDate string
var BuildCommit string

func Flag() {
	readme := flag.Bool("readme", false, "print readme")
	help := flag.Bool("help", false, "print this")

	PRESENCE_TTL := flag.String("presence-ttl", "", "The threshold of no activity duration to inform the home assistant device is not home")
	PRESENCE_INTERVAL := flag.String("presence-interval", "", "Interval of the presence verify system")
	ARP_TTL := flag.String("arp-ttl", "", "create an ARP query if the device is not present more than 35")
	PING_TTL := flag.String("ping-ttl", "", "create an ICMP query if the device is not present more than 12")
	INTERFACE := flag.String("interface", "", "interface from which network traffic will be listened to")
	ID_SOURCE := flag.String("id-source", "", "Home assistant ID combination based on network information")

	MQTT_SERVER := flag.String("mqtt-server", "", "Mqtt server address")
	MQTT_USERNAME := flag.String("mqtt-username", "", "Mqtt server username")
	MQTT_PASSWORD := flag.String("mqtt-password", "", "Mqtt server password")
	DNS_SERVER := flag.String("dns-server", "", "DNS server to resolve IP and hostname")

	LOG_LEVEL := flag.String("log-level", "", "Default is info, options are debug|info|error|fatal")
	LOG_FILE := flag.String("log-file", "", "Set output to file instead std")

	VERSION := flag.Bool("version", false, "print version")

	flag.Parse()

	if *PRESENCE_TTL != "" {
		os.Setenv("PRESENCE_TTL", *PRESENCE_TTL)
	}
	if *PRESENCE_INTERVAL != "" {
		os.Setenv("PRESENCE_INTERVAL", *PRESENCE_INTERVAL)
	}
	if *ARP_TTL != "" {
		os.Setenv("ARP_TTL", *ARP_TTL)
	}
	if *PING_TTL != "" {
		os.Setenv("PING_TTL", *PING_TTL)
	}
	if *INTERFACE != "" {
		os.Setenv("INTERFACE", *INTERFACE)
	}
	if *ID_SOURCE != "" {
		os.Setenv("ID_SOURCE", *ID_SOURCE)
	}
	if *MQTT_SERVER != "" {
		os.Setenv("MQTT_SERVER", *MQTT_SERVER)
	}
	if *MQTT_USERNAME != "" {
		os.Setenv("MQTT_USERNAME", *MQTT_USERNAME)
	}
	if *MQTT_PASSWORD != "" {
		os.Setenv("MQTT_PASSWORD", *MQTT_PASSWORD)
	}
	if *LOG_LEVEL != "" {
		os.Setenv("LOG_LEVEL", *LOG_LEVEL)
	}
	if *LOG_FILE != "" {
		os.Setenv("LOG_FILE", *LOG_FILE)
	}

	if *DNS_SERVER != "" {
		os.Setenv("DNS_SERVER", *DNS_SERVER)
	}

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *readme {
		fmt.Printf("%s\n", *Readme)
		os.Exit(0)
	}

	if *VERSION {
		fmt.Printf("version %s, commit %s, build date %s", BuildVersion, BuildCommit, BuildDate)
		os.Exit(0)
	}

}
