/**
 * Automated backup system for destructive operations.
 *
 * Provides intelligent backup management, automatic backup creation,
 * restoration capabilities, and backup lifecycle management.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: backup_engine.go
 * Description: Comprehensive backup system for file safety and recovery
 */

package backup

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"ena/internal/suggestions"
)

// BackupType defines the type of backup
type BackupType string

const (
	BackupTypeFile      BackupType = "file"
	BackupTypeDirectory BackupType = "directory"
	BackupTypeOperation BackupType = "operation"
	BackupTypeManual    BackupType = "manual"
	BackupTypeScheduled BackupType = "scheduled"
)

// BackupStatus defines the status of a backup
type BackupStatus string

const (
	BackupStatusCreated   BackupStatus = "created"
	BackupStatusVerified  BackupStatus = "verified"
	BackupStatusCorrupted BackupStatus = "corrupted"
	BackupStatusRestored  BackupStatus = "restored"
	BackupStatusExpired   BackupStatus = "expired"
)

// BackupMetadata contains metadata about a backup
type BackupMetadata struct {
	OriginalPath string       `json:"original_path"`
	BackupPath   string       `json:"backup_path"`
	Checksum     string       `json:"checksum"`
	Size         int64        `json:"size"`
	Type         BackupType   `json:"type"`
	Status       BackupStatus `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	ExpiresAt    *time.Time   `json:"expires_at,omitempty"`
	OperationID  string       `json:"operation_id,omitempty"`
	Description  string       `json:"description"`
	Tags         []string     `json:"tags"`
	Compressed   bool         `json:"compressed"`
	Encrypted    bool         `json:"encrypted"`
}

// BackupConfig defines backup configuration
type BackupConfig struct {
	Enabled         bool     `json:"enabled"`
	MaxBackups      int      `json:"max_backups"`
	RetentionDays   int      `json:"retention_days"`
	Compression     bool     `json:"compression"`
	Encryption      bool     `json:"encryption"`
	BackupDirectory string   `json:"backup_directory"`
	AutoCleanup     bool     `json:"auto_cleanup"`
	VerifyChecksums bool     `json:"verify_checksums"`
	ExcludePatterns []string `json:"exclude_patterns"`
	IncludePatterns []string `json:"include_patterns"`
	MaxBackupSize   int64    `json:"max_backup_size"`
	MinFreeSpace    int64    `json:"min_free_space"`
}

// BackupOperation represents a backup operation
type BackupOperation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Destination string                 `json:"destination,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	Backups     []BackupMetadata       `json:"backups"`
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	OperationID    string           `json:"operation_id"`
	BackupsCreated int              `json:"backups_created"`
	BackupsSkipped int              `json:"backups_skipped"`
	BackupsFailed  int              `json:"backups_failed"`
	TotalSize      int64            `json:"total_size"`
	Duration       time.Duration    `json:"duration"`
	Errors         []string         `json:"errors"`
	Backups        []BackupMetadata `json:"backups"`
}

// BackupEngine manages automated backup operations
type BackupEngine struct {
	config         BackupConfig
	analytics      *suggestions.UsageAnalytics
	mutex          sync.RWMutex
	operations     map[string]*BackupOperation
	backups        map[string]*BackupMetadata
	configFile     string
	operationsFile string
	backupsFile    string
	eventCallbacks map[string][]BackupEventCallback
	isRunning      bool
	stopChan       chan struct{}
	cleanupTicker  *time.Ticker
}

// BackupEventCallback is a function that gets called on backup events
type BackupEventCallback func(event BackupEvent)

