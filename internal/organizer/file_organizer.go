/**
 * Smart file organization system for automatic file sorting and management.
 *
 * Provides intelligent file classification, automatic organization rules,
 * and real-time file watching for seamless file management.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: file_organizer.go
 * Description: Core file organization engine with intelligent classification and sorting
 */

package organizer

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"ena/internal/suggestions"
)

// FileType represents the type of a file based on extension and content
type FileType struct {
	Category    string   `json:"category"`    // documents, images, videos, audio, archives, code, etc.
	Subcategory string   `json:"subcategory"` // pdf, jpg, mp4, zip, etc.
	Extensions  []string `json:"extensions"`  // [".pdf", ".doc", ".docx"]
	MimeTypes   []string `json:"mime_types"`  // ["application/pdf", "text/plain"]
	Keywords    []string `json:"keywords"`    // ["document", "text", "read"]
}

// OrganizationRule defines how files should be organized
type OrganizationRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Priority    int            `json:"priority"`     // Higher priority rules are applied first
	SourcePaths []string       `json:"source_paths"` // Paths to watch for files
	DestPath    string         `json:"dest_path"`    // Destination directory
	FileTypes   []string       `json:"file_types"`   // File types to organize
	Patterns    []string       `json:"patterns"`     // Regex patterns for file matching
	Conditions  RuleConditions `json:"conditions"`
	Actions     []RuleAction   `json:"actions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// RuleConditions defines conditions for when a rule should apply
type RuleConditions struct {
	MinSize     int64    `json:"min_size"`     // Minimum file size in bytes
	MaxSize     int64    `json:"max_size"`     // Maximum file size in bytes
	MinAge      string   `json:"min_age"`      // Minimum file age (e.g., "1h", "24h", "7d")
	MaxAge      string   `json:"max_age"`      // Maximum file age
	FileCount   int      `json:"file_count"`   // Minimum number of files to trigger rule
	ExcludeDirs []string `json:"exclude_dirs"` // Directories to exclude
}

// RuleAction defines what action to take when a rule matches
type RuleAction struct {
	Type        string            `json:"type"`        // move, copy, rename, delete
	Destination string            `json:"destination"` // For move/copy actions
	Template    string            `json:"template"`    // For rename actions
	Parameters  map[string]string `json:"parameters"`  // Additional parameters
}

// OrganizationResult represents the result of organizing files
type OrganizationResult struct {
	RuleID         string                `json:"rule_id"`
	FilesProcessed int                   `json:"files_processed"`
	FilesMoved     int                   `json:"files_moved"`
	FilesCopied    int                   `json:"files_copied"`
	FilesRenamed   int                   `json:"files_renamed"`
	FilesDeleted   int                   `json:"files_deleted"`
	Errors         []string              `json:"errors"`
	Duration       time.Duration         `json:"duration"`
	Details        []FileOperationDetail `json:"details"`
}

// FileOperationDetail contains details about a specific file operation
type FileOperationDetail struct {
	FilePath     string    `json:"file_path"`
	Action       string    `json:"action"`
	Destination  string    `json:"destination,omitempty"`
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
}

// FileOrganizer manages file organization operations
type FileOrganizer struct {
	rules          map[string]*OrganizationRule
	fileTypes      map[string]*FileType
	analytics      *suggestions.UsageAnalytics
	mutex          sync.RWMutex
	configFile     string
	rulesFile      string
	eventCallbacks map[string][]OrganizationEventCallback
	isRunning      bool
	stopChan       chan struct{}
}

// OrganizationEventCallback is a function that gets called on organization events
type OrganizationEventCallback func(event OrganizationEvent)

// OrganizationEvent represents an event that occurred during organization
type OrganizationEvent struct {
	Type      string                 `json:"type"` // rule_applied, file_organized, error, etc.
	RuleID    string                 `json:"rule_id,omitempty"`
	FilePath  string                 `json:"file_path,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewFileOrganizer creates a new file organizer instance
func NewFileOrganizer(analytics *suggestions.UsageAnalytics) *FileOrganizer {
	fo := &FileOrganizer{
		rules:          make(map[string]*OrganizationRule),
		fileTypes:      make(map[string]*FileType),
		analytics:      analytics,
		configFile:     "organizer_config.json",
		rulesFile:      "organizer_rules.json",
		eventCallbacks: make(map[string][]OrganizationEventCallback),
		stopChan:       make(chan struct{}),
	}

	// Initialize default file types
	fo.initializeDefaultFileTypes()

	// Load existing rules
	fo.loadRules()

	return fo
}

