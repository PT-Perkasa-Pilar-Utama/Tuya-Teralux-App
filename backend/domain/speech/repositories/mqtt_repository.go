package repositories

import (
	"sync"

	common_infra "teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttRepository struct {
	mqttSvc  common_infra.MqttService
	topic    string
	mu       sync.RWMutex
	callback func([]byte)
}

func NewMqttRepository(mqttSvc common_infra.MqttService, cfg *utils.Config) *MqttRepository {
	r := &MqttRepository{
		mqttSvc: mqttSvc,
		topic:   cfg.MqttTopic,
	}

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

	return r.mqttSvc.Subscribe(r.topic, 0, r.handleMessage)
}

func (r *MqttRepository) Publish(message string) error {
	return r.mqttSvc.Publish(r.topic, 0, false, message)
}

func (r *MqttRepository) Close() {
	// MqttService handles its own closing in common infrastructure
}
