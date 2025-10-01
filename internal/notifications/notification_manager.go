/**
 * Notification Manager Module
 *
 * Provides cross-platform desktop notifications for completed tasks and system events.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: notification_manager.go
 * Description: Desktop notification system with cross-platform support and customization
 */

package notifications

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationSuccess NotificationType = "success"
	NotificationError   NotificationType = "error"
	NotificationWarning NotificationType = "warning"
	NotificationInfo    NotificationType = "info"
	NotificationTask    NotificationType = "task"
)

// NotificationPriority represents notification priority levels
type NotificationPriority int

const (
	PriorityLow      NotificationPriority = 0
	PriorityNormal   NotificationPriority = 1
	PriorityHigh     NotificationPriority = 2
	PriorityCritical NotificationPriority = 3
)

// Notification represents a desktop notification
type Notification struct {
	ID        string
	Title     string
	Message   string
	Type      NotificationType
	Priority  NotificationPriority
	Duration  time.Duration
	Icon      string
	Sound     bool
	Actions   []NotificationAction
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NotificationAction represents a clickable action in the notification
type NotificationAction struct {
	ID    string
	Label string
	URL   string
}

// NotificationManager manages desktop notifications
type NotificationManager struct {
	enabled       bool
	platform      string
	notifications map[string]*Notification
	mutex         sync.RWMutex
	config        *NotificationConfig
	history       []*Notification
	historyMutex  sync.RWMutex
	maxHistory    int
	historyFile   string
}

// NotificationConfig holds notification configuration
type NotificationConfig struct {
	Enabled         bool
	DefaultDuration time.Duration
	MaxHistory      int
	SoundEnabled    bool
	IconPath        string
	Timeout         time.Duration
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	nm := &NotificationManager{
		enabled:       true,
		platform:      runtime.GOOS,
		notifications: make(map[string]*Notification),
		config: &NotificationConfig{
			Enabled:         true,
			DefaultDuration: 5 * time.Second,
			MaxHistory:      100,
			SoundEnabled:    true,
			IconPath:        "",
			Timeout:         10 * time.Second,
		},
		history:     make([]*Notification, 0),
		maxHistory:  100,
		historyFile: "notifications_history.json",
	}

	// Detect platform capabilities
	nm.detectPlatformCapabilities()

	// Load existing notification history
	nm.loadNotificationHistory()

	return nm
}

// detectPlatformCapabilities detects what notification systems are available
func (nm *NotificationManager) detectPlatformCapabilities() {
	switch nm.platform {
	case "darwin":
		// macOS - check for osascript
		if _, err := exec.LookPath("osascript"); err == nil {
			nm.config.IconPath = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericApplicationIcon.icns"
		}
	case "linux":
		// Linux - check for notify-send and find a suitable icon
		if _, err := exec.LookPath("notify-send"); err == nil {
			// Try to find a suitable icon
			possibleIcons := []string{
				"/usr/share/icons/gnome/32x32/apps/terminal.png",
				"/usr/share/icons/gnome/32x32/apps/utilities-terminal.png",
				"/usr/share/pixmaps/terminal.png",
				"/usr/share/icons/hicolor/32x32/apps/terminal.png",
				"/usr/share/icons/Adwaita/32x32/apps/utilities-terminal.png",
			}

			for _, iconPath := range possibleIcons {
				if _, err := os.Stat(iconPath); err == nil {
					nm.config.IconPath = iconPath
					break
				}
			}

			// If no icon found, leave empty (notify-send will work without icon)
			if nm.config.IconPath == "" {
				nm.config.IconPath = ""
			}
		}
	case "windows":
		// Windows - PowerShell notifications
		nm.config.IconPath = "C:\\Windows\\System32\\shell32.dll,1"
	}
}

// SetEnabled enables or disables notifications
func (nm *NotificationManager) SetEnabled(enabled bool) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.enabled = enabled
}

// IsEnabled returns whether notifications are enabled
func (nm *NotificationManager) IsEnabled() bool {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.enabled
}

// SetConfig updates the notification configuration
func (nm *NotificationManager) SetConfig(config *NotificationConfig) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.config = config
	nm.maxHistory = config.MaxHistory
}

// GetConfig returns the current notification configuration
func (nm *NotificationManager) GetConfig() *NotificationConfig {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.config
}

// SendNotification sends a desktop notification
func (nm *NotificationManager) SendNotification(notification *Notification) error {
	if !nm.enabled || !nm.config.Enabled {
		return fmt.Errorf("notifications are disabled")
	}

	// Set defaults
	if notification.Duration == 0 {
		notification.Duration = nm.config.DefaultDuration
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}
	if notification.ExpiresAt.IsZero() {
		notification.ExpiresAt = notification.CreatedAt.Add(notification.Duration)
	}

	// Generate ID if not provided
	if notification.ID == "" {
		notification.ID = fmt.Sprintf("ena_%d", time.Now().UnixNano())
	}

	// Store notification
	nm.mutex.Lock()
	nm.notifications[notification.ID] = notification
	nm.mutex.Unlock()

	// Add to history
	nm.addToHistory(notification)

	// Send platform-specific notification
	return nm.sendPlatformNotification(notification)
}

