package homeasistant

import (
	"os"
	"strconv"
	"strings"
	"time"

	b "github.com/ahmetozer/basiclog"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	publish func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token
	GetId   func(*DeviceInfo) string
)

func Connect() {

	mqttServer := "tcp://127.0.0.1:1883"

	if s := os.Getenv("MQTT_SERVER"); s != "" {
		mqttServer = s
	}
	b.Debug("mqtt server %s", mqttServer)

	client := mqtt.NewClient((&mqtt.ClientOptions{
		OnConnect:        connectHandler,
		OnConnectionLost: connectLostHandler,
		AutoReconnect:    true,
		Username:         os.Getenv("MQTT_USERNAME"),
		Password:         os.Getenv("MQTT_PASSWORD"),
		ClientID:         "mac_presence_" + strconv.Itoa(os.Getpid()),
		ConnectTimeout:   time.Second * 2,
	}).AddBroker(mqttServer))

	b.Debug("mqtt client %#s", client)

	if token := client.Connect(); token.Wait() {
		b.ErrNil(b.Fatal, token.Error())
	}

	publish = client.Publish

	switch s := strings.ToUpper(os.Getenv("ID_SOURCE")); s {
	case "", "IP":
		GetId = idByIp
	case "MAC":
		GetId = idByMac
	case "MAC_IP":
		GetId = idByMacIP
	default:
		b.Fatal("ID_SOURCE type is unknown, expected IP, MAC, MAC_IP but got '%s'", s)
	}

}

func ClientNew(device *DeviceInfo) error {
	id := GetId(device)
	return publish("homeassistant/device_tracker/"+id+"/config", 0, false, `{"state_topic": "presence/`+id+`/state", "name": "`+id+`", "payload_home": "home", "payload_not_home": "not_home"}`).Error()
}

func ClientAtHome(device *DeviceInfo) error {
	id := GetId(device)
	return publish("presence/"+id+"/state", 0, false, "home").Error()
}

func ClientNotHome(device *DeviceInfo) error {
	id := GetId(device)
	return publish("presence/"+id+"/state", 0, false, "not_home").Error()

}

func idByMac(d *DeviceInfo) string {

	return strings.Replace(d.Mac, "-", "_", -1)
}

func idByIp(d *DeviceInfo) string {

	return strings.Replace(d.LastIP.String(), ".", "_", -1)
}

func idByMacIP(d *DeviceInfo) string {

	return strings.Replace(d.Mac, "-", "_", -1) + strings.Replace(d.LastIP.String(), ".", "_", -1)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	b.Info("MQTT Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	b.ErrNil(b.Error, err)
}