// AddRule adds a new organization rule
func (fo *FileOrganizer) AddRule(rule *OrganizationRule) error {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	fo.rules[rule.ID] = rule

	fo.triggerEvent(OrganizationEvent{
		Type:      "rule_added",
		RuleID:    rule.ID,
		Message:   fmt.Sprintf("Added organization rule: %s", rule.Name),
		Timestamp: time.Now(),
	})

	return fo.saveRules()
}

// RemoveRule removes an organization rule
func (fo *FileOrganizer) RemoveRule(ruleID string) error {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()

	if _, exists := fo.rules[ruleID]; !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	delete(fo.rules, ruleID)

	fo.triggerEvent(OrganizationEvent{
		Type:      "rule_removed",
		RuleID:    ruleID,
		Message:   fmt.Sprintf("Removed organization rule: %s", ruleID),
		Timestamp: time.Now(),
	})

	return fo.saveRules()
}

// GetRules returns all organization rules
func (fo *FileOrganizer) GetRules() []*OrganizationRule {
	fo.mutex.RLock()
	defer fo.mutex.RUnlock()

	var rules []*OrganizationRule
	for _, rule := range fo.rules {
		rules = append(rules, rule)
	}

	return rules
}

// OrganizeFiles applies organization rules to files in specified directories
func (fo *FileOrganizer) OrganizeFiles(sourcePaths []string, dryRun bool) ([]OrganizationResult, error) {
	fo.mutex.RLock()
	rules := make([]*OrganizationRule, 0, len(fo.rules))
	for _, rule := range fo.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	fo.mutex.RUnlock()

	// Sort rules by priority (higher priority first)
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority < rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	var results []OrganizationResult

	for _, rule := range rules {
		// Check if rule applies to any of the source paths
		if !fo.ruleAppliesToPaths(rule, sourcePaths) {
			continue
		}

		result, err := fo.applyRule(rule, sourcePaths, dryRun)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		}

		results = append(results, result)
	}

	return results, nil
}

// OrganizeFile organizes a single file using applicable rules
func (fo *FileOrganizer) OrganizeFile(filePath string, dryRun bool) (*OrganizationResult, error) {
	fo.mutex.RLock()
	rules := make([]*OrganizationRule, 0, len(fo.rules))
	for _, rule := range fo.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	fo.mutex.RUnlock()

	// Sort rules by priority
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority < rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	// Find the first applicable rule
	for _, rule := range rules {
		if fo.ruleMatchesFile(rule, filePath) {
			result, err := fo.applyRuleToFile(rule, filePath, dryRun)
			if err != nil {
				result.Errors = append(result.Errors, err.Error())
			}
			return &result, err
		}
	}

	return &OrganizationResult{
		FilesProcessed: 1,
		Details: []FileOperationDetail{
			{
				FilePath: filePath,
				Action:   "none",
				Success:  true,
			},
		},
	}, nil
}

// StartWatching starts real-time file watching for automatic organization
func (fo *FileOrganizer) StartWatching() error {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()

	if fo.isRunning {
		return fmt.Errorf("file organizer is already running")
	}

	fo.isRunning = true
	go fo.watchFiles()

	fo.triggerEvent(OrganizationEvent{
		Type:      "watching_started",
		Message:   "Started real-time file organization",
		Timestamp: time.Now(),
	})

	return nil
}

// StopWatching stops real-time file watching
func (fo *FileOrganizer) StopWatching() error {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()

	if !fo.isRunning {
		return fmt.Errorf("file organizer is not running")
	}

	close(fo.stopChan)
	fo.isRunning = false

	fo.triggerEvent(OrganizationEvent{
		Type:      "watching_stopped",
		Message:   "Stopped real-time file organization",
		Timestamp: time.Now(),
	})

	return nil
}

// GetFileType classifies a file and returns its type information
func (fo *FileOrganizer) GetFileType(filePath string) (*FileType, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check by extension first
	for _, fileType := range fo.fileTypes {
		for _, typeExt := range fileType.Extensions {
			if ext == typeExt {
				return fileType, nil
			}
		}
	}

	// Try to determine by MIME type
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read first 512 bytes for MIME detection
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = http.DetectContentType(buffer)
	}

	// Check by MIME type
	for _, fileType := range fo.fileTypes {
		for _, typeMime := range fileType.MimeTypes {
			if strings.HasPrefix(mimeType, typeMime) {
				return fileType, nil
			}
		}
	}

	// Return unknown type
	return &FileType{
		Category:    "unknown",
		Subcategory: "unknown",
		Extensions:  []string{ext},
		MimeTypes:   []string{mimeType},
		Keywords:    []string{},
	}, nil
}

