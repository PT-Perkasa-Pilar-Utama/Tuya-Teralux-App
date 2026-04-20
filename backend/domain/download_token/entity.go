package download_token

import "time"

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
