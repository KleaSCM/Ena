/**
 * Smart application detection and management system.
 *
 * Provides intelligent detection of installed applications across different
 * platforms with comprehensive metadata, categorization, and management capabilities.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: app_scanner.go
 * Description: Advanced application detection and scanning engine
 */

package appdetect

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"ena/internal/suggestions"
)

// AppCategory defines the category of an application
type AppCategory string

const (
	CategoryDevelopment   AppCategory = "development"
	CategoryProductivity  AppCategory = "productivity"
	CategoryMedia         AppCategory = "media"
	CategoryCommunication AppCategory = "communication"
	CategoryGames         AppCategory = "games"
	CategorySystem        AppCategory = "system"
	CategoryUtilities     AppCategory = "utilities"
	CategoryGraphics      AppCategory = "graphics"
	CategoryOffice        AppCategory = "office"
	CategoryEducation     AppCategory = "education"
	CategorySecurity      AppCategory = "security"
	CategoryNetwork       AppCategory = "network"
	CategoryUnknown       AppCategory = "unknown"
)

// AppStatus defines the status of an application
type AppStatus string

const (
	StatusInstalled   AppStatus = "installed"
	StatusRunning     AppStatus = "running"
	StatusNotRunning  AppStatus = "not_running"
	StatusUnavailable AppStatus = "unavailable"
	StatusOutdated    AppStatus = "outdated"
	StatusCorrupted   AppStatus = "corrupted"
)

// AppInfo contains comprehensive information about an application
type AppInfo struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	DisplayName      string            `json:"display_name"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	Category         AppCategory       `json:"category"`
	Status           AppStatus         `json:"status"`
	ExecutablePath   string            `json:"executable_path"`
	InstallPath      string            `json:"install_path"`
	IconPath         string            `json:"icon_path"`
	Website          string            `json:"website"`
	Author           string            `json:"author"`
	License          string            `json:"license"`
	Size             int64             `json:"size"`
	InstallDate      *time.Time        `json:"install_date,omitempty"`
	LastUsed         *time.Time        `json:"last_used,omitempty"`
	IsDefaultApp     bool              `json:"is_default_app"`
	FileAssociations []string          `json:"file_associations"`
	CommandLineArgs  []string          `json:"command_line_args"`
	EnvironmentVars  map[string]string `json:"environment_vars"`
	Metadata         map[string]string `json:"metadata"`
	DetectedAt       time.Time         `json:"detected_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// AppDetectionResult contains the result of application detection
type AppDetectionResult struct {
	TotalScanned    int            `json:"total_scanned"`
	AppsFound       int            `json:"apps_found"`
	AppsUpdated     int            `json:"apps_updated"`
	AppsRemoved     int            `json:"apps_removed"`
	CategoriesFound map[string]int `json:"categories_found"`
	Platform        string         `json:"platform"`
	ScanDuration    time.Duration  `json:"scan_duration"`
	Errors          []string       `json:"errors"`
	Apps            []AppInfo      `json:"apps"`
}

// AppScanner manages application detection and scanning
type AppScanner struct {
	analytics       *suggestions.UsageAnalytics
	apps            map[string]*AppInfo
	mutex           sync.RWMutex
	configFile      string
	appsFile        string
	eventCallbacks  map[string][]AppEventCallback
	scanPaths       []string
	excludePaths    []string
	includePatterns []string
	excludePatterns []string
	platform        string
	lastScanTime    time.Time
	isScanning      bool
}

// AppEventCallback is a function that gets called on app events
type AppEventCallback func(event AppEvent)

