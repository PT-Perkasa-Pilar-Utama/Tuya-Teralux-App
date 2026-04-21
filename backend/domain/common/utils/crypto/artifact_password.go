package crypto

import "time"

type ArtifactType string

const (
	ArtifactTypeZIP ArtifactType = "zip"
	ArtifactTypePDF ArtifactType = "pdf"
)

// ArtifactPassword stores a generated password associated with an encrypted artifact.
// LinkedObjectKey is used to connect ZIP/PDF artifact pairs that belong to the same flow.
type ArtifactPassword struct {
	Password        string
	ArtifactType    ArtifactType
	CreatedAt       time.Time
	LinkedObjectKey string
}
