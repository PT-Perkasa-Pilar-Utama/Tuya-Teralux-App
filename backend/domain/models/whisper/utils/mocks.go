package utils

import (
	"context"
	"sensio/domain/models/whisper/dtos"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockBadgerStore struct {
	mock.Mock
}

func (m *MockBadgerStore) Set(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockBadgerStore) SetPreserveTTL(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockBadgerStore) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockBadgerStore) GetWithTTL(key string) ([]byte, time.Duration, error) {
	args := m.Called(key)
	var data []byte
	if val := args.Get(0); val != nil {
		data = val.([]byte)
	}
	return data, args.Get(1).(time.Duration), args.Error(2)
}

func (m *MockBadgerStore) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockBadgerStore) KeysWithPrefix(prefix string) ([]string, error) {
	args := m.Called(prefix)
	return args.Get(0).([]string), args.Error(1)
}

type GenericMockClient struct {
	mock.Mock
}

func (m *GenericMockClient) HealthCheck() bool {
	return m.Called().Bool(0)
}

func (m *GenericMockClient) WhisperHealthCheck() bool {
	return m.Called().Bool(0)
}

func (m *GenericMockClient) Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*dtos.WhisperResult, error) {
	args := m.Called(ctx, audioPath, language, diarize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.WhisperResult), args.Error(1)
}

func (m *GenericMockClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	args := m.Called(topic, qos, retained, payload)
	return args.Error(0)
}
