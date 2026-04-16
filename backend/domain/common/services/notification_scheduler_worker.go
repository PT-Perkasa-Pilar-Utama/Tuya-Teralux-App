package services

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/entities"
	"sensio/domain/common/repositories"
	"sensio/domain/common/utils"
	"strings"
	"sync"
	"time"
)

type NotificationSchedulerWorker struct {
	scheduledRepo   *repositories.ScheduledNotificationRepository
	waSvc           *WANotificationService
	templateService *WATemplateService
	stopCh          chan struct{}
	wg              sync.WaitGroup
	running         bool
	mu              sync.Mutex
}

func NewNotificationSchedulerWorker(waBaseURL string) *NotificationSchedulerWorker {
	return &NotificationSchedulerWorker{
		scheduledRepo:   repositories.NewScheduledNotificationRepository(),
		waSvc:           NewWANotificationService(waBaseURL),
		templateService: NewWATemplateService(),
		stopCh:          make(chan struct{}),
	}
}

func (w *NotificationSchedulerWorker) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.stopCh = make(chan struct{})
	w.mu.Unlock()

	w.wg.Add(1)
	go w.run()
	utils.LogInfo("NotificationSchedulerWorker: Started")
}

func (w *NotificationSchedulerWorker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	close(w.stopCh)
	w.mu.Unlock()

	w.wg.Wait()
	utils.LogInfo("NotificationSchedulerWorker: Stopped")
}

func (w *NotificationSchedulerWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processPendingNotifications()
		}
	}
}

func (w *NotificationSchedulerWorker) processPendingNotifications() {
	now := time.Now()
	notifications, err := w.scheduledRepo.GetPendingNotifications(100)
	if err != nil {
		utils.LogError("NotificationSchedulerWorker: Failed to get pending notifications: %v", err)
		return
	}

	for _, notification := range notifications {
		if notification.ScheduledAt.After(now) {
			continue
		}

		w.processNotification(&notification)
	}
}

func (w *NotificationSchedulerWorker) processNotification(notification *entities.ScheduledNotification) {
	var phoneNumbers []string
	if err := json.Unmarshal([]byte(notification.PhoneNumbers), &phoneNumbers); err != nil {
		utils.LogError("NotificationSchedulerWorker: Failed to parse phone numbers for %s: %v", notification.ID, err)
		w.scheduledRepo.UpdateStatus(notification.ID, entities.NotificationStatusFailed)
		return
	}

	var bookingInfo map[string]interface{}
	if err := json.Unmarshal([]byte(notification.BookingInfo), &bookingInfo); err != nil {
		utils.LogError("NotificationSchedulerWorker: Failed to parse booking info for %s: %v", notification.ID, err)
		bookingInfo = make(map[string]interface{})
	} else {
		utils.LogDebug("processNotification: ID=%s, BookingInfo=%+v, BookingTimeEnd=%s", notification.ID, bookingInfo, notification.BookingTimeEnd)
	}

	content := w.buildWAMessage(bookingInfo, notification)

	for _, phone := range phoneNumbers {
		if err := w.waSvc.SendMessage(phone, content); err != nil {
			utils.LogError("NotificationSchedulerWorker: Failed to send WA to %s: %v", phone, err)
			w.scheduledRepo.UpdateStatus(notification.ID, entities.NotificationStatusFailed)
			return
		}
		utils.LogInfo("NotificationSchedulerWorker: WA sent to %s for notification %s", phone, notification.ID)
	}

	w.scheduledRepo.UpdateStatus(notification.ID, entities.NotificationStatusSent)
	utils.LogInfo("NotificationSchedulerWorker: Notification %s processed successfully", notification.ID)
}

func (w *NotificationSchedulerWorker) buildWAMessage(bookingInfo map[string]interface{}, notification *entities.ScheduledNotification) string {
	customerName := w.getStringValue(bookingInfo, "SDTGetRoomTeraluxCustomerName")
	buildingName := w.getStringValue(bookingInfo, "SDTGetRoomTeraluxBuildingsName")
	roomName := w.getStringValue(bookingInfo, "SDTGetRoomTeraluxRoomName")
	password := w.getStringValue(bookingInfo, "SDTGetRoomTeraluxItemRoomPassword")
	bookingTime := w.getStringValue(bookingInfo, "SDTGetRoomTeraluxBookingtimeChar")

	remainingMinutes := w.calculateRemainingMinutes(notification)

	var dateStr string
	if notification.BookingTimeEnd != "" {
		if t, err := time.Parse(time.RFC3339, notification.BookingTimeEnd); err == nil {
			dateStr = t.Format("02 January 2006")
		} else {
			dateStr = notification.ScheduledAt.Format("02 January 2006")
		}
	} else {
		dateStr = notification.ScheduledAt.Format("02 January 2006")
	}

	templateData := &WATemplateData{
		CustomerName:     customerName,
		BuildingName:     buildingName,
		RoomName:         roomName,
		DateStr:          dateStr,
		BookingTime:      bookingTime,
		Password:         password,
		RemainingMinutes: remainingMinutes,
		CompanyName:      "Teralux Team",
	}

	templateName := notification.Template
	if templateName == "" {
		templateName = "start_meeting"
	}

	message, err := w.templateService.RenderTemplate(templateName, templateData)
	if err != nil {
		utils.LogWarn("NotificationSchedulerWorker: Failed to render template %s, using fallback: %v", templateName, err)
		return w.buildFallbackMessage(templateData)
	}

	return message
}

func (w *NotificationSchedulerWorker) buildFallbackMessage(data *WATemplateData) string {
	return fmt.Sprintf(`[PENGINGAT JADWAL PERTEMUAN]

Yth. %s,

Melalui pesan ini, kami ingin mengingatkan bahwa jadwal pertemuan Anda akan dimulai dalam %s menit. Berikut adalah rincian pertemuan tersebut:

🏢 Lokasi: %s
🚪 Ruang: %s
📅 Tanggal: %s
⏰ Waktu: %s
🔐 Password: %s

Kami mohon kesediaan %s untuk bersiap sebelum waktu pertemuan dimulai. Terima kasih atas perhatian dan kerja samanya.

Salam hangat,
Teralux Team`,
		data.CustomerName,
		data.RemainingMinutes,
		data.BuildingName,
		data.RoomName,
		data.DateStr,
		data.BookingTime,
		data.Password,
		data.CustomerName,
	)
}

func (w *NotificationSchedulerWorker) getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}

func (w *NotificationSchedulerWorker) calculateRemainingMinutes(notification *entities.ScheduledNotification) string {
	if notification.BookingTimeEnd == "" {
		return "15"
	}

	bookingTimeEnd, err := time.Parse(time.RFC3339, notification.BookingTimeEnd)
	if err != nil {
		return "15"
	}

	scheduledAt := notification.ScheduledAt
	diff := scheduledAt.Sub(bookingTimeEnd.Add(-time.Duration(15) * time.Minute))

	minutes := int(diff.Minutes())
	if minutes <= 0 {
		minutes = int(time.Until(bookingTimeEnd).Minutes())
	}

	if minutes < 1 {
		minutes = 1
	}

	return fmt.Sprintf("%d", minutes)
}
