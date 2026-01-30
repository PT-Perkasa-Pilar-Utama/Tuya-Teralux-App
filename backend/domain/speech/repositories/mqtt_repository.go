package repositories

import (
	"fmt"
	"sync"
	"time"

	"teralux_app/domain/common/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttRepository struct {
	client   mqtt.Client
	topic    string
	mu       sync.RWMutex
	callback func([]byte)
}

func NewMqttRepository(cfg *utils.Config) *MqttRepository {
	r := &MqttRepository{
		topic: cfg.MqttTopic,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MqttBroker)
	opts.SetClientID(cfg.MqttUsername + "_client_" + fmt.Sprintf("%d", time.Now().Unix()))
	opts.SetUsername(cfg.MqttUsername)
	opts.SetPassword(cfg.MqttPassword)
	opts.SetAutoReconnect(true)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		utils.LogInfo("MQTT Connected to %s", cfg.MqttBroker)
		if token := c.Subscribe(r.topic, 0, r.handleMessage); token.Wait() && token.Error() != nil {
			utils.LogError("Failed to subscribe to topic %s: %v", r.topic, token.Error())
		} else {
			utils.LogInfo("Subscribed to MQTT topic: %s", r.topic)
		}
	})

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		utils.LogInfo("MQTT Connection lost: %v", err)
	})

	r.client = mqtt.NewClient(opts)

	go func() {
		if token := r.client.Connect(); token.Wait() && token.Error() != nil {
			utils.LogError("MQTT Initial Connection Error: %v", token.Error())
		}
	}()

	return r
}

func (r *MqttRepository) handleMessage(client mqtt.Client, msg mqtt.Message) {
	r.mu.RLock()
	cb := r.callback
	r.mu.RUnlock()

	payload := msg.Payload()
	utils.LogDebug("Received MQTT message on topic %s (size: %d bytes)", msg.Topic(), len(payload))

	if cb != nil {
		cb(payload)
	}
}

func (r *MqttRepository) Subscribe(callback func([]byte)) error {
	r.mu.Lock()
	r.callback = callback
	r.mu.Unlock()

	if r.client.IsConnected() {
		_ = r.client.Subscribe(r.topic, 0, r.handleMessage)
	}

	return nil
}

func (r *MqttRepository) Publish(message string) error {
	if !r.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	token := r.client.Publish(r.topic, 0, false, message)
	token.Wait()
	return token.Error()
}

func (r *MqttRepository) Close() {
	if r.client.IsConnected() {
		r.client.Disconnect(250)
	}
}
