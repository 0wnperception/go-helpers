package mqtt_connector

import (
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttConfig struct {
	Host              string
	Port              string
	ConnectionType    string
	ClientId          string
	UserName          string
	Password          string
	PingTimeout       time.Duration
	ReconnectAttempts int
	ReconnectTimeout  time.Duration
	DisconnectTimeout uint
}

type Observer interface {
	OnConnectionLostHandler(err error)
}

type MqttConnector struct {
	sync.Locker
	cfg             *MqttConfig
	client          mqtt.Client
	opts            *mqtt.ClientOptions
	cancelReconnect chan struct{}
	observer        Observer
}

func NewMqttConnector(cfg *MqttConfig) (*MqttConnector, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.ConnectionType + "://" + cfg.Host + ":" + cfg.Port)
	opts.SetClientID(cfg.ClientId)
	if cfg.UserName != "" {
		opts.SetUsername(cfg.UserName)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	opts.SetPingTimeout(cfg.PingTimeout)
	opts.SetAutoReconnect(false)

	mc := &MqttConnector{
		Locker:          &sync.RWMutex{},
		cfg:             cfg,
		client:          mqtt.NewClient(opts),
		opts:            opts,
		cancelReconnect: make(chan struct{}, 1),
	}

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		mc.onConnectionLostHandler(err)
	})

	return mc, nil
}

func (mc *MqttConnector) Handle(observer Observer) error {
	mc.observer = observer
	return mc.reconnect()
}

func (mc *MqttConnector) Subscripe(topic string, callback func([]byte)) error {
	if token := mc.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		go callback(msg.Payload())
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mc *MqttConnector) Publish(topic string, payload []byte) error {
	mc.Lock()
	defer mc.Unlock()
	if token := mc.client.Publish(topic, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mc *MqttConnector) StopHandle() {
	close(mc.cancelReconnect)
	if mc.client != nil && mc.client.IsConnected() {
		mc.client.Disconnect(mc.cfg.DisconnectTimeout)
	}
	mc.client = nil
}

func (mc *MqttConnector) reconnect() error {
	timer := time.NewTicker(mc.cfg.ReconnectTimeout)
	counter := mc.cfg.ReconnectAttempts
	for {
		select {
		case <-timer.C:
			mc.client = mqtt.NewClient(mc.opts)
			token := mc.client.Connect()
			token.Wait()
			if token.Error() != nil {
				if counter--; counter == 0 {
					return token.Error()
				}
			} else {
				return nil
			}
		case <-mc.cancelReconnect:
			return nil
		}
	}
}

func (mc *MqttConnector) onConnectionLostHandler(err error) {
	if err := mc.reconnect(); err != nil {
		mc.observer.OnConnectionLostHandler(err)
	}
}
