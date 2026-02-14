package usecases_test

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"teralux_app/domain/recordings/entities"
	"teralux_app/domain/recordings/usecases"
)

// Mocks
type MockRecordingRepository struct {
	mock.Mock
}

func (m *MockRecordingRepository) Save(recording *entities.Recording) error {
	args := m.Called(recording)
	return args.Error(0)
}

func (m *MockRecordingRepository) GetAll(page, limit int) ([]entities.Recording, int64, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]entities.Recording), args.Get(1).(int64), args.Error(2)
}

func (m *MockRecordingRepository) GetByID(id string) (*entities.Recording, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Recording), args.Error(1)
}

func (m *MockRecordingRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	args := m.Called(file, dst)
	return args.Error(0)
}

func (m *MockFileService) EnsureDir(dirName string) error {
	args := m.Called(dirName)
	return args.Error(0)
}

func createTestFileHeader() (*multipart.FileHeader, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.mp3")
	if err != nil {
		return nil, err
	}
	part.Write([]byte("dummy content"))
	writer.Close()

	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, header, err := req.FormFile("file")
	return header, err
}

func TestSaveRecordingUseCase_SaveRecording(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		mockFileService := new(MockFileService)
		uc := usecases.NewSaveRecordingUseCase(mockRepo, mockFileService)

		fileHeader, err := createTestFileHeader()
		assert.NoError(t, err)
		
		// Mock FileService expectations
		// Note: The fileHeader passed to SaveUploadedFile will be the same pointer
		mockFileService.On("SaveUploadedFile", fileHeader, mock.AnythingOfType("string")).Return(nil)

		// Mock Repository expectations
		mockRepo.On("Save", mock.AnythingOfType("*entities.Recording")).Return(nil)

		recording, err := uc.SaveRecording(fileHeader)

		assert.NoError(t, err)
		assert.NotNil(t, recording)
		assert.Contains(t, recording.Filename, ".mp3")
		assert.Equal(t, "test.mp3", recording.OriginalName)
		assert.Contains(t, recording.AudioUrl, "/uploads/audio/")
		
		mockFileService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("File Save Error", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		mockFileService := new(MockFileService)
		uc := usecases.NewSaveRecordingUseCase(mockRepo, mockFileService)

		fileHeader, _ := createTestFileHeader()

		mockFileService.On("SaveUploadedFile", fileHeader, mock.AnythingOfType("string")).Return(errors.New("disk full"))

		recording, err := uc.SaveRecording(fileHeader)

		assert.Error(t, err)
		assert.Nil(t, recording)
		assert.Contains(t, err.Error(), "disk full")

		mockFileService.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Save", mock.Anything)
	})

	t.Run("Database Save Error", func(t *testing.T) {
		mockRepo := new(MockRecordingRepository)
		mockFileService := new(MockFileService)
		uc := usecases.NewSaveRecordingUseCase(mockRepo, mockFileService)

		fileHeader, _ := createTestFileHeader()

		mockFileService.On("SaveUploadedFile", fileHeader, mock.AnythingOfType("string")).Return(nil)
		mockRepo.On("Save", mock.AnythingOfType("*entities.Recording")).Return(errors.New("db error"))

		recording, err := uc.SaveRecording(fileHeader)

		assert.Error(t, err)
		assert.Nil(t, recording)
		assert.Equal(t, "db error", err.Error())

		mockFileService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