// Private helper methods

func (fo *FileOrganizer) initializeDefaultFileTypes() {
	defaultTypes := []*FileType{
		{
			Category:    "documents",
			Subcategory: "pdf",
			Extensions:  []string{".pdf"},
			MimeTypes:   []string{"application/pdf"},
			Keywords:    []string{"document", "pdf", "text"},
		},
		{
			Category:    "documents",
			Subcategory: "word",
			Extensions:  []string{".doc", ".docx"},
			MimeTypes:   []string{"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
			Keywords:    []string{"document", "word", "text"},
		},
		{
			Category:    "images",
			Subcategory: "jpeg",
			Extensions:  []string{".jpg", ".jpeg"},
			MimeTypes:   []string{"image/jpeg"},
			Keywords:    []string{"image", "photo", "picture"},
		},
		{
			Category:    "images",
			Subcategory: "png",
			Extensions:  []string{".png"},
			MimeTypes:   []string{"image/png"},
			Keywords:    []string{"image", "photo", "picture"},
		},
		{
			Category:    "videos",
			Subcategory: "mp4",
			Extensions:  []string{".mp4"},
			MimeTypes:   []string{"video/mp4"},
			Keywords:    []string{"video", "movie", "film"},
		},
		{
			Category:    "audio",
			Subcategory: "mp3",
			Extensions:  []string{".mp3"},
			MimeTypes:   []string{"audio/mpeg"},
			Keywords:    []string{"audio", "music", "sound"},
		},
		{
			Category:    "archives",
			Subcategory: "zip",
			Extensions:  []string{".zip"},
			MimeTypes:   []string{"application/zip"},
			Keywords:    []string{"archive", "compressed", "zip"},
		},
		{
			Category:    "code",
			Subcategory: "go",
			Extensions:  []string{".go"},
			MimeTypes:   []string{"text/x-go"},
			Keywords:    []string{"code", "programming", "go"},
		},
	}

	for _, fileType := range defaultTypes {
		fo.fileTypes[fileType.Subcategory] = fileType
	}
}

func (fo *FileOrganizer) ruleAppliesToPaths(rule *OrganizationRule, sourcePaths []string) bool {
	for _, sourcePath := range sourcePaths {
		for _, rulePath := range rule.SourcePaths {
			if strings.HasPrefix(sourcePath, rulePath) || strings.HasPrefix(rulePath, sourcePath) {
				return true
			}
		}
	}
	return false
}

func (fo *FileOrganizer) ruleMatchesFile(rule *OrganizationRule, filePath string) bool {
	// Check file type
	fileType, err := fo.GetFileType(filePath)
	if err == nil {
		for _, ruleFileType := range rule.FileTypes {
			if fileType.Subcategory == ruleFileType || fileType.Category == ruleFileType {
				return true
			}
		}
	}

	// Check patterns
	for _, pattern := range rule.Patterns {
		matched, err := regexp.MatchString(pattern, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}
	}

	return false
}

func (fo *FileOrganizer) applyRule(rule *OrganizationRule, sourcePaths []string, dryRun bool) (OrganizationResult, error) {
	result := OrganizationResult{
		RuleID: rule.ID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Collect files to process
	var filesToProcess []string
	for _, sourcePath := range sourcePaths {
		err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if fo.ruleMatchesFile(rule, path) {
				filesToProcess = append(filesToProcess, path)
			}

			return nil
		})

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("error walking %s: %v", sourcePath, err))
		}
	}

	result.FilesProcessed = len(filesToProcess)

	// Process files
	for _, filePath := range filesToProcess {
		detail, err := fo.processFile(rule, filePath, dryRun)
		if err != nil {
			detail.Error = err.Error()
			result.Errors = append(result.Errors, err.Error())
		}

		result.Details = append(result.Details, detail)

		switch detail.Action {
		case "move":
			if detail.Success {
				result.FilesMoved++
			}
		case "copy":
			if detail.Success {
				result.FilesCopied++
			}
		case "rename":
			if detail.Success {
				result.FilesRenamed++
			}
		case "delete":
			if detail.Success {
				result.FilesDeleted++
			}
		}
	}

	return result, nil
}

