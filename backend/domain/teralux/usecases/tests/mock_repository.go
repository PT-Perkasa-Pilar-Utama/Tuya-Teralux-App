package tests

import (
	"teralux_app/domain/teralux/entities"
	tuya_dtos "teralux_app/domain/tuya/dtos"
)

// MockDeviceRepository is a manual mock for DeviceRepository
type MockDeviceRepository struct {
	CreateFunc         func(device *entities.Device) error
	GetAllFunc         func() ([]entities.Device, error)
	GetByTeraluxIDFunc func(teraluxID string) ([]entities.Device, error)
	GetByIDFunc        func(id string) (*entities.Device, error)
	UpdateFunc         func(device *entities.Device) error
	DeleteFunc         func(id string) error
}

func (m *MockDeviceRepository) Create(device *entities.Device) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(device)
	}
	return nil
}

func (m *MockDeviceRepository) GetAll() ([]entities.Device, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []entities.Device{}, nil
}

func (m *MockDeviceRepository) GetByTeraluxID(teraluxID string) ([]entities.Device, error) {
	if m.GetByTeraluxIDFunc != nil {
		return m.GetByTeraluxIDFunc(teraluxID)
	}
	return []entities.Device{}, nil
}

func (m *MockDeviceRepository) GetByID(id string) (*entities.Device, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockDeviceRepository) Update(device *entities.Device) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(device)
	}
	return nil
}

func (m *MockDeviceRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

// MockDeviceStatusRepository is a manual mock for DeviceStatusRepository
type MockDeviceStatusRepository struct {
	CreateFunc                  func(status *entities.DeviceStatus) error
	GetAllFunc                  func() ([]entities.DeviceStatus, error)
	GetByDeviceIDFunc           func(deviceID string) ([]entities.DeviceStatus, error)
	GetByDeviceIDAndCodeFunc    func(deviceID, code string) (*entities.DeviceStatus, error)
	UpsertFunc                  func(status *entities.DeviceStatus) error
	DeleteByDeviceIDAndCodeFunc func(deviceID, code string) error
	DeleteByDeviceIDFunc        func(deviceID string) error
	UpsertDeviceStatusesFunc    func(deviceID string, statuses []entities.DeviceStatus) error
}

func (m *MockDeviceStatusRepository) Create(status *entities.DeviceStatus) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(status)
	}
	return nil
}

func (m *MockDeviceStatusRepository) GetAll() ([]entities.DeviceStatus, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []entities.DeviceStatus{}, nil
}

func (m *MockDeviceStatusRepository) GetByDeviceID(deviceID string) ([]entities.DeviceStatus, error) {
	if m.GetByDeviceIDFunc != nil {
		return m.GetByDeviceIDFunc(deviceID)
	}
	return []entities.DeviceStatus{}, nil
}

func (m *MockDeviceStatusRepository) GetByDeviceIDAndCode(deviceID, code string) (*entities.DeviceStatus, error) {
	if m.GetByDeviceIDAndCodeFunc != nil {
		return m.GetByDeviceIDAndCodeFunc(deviceID, code)
	}
	return nil, nil
}

func (m *MockDeviceStatusRepository) Upsert(status *entities.DeviceStatus) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(status)
	}
	return nil
}

func (m *MockDeviceStatusRepository) DeleteByDeviceIDAndCode(deviceID, code string) error {
	if m.DeleteByDeviceIDAndCodeFunc != nil {
		return m.DeleteByDeviceIDAndCodeFunc(deviceID, code)
	}
	return nil
}

func (m *MockDeviceStatusRepository) DeleteByDeviceID(deviceID string) error {
	if m.DeleteByDeviceIDFunc != nil {
		return m.DeleteByDeviceIDFunc(deviceID)
	}
	return nil
}

func (m *MockDeviceStatusRepository) UpsertDeviceStatuses(deviceID string, statuses []entities.DeviceStatus) error {
	if m.UpsertDeviceStatusesFunc != nil {
		return m.UpsertDeviceStatusesFunc(deviceID, statuses)
	}
	return nil
}

// MockTuyaAuthUseCase
type MockTuyaAuthUseCase struct {
	AuthenticateFunc func() (*tuya_dtos.TuyaAuthResponseDTO, error)
}

func (m *MockTuyaAuthUseCase) Authenticate() (*tuya_dtos.TuyaAuthResponseDTO, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc()
	}
	return nil, nil
}

// MockTuyaGetDeviceByIDUseCase
type MockTuyaGetDeviceByIDUseCase struct {
	GetDeviceByIDFunc func(accessToken, deviceID string) (*tuya_dtos.TuyaDeviceDTO, error)
}

func (m *MockTuyaGetDeviceByIDUseCase) GetDeviceByID(accessToken, deviceID string) (*tuya_dtos.TuyaDeviceDTO, error) {
	if m.GetDeviceByIDFunc != nil {
		return m.GetDeviceByIDFunc(accessToken, deviceID)
	}
	return nil, nil
}

// MockTeraluxRepository is a manual mock for TeraluxRepository
type MockTeraluxRepository struct {
	CreateFunc          func(teralux *entities.Teralux) error
	GetAllFunc          func() ([]entities.Teralux, error)
	GetByIDFunc         func(id string) (*entities.Teralux, error)
	GetByMacAddressFunc func(macAddress string) (*entities.Teralux, error)
	UpdateFunc          func(teralux *entities.Teralux) error
	DeleteFunc          func(id string) error
}

func (m *MockTeraluxRepository) Create(teralux *entities.Teralux) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(teralux)
	}
	return nil
}

func (m *MockTeraluxRepository) GetAll() ([]entities.Teralux, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []entities.Teralux{}, nil
}

func (m *MockTeraluxRepository) GetByID(id string) (*entities.Teralux, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockTeraluxRepository) GetByMacAddress(macAddress string) (*entities.Teralux, error) {
	if m.GetByMacAddressFunc != nil {
		return m.GetByMacAddressFunc(macAddress)
	}
	return nil, nil
}

func (m *MockTeraluxRepository) Update(teralux *entities.Teralux) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(teralux)
	}
	return nil
}

func (m *MockTeraluxRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}
