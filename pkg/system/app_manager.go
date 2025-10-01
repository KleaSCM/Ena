/**
 * Application Manager Package
 *
 * Provides application management capabilities including starting, stopping,
 * restarting applications, and listing running processes with proper error handling.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: app_manager.go
 * Description: Application lifecycle management and process control
 */

package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// AppManager handles all application-related operations
type AppManager struct {
	sync.RWMutex                     // Safe lock for concurrent access protection
	RunningApps  map[string]*AppInfo // Information about managed applications
	stopCh       chan struct{}       // Channel to stop monitoring
}

// AppInfo represents information about a running application
type AppInfo struct {
	Name        string      `json:"name"`
	PID         int         `json:"pid"`
	StartTime   time.Time   `json:"start_time"`
	Status      string      `json:"status"`
	CommandLine string      `json:"command_line"`
	Process     *os.Process `json:"-"` // Process reference (excluded from JSON)
}

// NewAppManager creates a new application manager instance
func NewAppManager() *AppManager {
	// Application management with care ✨
	manager := &AppManager{
		RunningApps: make(map[string]*AppInfo),
		stopCh:      make(chan struct{}),
	}

	// Start background monitoring to watch process lifecycle
	go manager.monitorProcesses()

	return manager
}

// StartApplication starts an application by name
func (am *AppManager) StartApplication(appName string) (string, error) {
	// Start application gently ✨

	// Acquire write lock
	am.Lock()
	defer am.Unlock()

	// Check if already running
	if _, exists := am.RunningApps[appName]; exists {
		return fmt.Sprintf("Application \"%s\" is already running! 😅", appName), nil
	}

	// Determine launch command based on app name
	appCommand := am.getAppCommand(appName)
	if appCommand == "" {
		// Use app name directly as command if not found
		appCommand = appName
	}

	// Launch application (properly split arguments)
	parts := am.parseCommand(appCommand)
	cmd := exec.Command(parts[0], parts[1:]...)

	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("Failed to start application \"%s\": %v", appName, err)
	}

	// Record application information
	am.RunningApps[appName] = &AppInfo{
		Name:        appName,
		PID:         cmd.Process.Pid,
		StartTime:   time.Now(),
		Status:      "running",
		CommandLine: appCommand,
		Process:     cmd.Process,
	}

	return fmt.Sprintf("Started application \"%s\"! (PID: %d) ✨", appName, cmd.Process.Pid), nil
}

// StopApplication stops a running application
func (am *AppManager) StopApplication(appName string) (string, error) {
	// Stop application gently

	// Acquire write lock
	am.Lock()
	defer am.Unlock()

	appInfo, exists := am.RunningApps[appName]
	if !exists {
		return "", fmt.Errorf("Application \"%s\" is not running! 😅", appName)
	}

	// Terminate process
	err := appInfo.Process.Kill()
	if err != nil {
		// Fallback to kill command
		cmd := exec.Command("kill", fmt.Sprintf("%d", appInfo.PID))
		err = cmd.Run()
		if err != nil {
			return "", fmt.Errorf("Failed to stop application \"%s\": %v", appName, err)
		}
	}

	// Remove application information
	delete(am.RunningApps, appName)

	return fmt.Sprintf("Stopped application \"%s\" 💤", appName), nil
}

// RestartApplication restarts an application
func (am *AppManager) RestartApplication(appName string) (string, error) {
	// Restart application with a fresh start ✨
	// First stop
	if _, exists := am.RunningApps[appName]; exists {
		_, err := am.StopApplication(appName)
		if err != nil {
			return "", fmt.Errorf("Failed to stop application \"%s\": %v", appName, err)
		}

		// Wait a bit
		time.Sleep(2 * time.Second)
	}

	// Restart
	_, err := am.StartApplication(appName)
	if err != nil {
		return "", fmt.Errorf("Failed to restart application \"%s\": %v", appName, err)
	}

	return fmt.Sprintf("Restarted application \"%s\"! ✨", appName), nil
}

