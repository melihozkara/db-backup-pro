package notification

import (
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dbbackup/internal/database"
	"dbbackup/internal/i18n"
)

// TelegramNotifier handles Telegram notifications
type TelegramNotifier struct {
	botToken string
	chatID   int64
	enabled  bool
}

// NewTelegramNotifier creates a new Telegram notifier
func NewTelegramNotifier(botToken, chatID string, enabled bool) (*TelegramNotifier, error) {
	var chatIDInt int64
	if chatID != "" {
		var err error
		chatIDInt, err = strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat ID: %w", err)
		}
	}

	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatIDInt,
		enabled:  enabled,
	}, nil
}

// TestConnection tests the bot token
func (t *TelegramNotifier) TestConnection() error {
	if t.botToken == "" {
		return fmt.Errorf("bot token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(t.botToken)
	if err != nil {
		return fmt.Errorf("invalid bot token: %w", err)
	}

	// Try to get bot info
	_, err = bot.GetMe()
	if err != nil {
		return fmt.Errorf("failed to verify bot: %w", err)
	}

	return nil
}

// SendTestMessage sends a test message
func (t *TelegramNotifier) SendTestMessage() error {
	if t.botToken == "" || t.chatID == 0 {
		return fmt.Errorf("bot token or chat ID is empty")
	}

	bot, err := tgbotapi.NewBotAPI(t.botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	message := i18n.T("telegram.testMessage")
	msg := tgbotapi.NewMessage(t.chatID, message)
	msg.ParseMode = "Markdown"

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendBackupSuccess sends a success notification
func (t *TelegramNotifier) SendBackupSuccess(job database.BackupJob, history database.BackupHistory) error {
	if !t.enabled || t.botToken == "" || t.chatID == 0 {
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(t.botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	duration := "-"
	if history.CompletedAt != nil {
		d := history.CompletedAt.Sub(history.StartedAt)
		duration = formatDuration(d)
	}

	message := i18n.T("telegram.backupSuccess",
		job.Name,
		history.FileName,
		formatSize(history.FileSize),
		duration,
		history.StoragePath,
		history.StartedAt.Format(i18n.T("telegram.dateFormat")),
	)

	msg := tgbotapi.NewMessage(t.chatID, message)
	msg.ParseMode = "Markdown"

	_, err = bot.Send(msg)
	return err
}

// SendBackupFailure sends a failure notification
func (t *TelegramNotifier) SendBackupFailure(job database.BackupJob, errorMsg string) error {
	if !t.enabled || t.botToken == "" || t.chatID == 0 {
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(t.botToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	message := i18n.T("telegram.backupFailed",
		job.Name,
		errorMsg,
		time.Now().Format(i18n.T("telegram.dateFormat")),
	)

	msg := tgbotapi.NewMessage(t.chatID, message)
	msg.ParseMode = "Markdown"

	_, err = bot.Send(msg)
	return err
}

// formatSize formats bytes to human readable format
func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const k = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	size := float64(bytes)

	for size >= k && i < len(sizes)-1 {
		size /= k
		i++
	}

	return fmt.Sprintf("%.2f %s", size, sizes[i])
}

// formatDuration formats duration to human readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return i18n.T("telegram.durationSeconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return i18n.T("telegram.durationMinutes", int(d.Minutes()))
	}
	return i18n.T("telegram.durationHoursMinutes", int(d.Hours()), int(d.Minutes())%60)
}

// NotifyFromSettings creates a notifier from app settings and sends notification
func NotifyBackupResult(job database.BackupJob, history database.BackupHistory) {
	// Get telegram settings
	botToken, _ := database.GetSetting("telegram_bot_token")
	chatID, _ := database.GetSetting("telegram_chat_id")
	enabledStr, _ := database.GetSetting("telegram_enabled")
	enabled := enabledStr == "true"

	notifier, err := NewTelegramNotifier(botToken, chatID, enabled)
	if err != nil {
		return
	}

	if history.Status == "success" {
		notifier.SendBackupSuccess(job, history)
	} else if history.Status == "failed" {
		notifier.SendBackupFailure(job, history.ErrorMessage)
	}
}