// AppEvent represents an event that occurred during app scanning
type AppEvent struct {
	Type      string                 `json:"type"`
	AppID     string                 `json:"app_id,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewAppScanner creates a new application scanner instance
func NewAppScanner(analytics *suggestions.UsageAnalytics) *AppScanner {
	as := &AppScanner{
		analytics:       analytics,
		apps:            make(map[string]*AppInfo),
		configFile:      "app_scanner_config.json",
		appsFile:        "detected_apps.json",
		eventCallbacks:  make(map[string][]AppEventCallback),
		platform:        runtime.GOOS,
		scanPaths:       getDefaultScanPaths(),
		excludePaths:    getDefaultExcludePaths(),
		includePatterns: getDefaultIncludePatterns(),
		excludePatterns: getDefaultExcludePatterns(),
	}

	// Load existing app data
	as.loadApps()

	return as
}

// ScanForApps performs a comprehensive scan for installed applications
func (as *AppScanner) ScanForApps(deepScan bool) (*AppDetectionResult, error) {
	if as.isScanning {
		return nil, fmt.Errorf("scan already in progress")
	}

	as.mutex.Lock()
	as.isScanning = true
	as.mutex.Unlock()

	defer func() {
		as.mutex.Lock()
		as.isScanning = false
		as.lastScanTime = time.Now()
		as.mutex.Unlock()
	}()

	startTime := time.Now()
	result := &AppDetectionResult{
		Platform:        as.platform,
		CategoriesFound: make(map[string]int),
		Errors:          make([]string, 0),
		Apps:            make([]AppInfo, 0),
	}

	as.triggerEvent(AppEvent{
		Type:      "scan_started",
		Message:   "Started application detection scan",
		Timestamp: time.Now(),
	})

	// Perform platform-specific scanning
	var apps []AppInfo
	var err error

	switch as.platform {
	case "linux":
		apps, err = as.scanLinuxApps(deepScan)
	case "darwin":
		apps, err = as.scanMacOSApps(deepScan)
	case "windows":
		apps, err = as.scanWindowsApps(deepScan)
	default:
		err = fmt.Errorf("unsupported platform: %s", as.platform)
	}

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		as.triggerEvent(AppEvent{
			Type:      "scan_failed",
			Message:   fmt.Sprintf("Scan failed: %v", err),
			Timestamp: time.Now(),
		})
		return result, err
	}

	// Process detected apps
	as.mutex.Lock()
	for _, app := range apps {
		// Update or add app
		existingApp, exists := as.apps[app.ID]
		if exists {
			existingApp.UpdatedAt = time.Now()
			// Update fields that might have changed
			existingApp.ExecutablePath = app.ExecutablePath
			existingApp.Status = app.Status
			existingApp.Version = app.Version
			existingApp.Size = app.Size
			result.AppsUpdated++
		} else {
			as.apps[app.ID] = &app
			result.AppsFound++
		}

		// Update category counts
		result.CategoriesFound[string(app.Category)]++
	}

	// Remove apps that no longer exist
	removedApps := as.removeNonExistentApps()
	result.AppsRemoved = len(removedApps)

	// Convert to slice for result
	for _, app := range as.apps {
		result.Apps = append(result.Apps, *app)
	}

	as.mutex.Unlock()

	// Sort apps by name
	sort.Slice(result.Apps, func(i, j int) bool {
		return result.Apps[i].DisplayName < result.Apps[j].DisplayName
	})

	result.TotalScanned = len(result.Apps)
	result.ScanDuration = time.Since(startTime)

	// Save updated app data
	as.saveApps()

	as.triggerEvent(AppEvent{
		Type:    "scan_completed",
		Message: fmt.Sprintf("Scan completed: found %d apps", result.AppsFound),
		Data: map[string]interface{}{
			"apps_found":    result.AppsFound,
			"apps_updated":  result.AppsUpdated,
			"apps_removed":  result.AppsRemoved,
			"scan_duration": result.ScanDuration.String(),
		},
		Timestamp: time.Now(),
	})

	return result, nil
}

// GetApps returns all detected applications, optionally filtered
func (as *AppScanner) GetApps(filter map[string]interface{}) []AppInfo {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	var apps []AppInfo
	for _, app := range as.apps {
		if as.matchesFilter(app, filter) {
			apps = append(apps, *app)
		}
	}

	// Sort by display name
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].DisplayName < apps[j].DisplayName
	})

	return apps
}

// GetAppByID returns a specific application by ID
func (as *AppScanner) GetAppByID(appID string) (*AppInfo, error) {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	app, exists := as.apps[appID]
	if !exists {
		return nil, fmt.Errorf("app %s not found", appID)
	}

	return app, nil
}

// GetAppsByCategory returns applications in a specific category
func (as *AppScanner) GetAppsByCategory(category AppCategory) []AppInfo {
	return as.GetApps(map[string]interface{}{
		"category": category,
	})
}

// GetRunningApps returns currently running applications
func (as *AppScanner) GetRunningApps() []AppInfo {
	return as.GetApps(map[string]interface{}{
		"status": StatusRunning,
	})
}

// GetDefaultApps returns applications that are set as defaults for file types
func (as *AppScanner) GetDefaultApps() []AppInfo {
	return as.GetApps(map[string]interface{}{
		"is_default_app": true,
	})
}

// UpdateAppStatus updates the status of an application
func (as *AppScanner) UpdateAppStatus(appID string, status AppStatus) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	app, exists := as.apps[appID]
	if !exists {
		return fmt.Errorf("app %s not found", appID)
	}

	oldStatus := app.Status
	app.Status = status
	app.UpdatedAt = time.Now()

	as.triggerEvent(AppEvent{
		Type:    "app_status_changed",
		AppID:   appID,
		Message: fmt.Sprintf("App %s status changed from %s to %s", app.Name, oldStatus, status),
		Data: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": status,
		},
		Timestamp: time.Now(),
	})

	as.saveApps()
	return nil
}

// GetAppStats returns statistics about detected applications
func (as *AppScanner) GetAppStats() map[string]interface{} {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_apps":     len(as.apps),
		"categories":     make(map[string]int),
		"status_counts":  make(map[string]int),
		"platform":       as.platform,
		"last_scan":      as.lastScanTime,
		"running_apps":   0,
		"default_apps":   0,
		"total_size":     int64(0),
		"oldest_install": time.Time{},
		"newest_install": time.Time{},
	}

	categoryCounts := make(map[string]int)
	statusCounts := make(map[string]int)
	var totalSize int64
	var oldestTime, newestTime time.Time

	for _, app := range as.apps {
		// Category counts
		categoryCounts[string(app.Category)]++

		// Status counts
		statusCounts[string(app.Status)]++

		// Size
		totalSize += app.Size

		// Running apps
		if app.Status == StatusRunning {
			stats["running_apps"] = stats["running_apps"].(int) + 1
		}

		// Default apps
		if app.IsDefaultApp {
			stats["default_apps"] = stats["default_apps"].(int) + 1
		}

		// Install dates
		if app.InstallDate != nil {
			if oldestTime.IsZero() || app.InstallDate.Before(oldestTime) {
				oldestTime = *app.InstallDate
			}
			if newestTime.IsZero() || app.InstallDate.After(newestTime) {
				newestTime = *app.InstallDate
			}
		}
	}

	stats["categories"] = categoryCounts
	stats["status_counts"] = statusCounts
	stats["total_size"] = totalSize
	stats["oldest_install"] = oldestTime
	stats["newest_install"] = newestTime

	return stats
}

// Platform-specific scanning methods

func (as *AppScanner) scanLinuxApps(deepScan bool) ([]AppInfo, error) {
	var apps []AppInfo

	// Scan common application directories
	scanDirs := []string{
		"/usr/share/applications",
		"/usr/local/share/applications",
		"/home/*/.local/share/applications",
		"/var/lib/flatpak/exports/share/applications",
		"/var/lib/snapd/desktop/applications",
	}

	for _, dir := range scanDirs {
		desktopFiles, err := filepath.Glob(filepath.Join(dir, "*.desktop"))
		if err != nil {
			continue
		}

		for _, desktopFile := range desktopFiles {
			app, err := as.parseDesktopFile(desktopFile)
			if err == nil && app != nil {
				apps = append(apps, *app)
			}
		}
	}

	// Scan installed packages if deep scan is enabled
	if deepScan {
		packageApps, err := as.scanLinuxPackages()
		if err == nil {
			apps = append(apps, packageApps...)
		}
	}

	return apps, nil
}

func (as *AppScanner) scanMacOSApps(deepScan bool) ([]AppInfo, error) {
	var apps []AppInfo

	// Scan /Applications directory
	applicationsDir := "/Applications"
	err := filepath.Walk(applicationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if strings.HasSuffix(path, ".app") && info.IsDir() {
			app, err := as.parseMacOSApp(path)
			if err == nil && app != nil {
				apps = append(apps, *app)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Scan user Applications directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userAppsDir := filepath.Join(homeDir, "Applications")
		filepath.Walk(userAppsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if strings.HasSuffix(path, ".app") && info.IsDir() {
				app, err := as.parseMacOSApp(path)
				if err == nil && app != nil {
					apps = append(apps, *app)
				}
			}

			return nil
		})
	}

	return apps, nil
}

func (as *AppScanner) scanWindowsApps(deepScan bool) ([]AppInfo, error) {
	var apps []AppInfo

	// Scan registry for installed applications
	registryApps, err := as.scanWindowsRegistry()
	if err == nil {
		apps = append(apps, registryApps...)
	}

	// Scan Program Files directories
	programFiles := []string{
		"C:\\Program Files",
		"C:\\Program Files (x86)",
	}

	for _, dir := range programFiles {
		dirApps, err := as.scanWindowsDirectory(dir)
		if err == nil {
			apps = append(apps, dirApps...)
		}
	}

	return apps, nil
}

// Helper methods for parsing application files

func (as *AppScanner) parseDesktopFile(desktopFile string) (*AppInfo, error) {
	file, err := os.Open(desktopFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	app := &AppInfo{
		ID:         generateAppID(filepath.Base(desktopFile)),
		DetectedAt: time.Now(),
		UpdatedAt:  time.Now(),
		Status:     StatusInstalled,
	}

	lines := strings.Split(string(content), "\n")
	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		if currentSection == "Desktop Entry" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Name":
				app.DisplayName = value
				if app.Name == "" {
					app.Name = value
				}
			case "GenericName":
				if app.Description == "" {
					app.Description = value
				}
			case "Comment":
				app.Description = value
			case "Exec":
				app.ExecutablePath = value
			case "Icon":
				app.IconPath = value
			case "Categories":
				app.Category = as.categorizeFromKeywords(strings.Split(value, ";"))
			}
		}
	}

	// Skip if no executable or name
	if app.ExecutablePath == "" || app.DisplayName == "" {
		return nil, fmt.Errorf("invalid desktop file: missing name or exec")
	}

	// Set install path
	app.InstallPath = filepath.Dir(desktopFile)

	// Check if app is running
	if as.isAppRunning(app.ExecutablePath) {
		app.Status = StatusRunning
	}

	return app, nil
}

func (as *AppScanner) parseMacOSApp(appPath string) (*AppInfo, error) {
	app := &AppInfo{
		ID:          generateAppID(filepath.Base(appPath)),
		DetectedAt:  time.Now(),
		UpdatedAt:   time.Now(),
		Status:      StatusInstalled,
		InstallPath: appPath,
	}

	// Parse Info.plist
	infoPlistPath := filepath.Join(appPath, "Contents", "Info.plist")
	if _, err := os.Stat(infoPlistPath); err == nil {
		// Use plutil to parse plist (simplified approach)
		cmd := exec.Command("plutil", "-p", infoPlistPath)
		output, err := cmd.Output()
		if err == nil {
			as.parsePlistOutput(string(output), app)
		}
	}

	// Set executable path
	executablePath := filepath.Join(appPath, "Contents", "MacOS", filepath.Base(appPath))
	if _, err := os.Stat(executablePath); err == nil {
		app.ExecutablePath = executablePath
	}

	// Set default values if not found
	if app.DisplayName == "" {
		app.DisplayName = strings.TrimSuffix(filepath.Base(appPath), ".app")
	}
	if app.Name == "" {
		app.Name = app.DisplayName
	}

	// Check if app is running
	if app.ExecutablePath != "" && as.isAppRunning(app.ExecutablePath) {
		app.Status = StatusRunning
	}

	return app, nil
}

// Utility methods

func (as *AppScanner) isAppRunning(executablePath string) bool {
	// Simplified running check - in production, this would be more sophisticated
	cmd := exec.Command("pgrep", "-f", filepath.Base(executablePath))
	err := cmd.Run()
	return err == nil
}

func (as *AppScanner) categorizeFromKeywords(categories []string) AppCategory {
	for _, cat := range categories {
		cat = strings.ToLower(cat)
		switch {
		case strings.Contains(cat, "development") || strings.Contains(cat, "programming"):
			return CategoryDevelopment
		case strings.Contains(cat, "productivity") || strings.Contains(cat, "office"):
			return CategoryProductivity
		case strings.Contains(cat, "graphics") || strings.Contains(cat, "image"):
			return CategoryGraphics
		case strings.Contains(cat, "multimedia") || strings.Contains(cat, "video") || strings.Contains(cat, "audio"):
			return CategoryMedia
		case strings.Contains(cat, "game"):
			return CategoryGames
		case strings.Contains(cat, "network") || strings.Contains(cat, "internet"):
			return CategoryNetwork
		case strings.Contains(cat, "system") || strings.Contains(cat, "utility"):
			return CategorySystem
		}
	}
	return CategoryUnknown
}

func (as *AppScanner) matchesFilter(app *AppInfo, filter map[string]interface{}) bool {
	for key, value := range filter {
		switch key {
		case "category":
			if app.Category != AppCategory(fmt.Sprintf("%v", value)) {
				return false
			}
		case "status":
			if app.Status != AppStatus(fmt.Sprintf("%v", value)) {
				return false
			}
		case "is_default_app":
			if app.IsDefaultApp != value.(bool) {
				return false
			}
		case "name":
			if !strings.Contains(strings.ToLower(app.Name), strings.ToLower(fmt.Sprintf("%v", value))) {
				return false
			}
		}
	}
	return true
}

func (as *AppScanner) removeNonExistentApps() []string {
	var removed []string
	for appID, app := range as.apps {
		if app.ExecutablePath != "" {
			if _, err := os.Stat(app.ExecutablePath); os.IsNotExist(err) {
				delete(as.apps, appID)
				removed = append(removed, appID)
			}
		}
	}
	return removed
}

// Event handling
func (as *AppScanner) triggerEvent(event AppEvent) {
	callbacks := as.eventCallbacks[event.Type]
	for _, callback := range callbacks {
		go callback(event) // Run callbacks asynchronously
	}
}

// AddEventCallback adds a callback for app events
func (as *AppScanner) AddEventCallback(eventType string, callback AppEventCallback) {
	as.eventCallbacks[eventType] = append(as.eventCallbacks[eventType], callback)
}

// Data persistence
func (as *AppScanner) loadApps() error {
	if _, err := os.Stat(as.appsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(as.appsFile)
	if err != nil {
		return err
	}

	var apps []AppInfo
	if err := json.Unmarshal(data, &apps); err != nil {
		return err
	}

	for _, app := range apps {
		as.apps[app.ID] = &app
	}

	return nil
}

func (as *AppScanner) saveApps() error {
	var apps []AppInfo
	for _, app := range as.apps {
		apps = append(apps, *app)
	}

	data, err := json.MarshalIndent(apps, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(as.appsFile, data, 0644)
}

// Helper functions

func generateAppID(name string) string {
	// Generate a unique ID from the app name
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	id := reg.ReplaceAllString(strings.ToLower(name), "_")
	return fmt.Sprintf("app_%s_%d", id, time.Now().Unix())
}

func getDefaultScanPaths() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{
			"/usr/share/applications",
			"/usr/local/share/applications",
			"/home/*/.local/share/applications",
		}
	case "darwin":
		return []string{
			"/Applications",
			"~/Applications",
		}
	case "windows":
		return []string{
			"C:\\Program Files",
			"C:\\Program Files (x86)",
		}
	default:
		return []string{}
	}
}

func getDefaultExcludePaths() []string {
	return []string{
		"/System",
		"/tmp",
		"/var/tmp",
		"node_modules",
		".git",
	}
}

func getDefaultIncludePatterns() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{"*.desktop"}
	case "darwin":
		return []string{"*.app"}
	case "windows":
		return []string{"*.exe", "*.msi"}
	default:
		return []string{}
	}
}

func getDefaultExcludePatterns() []string {
	return []string{
		"*.tmp",
		"*.temp",
		"*.log",
		"*.cache",
	}
}

// Simplified plist parsing (in production, use a proper plist library)
func (as *AppScanner) parsePlistOutput(output string, app *AppInfo) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "CFBundleName") && strings.Contains(line, "=>") {
			parts := strings.Split(line, "=>")
			if len(parts) > 1 {
				name := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				if app.DisplayName == "" {
					app.DisplayName = name
				}
			}
		}
		if strings.Contains(line, "CFBundleVersion") && strings.Contains(line, "=>") {
			parts := strings.Split(line, "=>")
			if len(parts) > 1 {
				version := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				app.Version = version
			}
		}
	}
}

// Placeholder methods for platform-specific scanning
func (as *AppScanner) scanLinuxPackages() ([]AppInfo, error) {
	// In production, this would scan package managers like apt, yum, pacman, etc.
	return []AppInfo{}, nil
}

func (as *AppScanner) scanWindowsRegistry() ([]AppInfo, error) {
	// In production, this would scan Windows registry for installed applications
	return []AppInfo{}, nil
}

func (as *AppScanner) scanWindowsDirectory(dir string) ([]AppInfo, error) {
	// In production, this would scan Windows directories for applications
	return []AppInfo{}, nil
}
