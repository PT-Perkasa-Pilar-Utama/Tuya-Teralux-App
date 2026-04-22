package interfaces

import "context"

type ITerminalRepository interface {
	Create(ctx context.Context, terminal *Terminal) error
	GetAll(ctx context.Context) ([]Terminal, error)
	GetAllPaginated(ctx context.Context, offset, limit int, roomID *string) ([]Terminal, int64, error)
	GetByID(ctx context.Context, id string) (*Terminal, error)
	GetByMacAddress(ctx context.Context, macAddress string) (*Terminal, error)
	GetByRoomID(ctx context.Context, roomID string) ([]Terminal, error)
	Update(ctx context.Context, terminal *Terminal) error
	Delete(ctx context.Context, id string) error
}

type Terminal struct {
	ID              string
	MacAddress      string
	RoomID          string
	TuyaUID         string
	Name            string
	DeviceTypeID    int
	AiProvider      string
	AiEngineProfile string
}