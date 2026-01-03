package tests

import (
	"teralux_app/domain/teralux/entities"
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
	CreateFunc               func(status *entities.DeviceStatus) error
	GetAllFunc               func() ([]entities.DeviceStatus, error)
	GetByDeviceIDFunc        func(deviceID string) ([]entities.DeviceStatus, error)
	GetByDeviceIDAndCodeFunc func(deviceID, code string) (*entities.DeviceStatus, error)
	GetByIDFunc              func(id string) (*entities.DeviceStatus, error)
	UpdateFunc               func(status *entities.DeviceStatus) error
	DeleteFunc               func(id string) error
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

func (m *MockDeviceStatusRepository) GetByID(id string) (*entities.DeviceStatus, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockDeviceStatusRepository) Update(status *entities.DeviceStatus) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(status)
	}
	return nil
}

func (m *MockDeviceStatusRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
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
