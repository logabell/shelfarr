package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
	"gorm.io/gorm"
)

// NotificationRequest represents a notification configuration request
type NotificationRequest struct {
	Name            string `json:"name"`
	Type            string `json:"type"` // webhook, discord, telegram, email
	Enabled         bool   `json:"enabled"`
	WebhookURL      string `json:"webhookUrl,omitempty"`
	DiscordWebhook  string `json:"discordWebhook,omitempty"`
	TelegramBotToken string `json:"telegramBotToken,omitempty"`
	TelegramChatID  string `json:"telegramChatId,omitempty"`
	EmailTo         string `json:"emailTo,omitempty"`
	OnGrab          bool   `json:"onGrab"`
	OnDownload      bool   `json:"onDownload"`
	OnUpgrade       bool   `json:"onUpgrade"`
	OnImport        bool   `json:"onImport"`
	OnDelete        bool   `json:"onDelete"`
	OnHealthIssue   bool   `json:"onHealthIssue"`
}

// getNotifications returns all notification configurations
func (s *Server) getNotifications(c echo.Context) error {
	var notifications []db.Notification
	s.db.Find(&notifications)
	return c.JSON(http.StatusOK, notifications)
}

// addNotification creates a new notification configuration
func (s *Server) addNotification(c echo.Context) error {
	var req NotificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	notification := db.Notification{
		Name:             req.Name,
		Type:             req.Type,
		Enabled:          req.Enabled,
		WebhookURL:       req.WebhookURL,
		DiscordWebhook:   req.DiscordWebhook,
		TelegramBotToken: req.TelegramBotToken,
		TelegramChatID:   req.TelegramChatID,
		EmailTo:          req.EmailTo,
		OnGrab:           req.OnGrab,
		OnDownload:       req.OnDownload,
		OnUpgrade:        req.OnUpgrade,
		OnImport:         req.OnImport,
		OnDelete:         req.OnDelete,
		OnHealthIssue:    req.OnHealthIssue,
	}

	if err := s.db.Create(&notification).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create notification"})
	}

	return c.JSON(http.StatusCreated, notification)
}

// updateNotification updates an existing notification configuration
func (s *Server) updateNotification(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var notification db.Notification
	if err := s.db.First(&notification, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Notification not found"})
	}

	var req NotificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	notification.Name = req.Name
	notification.Type = req.Type
	notification.Enabled = req.Enabled
	notification.WebhookURL = req.WebhookURL
	notification.DiscordWebhook = req.DiscordWebhook
	notification.TelegramBotToken = req.TelegramBotToken
	notification.TelegramChatID = req.TelegramChatID
	notification.EmailTo = req.EmailTo
	notification.OnGrab = req.OnGrab
	notification.OnDownload = req.OnDownload
	notification.OnUpgrade = req.OnUpgrade
	notification.OnImport = req.OnImport
	notification.OnDelete = req.OnDelete
	notification.OnHealthIssue = req.OnHealthIssue

	if err := s.db.Save(&notification).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update notification"})
	}

	return c.JSON(http.StatusOK, notification)
}

// deleteNotification removes a notification configuration
func (s *Server) deleteNotification(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := s.db.Delete(&db.Notification{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete notification"})
	}

	return c.NoContent(http.StatusNoContent)
}

// testNotification sends a test notification
func (s *Server) testNotification(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var notification db.Notification
	if err := s.db.First(&notification, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Notification not found"})
	}

	// Send test notification based on type
	var testErr error
	switch notification.Type {
	case "webhook":
		testErr = sendWebhookTest(notification.WebhookURL)
	case "discord":
		testErr = sendDiscordTest(notification.DiscordWebhook)
	case "telegram":
		testErr = sendTelegramTest(notification.TelegramBotToken, notification.TelegramChatID)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unknown notification type"})
	}

	if testErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": testErr.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Test notification sent successfully"})
}

// sendWebhookTest sends a test webhook
func sendWebhookTest(url string) error {
	payload := map[string]interface{}{
		"event": "test",
		"message": "Test notification from Shelfarr",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// sendDiscordTest sends a test Discord notification
func sendDiscordTest(webhookURL string) error {
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       "Shelfarr Test",
				"description": "This is a test notification from Shelfarr",
				"color":       3447003, // Blue
				"footer": map[string]string{
					"text": "Shelfarr",
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Discord returned status %d", resp.StatusCode)
	}

	return nil
}

// sendTelegramTest sends a test Telegram notification
func sendTelegramTest(botToken, chatID string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    "🧪 *Test notification from Shelfarr*\n\nIf you see this message, notifications are working!",
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send Telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Telegram returned status %d", resp.StatusCode)
	}

	return nil
}

// NotificationService provides methods for sending notifications
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService creates a new notification service
func NewNotificationService(database *gorm.DB) *NotificationService {
	return &NotificationService{db: database}
}

// SendNotification sends a notification to all enabled notification configs
func (ns *NotificationService) SendNotification(eventType string, data map[string]interface{}) {
	var notifications []db.Notification
	ns.db.Where("enabled = ?", true).Find(&notifications)

	for _, n := range notifications {
		// Check if this notification should receive this event type
		shouldSend := false
		switch eventType {
		case "grab":
			shouldSend = n.OnGrab
		case "download":
			shouldSend = n.OnDownload
		case "upgrade":
			shouldSend = n.OnUpgrade
		case "import":
			shouldSend = n.OnImport
		case "delete":
			shouldSend = n.OnDelete
		case "health":
			shouldSend = n.OnHealthIssue
		}

		if !shouldSend {
			continue
		}

		// Send notification based on type
		switch n.Type {
		case "webhook":
			go sendWebhook(n.WebhookURL, eventType, data)
		case "discord":
			go sendDiscord(n.DiscordWebhook, eventType, data)
		case "telegram":
			go sendTelegram(n.TelegramBotToken, n.TelegramChatID, eventType, data)
		}
	}
}

func sendWebhook(url, eventType string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"event": eventType,
		"data":  data,
	}
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}

func sendDiscord(webhookURL, eventType string, data map[string]interface{}) {
	title := data["title"].(string)
	message := fmt.Sprintf("Event: %s", eventType)
	if msg, ok := data["message"].(string); ok {
		message = msg
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": message,
				"color":       3447003,
			},
		},
	}
	body, _ := json.Marshal(payload)
	http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
}

func sendTelegram(botToken, chatID, eventType string, data map[string]interface{}) {
	title := data["title"].(string)
	message := fmt.Sprintf("📚 *%s*\n\nEvent: %s", title, eventType)
	if msg, ok := data["message"].(string); ok {
		message = fmt.Sprintf("📚 *%s*\n\n%s", title, msg)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}

