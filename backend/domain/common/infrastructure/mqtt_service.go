package infrastructure

import (
	"fmt"
	"time"

	"teralux_app/domain/common/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MqttService defines the interface for MQTT operations
type MqttService interface {
	Connect() error
	Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error
	Unsubscribe(topic string) error
	Publish(topic string, qos byte, retained bool, payload interface{}) error
	IsConnected() bool
	Close()
}

// mqttService manages the MQTT client connection
type mqttService struct {
	client mqtt.Client
	config *utils.Config
}

// NewMqttService initializes a new MQTT service
func NewMqttService(cfg *utils.Config) MqttService {
	s := &mqttService{
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
func (s *mqttService) Connect() error {
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Subscribe subscribes to a topic
func (s *mqttService) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	if token := s.client.Subscribe(topic, qos, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	utils.LogDebug("Subscribed to MQTT topic: %s", topic)
	return nil
}

// Unsubscribe unsubscribes from a topic
func (s *mqttService) Unsubscribe(topic string) error {
	if token := s.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Publish publishes a message to a topic
func (s *mqttService) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if !s.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	token := s.client.Publish(topic, qos, retained, payload)
	token.Wait()
	return token.Error()
}

// IsConnected checks if the client is connected
func (s *mqttService) IsConnected() bool {
	return s.client.IsConnected()
}

// Close disconnects the client
func (s *mqttService) Close() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
		utils.LogInfo("Common MQTT Disconnected")
	}
}
