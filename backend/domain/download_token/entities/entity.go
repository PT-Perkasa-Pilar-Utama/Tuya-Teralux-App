package entities

import (
	"errors"
	"time"
)

var (
	ErrTokenNotFound = errors.New("download token not found")
	ErrTokenExpired  = errors.New("download token expired")
	ErrTokenConsumed = errors.New("download token consumed")
	ErrTokenRevoked  = errors.New("download token revoked")
	ErrObjectNotFound = errors.New("object not found in storage")
)

type Token struct {
	TokenID    string
	Recipient  string
	ObjectKey  string
	Purpose    string
	Password   string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	RevokedAt  *time.Time
}
