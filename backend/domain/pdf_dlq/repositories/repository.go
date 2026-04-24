package repositories

import (
	"log"
	"time"

	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	"sensio/domain/pdf_dlq/entities"

	"gorm.io/gorm"
)

type PDFDeadLetterRepository interface {
	Create(entry *entities.PDFDeadLetter) error
	GetByJobID(jobID string) (*entities.PDFDeadLetter, error)
}

type pdfDeadLetterRepository struct {
	db *gorm.DB
}

func NewPDFDeadLetterRepository() PDFDeadLetterRepository {
	return &pdfDeadLetterRepository{
		db: infrastructure.DB,
	}
}

func (r *pdfDeadLetterRepository) Create(entry *entities.PDFDeadLetter) error {
	start := time.Now()
	result := r.db.Create(entry)
	duration := time.Since(start)

	if result.Error != nil {
		log.Printf("PDFDeadLetterRepository: Create failed | duration_ms=%d | error=%v", duration.Milliseconds(), result.Error)
		return result.Error
	}

	log.Printf("PDFDeadLetterRepository: Create completed | job_id=%s | duration_ms=%d | rows=%d", entry.JobID, duration.Milliseconds(), result.RowsAffected)
	return nil
}

func (r *pdfDeadLetterRepository) GetByJobID(jobID string) (*entities.PDFDeadLetter, error) {
	start := time.Now()

	var entry entities.PDFDeadLetter
	if err := r.db.Where("job_id = ?", jobID).First(&entry).Error; err != nil {
		utils.LogDebug("PDFDeadLetterRepository: GetByJobID failed | job_id=%s | db_duration_ms=%d | error=%v", jobID, time.Since(start).Milliseconds(), err)
		return nil, err
	}

	utils.LogDebug("PDFDeadLetterRepository: GetByJobID completed | job_id=%s | db_duration_ms=%d", jobID, time.Since(start).Milliseconds())
	return &entry, nil
}