// ListApplications returns a list of running applications
func (am *AppManager) ListApplications() (string, error) {
	// Show all running applications

	// Acquire read lock
	am.RLock()
	defer am.RUnlock()

	if len(am.RunningApps) == 0 {
		return "No applications are currently running 😅", nil
	}

	result := []string{
		"📱 Running Applications:",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	}

	for name, app := range am.RunningApps {
		uptime := time.Since(app.StartTime)
		result = append(result, fmt.Sprintf("📱 %s", name))
		result = append(result, fmt.Sprintf("   PID: %d", app.PID))
		result = append(result, fmt.Sprintf("   Started: %s", app.StartTime.Format("15:04:05")))
		result = append(result, fmt.Sprintf("   Uptime: %s", uptime.Round(time.Second)))
		result = append(result, fmt.Sprintf("   Status: %s", app.Status))
		result = append(result, "")
	}

	return strings.Join(result, "\n"), nil
}

// GetApplicationInfo returns detailed information about a specific application
func (am *AppManager) GetApplicationInfo(appName string) (string, error) {
	// Get detailed application information

	// Acquire read lock
	am.RLock()
	defer am.RUnlock()

	appInfo, exists := am.RunningApps[appName]
	if !exists {
		return "", fmt.Errorf("Application \"%s\" is not running! 😅", appName)
	}

	uptime := time.Since(appInfo.StartTime)

	result := []string{
		fmt.Sprintf("📱 Application Information: %s", appName),
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		fmt.Sprintf("Name: %s", appInfo.Name),
		fmt.Sprintf("PID: %d", appInfo.PID),
		fmt.Sprintf("Started: %s", appInfo.StartTime.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Uptime: %s", uptime.Round(time.Second)),
		fmt.Sprintf("Status: %s", appInfo.Status),
		fmt.Sprintf("Command: %s", appInfo.CommandLine),
	}

	return strings.Join(result, "\n"), nil
}

// getAppCommand returns the command to start a specific application
func (am *AppManager) getAppCommand(appName string) string {
	// Determine launch command based on app name and OS
	appCommands := map[string]map[string]string{
		"linux": {
			// Text editors
			"vim":    "vim",
			"nano":   "nano",
			"emacs":  "emacs",
			"code":   "code",
			"cursor": "cursor",

			// Browsers
			"firefox":  "firefox",
			"chrome":   "google-chrome",
			"chromium": "chromium",

			// File managers
			"nautilus": "nautilus",
			"dolphin":  "dolphin",
			"thunar":   "thunar",

			// Development tools
			"git":    "git",
			"docker": "docker",
			"node":   "node",
			"python": "python3",
			"go":     "go",

			// Media
			"vlc":     "vlc",
			"mpv":     "mpv",
			"spotify": "spotify",

			// System tools
			"htop":           "htop",
			"top":            "top",
			"system-monitor": "gnome-system-monitor",
			"task-manager":   "gnome-system-monitor",

			// Terminals
			"terminal":  "gnome-terminal",
			"kitty":     "kitty",
			"alacritty": "alacritty",
			"wezterm":   "wezterm",
		},
		"darwin": {
			// Text editors
			"vim":    "vim",
			"nano":   "nano",
			"emacs":  "emacs",
			"code":   "code",
			"cursor": "cursor",

			// Browsers
			"firefox": "firefox",
			"chrome":  "google-chrome",
			"safari":  "open -a Safari",

			// File managers
			"finder": "open -a Finder",

			// Development tools
			"git":    "git",
			"docker": "docker",
			"node":   "node",
			"python": "python3",
			"go":     "go",

			// Media
			"vlc":     "vlc",
			"mpv":     "mpv",
			"spotify": "spotify",

			// System tools
			"htop":             "htop",
			"top":              "top",
			"activity-monitor": "open -a 'Activity Monitor'",

			// Terminals
			"terminal":  "open -a Terminal",
			"kitty":     "kitty",
			"alacritty": "alacritty",
			"wezterm":   "wezterm",
		},
		"windows": {
			// Text editors
			"notepad": "notepad",
			"code":    "code",
			"cursor":  "cursor",

			// Browsers
			"firefox": "firefox",
			"chrome":  "chrome",
			"edge":    "msedge",

			// File managers
			"explorer": "explorer",

			// Development tools
			"git":    "git",
			"docker": "docker",
			"node":   "node",
			"python": "python",
			"go":     "go",

			// Media
			"vlc":     "vlc",
			"spotify": "spotify",

			// System tools
			"task-manager": "taskmgr",

			// Terminals
			"terminal":   "cmd",
			"powershell": "powershell",
		},
	}

	// Get current OS
	osCommands, exists := appCommands[runtime.GOOS]
	if !exists {
		// Default: use app name directly as command
		return appName
	}

	// Case-insensitive search
	appNameLower := strings.ToLower(appName)

	// Direct match
	if command, exists := osCommands[appNameLower]; exists {
		return command
	}

	// Partial match
	for name, command := range osCommands {
		if strings.Contains(appNameLower, name) {
			return command
		}
	}

	// Default: use app name directly as command
	return appName
}

// parseCommand parses a command string with proper handling of quoted arguments
func (am *AppManager) parseCommand(command string) []string {
	// Parse command with proper quote handling
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := '"'

	for _, char := range command {
		switch char {
		case '"', '\'':
			if !inQuotes {
				// Start quote
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				// End quote with same character
				inQuotes = false
			} else {
				// Different quote character inside quotes
				current.WriteRune(char)
			}
		case ' ':
			if inQuotes {
				// Keep spaces inside quotes
				current.WriteRune(char)
			} else {
				// Spaces outside quotes are separators
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// KillAllApplications stops all running applications
func (am *AppManager) KillAllApplications() (string, error) {
	// Stop all applications carefully

	// 書き込みロックを取得
	am.Lock()
	defer am.Unlock()

	if len(am.RunningApps) == 0 {
		return "No applications to stop 😅", nil
	}

	// Create copy of keys to iterate safely
	appNames := make([]string, 0, len(am.RunningApps))
	for appName := range am.RunningApps {
		appNames = append(appNames, appName)
	}

	var stoppedApps []string
	for _, appName := range appNames {
		appInfo, exists := am.RunningApps[appName]
		if !exists {
			continue // If already deleted
		}

		// Terminate process
		err := appInfo.Process.Kill()
		if err != nil {
			// Fallback to kill command
			cmd := exec.Command("kill", fmt.Sprintf("%d", appInfo.PID))
			err = cmd.Run()
			if err != nil {
				continue // Continue even with errors
			}
		}

		// Remove application information
		delete(am.RunningApps, appName)
		stoppedApps = append(stoppedApps, appName)
	}

	if len(stoppedApps) == 0 {
		return "Failed to stop applications 😅", nil
	}

	return fmt.Sprintf("Stopped the following applications: %s 💤", strings.Join(stoppedApps, ", ")), nil
}

// GetSystemProcesses returns a list of system processes
func (am *AppManager) GetSystemProcesses() (string, error) {
	// Show system process list
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("ps", "aux")
	case "darwin":
		cmd = exec.Command("ps", "aux")
	case "windows":
		cmd = exec.Command("tasklist")
	default:
		return "", fmt.Errorf("Unsupported OS: %s 😅", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get process list: %v", err)
	}

	return fmt.Sprintf("🖥️  System Process List:\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n%s", string(output)), nil
}

// monitorProcesses monitors running processes in the background
func (am *AppManager) monitorProcesses() {
	// Background monitoring - watching process lifecycle
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.checkProcessStatus()
		case <-am.stopCh:
			// Stop monitoring - thanks for your service
			return
		}
	}
}

// checkProcessStatus checks the status of all tracked processes
func (am *AppManager) checkProcessStatus() {
	// Check process status - monitoring for dead processes
	am.Lock()
	defer am.Unlock()

	var deadApps []string
	for appName, appInfo := range am.RunningApps {
		// Check if process exists using system command
		cmd := exec.Command("kill", "-0", fmt.Sprintf("%d", appInfo.PID))
		err := cmd.Run()
		if err != nil {
			// Process does not exist
			appInfo.Status = "stopped"
			deadApps = append(deadApps, appName)
		} else {
			// Update status if process is alive
			appInfo.Status = "running"
		}
	}

	// Remove dead processes
	for _, appName := range deadApps {
		delete(am.RunningApps, appName)
	}
}

// StopMonitoring stops the background process monitoring
func (am *AppManager) StopMonitoring() {
	// Stop monitoring gracefully
	close(am.stopCh)
}
