package interfaces

import (
	"sensio/domain/tuya/dtos"
)

type AuthUseCase interface {
	Authenticate() (*dtos.TuyaAuthResponseDTO, error)
	GetTuyaAccessToken() (string, error)
}
