package mocks

import (
	"encoding/json"
	"sync"
)

type MockMQTTPublisher struct {
	mu         sync.RWMutex
	Published  []MQTTMessage
	ShouldFail bool
}

type MQTTMessage struct {
	Topic   string
	Payload interface{}
	QOS     byte
}

func NewMockMQTTPublisher() *MockMQTTPublisher {
	return &MockMQTTPublisher{
		Published: make([]MQTTMessage, 0),
	}
}

func (m *MockMQTTPublisher) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if m.ShouldFail {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.Published = append(m.Published, MQTTMessage{
		Topic:   topic,
		Payload: payload,
		QOS:     qos,
	})
	return nil
}

func (m *MockMQTTPublisher) GetPublished(topic string) []MQTTMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var messages []MQTTMessage
	for _, msg := range m.Published {
		if msg.Topic == topic {
			messages = append(messages, msg)
		}
	}
	return messages
}

func (m *MockMQTTPublisher) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Published = make([]MQTTMessage, 0)
}

func (m *MockMQTTPublisher) GetLastPayload(topic string) interface{} {
	messages := m.GetPublished(topic)
	if len(messages) == 0 {
		return nil
	}
	return messages[len(messages)-1].Payload
}

func (m *MockMQTTPublisher) PayloadToJSON(topic string) ([]byte, error) {
	payload := m.GetLastPayload(topic)
	return json.Marshal(payload)
}
