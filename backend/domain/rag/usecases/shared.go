package usecases

type mqttPublisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}
