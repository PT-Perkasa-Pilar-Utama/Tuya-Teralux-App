package infrastructure

import (
	"fmt"
	"strings"
	"time"

	"sensio/domain/common/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MqttService manages the MQTT client connection
type MqttService struct {
	client mqtt.Client
	config *utils.Config
}

// NewMqttService initializes a new MQTT service
func NewMqttService(cfg *utils.Config) *MqttService {
	s := &MqttService{
		config: cfg,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MqttBroker)
	opts.SetClientID(fmt.Sprintf("%s_backend_%d", cfg.MqttUsername, time.Now().Unix()))
	opts.SetUsername(cfg.MqttUsername)
	opts.SetPassword(cfg.MqttPassword)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(false) // Maintain session for reliability

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		utils.LogInfo("Common MQTT Connected to %s", cfg.MqttBroker)
	})

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		utils.LogInfo("Common MQTT Connection lost: %v", err)
	})

	s.client = mqtt.NewClient(opts)
	return s
}

// Connect initiates the connection to the MQTT broker
func (s *MqttService) Connect() error {
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Subscribe subscribes to a topic. If the topic contains wildcards (+ or #), it uses Shared Subscription with group 'sensio'.
func (s *MqttService) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	modifiedTopic := topic
	if (strings.Contains(topic, "+") || strings.Contains(topic, "#")) && !strings.HasPrefix(topic, "$share/") {
		modifiedTopic = "$share/sensio/" + topic
	}

	if token := s.client.Subscribe(modifiedTopic, qos, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	utils.LogDebug("Subscribed to MQTT topic: %s", modifiedTopic)
	return nil
}

// Unsubscribe unsubscribes from a topic. Handles shared subscription topics automatically.
func (s *MqttService) Unsubscribe(topic string) error {
	modifiedTopic := topic
	if (strings.Contains(topic, "+") || strings.Contains(topic, "#")) && !strings.HasPrefix(topic, "$share/") {
		modifiedTopic = "$share/sensio/" + topic
	}

	if token := s.client.Unsubscribe(modifiedTopic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Publish publishes a message to a topic
func (s *MqttService) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if !s.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	token := s.client.Publish(topic, qos, retained, payload)
	token.Wait()
	return token.Error()
}

// IsConnected checks if the client is connected
func (s *MqttService) IsConnected() bool {
	return s.client.IsConnected()
}

// Close disconnects the client
func (s *MqttService) Close() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
		utils.LogInfo("Common MQTT Disconnected")
	}
}