// sendPlatformNotification sends notification using platform-specific method
func (nm *NotificationManager) sendPlatformNotification(notification *Notification) error {
	switch nm.platform {
	case "darwin":
		return nm.sendMacOSNotification(notification)
	case "linux":
		return nm.sendLinuxNotification(notification)
	case "windows":
		return nm.sendWindowsNotification(notification)
	default:
		return fmt.Errorf("unsupported platform: %s", nm.platform)
	}
}

// sendMacOSNotification sends notification on macOS using osascript
func (nm *NotificationManager) sendMacOSNotification(notification *Notification) error {
	script := fmt.Sprintf(`
		display notification "%s" with title "%s" subtitle "Ena Assistant"
	`,
		strings.ReplaceAll(notification.Message, `"`, `\"`),
		strings.ReplaceAll(notification.Title, `"`, `\"`))

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// sendLinuxNotification sends notification on Linux using notify-send
func (nm *NotificationManager) sendLinuxNotification(notification *Notification) error {
	args := []string{
		"-t", fmt.Sprintf("%d", int(notification.Duration.Seconds()*1000)),
		notification.Title,
		notification.Message,
	}

	// Add icon only if it exists
	if nm.config.IconPath != "" {
		if _, err := os.Stat(nm.config.IconPath); err == nil {
			args = append([]string{"-i", nm.config.IconPath}, args...)
		}
	}

	// Add urgency based on priority (notify-send only supports low, normal, critical)
	switch notification.Priority {
	case PriorityCritical:
		args = append([]string{"-u", "critical"}, args...)
	case PriorityHigh:
		args = append([]string{"-u", "critical"}, args...) // Map high to critical
	case PriorityLow:
		args = append([]string{"-u", "low"}, args...)
	default:
		args = append([]string{"-u", "normal"}, args...)
	}

	cmd := exec.Command("notify-send", args...)
	return cmd.Run()
}

// sendWindowsNotification sends notification on Windows using PowerShell
func (nm *NotificationManager) sendWindowsNotification(notification *Notification) error {
	script := fmt.Sprintf(`
		[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

		$template = @"
		<toast>
			<visual>
				<binding template="ToastGeneric">
					<text>%s</text>
					<text>%s</text>
				</binding>
			</visual>
		</toast>
"@

		$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
		$xml.LoadXml($template)
		$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
		$toast.ExpirationTime = [DateTimeOffset]::Now.AddSeconds(%d)
		$toast.Tag = "%s"
		$toast.Group = "EnaAssistant"

		$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Ena Assistant")
		$notifier.Show($toast)
	`,
		notification.Title,
		notification.Message,
		int(notification.Duration.Seconds()),
		notification.ID)

	cmd := exec.Command("powershell", "-Command", script)
	return cmd.Run()
}

// addToHistory adds notification to history
func (nm *NotificationManager) addToHistory(notification *Notification) {
	nm.historyMutex.Lock()
	defer nm.historyMutex.Unlock()

	nm.history = append(nm.history, notification)

	// Trim history if it exceeds max size
	if len(nm.history) > nm.maxHistory {
		nm.history = nm.history[len(nm.history)-nm.maxHistory:]
	}

	// Save history to disk
	if err := nm.saveNotificationHistory(); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to save notification history: %v\n", err)
	}
}

// GetHistory returns notification history
func (nm *NotificationManager) GetHistory() []*Notification {
	nm.historyMutex.RLock()
	defer nm.historyMutex.RUnlock()

	// Return a copy to prevent external modification
	history := make([]*Notification, len(nm.history))
	copy(history, nm.history)
	return history
}

// ClearHistory clears notification history
func (nm *NotificationManager) ClearHistory() {
	nm.historyMutex.Lock()
	defer nm.historyMutex.Unlock()
	nm.history = make([]*Notification, 0)

	// Also clear the history file
	if err := nm.saveNotificationHistory(); err != nil {
		fmt.Printf("Warning: Failed to clear notification history file: %v\n", err)
	}
}

// GetActiveNotifications returns currently active notifications
func (nm *NotificationManager) GetActiveNotifications() []*Notification {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	now := time.Now()
	var active []*Notification

	for _, notification := range nm.notifications {
		if notification.ExpiresAt.After(now) {
			active = append(active, notification)
		}
	}

	return active
}

// DismissNotification dismisses a specific notification
func (nm *NotificationManager) DismissNotification(id string) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if _, exists := nm.notifications[id]; !exists {
		return fmt.Errorf("notification '%s' not found", id)
	}

	delete(nm.notifications, id)
	return nil
}

// DismissAllNotifications dismisses all active notifications
func (nm *NotificationManager) DismissAllNotifications() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.notifications = make(map[string]*Notification)
}

// CreateSuccessNotification creates a success notification
func (nm *NotificationManager) CreateSuccessNotification(title, message string) *Notification {
	return &Notification{
		Title:    title,
		Message:  message,
		Type:     NotificationSuccess,
		Priority: PriorityNormal,
		Duration: nm.config.DefaultDuration,
		Sound:    nm.config.SoundEnabled,
	}
}

