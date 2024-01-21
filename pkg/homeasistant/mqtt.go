package homeasistant

import (
	"encoding/json"
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

type deviceConfig struct {
	StateTopic string `json:"state_topic"`
	// JsonAttributesTopic string `json:"json_attributes_topic,omitempty"`
	ObjectId       string `json:"object_id,omitempty"`
	Name           string `json:"name,omitempty"`
	PayloadHome    string `json:"payload_home,omitempty"`
	PayloadNotHome string `json:"payload_not_home,omitempty"`
	SourceType     string `json:"source_type,omitempty"`
	// ValueTemplate  string `json:"value_template,omitempty"`
	// Device         struct {
	// 	Connections [][]string `json:"connections,omitempty"`
	// } `json:"device,omitempty"`
}

const (
	PayloadHome    = "home"
	PayloadNotHome = "not_home"
	SourceType     = "router"
)

func ClientNew(device *DeviceInfo) error {
	id := GetId(device)

	data, err := json.Marshal(deviceConfig{
		StateTopic:     "presence/" + id + "/state",
		Name:           "presence_" + id,
		ObjectId:       "net_presence_" + id,
		PayloadHome:    PayloadHome,
		PayloadNotHome: PayloadNotHome,
		SourceType:     SourceType,
	})
	b.ErrNil(b.Error, err)
	token := publish("homeassistant/device_tracker/"+id+"/config", 0, false, string(data))
	token.Wait()
	return token.Error()
}

func State(device *DeviceInfo, state string) error {
	token := publish("presence/"+GetId(device)+"/state", 0, false, state)
	token.Wait()
	return token.Error()
}

func idByMac(d *DeviceInfo) string {

	return strings.Replace(d.Mac, "-", "_", -1)
}

func idByIp(d *DeviceInfo) string {

	return strings.Replace(d.IP.String(), ".", "_", -1)
}

func idByMacIP(d *DeviceInfo) string {

	return strings.Replace(d.Mac, "-", "_", -1) + strings.Replace(d.IP.String(), ".", "_", -1)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	b.Info("MQTT Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	b.ErrNil(b.Error, err)
}
