package interfaces

// DownloadTokenCreator defines the interface for creating download tokens
type DownloadTokenCreator interface {
	CreateToken(recipient, objectKey, purpose string, password ...string) (string, error)
}