// CreateErrorNotification creates an error notification
func (nm *NotificationManager) CreateErrorNotification(title, message string) *Notification {
	return &Notification{
		Title:    title,
		Message:  message,
		Type:     NotificationError,
		Priority: PriorityHigh,
		Duration: nm.config.DefaultDuration,
		Sound:    nm.config.SoundEnabled,
	}
}

// CreateWarningNotification creates a warning notification
func (nm *NotificationManager) CreateWarningNotification(title, message string) *Notification {
	return &Notification{
		Title:    title,
		Message:  message,
		Type:     NotificationWarning,
		Priority: PriorityNormal,
		Duration: nm.config.DefaultDuration,
		Sound:    nm.config.SoundEnabled,
	}
}

// CreateInfoNotification creates an info notification
func (nm *NotificationManager) CreateInfoNotification(title, message string) *Notification {
	return &Notification{
		Title:    title,
		Message:  message,
		Type:     NotificationInfo,
		Priority: PriorityLow,
		Duration: nm.config.DefaultDuration,
		Sound:    nm.config.SoundEnabled,
	}
}

// CreateTaskNotification creates a task completion notification
func (nm *NotificationManager) CreateTaskNotification(title, message string) *Notification {
	return &Notification{
		Title:    title,
		Message:  message,
		Type:     NotificationTask,
		Priority: PriorityNormal,
		Duration: nm.config.DefaultDuration,
		Sound:    nm.config.SoundEnabled,
	}
}

// TestNotification sends a test notification
func (nm *NotificationManager) TestNotification() error {
	testNotification := &Notification{
		Title:    "Ena Assistant",
		Message:  "Desktop notifications are working! ✨",
		Type:     NotificationInfo,
		Priority: PriorityNormal,
		Duration: 3 * time.Second,
		Sound:    true,
	}

	return nm.SendNotification(testNotification)
}

// GetPlatformInfo returns information about the current platform
func (nm *NotificationManager) GetPlatformInfo() map[string]interface{} {
	info := map[string]interface{}{
		"platform": nm.platform,
		"enabled":  nm.enabled,
		"config":   nm.config,
	}

	// Check for platform-specific tools
	switch nm.platform {
	case "darwin":
		_, hasOsascript := exec.LookPath("osascript")
		info["osascript_available"] = hasOsascript == nil
	case "linux":
		_, hasNotifySend := exec.LookPath("notify-send")
		info["notify_send_available"] = hasNotifySend == nil
	case "windows":
		info["powershell_available"] = true // Assume PowerShell is available on Windows
	}

	return info
}

// CleanupExpiredNotifications removes expired notifications
func (nm *NotificationManager) CleanupExpiredNotifications() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	now := time.Now()
	for id, notification := range nm.notifications {
		if notification.ExpiresAt.Before(now) {
			delete(nm.notifications, id)
		}
	}
}

// StartCleanupRoutine starts a background routine to clean up expired notifications
func (nm *NotificationManager) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			nm.CleanupExpiredNotifications()
		}
	}()
}

// NotificationHistoryData represents the structure of the history file
type NotificationHistoryData struct {
	Version       string          `json:"version"`
	LastUpdated   time.Time       `json:"last_updated"`
	Notifications []*Notification `json:"notifications"`
}

// saveNotificationHistory saves notification history to disk
func (nm *NotificationManager) saveNotificationHistory() error {
	historyData := NotificationHistoryData{
		Version:       "1.0",
		LastUpdated:   time.Now(),
		Notifications: nm.history,
	}

	data, err := json.MarshalIndent(historyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal notification history: %v", err)
	}

	if err := os.WriteFile(nm.historyFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write notification history file: %v", err)
	}

	return nil
}

// loadNotificationHistory loads notification history from disk
func (nm *NotificationManager) loadNotificationHistory() {
	if _, err := os.Stat(nm.historyFile); os.IsNotExist(err) {
		// History file doesn't exist, start with empty history
		return
	}

	data, err := os.ReadFile(nm.historyFile)
	if err != nil {
		fmt.Printf("Warning: Failed to read notification history file: %v\n", err)
		return
	}

	var historyData NotificationHistoryData
	if err := json.Unmarshal(data, &historyData); err != nil {
		fmt.Printf("Warning: Failed to parse notification history file: %v\n", err)
		return
	}

	// Load the notifications into history
	nm.historyMutex.Lock()
	nm.history = historyData.Notifications
	nm.historyMutex.Unlock()

	// Trim history if it exceeds max size
	if len(nm.history) > nm.maxHistory {
		nm.historyMutex.Lock()
		nm.history = nm.history[len(nm.history)-nm.maxHistory:]
		nm.historyMutex.Unlock()
	}
}

// GetHistoryFile returns the path to the history file
func (nm *NotificationManager) GetHistoryFile() string {
	return nm.historyFile
}

// SetHistoryFile sets the path to the history file
func (nm *NotificationManager) SetHistoryFile(path string) {
	nm.historyMutex.Lock()
	defer nm.historyMutex.Unlock()
	nm.historyFile = path
}
