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

The system configuration is done by setting environment variables or arguments.

```bash
PRESENCE_TTL=60 #The threshold of no activity duration to inform the home assistant device is not home.
PRESENCE_INTERVAL=15 # Interval of the presence verify system 
ARP_TTL=35 # create an ARP query if the device is not present more than 35.
PING_TTL=12 # create an ICMP query if the device is not present more than 12.
INTERFACE=eth0 # The interface from which network traffic will be listened to.

ID_SOURCE=IP # Home assistant ID combination based on network information. 
# If you have a flat network (which means wifi repeaters work at bridge mode and no NAT present) you can use MAC address instead of IP.
# If you have cheap Wifi repeaters, which do not have MESH support and bridge, you can use IP mode.
# The last mode MAC_IP combination can be used for debugging purposes.

MQTT_SERVER=tcp://127.0.0.1:1883 # Mqtt server address.
MQTT_USERNAME=""
MQTT_PASSWORD=""
LOG_LEVEL=info # Default is info, options are debug|info|error|fatal.

DNS_SERVER="" # Set DNS server which integrated to DNS server (IP:PORT) to resolve IP to hostname. IF its installed to raspberry pi, you can set 127.0.0.1:53 or your routers address 192.168.1.1:53

LOG_FILE="" # Set log file
```

For arguments, you can use `-help` to list arguments.

### Home Asistant

When you connect your home assistant to the MQTT server, the application sends auto-configuration packets and the new devices will appear at the Home Assistant entity part.

## Install

Libpcap library is require in your system. You can get this library at Debian with `apt install -y libpcap-dev` command.

### Prebuilt

Visit [https://github.com/ahmetozer/net-presence/releases](https://github.com/ahmetozer/net-presence/releases) and download your binnary.

```bash
wget https://github.com/ahmetozer/net-presence/releases/download/xxx/x.gz -O net-presence.gz # replace x with version
gunzip net-presence.gz
chmod +x net-presence
mv net-presence /usr/bin/
```

### From source

Install mac-presence with the below command.

```bash
go install github.com/ahmetozer/mac-presence@latest
```