func (fo *FileOrganizer) applyRuleToFile(rule *OrganizationRule, filePath string, dryRun bool) (OrganizationResult, error) {
	result := OrganizationResult{
		RuleID: rule.ID,
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime)
	}()

	detail, err := fo.processFile(rule, filePath, dryRun)
	if err != nil {
		detail.Error = err.Error()
		result.Errors = append(result.Errors, err.Error())
	}

	result.Details = append(result.Details, detail)
	result.FilesProcessed = 1

	if detail.Success {
		switch detail.Action {
		case "move":
			result.FilesMoved = 1
		case "copy":
			result.FilesCopied = 1
		case "rename":
			result.FilesRenamed = 1
		case "delete":
			result.FilesDeleted = 1
		}
	}

	return result, nil
}

func (fo *FileOrganizer) processFile(rule *OrganizationRule, filePath string, dryRun bool) (FileOperationDetail, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return FileOperationDetail{
			FilePath: filePath,
			Action:   "error",
			Success:  false,
			Error:    err.Error(),
		}, err
	}

	detail := FileOperationDetail{
		FilePath:     filePath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		Success:      false,
	}

	// Apply actions
	for _, action := range rule.Actions {
		switch action.Type {
		case "move":
			if !dryRun {
				destPath := fo.buildDestinationPath(rule, filePath, action.Destination)
				err = fo.moveFile(filePath, destPath)
				if err != nil {
					detail.Error = err.Error()
					return detail, err
				}
				detail.Destination = destPath
			}
			detail.Action = "move"
			detail.Success = true

		case "copy":
			if !dryRun {
				destPath := fo.buildDestinationPath(rule, filePath, action.Destination)
				err = fo.copyFile(filePath, destPath)
				if err != nil {
					detail.Error = err.Error()
					return detail, err
				}
				detail.Destination = destPath
			}
			detail.Action = "copy"
			detail.Success = true

		case "rename":
			if !dryRun {
				newName := fo.buildFileName(filePath, action.Template)
				err = fo.renameFile(filePath, newName)
				if err != nil {
					detail.Error = err.Error()
					return detail, err
				}
				detail.Destination = newName
			}
			detail.Action = "rename"
			detail.Success = true

		case "delete":
			if !dryRun {
				err = os.Remove(filePath)
				if err != nil {
					detail.Error = err.Error()
					return detail, err
				}
			}
			detail.Action = "delete"
			detail.Success = true
		}
	}

	return detail, nil
}

func (fo *FileOrganizer) buildDestinationPath(rule *OrganizationRule, filePath, template string) string {
	if template == "" {
		template = rule.DestPath
	}

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	// Replace placeholders
	destPath := strings.ReplaceAll(template, "{filename}", fileName)
	destPath = strings.ReplaceAll(destPath, "{name}", nameWithoutExt)
	destPath = strings.ReplaceAll(destPath, "{ext}", ext)
	destPath = strings.ReplaceAll(destPath, "{date}", time.Now().Format("2006-01-02"))
	destPath = strings.ReplaceAll(destPath, "{year}", time.Now().Format("2006"))
	destPath = strings.ReplaceAll(destPath, "{month}", time.Now().Format("01"))

	// Get file type for category-based organization
	fileType, err := fo.GetFileType(filePath)
	if err == nil {
		destPath = strings.ReplaceAll(destPath, "{category}", fileType.Category)
		destPath = strings.ReplaceAll(destPath, "{subcategory}", fileType.Subcategory)
	}

	return destPath
}

func (fo *FileOrganizer) buildFileName(filePath, template string) string {
	if template == "" {
		return filepath.Base(filePath)
	}

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	// Replace placeholders
	newName := strings.ReplaceAll(template, "{filename}", fileName)
	newName = strings.ReplaceAll(newName, "{name}", nameWithoutExt)
	newName = strings.ReplaceAll(newName, "{ext}", ext)
	newName = strings.ReplaceAll(newName, "{date}", time.Now().Format("2006-01-02"))
	newName = strings.ReplaceAll(newName, "{timestamp}", fmt.Sprintf("%d", time.Now().Unix()))

	return newName
}

func (fo *FileOrganizer) moveFile(src, dest string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return os.Rename(src, dest)
}

