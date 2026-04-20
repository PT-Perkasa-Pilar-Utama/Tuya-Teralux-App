package usecases

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	"sensio/domain/recordings/entities"
	"sensio/domain/recordings/repositories"
	"sensio/domain/recordings/services"
)

type SaveRecordingUseCase interface {
	SaveRecording(file *multipart.FileHeader, macAddress, baseURL string, opts ...SaveRecordingOption) (*entities.Recording, error)
	SaveRecordingFromBytes(data []byte, originalName, macAddress, baseURL string, opts ...SaveRecordingOption) (*entities.Recording, error)
	SaveRecordingFromPath(path, originalName, macAddress, baseURL string, opts ...SaveRecordingOption) (*entities.Recording, error)
}

type SaveRecordingOption struct {
	NotifyBIG bool
}

type saveRecordingUseCase struct {
	repo        repositories.RecordingRepository
	fileService infrastructure.FileService
	bigService  services.BIGRoomAudioUpdateService
}

func NewSaveRecordingUseCase(repo repositories.RecordingRepository, fileService infrastructure.FileService, bigService services.BIGRoomAudioUpdateService) SaveRecordingUseCase {
	return &saveRecordingUseCase{
		repo:        repo,
		fileService: fileService,
		bigService:  bigService,
	}
}

func shouldNotifyBIG(opts []SaveRecordingOption) bool {
	for _, opt := range opts {
		if opt.NotifyBIG {
			return true
		}
	}

	return false
}

func (uc *saveRecordingUseCase) SaveRecording(fileHeader *multipart.FileHeader, macAddress, baseURL string, opts ...SaveRecordingOption) (*entities.Recording, error) {
	// 1. Generate UUIDv7 for filename
	fileExt := filepath.Ext(fileHeader.Filename)
	uuidFilename, _ := uuid.NewV7()
	newFilename := uuidFilename.String() + fileExt

	// 2. Define paths
	uploadPath := filepath.Join("uploads", "audio", newFilename)

	// 3. Save physical file (using gin context indirectly or standard io)
	// Since we are in Usecase, we need to handle file saving responsibly.
	// In a clean architecture, file saving might be in infrastructure service,
	// but for simplicity here we assume the controller passes the file handler or we use a helper.
	// HOWEVER, the standard way in Gin is via Context.SaveUploadedFile.
	// To keep Usecase pure, we should receive the *multipart.FileHeader and save it manually.

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()

	if err := uc.fileService.SaveUploadedFile(fileHeader, uploadPath); err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// 4. Construct Public URL
	// Prepend baseURL if provided to ensure full domain is stored
	publicUrl := fmt.Sprintf("/uploads/audio/%s", newFilename)
	if baseURL != "" {
		publicUrl = fmt.Sprintf("%s%s", baseURL, publicUrl)
	}

	uuidEntity, _ := uuid.NewV7()
	// 5. Create Entity
	recording := &entities.Recording{
		ID:           uuidEntity.String(),
		Filename:     newFilename,
		OriginalName: fileHeader.Filename,
		AudioUrl:     publicUrl,
		MacAddress:   macAddress,
		CreatedAt:    time.Now(),
	}

	// 6. Save Metadata
	if err := uc.repo.Save(recording); err != nil {
		return nil, err
	}

	// 7. Trigger BIG Room Audio Update
	if macAddress != "" && shouldNotifyBIG(opts) {
		go func() {
			if err := uc.bigService.UpdateRoomOccupiedAudio(macAddress, publicUrl); err != nil {
				utils.LogError("SaveRecordingUseCase.SaveRecording: Failed to update room occupied audio: %v", err)
			}
		}()
	}

	return recording, nil
}

func (uc *saveRecordingUseCase) SaveRecordingFromBytes(data []byte, originalName, macAddress, baseURL string, opts ...SaveRecordingOption) (*entities.Recording, error) {
	// 1. Generate UUIDv4 for filename
	fileExt := filepath.Ext(originalName)
	if fileExt == "" {
		fileExt = ".wav" // default
	}
	uuidFilename, _ := uuid.NewV7()
	newFilename := uuidFilename.String() + fileExt

	// 2. Define paths
	uploadPath := filepath.Join("uploads", "audio", newFilename)

	// 3. Save physical file
	if err := uc.fileService.SaveFile(data, uploadPath); err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// 4. Construct Public URL
	publicUrl := fmt.Sprintf("/uploads/audio/%s", newFilename)
	if baseURL != "" {
		publicUrl = fmt.Sprintf("%s%s", baseURL, publicUrl)
	}

	uuidEntity, _ := uuid.NewV7()
	// 5. Create Entity
	recording := &entities.Recording{
		ID:           uuidEntity.String(),
		Filename:     newFilename,
		OriginalName: originalName,
		AudioUrl:     publicUrl,
		MacAddress:   macAddress,
		CreatedAt:    time.Now(),
	}

	// 6. Save Metadata
	if err := uc.repo.Save(recording); err != nil {
		return nil, err
	}

	// 7. Trigger BIG Room Audio Update
	if macAddress != "" && shouldNotifyBIG(opts) {
		go func() {
			if err := uc.bigService.UpdateRoomOccupiedAudio(macAddress, publicUrl); err != nil {
				utils.LogError("SaveRecordingUseCase.SaveRecordingFromBytes: Failed to update room occupied audio: %v", err)
			}
		}()
	}

	return recording, nil
}

func (uc *saveRecordingUseCase) SaveRecordingFromPath(srcPath, originalName, macAddress, baseURL string, opts ...SaveRecordingOption) (*entities.Recording, error) {
	// 1. Generate UUIDv7 for filename
	fileExt := filepath.Ext(originalName)
	if fileExt == "" {
		fileExt = ".wav" // default
	}
	uuidFilename, _ := uuid.NewV7()
	newFilename := uuidFilename.String() + fileExt

	// 2. Define paths
	uploadPath := filepath.Join("uploads", "audio", newFilename)

	// 3. Move physical file
	if err := uc.fileService.MoveFile(srcPath, uploadPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %v", err)
	}

	// 4. Construct Public URL
	publicUrl := fmt.Sprintf("/uploads/audio/%s", newFilename)
	if baseURL != "" {
		publicUrl = fmt.Sprintf("%s%s", baseURL, publicUrl)
	}

	uuidEntity, _ := uuid.NewV7()
	// 5. Create Entity
	recording := &entities.Recording{
		ID:           uuidEntity.String(),
		Filename:     newFilename,
		OriginalName: originalName,
		AudioUrl:     publicUrl,
		MacAddress:   macAddress,
		CreatedAt:    time.Now(),
	}

	// 6. Save Metadata
	if err := uc.repo.Save(recording); err != nil {
		return nil, err
	}

	// 7. Trigger BIG Room Audio Update
	if macAddress != "" && shouldNotifyBIG(opts) {
		go func() {
			if err := uc.bigService.UpdateRoomOccupiedAudio(macAddress, publicUrl); err != nil {
				utils.LogError("SaveRecordingUseCase.SaveRecordingFromPath: Failed to update room occupied audio: %v", err)
			}
		}()
	}

	return recording, nil
}