// BackupEvent represents an event that occurred during backup operations
type BackupEvent struct {
	Type        string                 `json:"type"`
	OperationID string                 `json:"operation_id,omitempty"`
	BackupID    string                 `json:"backup_id,omitempty"`
	FilePath    string                 `json:"file_path,omitempty"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewBackupEngine creates a new backup engine instance
func NewBackupEngine(analytics *suggestions.UsageAnalytics) *BackupEngine {
	be := &BackupEngine{
		config: BackupConfig{
			Enabled:         true,
			MaxBackups:      100,
			RetentionDays:   30,
			Compression:     true,
			Encryption:      false,
			BackupDirectory: "~/.ena/backups",
			AutoCleanup:     true,
			VerifyChecksums: true,
			ExcludePatterns: []string{".git/", "node_modules/", ".DS_Store"},
			IncludePatterns: []string{},
			MaxBackupSize:   100 * 1024 * 1024 * 1024, // 100GB
			MinFreeSpace:    1 * 1024 * 1024 * 1024,   // 1GB
		},
		analytics:      analytics,
		operations:     make(map[string]*BackupOperation),
		backups:        make(map[string]*BackupMetadata),
		configFile:     "backup_config.json",
		operationsFile: "backup_operations.json",
		backupsFile:    "backup_metadata.json",
		eventCallbacks: make(map[string][]BackupEventCallback),
		stopChan:       make(chan struct{}),
	}

	// Expand backup directory path
	be.expandBackupDirectory()

	// Load existing data
	be.loadConfig()
	be.loadOperations()
	be.loadBackups()

	// Start cleanup routine if enabled
	if be.config.AutoCleanup {
		be.startCleanupRoutine()
	}

	return be
}

// CreateBackup creates a backup of the specified file or directory
func (be *BackupEngine) CreateBackup(sourcePath, operationID, description string, tags []string) (*BackupMetadata, error) {
	be.mutex.Lock()
	defer be.mutex.Unlock()

	if !be.config.Enabled {
		return nil, fmt.Errorf("backup system is disabled")
	}

	// Check if backup is needed
	if !be.shouldBackup(sourcePath) {
		return nil, fmt.Errorf("backup not needed for %s", sourcePath)
	}

	// Generate backup ID
	backupID := fmt.Sprintf("backup_%d", time.Now().UnixNano())

	// Create backup metadata
	metadata := &BackupMetadata{
		OriginalPath: sourcePath,
		BackupPath:   be.generateBackupPath(sourcePath, backupID),
		Type:         be.detectBackupType(sourcePath),
		Status:       BackupStatusCreated,
		CreatedAt:    time.Now(),
		OperationID:  operationID,
		Description:  description,
		Tags:         tags,
		Compressed:   be.config.Compression,
		Encrypted:    be.config.Encryption,
	}

	// Set expiration time
	if be.config.RetentionDays > 0 {
		expiresAt := time.Now().Add(time.Duration(be.config.RetentionDays) * 24 * time.Hour)
		metadata.ExpiresAt = &expiresAt
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(metadata.BackupPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Perform the backup
	if err := be.performBackup(metadata); err != nil {
		return nil, fmt.Errorf("failed to perform backup: %v", err)
	}

	// Verify backup if enabled
	if be.config.VerifyChecksums {
		if err := be.verifyBackup(metadata); err != nil {
			metadata.Status = BackupStatusCorrupted
			be.backups[backupID] = metadata
			be.saveBackups()
			return metadata, fmt.Errorf("backup verification failed: %v", err)
		}
		metadata.Status = BackupStatusVerified
	}

	// Store backup metadata
	be.backups[backupID] = metadata

	// Update operation if exists
	if operation, exists := be.operations[operationID]; exists {
		operation.Backups = append(operation.Backups, *metadata)
	}

	// Save data
	be.saveBackups()
	be.saveOperations()

	// Trigger event
	be.triggerEvent(BackupEvent{
		Type:        "backup_created",
		OperationID: operationID,
		BackupID:    backupID,
		FilePath:    sourcePath,
		Message:     fmt.Sprintf("Backup created for %s", sourcePath),
		Data: map[string]interface{}{
			"backup_path": metadata.BackupPath,
			"size":        metadata.Size,
			"checksum":    metadata.Checksum,
		},
		Timestamp: time.Now(),
	})

	return metadata, nil
}

// CreateOperationBackup creates backups for a complete operation
func (be *BackupEngine) CreateOperationBackup(operationID, operationType, source string, files []string) (*BackupResult, error) {
	startTime := time.Now()
	result := &BackupResult{
		OperationID: operationID,
		Backups:     make([]BackupMetadata, 0),
	}

	// Create operation record
	operation := &BackupOperation{
		ID:        operationID,
		Type:      operationType,
		Source:    source,
		CreatedAt: time.Now(),
		Backups:   make([]BackupMetadata, 0),
	}

	be.mutex.Lock()
	be.operations[operationID] = operation
	be.mutex.Unlock()

	// Create backups for each file
	for _, filePath := range files {
		metadata, err := be.CreateBackup(filePath, operationID, fmt.Sprintf("Operation: %s", operationType), []string{operationType})
		if err != nil {
			result.BackupsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("failed to backup %s: %v", filePath, err))
			continue
		}

		result.Backups = append(result.Backups, *metadata)
		result.BackupsCreated++
		result.TotalSize += metadata.Size
	}

	result.Duration = time.Since(startTime)

	// Trigger event
	be.triggerEvent(BackupEvent{
		Type:        "operation_backup_completed",
		OperationID: operationID,
		Message:     fmt.Sprintf("Operation backup completed: %d backups created", result.BackupsCreated),
		Data: map[string]interface{}{
			"backups_created": result.BackupsCreated,
			"backups_failed":  result.BackupsFailed,
			"total_size":      result.TotalSize,
			"duration":        result.Duration.String(),
		},
		Timestamp: time.Now(),
	})

	return result, nil
}

// RestoreBackup restores a backup to its original location or a new location
func (be *BackupEngine) RestoreBackup(backupID, destinationPath string, overwrite bool) error {
	be.mutex.RLock()
	metadata, exists := be.backups[backupID]
	be.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("backup %s not found", backupID)
	}

	if metadata.Status == BackupStatusCorrupted {
		return fmt.Errorf("cannot restore corrupted backup %s", backupID)
	}

	// Determine destination path
	if destinationPath == "" {
		destinationPath = metadata.OriginalPath
	}

	// Check if destination exists and overwrite is not allowed
	if !overwrite {
		if _, err := os.Stat(destinationPath); err == nil {
			return fmt.Errorf("destination %s already exists and overwrite is disabled", destinationPath)
		}
	}

	// Verify backup integrity
	if be.config.VerifyChecksums {
		if err := be.verifyBackup(metadata); err != nil {
			return fmt.Errorf("backup verification failed: %v", err)
		}
	}

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Perform restoration
	if err := be.performRestore(metadata, destinationPath); err != nil {
		return fmt.Errorf("failed to restore backup: %v", err)
	}

	// Update metadata
	be.mutex.Lock()
	metadata.Status = BackupStatusRestored
	be.mutex.Unlock()

	// Save metadata
	be.saveBackups()

	// Trigger event
	be.triggerEvent(BackupEvent{
		Type:      "backup_restored",
		BackupID:  backupID,
		FilePath:  destinationPath,
		Message:   fmt.Sprintf("Backup restored to %s", destinationPath),
		Timestamp: time.Now(),
	})

	return nil
}

// ListBackups returns all backups, optionally filtered
func (be *BackupEngine) ListBackups(filter map[string]interface{}) []BackupMetadata {
	be.mutex.RLock()
	defer be.mutex.RUnlock()

	var backups []BackupMetadata
	for _, backup := range be.backups {
		if be.matchesFilter(backup, filter) {
			backups = append(backups, *backup)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups
}

// DeleteBackup deletes a backup and its associated files
func (be *BackupEngine) DeleteBackup(backupID string) error {
	be.mutex.Lock()
	defer be.mutex.Unlock()

	metadata, exists := be.backups[backupID]
	if !exists {
		return fmt.Errorf("backup %s not found", backupID)
	}

	// Delete backup file
	if err := os.Remove(metadata.BackupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete backup file: %v", err)
	}

	// Remove from metadata
	delete(be.backups, backupID)

	// Remove from operations
	for _, operation := range be.operations {
		for i, backup := range operation.Backups {
			if backup.BackupPath == metadata.BackupPath {
				operation.Backups = append(operation.Backups[:i], operation.Backups[i+1:]...)
				break
			}
		}
	}

	// Save data
	be.saveBackups()
	be.saveOperations()

	// Trigger event
	be.triggerEvent(BackupEvent{
		Type:      "backup_deleted",
		BackupID:  backupID,
		FilePath:  metadata.OriginalPath,
		Message:   fmt.Sprintf("Backup deleted for %s", metadata.OriginalPath),
		Timestamp: time.Now(),
	})

	return nil
}

// CleanupExpiredBackups removes expired backups
func (be *BackupEngine) CleanupExpiredBackups() (int, error) {
	be.mutex.Lock()
	defer be.mutex.Unlock()

	var expiredBackups []string
	now := time.Now()

	for backupID, metadata := range be.backups {
		if metadata.ExpiresAt != nil && now.After(*metadata.ExpiresAt) {
			expiredBackups = append(expiredBackups, backupID)
		}
	}

	cleanedCount := 0
	for _, backupID := range expiredBackups {
		metadata := be.backups[backupID]

		// Delete backup file
		if err := os.Remove(metadata.BackupPath); err != nil && !os.IsNotExist(err) {
			continue // Skip if file deletion fails
		}

		// Remove from metadata
		delete(be.backups, backupID)
		cleanedCount++
	}

	if cleanedCount > 0 {
		be.saveBackups()

		// Trigger event
		be.triggerEvent(BackupEvent{
			Type:      "backup_cleanup",
			Message:   fmt.Sprintf("Cleaned up %d expired backups", cleanedCount),
			Timestamp: time.Now(),
		})
	}

	return cleanedCount, nil
}

// GetBackupStats returns backup statistics
func (be *BackupEngine) GetBackupStats() map[string]interface{} {
	be.mutex.RLock()
	defer be.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_backups":    len(be.backups),
		"total_operations": len(be.operations),
		"total_size":       int64(0),
		"status_counts":    make(map[string]int),
		"type_counts":      make(map[string]int),
		"oldest_backup":    time.Time{},
		"newest_backup":    time.Time{},
	}

	var totalSize int64
	var oldestTime, newestTime time.Time
	statusCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, backup := range be.backups {
		totalSize += backup.Size

		// Status counts
		statusCounts[string(backup.Status)]++

		// Type counts
		typeCounts[string(backup.Type)]++

		// Time tracking
		if oldestTime.IsZero() || backup.CreatedAt.Before(oldestTime) {
			oldestTime = backup.CreatedAt
		}
		if newestTime.IsZero() || backup.CreatedAt.After(newestTime) {
			newestTime = backup.CreatedAt
		}
	}

	stats["total_size"] = totalSize
	stats["status_counts"] = statusCounts
	stats["type_counts"] = typeCounts
	stats["oldest_backup"] = oldestTime
	stats["newest_backup"] = newestTime

	return stats
}

// Private helper methods

func (be *BackupEngine) expandBackupDirectory() {
	if strings.HasPrefix(be.config.BackupDirectory, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			be.config.BackupDirectory = filepath.Join(homeDir, be.config.BackupDirectory[2:])
		}
	}
}

func (be *BackupEngine) shouldBackup(sourcePath string) bool {
	// Check exclude patterns
	for _, pattern := range be.config.ExcludePatterns {
		if strings.Contains(sourcePath, pattern) {
			return false
		}
	}

	// Check include patterns
	if len(be.config.IncludePatterns) > 0 {
		matched := false
		for _, pattern := range be.config.IncludePatterns {
			if strings.Contains(sourcePath, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check file size
	if info, err := os.Stat(sourcePath); err == nil {
		if info.Size() > be.config.MaxBackupSize {
			return false
		}
	}

	return true
}

func (be *BackupEngine) detectBackupType(sourcePath string) BackupType {
	if info, err := os.Stat(sourcePath); err == nil {
		if info.IsDir() {
			return BackupTypeDirectory
		}
		return BackupTypeFile
	}
	return BackupTypeFile
}

func (be *BackupEngine) generateBackupPath(sourcePath, backupID string) string {
	// Create a safe filename from the source path
	safePath := strings.ReplaceAll(sourcePath, "/", "_")
	safePath = strings.ReplaceAll(safePath, " ", "_")
	safePath = strings.ReplaceAll(safePath, "..", "_")

	// Generate backup filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s_%s", timestamp, safePath, backupID)

	return filepath.Join(be.config.BackupDirectory, filename)
}

func (be *BackupEngine) performBackup(metadata *BackupMetadata) error {
	sourceFile, err := os.Open(metadata.OriginalPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(metadata.BackupPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy file and calculate checksum
	hash := md5.New()
	multiWriter := io.MultiWriter(destFile, hash)

	size, err := io.Copy(multiWriter, sourceFile)
	if err != nil {
		return err
	}

	metadata.Size = size
	metadata.Checksum = hex.EncodeToString(hash.Sum(nil))

	return nil
}

func (be *BackupEngine) verifyBackup(metadata *BackupMetadata) error {
	file, err := os.Open(metadata.BackupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	calculatedChecksum := hex.EncodeToString(hash.Sum(nil))
	if calculatedChecksum != metadata.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", metadata.Checksum, calculatedChecksum)
	}

	return nil
}

func (be *BackupEngine) performRestore(metadata *BackupMetadata, destinationPath string) error {
	sourceFile, err := os.Open(metadata.BackupPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (be *BackupEngine) matchesFilter(backup *BackupMetadata, filter map[string]interface{}) bool {
	for key, value := range filter {
		switch key {
		case "operation_id":
			if backup.OperationID != value {
				return false
			}
		case "type":
			if string(backup.Type) != value {
				return false
			}
		case "status":
			if string(backup.Status) != value {
				return false
			}
		case "tag":
			if valueStr, ok := value.(string); ok {
				found := false
				for _, tag := range backup.Tags {
					if tag == valueStr {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	return true
}

func (be *BackupEngine) startCleanupRoutine() {
	be.cleanupTicker = time.NewTicker(24 * time.Hour) // Run daily
	go func() {
		for {
			select {
			case <-be.cleanupTicker.C:
				be.CleanupExpiredBackups()
			case <-be.stopChan:
				return
			}
		}
	}()
}

func (be *BackupEngine) loadConfig() error {
	if _, err := os.Stat(be.configFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(be.configFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &be.config)
}

func (be *BackupEngine) saveConfig() error {
	data, err := json.MarshalIndent(be.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(be.configFile, data, 0644)
}

func (be *BackupEngine) loadOperations() error {
	if _, err := os.Stat(be.operationsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(be.operationsFile)
	if err != nil {
		return err
	}

	var operations []*BackupOperation
	if err := json.Unmarshal(data, &operations); err != nil {
		return err
	}

	for _, operation := range operations {
		be.operations[operation.ID] = operation
	}

	return nil
}

func (be *BackupEngine) saveOperations() error {
	var operations []*BackupOperation
	for _, operation := range be.operations {
		operations = append(operations, operation)
	}

	data, err := json.MarshalIndent(operations, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(be.operationsFile, data, 0644)
}

func (be *BackupEngine) loadBackups() error {
	if _, err := os.Stat(be.backupsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(be.backupsFile)
	if err != nil {
		return err
	}

	var backups []*BackupMetadata
	if err := json.Unmarshal(data, &backups); err != nil {
		return err
	}

	for _, backup := range backups {
		// Generate backup ID from backup path
		backupID := filepath.Base(backup.BackupPath)
		be.backups[backupID] = backup
	}

	return nil
}

func (be *BackupEngine) saveBackups() error {
	var backups []*BackupMetadata
	for _, backup := range be.backups {
		backups = append(backups, backup)
	}

	data, err := json.MarshalIndent(backups, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(be.backupsFile, data, 0644)
}

func (be *BackupEngine) triggerEvent(event BackupEvent) {
	be.mutex.RLock()
	callbacks := be.eventCallbacks[event.Type]
	be.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(event) // Run callbacks asynchronously
	}
}

// AddEventCallback adds a callback for backup events
func (be *BackupEngine) AddEventCallback(eventType string, callback BackupEventCallback) {
	be.mutex.Lock()
	defer be.mutex.Unlock()

	be.eventCallbacks[eventType] = append(be.eventCallbacks[eventType], callback)
}

// UpdateConfig updates the backup configuration
func (be *BackupEngine) UpdateConfig(config BackupConfig) error {
	be.mutex.Lock()
	defer be.mutex.Unlock()

	be.config = config
	be.expandBackupDirectory()

	return be.saveConfig()
}

// GetConfig returns the current backup configuration
func (be *BackupEngine) GetConfig() BackupConfig {
	be.mutex.RLock()
	defer be.mutex.RUnlock()

	return be.config
}

// Shutdown gracefully shuts down the backup engine
func (be *BackupEngine) Shutdown() error {
	if be.cleanupTicker != nil {
		be.cleanupTicker.Stop()
	}

	close(be.stopChan)
	be.isRunning = false

	return nil
}