func (fo *FileOrganizer) copyFile(src, dest string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (fo *FileOrganizer) renameFile(src, newName string) error {
	dest := filepath.Join(filepath.Dir(src), newName)
	return os.Rename(src, dest)
}

func (fo *FileOrganizer) watchFiles() {
	// Use the existing comprehensive file watcher system
	// This will be called by the main file watcher when files are detected
	fo.triggerEvent(OrganizationEvent{
		Type:      "watching_started",
		Message:   "Real-time file organization ready",
		Timestamp: time.Now(),
	})

	// Keep running until stopped
	<-fo.stopChan

	fo.triggerEvent(OrganizationEvent{
		Type:      "watching_stopped",
		Message:   "Real-time file organization stopped",
		Timestamp: time.Now(),
	})
}

func (fo *FileOrganizer) loadRules() error {
	if _, err := os.Stat(fo.rulesFile); os.IsNotExist(err) {
		return nil // No rules file exists yet
	}

	data, err := os.ReadFile(fo.rulesFile)
	if err != nil {
		return fmt.Errorf("error reading rules file: %v", err)
	}

	var rules []*OrganizationRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("error parsing rules file: %v", err)
	}

	for _, rule := range rules {
		fo.rules[rule.ID] = rule
	}

	return nil
}

func (fo *FileOrganizer) saveRules() error {
	var rules []*OrganizationRule
	for _, rule := range fo.rules {
		rules = append(rules, rule)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling rules: %v", err)
	}

	if err := os.WriteFile(fo.rulesFile, data, 0644); err != nil {
		return fmt.Errorf("error writing rules file: %v", err)
	}

	return nil
}

func (fo *FileOrganizer) triggerEvent(event OrganizationEvent) {
	fo.mutex.RLock()
	callbacks := fo.eventCallbacks[event.Type]
	fo.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(event) // Run callbacks asynchronously
	}
}

// AddEventCallback adds a callback for organization events
func (fo *FileOrganizer) AddEventCallback(eventType string, callback OrganizationEventCallback) {
	fo.mutex.Lock()
	defer fo.mutex.Unlock()

	fo.eventCallbacks[eventType] = append(fo.eventCallbacks[eventType], callback)
}

// HandleFileEvent processes file events from the existing file watcher
func (fo *FileOrganizer) HandleFileEvent(filePath string, eventType string) {
	fo.triggerEvent(OrganizationEvent{
		Type:      "file_detected",
		FilePath:  filePath,
		Message:   fmt.Sprintf("New file detected: %s", filepath.Base(filePath)),
		Timestamp: time.Now(),
	})

	// Only process file creation and move events for organization
	if eventType == "CREATE" || eventType == "MOVE" {
		// Organize the file automatically
		result, err := fo.OrganizeFile(filePath, false) // false = not dry run
		if err != nil {
			fo.triggerEvent(OrganizationEvent{
				Type:      "error",
				FilePath:  filePath,
				Message:   fmt.Sprintf("Failed to organize file %s: %v", filePath, err),
				Timestamp: time.Now(),
			})
			return
		}

		// Report organization result
		if result.FilesProcessed > 0 {
			fo.triggerEvent(OrganizationEvent{
				Type:     "file_organized",
				FilePath: filePath,
				Message:  fmt.Sprintf("File organized: %s", filepath.Base(filePath)),
				Data: map[string]interface{}{
					"rule_id":     result.RuleID,
					"action":      result.Details[0].Action,
					"destination": result.Details[0].Destination,
					"success":     result.Details[0].Success,
				},
				Timestamp: time.Now(),
			})
		}
	}
}

// GetWatchedPaths returns all paths that should be watched based on rules
func (fo *FileOrganizer) GetWatchedPaths() []string {
	fo.mutex.RLock()
	defer fo.mutex.RUnlock()

	var paths []string
	pathMap := make(map[string]bool)

	for _, rule := range fo.rules {
		if rule.Enabled {
			for _, sourcePath := range rule.SourcePaths {
				if !pathMap[sourcePath] {
					paths = append(paths, sourcePath)
					pathMap[sourcePath] = true
				}
			}
		}
	}

	return paths
}

// GetAllFileExtensions returns all file extensions from file types
func (fo *FileOrganizer) GetAllFileExtensions() []string {
	fo.mutex.RLock()
	defer fo.mutex.RUnlock()

	var extensions []string
	extMap := make(map[string]bool)

	for _, fileType := range fo.fileTypes {
		for _, ext := range fileType.Extensions {
			if !extMap[ext] {
				extensions = append(extensions, ext)
				extMap[ext] = true
			}
		}
	}

	return extensions
}
