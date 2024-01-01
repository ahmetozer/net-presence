# Mac Presence

Device presence sensor for Home Assistant based on network traffic.

## How it works?

A variety of media services and devices use neighbor discovery protocols which create multicast and broadcast packages at the local network. Those packages arrive at multiple or all of the devices in the network and this system primarily uses those packets for presence detection.

## What is the difference?

At Legacy nmap methods, auto device detection is done by network scan which leads high count of ARP packages at your network and invokes your mobile devices or hits firewall limits that result in no response or battery drainage of pocket devices.  
This approach improves the result confidence of the system compared to scan-only methods of false negative (not_home) detection and prevents false triggers at Home Assistant this system overcomes those issues by listening to your local network always, so it does not require scan packages by default.
Broadcast query is only invoked if the device does not create traffic in a certain time and has no response for ICMP packets, ARP will used as a fallback method to overcome firewall limitations.

## Configuration

### Mac Presence

The system configuration is done by setting environment variables.

```bash
MAC_TTL=30 # Informing the Home Assistant of the device status as away.
ARP_TTL=15 # Creating ARP packages that query MAC address based on the last seen IP address.
PING_TTL=12 # Pinging of the device with ICMP packages after the duration of no packages arriving in the system.
INTERFACE=eth0 # The interface from which network traffic will be listened to.
MQTT_SERVER=tcp://127.0.0.1:1883 # Mqtt server address.
MQTT_USERNAME=""
MQTT_PASSWORD=""
LOG_LEVEL=info # Default is info, options are debug|info|error.
```


### Home Asistant

When you connect your home assistant to the MQTT server, the application sends auto-configuration packets and the new devices will appear at the Home Assistant entity part.

## Install

Before getting the repository, you might need libpcap library in your system. You can get this library at Debian with `apt install -y libpcap-dev` command.

Install mac-presence with the below command.

```bash
go install github.com/ahmetozer/mac-presence@latest
```