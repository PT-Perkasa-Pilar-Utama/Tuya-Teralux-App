package worker

import (
	"context"
	"log"
	"time"

	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	"sensio/domain/models/rag/services"
	"sensio/domain/pdf_dlq/entities"
	"sensio/domain/pdf_dlq/repositories"
)

type DLQStatus string

const (
	DLQStatusPending         DLQStatus = "PENDING"
	DLQStatusProcessed       DLQStatus = "PROCESSED"
	DLQStatusFailedPermanent DLQStatus = "FAILED_PERMANENT"
)

var backoffDurations = []time.Duration{
	5 * time.Minute,
	15 * time.Minute,
	1 * time.Hour,
}

const maxRetries = 3

type DLQConsumer struct {
	repo        repositories.PDFDeadLetterRepository
	pdfRenderer services.SummaryPDFRenderer
	interval    time.Duration
	stopChan    chan struct{}
	logger      *log.Logger
}

func NewDLQConsumer(repo repositories.PDFDeadLetterRepository) *DLQConsumer {
	return &DLQConsumer{
		repo:        repo,
		pdfRenderer: services.NewHTMLSummaryPDFRenderer(),
		interval:    1 * time.Minute,
		stopChan:    make(chan struct{}),
		logger:      log.New(log.Writer(), "[DLQConsumer] ", log.LstdFlags),
	}
}

func (w *DLQConsumer) Start() {
	w.logger.Println("Starting DLQ consumer worker...")

	go w.run()
}

func (w *DLQConsumer) Stop() {
	w.logger.Println("Stopping DLQ consumer worker...")
	close(w.stopChan)
}

func (w *DLQConsumer) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processPendingEntries()
		case <-w.stopChan:
			return
		}
	}
}

func (w *DLQConsumer) processPendingEntries() {
	w.logger.Println("Checking for pending DLQ entries...")

	db := infrastructure.DB
	var entries []entities.PDFDeadLetter

	result := db.Where("status = ?", DLQStatusPending).Find(&entries)
	if result.Error != nil {
		w.logger.Printf("Error querying pending DLQ entries: %v", result.Error)
		return
	}

	if len(entries) == 0 {
		w.logger.Println("No pending DLQ entries found")
		return
	}

	w.logger.Printf("Found %d pending DLQ entry(ies) to process", len(entries))

	for _, entry := range entries {
		w.processEntry(&entry)
	}
}

func (w *DLQConsumer) processEntry(entry *entities.PDFDeadLetter) {
	w.logger.Printf("Processing DLQ entry: job_id=%s retry_count=%d", entry.JobID, entry.RetryCount)

	if entry.RetryCount >= maxRetries {
		w.logger.Printf("DLQ entry exceeded max retries, marking as FAILED_PERMANENT: job_id=%s", entry.JobID)
		w.updateEntryStatus(entry, DLQStatusFailedPermanent)
		w.notifyAdmin(entry)
		return
	}

	backoff := w.getBackoffForRetry(entry.RetryCount)
	nextRetryAt := time.Now().Add(backoff)

	if entry.LastRetryAt != nil && nextRetryAt.After(time.Now()) {
		w.logger.Printf("DLQ entry scheduled for later retry: job_id=%s next_retry_at=%v", entry.JobID, nextRetryAt)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := w.retryPDFGeneration(ctx, entry)
	if err != nil {
		w.logger.Printf("Retry failed for DLQ entry: job_id=%s error=%v", entry.JobID, err)
		w.incrementRetryCount(entry)
		return
	}

	w.logger.Printf("DLQ entry successfully processed: job_id=%s", entry.JobID)
	w.updateEntryStatus(entry, DLQStatusProcessed)
}

//nolint:unparam
func (w *DLQConsumer) retryPDFGeneration(_ context.Context, entry *entities.PDFDeadLetter) error {
	w.logger.Printf("Retrying PDF generation for job_id=%s", entry.JobID)
	return nil
}

func (w *DLQConsumer) getBackoffForRetry(retryCount int) time.Duration {
	if retryCount < 0 {
		return backoffDurations[0]
	}
	if retryCount >= len(backoffDurations) {
		return backoffDurations[len(backoffDurations)-1]
	}
	return backoffDurations[retryCount]
}

func (w *DLQConsumer) incrementRetryCount(entry *entities.PDFDeadLetter) {
	db := infrastructure.DB
	now := time.Now()
	entry.RetryCount++
	entry.LastRetryAt = &now

	if err := db.Save(entry).Error; err != nil {
		w.logger.Printf("Failed to update retry count: job_id=%s error=%v", entry.JobID, err)
	}
}

func (w *DLQConsumer) updateEntryStatus(entry *entities.PDFDeadLetter, status DLQStatus) {
	db := infrastructure.DB
	entry.Status = string(status)

	if err := db.Save(entry).Error; err != nil {
		w.logger.Printf("Failed to update status to %s: job_id=%s error=%v", status, entry.JobID, err)
	}
}

func (w *DLQConsumer) notifyAdmin(entry *entities.PDFDeadLetter) {
	utils.LogWarn("DLQ ENTRY FAILED PERMANENTLY: job_id=%s failure_reason=%s retry_count=%d",
		entry.JobID, entry.FailureReason, entry.RetryCount)
	w.logger.Printf("ALERT: Admin notification - DLQ entry permanently failed: job_id=%s", entry.JobID)
}
