package infrastructure

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"sensio/domain/common/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// IMqttService defines the interface for MQTT operations
type IMqttService interface {
	Connect() error
	Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error
	Unsubscribe(topic string) error
	Publish(topic string, qos byte, retained bool, payload interface{}) error
	IsConnected() bool
	Close()
}

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

	// Support WSS (WebSocket Secure) and MQTT over TLS
	brokerURL := cfg.MqttBroker
	
	// Check if using WSS (WebSocket Secure)
	if strings.HasPrefix(brokerURL, "wss://") {
		// WSS URL - use as-is for WebSocket connection
		opts.AddBroker(brokerURL)
		utils.LogInfo("MQTT WSS mode: %s", brokerURL)
	} else if strings.HasPrefix(brokerURL, "ws://") {
		// WS URL - use as-is
		opts.AddBroker(brokerURL)
		utils.LogInfo("MQTT WS mode: %s", brokerURL)
	} else {
		// Traditional MQTT/mqtts - extract host:port
		if strings.HasPrefix(brokerURL, "ssl://") ||
		   strings.HasPrefix(brokerURL, "tcps://") ||
		   strings.HasPrefix(brokerURL, "mqtts://") ||
		   strings.HasPrefix(brokerURL, "mqtt://") ||
		   strings.HasPrefix(brokerURL, "tcp://") {
			brokerURL = strings.SplitN(brokerURL, "://", 2)[1]
		}
		opts.AddBroker(brokerURL)
	}

	opts.SetClientID(fmt.Sprintf("%s_backend_%d", cfg.MqttUsername, time.Now().UnixNano()))
	opts.SetUsername(cfg.MqttUsername)
	opts.SetPassword(cfg.MqttPassword)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true) // Use clean session for fresh connection
	opts.SetConnectTimeout(30 * time.Second) // Longer timeout for TLS handshake
	opts.SetKeepAlive(30 * time.Second) // MQTT keepalive (lowered to prevent proxy idle timeouts)
	opts.SetOrderMatters(false) // Allow parallel publishes
	opts.SetMaxReconnectInterval(5 * time.Second)

	// Enable TLS if the broker URL starts with ssl://, tcps://, or mqtts://
	if strings.HasPrefix(cfg.MqttBroker, "ssl://") ||
	   strings.HasPrefix(cfg.MqttBroker, "tcps://") ||
	   strings.HasPrefix(cfg.MqttBroker, "mqtts://") {
		// Extract hostname for SNI (Server Name Indication)
		brokerHost := strings.SplitN(strings.SplitN(cfg.MqttBroker, "://", 2)[1], ":", 2)[0]
		if brokerHost != "" {
			// Load CA certificate from standard path (same as mosquitto --cafile)
			rootCAs := x509.NewCertPool()
			caCertPath := "/etc/ssl/certs/ca-certificates.crt"
			caCerts, err := os.ReadFile(caCertPath)
			if err != nil {
				utils.LogError("Failed to read CA certs from %s: %v", caCertPath, err)
			} else {
				if !rootCAs.AppendCertsFromPEM(caCerts) {
					utils.LogError("Failed to append CA certs from %s", caCertPath)
				}
			}
			
			// Load client certificate and key for mutual TLS (mTLS)
			tlsConfig := &tls.Config{
				MinVersion:         tls.VersionTLS12,
				ServerName:         brokerHost,
				RootCAs:            rootCAs,
				InsecureSkipVerify: false,
			}
			
			// Try to load client certificate for mTLS
			clientCertPath := "/tmp/mqtt_client.crt"
			clientKeyPath := "/tmp/mqtt_client.key"
			cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
			if err != nil {
				utils.LogError("Failed to load client certificate: %v (mTLS may be required)", err)
			} else {
				tlsConfig.Certificates = []tls.Certificate{cert}
				utils.LogDebug("Client certificate loaded for mTLS")
			}
			
			opts.SetTLSConfig(tlsConfig)
			utils.LogDebug("MQTT TLS configured: host=%s, SNI=%s", brokerHost, brokerHost)
		}
	}

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		utils.LogInfo("Common MQTT Connected to %s", cfg.MqttBroker)
	})

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		utils.LogInfo("Common MQTT Connection lost: %v", err)
	})

	opts.SetReconnectingHandler(func(c mqtt.Client, opts *mqtt.ClientOptions) {
		utils.LogInfo("Common MQTT attempting reconnect to %s", cfg.MqttBroker)
	})

	s.client = mqtt.NewClient(opts)
	return s
}

// Connect initiates the connection to the MQTT broker with retry logic
func (s *MqttService) Connect() error {
	utils.LogInfo("Connecting to MQTT broker: %s", s.config.MqttBroker)
	
	maxRetries := 3
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		utils.LogDebug("MQTT connect attempt %d/%d", attempt, maxRetries)
		
		token := s.client.Connect()
		token.Wait()
		
		if token.Error() == nil {
			utils.LogInfo("Successfully connected to MQTT broker: %s", s.config.MqttBroker)
			return nil
		}
		
		lastErr = token.Error()
		utils.LogError("MQTT connect attempt %d failed: %v", attempt, lastErr)
		
		if attempt < maxRetries {
			time.Sleep(2 * time.Second)
		}
	}
	
	return fmt.Errorf("failed to connect to MQTT after %d attempts: %w", maxRetries, lastErr)
}

// Subscribe subscribes to a topic. If the topic contains wildcards (+ or #), it uses Shared Subscription with group 'sensio'.
func (s *MqttService) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	finalTopic := topic
	// If it contains wildcards and is not already a shared subscription
	if (strings.Contains(topic, "+") || strings.Contains(topic, "#")) && !strings.HasPrefix(topic, "$share/") {
		finalTopic = fmt.Sprintf("$share/sensio/%s", topic)
	}

	if token := s.client.Subscribe(finalTopic, qos, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	utils.LogDebug("Subscribed to MQTT topic: %s", finalTopic)
	return nil
}

// Unsubscribe unsubscribes from a topic. Handles shared subscription topics automatically.
func (s *MqttService) Unsubscribe(topic string) error {
	if token := s.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Publish publishes a message to a topic
func (s *MqttService) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if !s.client.IsConnected() {
		// Debug log only - MQTT is optional for push notifications
		utils.LogDebug("MQTT publish skipped: client not connected (broker: %s)", s.config.MqttBroker)
		return fmt.Errorf("MQTT client not connected")
	}

	utils.LogDebug("MQTT publishing to topic: %s (payload size: %d bytes)", topic, len(fmt.Sprintf("%v", payload)))

	token := s.client.Publish(topic, qos, retained, payload)
	token.Wait()

	if token.Error() != nil {
		utils.LogError("MQTT publish to %s failed: %v", topic, token.Error())
		return token.Error()
	}

	utils.LogDebug("MQTT publish success to topic: %s", topic)
	return nil
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
