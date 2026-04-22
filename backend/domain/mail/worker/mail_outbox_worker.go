package worker

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"sensio/domain/common/utils"
	"sensio/domain/mail/entities"
	"sensio/domain/mail/repositories"
	"sensio/domain/mail/services"
	"sensio/domain/download_token"
)

type MailOutboxWorker struct {
	repo          repositories.MailOutboxRepository
	mailService   *services.MailService
	tokenService  *download_token.DownloadTokenService
	interval      time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	running       uint32
	logger        *log.Logger
}

func NewMailOutboxWorker(
	mailService *services.MailService,
	tokenService *download_token.DownloadTokenService,
) *MailOutboxWorker {
	return &MailOutboxWorker{
		repo:          repositories.NewMailOutboxRepository(),
		mailService:   mailService,
		tokenService:  tokenService,
		interval:      30 * time.Second,
		stopChan:      make(chan struct{}),
		logger:        log.New(log.Writer(), "[MailOutboxWorker] ", log.LstdFlags),
	}
}

func (w *MailOutboxWorker) Start() {
	if !atomic.CompareAndSwapUint32(&w.running, 0, 1) {
		return
	}
	w.stopChan = make(chan struct{})

	w.wg.Add(1)
	go w.run()
	utils.LogInfo("MailOutboxWorker: Started")
}

func (w *MailOutboxWorker) Stop() {
	if !atomic.CompareAndSwapUint32(&w.running, 1, 0) {
		return
	}
	close(w.stopChan)

	w.wg.Wait()
	utils.LogInfo("MailOutboxWorker: Stopped")
}

func (w *MailOutboxWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.processPendingEntries()
		}
	}
}

func (w *MailOutboxWorker) processPendingEntries() {
	entries, err := w.repo.GetPendingEntries(50)
	if err != nil {
		w.logger.Printf("Error querying pending entries: %v", err)
		return
	}

	if len(entries) == 0 {
		return
	}

	w.logger.Printf("Found %d pending mail outbox entry(ies) to process", len(entries))

	for _, entry := range entries {
		w.processEntry(&entry)
	}
}

func (w *MailOutboxWorker) processEntry(entry *entities.MailOutbox) {
	w.logger.Printf("Processing mail outbox entry: task_id=%s retry_count=%d", entry.TaskID, entry.RetryCount)

	if entry.RetryCount >= 5 {
		w.logger.Printf("Mail outbox entry exceeded max retries, marking as failed: task_id=%s", entry.TaskID)
		w.repo.UpdateStatus(entry.TaskID, entities.MailOutboxStatusFailed, "max retries exceeded")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := w.processMailJob(ctx, entry)
	if err != nil {
		w.logger.Printf("Failed to process mail outbox entry: task_id=%s error=%v", entry.TaskID, err)
		w.repo.IncrementRetry(entry.TaskID)
		w.repo.UpdateStatus(entry.TaskID, entities.MailOutboxStatusFailed, err.Error())
		return
	}

	w.logger.Printf("Mail outbox entry processed successfully: task_id=%s", entry.TaskID)
	w.repo.UpdateStatus(entry.TaskID, entities.MailOutboxStatusSent, "")
}

func (w *MailOutboxWorker) processMailJob(ctx context.Context, entry *entities.MailOutbox) error {
	token, err := w.tokenService.CreateToken(entry.Recipient, entry.ObjectKey, entry.Purpose)
	if err != nil {
		return err
	}

	linkData := map[string]interface{}{
		"download_link": "/api/download/resolve?state=" + token + "&purpose=" + entry.Purpose,
	}

	err = w.mailService.SendEmailWithTemplate(
		[]string{entry.Recipient},
		entry.Subject+" - Download Link",
		"secure_link",
		linkData,
		nil,
	)

	return err
}