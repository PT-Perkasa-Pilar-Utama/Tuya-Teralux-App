package download_token

import (
	"context"
	"testing"
	"time"

	"sensio/domain/common/infrastructure"
	"sensio/domain/crypto"

	"github.com/stretchr/testify/assert"
)

func TestToken_CreateAndResolve(t *testing.T) {
	store := NewStore()
	password, _ := crypto.GenerateStrongPassword()
	service := &DownloadTokenService{
		store:           store,
		now:             func() time.Time { return time.Now() },
		storageProvider: &mockStorageProvider{},
	}

	token, err := service.CreateToken("user@example.com", "audio/test.zip", "audio_zip", password)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.TokenID)
	assert.Equal(t, "user@example.com", token.Recipient)
	assert.Equal(t, "audio/test.zip", token.ObjectKey)
	assert.Equal(t, password, token.Password)
	assert.True(t, token.ExpiresAt.After(time.Now()))

	_, err = service.ResolveToken(token.TokenID)
	assert.NoError(t, err)

	_, err = service.ResolveToken(token.TokenID)
	assert.Error(t, err)
	assert.Equal(t, ErrTokenConsumed, err)
}

type mockStorageProvider struct{}

func (m *mockStorageProvider) Put(ctx context.Context, key string, data []byte, contentType string) error {
	return nil
}

func (m *mockStorageProvider) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (m *mockStorageProvider) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorageProvider) PresignPut(ctx context.Context, key string, contentType string, ttl int64) (string, error) {
	return "https://example.com-signed-url", nil
}

var _ infrastructure.StorageProvider = (*mockStorageProvider)(nil)

func TestToken_RevokeRejected(t *testing.T) {
	store := NewStore()
	service := &DownloadTokenService{
		store: store,
		now:   func() time.Time { return time.Now() },
	}

	token, err := service.CreateToken("user@example.com", "audio/test.zip", "audio_zip")
	assert.NoError(t, err)

	err = service.RevokeToken(token.TokenID)
	assert.NoError(t, err)

	_, err = service.ResolveToken(token.TokenID)
	assert.Error(t, err)
	assert.Equal(t, ErrTokenRevoked, err)
}
