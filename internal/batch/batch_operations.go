/**
 * Batch operations system for handling multiple files and folders efficiently.
 *
 * Provides batch delete, copy, move, and other operations with progress tracking,
 * error handling, and comprehensive logging for improved user experience.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: batch_operations.go
 * Description: Core batch operations engine with progress tracking and error handling
 */

package batch

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ena/internal/progress"
	"ena/internal/suggestions"
)

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // delete, copy, move, create
	Source      string                 `json:"source"`
	Destination string                 `json:"destination,omitempty"`
	Size        int64                  `json:"size"`
	Status      string                 `json:"status"` // pending, running, completed, failed, skipped
	Error       string                 `json:"error,omitempty"`
	StartTime   time.Time              `json:"start_time,omitempty"`
	EndTime     time.Time              `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BatchJob represents a collection of operations to be executed
type BatchJob struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Operations    []BatchOperation       `json:"operations"`
	Status        string                 `json:"status"`   // pending, running, completed, failed, cancelled
	Progress      float64                `json:"progress"` // 0.0 to 1.0
	TotalSize     int64                  `json:"total_size"`
	ProcessedSize int64                  `json:"processed_size"`
	StartTime     time.Time              `json:"start_time,omitempty"`
	EndTime       time.Time              `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	ErrorCount    int                    `json:"error_count"`
	SuccessCount  int                    `json:"success_count"`
	SkippedCount  int                    `json:"skipped_count"`
	Config        BatchConfig            `json:"config"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BatchConfig contains configuration for batch operations
type BatchConfig struct {
	MaxConcurrency      int           `json:"max_concurrency"`      // Maximum concurrent operations
	SkipErrors          bool          `json:"skip_errors"`          // Continue on errors
	DryRun              bool          `json:"dry_run"`              // Preview mode
	ConfirmEach         bool          `json:"confirm_each"`         // Confirm each operation
	ProgressInterval    time.Duration `json:"progress_interval"`    // Progress update interval
	RetryCount          int           `json:"retry_count"`          // Number of retries on failure
	RetryDelay          time.Duration `json:"retry_delay"`          // Delay between retries
	PreserveTimestamps  bool          `json:"preserve_timestamps"`  // Preserve file timestamps
	PreservePermissions bool          `json:"preserve_permissions"` // Preserve file permissions
	FollowSymlinks      bool          `json:"follow_symlinks"`      // Follow symbolic links
	ExcludePatterns     []string      `json:"exclude_patterns"`     // Patterns to exclude
	IncludePatterns     []string      `json:"include_patterns"`     // Patterns to include
}

// BatchManager manages batch operations and provides the main interface
type BatchManager struct {
	jobs           map[string]*BatchJob
	mutex          sync.RWMutex
	progressBars   map[string]*progress.ProgressBar
	analytics      *suggestions.UsageAnalytics
	defaultConfig  BatchConfig
	eventCallbacks map[string][]BatchEventCallback
}

// BatchEventCallback is a function that gets called on batch events
type BatchEventCallback func(event BatchEvent)

// BatchEvent represents an event that occurred during batch processing
type BatchEvent struct {
	Type      string                 `json:"type"` // job_started, job_completed, operation_started, operation_completed, error, progress
	JobID     string                 `json:"job_id"`
	Operation *BatchOperation        `json:"operation,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewBatchManager creates a new batch manager instance
func NewBatchManager(analytics *suggestions.UsageAnalytics) *BatchManager {
	return &BatchManager{
		jobs:         make(map[string]*BatchJob),
		progressBars: make(map[string]*progress.ProgressBar),
		analytics:    analytics,
		defaultConfig: BatchConfig{
			MaxConcurrency:      4,
			SkipErrors:          true,
			DryRun:              false,
			ConfirmEach:         false,
			ProgressInterval:    500 * time.Millisecond,
			RetryCount:          2,
			RetryDelay:          1 * time.Second,
			PreserveTimestamps:  true,
			PreservePermissions: true,
			FollowSymlinks:      false,
			ExcludePatterns:     []string{},
			IncludePatterns:     []string{},
		},
		eventCallbacks: make(map[string][]BatchEventCallback),
	}
}

// CreateBatchJob creates a new batch job
func (bm *BatchManager) CreateBatchJob(name, description string, operations []BatchOperation, config BatchConfig) *BatchJob {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	jobID := fmt.Sprintf("batch_%d", time.Now().UnixNano())

	// Merge with default config
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = bm.defaultConfig.MaxConcurrency
	}
	if config.ProgressInterval == 0 {
		config.ProgressInterval = bm.defaultConfig.ProgressInterval
	}
	if config.RetryCount == 0 {
		config.RetryCount = bm.defaultConfig.RetryCount
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = bm.defaultConfig.RetryDelay
	}

	job := &BatchJob{
		ID:            jobID,
		Name:          name,
		Description:   description,
		Operations:    operations,
		Status:        "pending",
		Progress:      0.0,
		TotalSize:     bm.calculateTotalSize(operations),
		ProcessedSize: 0,
		ErrorCount:    0,
		SuccessCount:  0,
		SkippedCount:  0,
		Config:        config,
		Metadata:      make(map[string]interface{}),
	}

	bm.jobs[jobID] = job
	return job
}

// ExecuteBatchJob executes a batch job with progress tracking
func (bm *BatchManager) ExecuteBatchJob(jobID string) error {
	bm.mutex.Lock()
	job, exists := bm.jobs[jobID]
	if !exists {
		bm.mutex.Unlock()
		return fmt.Errorf("batch job %s not found", jobID)
	}
	bm.mutex.Unlock()

	// Create progress bar
	pb := progress.NewProgressBar(int64(len(job.Operations)), &progress.ProgressBarConfig{
		RefreshRate: job.Config.ProgressInterval,
		ShowETA:     true,
		ShowSpeed:   true,
	})

	bm.mutex.Lock()
	bm.progressBars[jobID] = pb
	bm.mutex.Unlock()

	// Start job
	job.Status = "running"
	job.StartTime = time.Now()
	bm.triggerEvent(BatchEvent{
		Type:      "job_started",
		JobID:     jobID,
		Message:   fmt.Sprintf("Started batch job: %s", job.Name),
		Timestamp: time.Now(),
	})

	// Execute operations
	err := bm.executeOperations(job, pb)

	// Complete job
	job.Status = "completed"
	if err != nil {
		job.Status = "failed"
	}
	job.EndTime = time.Now()
	job.Duration = job.EndTime.Sub(job.StartTime)
	job.Progress = 1.0

	bm.triggerEvent(BatchEvent{
		Type:    "job_completed",
		JobID:   jobID,
		Message: fmt.Sprintf("Completed batch job: %s", job.Name),
		Data: map[string]interface{}{
			"success_count": job.SuccessCount,
			"error_count":   job.ErrorCount,
			"skipped_count": job.SkippedCount,
			"duration":      job.Duration.String(),
		},
		Timestamp: time.Now(),
	})

	// Clean up progress bar
	bm.mutex.Lock()
	delete(bm.progressBars, jobID)
	bm.mutex.Unlock()

	pb.Finish()

	return err
}

// BatchDelete creates a batch delete job for multiple files/folders
func (bm *BatchManager) BatchDelete(paths []string, config BatchConfig) (*BatchJob, error) {
	var operations []BatchOperation

	for _, path := range paths {
		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip non-existent paths
			}
			return nil, fmt.Errorf("error checking path %s: %v", path, err)
		}

		// Create delete operation
		operation := BatchOperation{
			ID:     fmt.Sprintf("delete_%d", time.Now().UnixNano()),
			Type:   "delete",
			Source: path,
			Size:   info.Size(),
			Status: "pending",
			Metadata: map[string]interface{}{
				"is_directory": info.IsDir(),
				"permissions":  info.Mode(),
				"mod_time":     info.ModTime(),
			},
		}

		operations = append(operations, operation)
	}

	if len(operations) == 0 {
		return nil, fmt.Errorf("no valid paths to delete")
	}

	job := bm.CreateBatchJob(
		fmt.Sprintf("Delete %d items", len(operations)),
		fmt.Sprintf("Delete %d files/folders", len(operations)),
		operations,
		config,
	)

	return job, nil
}

// BatchCopy creates a batch copy job for recursive copying
func (bm *BatchManager) BatchCopy(sourcePaths []string, destination string, config BatchConfig) (*BatchJob, error) {
	var operations []BatchOperation

	// Ensure destination exists
	if err := os.MkdirAll(destination, 0755); err != nil {
		return nil, fmt.Errorf("error creating destination directory: %v", err)
	}

	for _, sourcePath := range sourcePaths {
		// Walk through source path recursively
		err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip if excluded
			if bm.shouldExclude(path, config.ExcludePatterns) {
				return nil
			}

			// Calculate relative path
			relPath, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}

			destPath := filepath.Join(destination, relPath)

			// Create copy operation
			operation := BatchOperation{
				ID:          fmt.Sprintf("copy_%d", time.Now().UnixNano()),
				Type:        "copy",
				Source:      path,
				Destination: destPath,
				Size:        info.Size(),
				Status:      "pending",
				Metadata: map[string]interface{}{
					"is_directory":  info.IsDir(),
					"permissions":   info.Mode(),
					"mod_time":      info.ModTime(),
					"relative_path": relPath,
				},
			}

			operations = append(operations, operation)
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking source path %s: %v", sourcePath, err)
		}
	}

	if len(operations) == 0 {
		return nil, fmt.Errorf("no files to copy")
	}

	job := bm.CreateBatchJob(
		fmt.Sprintf("Copy %d items", len(operations)),
		fmt.Sprintf("Copy %d files/folders to %s", len(operations), destination),
		operations,
		config,
	)

	return job, nil
}

// BatchMove creates a batch move job
func (bm *BatchManager) BatchMove(sourcePaths []string, destination string, config BatchConfig) (*BatchJob, error) {
	var operations []BatchOperation

	// Ensure destination exists
	if err := os.MkdirAll(destination, 0755); err != nil {
		return nil, fmt.Errorf("error creating destination directory: %v", err)
	}

	for _, sourcePath := range sourcePaths {
		info, err := os.Stat(sourcePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("error checking source path %s: %v", sourcePath, err)
		}

		// Calculate destination path
		destPath := filepath.Join(destination, filepath.Base(sourcePath))

		// Create move operation
		operation := BatchOperation{
			ID:          fmt.Sprintf("move_%d", time.Now().UnixNano()),
			Type:        "move",
			Source:      sourcePath,
			Destination: destPath,
			Size:        info.Size(),
			Status:      "pending",
			Metadata: map[string]interface{}{
				"is_directory": info.IsDir(),
				"permissions":  info.Mode(),
				"mod_time":     info.ModTime(),
			},
		}

		operations = append(operations, operation)
	}

	if len(operations) == 0 {
		return nil, fmt.Errorf("no valid paths to move")
	}

	job := bm.CreateBatchJob(
		fmt.Sprintf("Move %d items", len(operations)),
		fmt.Sprintf("Move %d files/folders to %s", len(operations), destination),
		operations,
		config,
	)

	return job, nil
}

// GetJobStatus returns the status of a batch job
func (bm *BatchManager) GetJobStatus(jobID string) (*BatchJob, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	job, exists := bm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("batch job %s not found", jobID)
	}

	return job, nil
}

// ListJobs returns all batch jobs
func (bm *BatchManager) ListJobs() []*BatchJob {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	var jobs []*BatchJob
	for _, job := range bm.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// CancelJob cancels a running batch job
func (bm *BatchManager) CancelJob(jobID string) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	job, exists := bm.jobs[jobID]
	if !exists {
		return fmt.Errorf("batch job %s not found", jobID)
	}

	if job.Status != "running" {
		return fmt.Errorf("job %s is not running", jobID)
	}

	job.Status = "cancelled"
	job.EndTime = time.Now()
	job.Duration = job.EndTime.Sub(job.StartTime)

	// Cancel progress bar
	if pb, exists := bm.progressBars[jobID]; exists {
		pb.Finish()
		delete(bm.progressBars, jobID)
	}

	bm.triggerEvent(BatchEvent{
		Type:      "job_cancelled",
		JobID:     jobID,
		Message:   fmt.Sprintf("Cancelled batch job: %s", job.Name),
		Timestamp: time.Now(),
	})

	return nil
}

// Private helper methods

func (bm *BatchManager) executeOperations(job *BatchJob, pb *progress.ProgressBar) error {
	semaphore := make(chan struct{}, job.Config.MaxConcurrency)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for i, operation := range job.Operations {
		wg.Add(1)
		go func(op BatchOperation, index int) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			bm.executeOperation(job, &op, index, pb)

			mutex.Lock()
			job.Operations[index] = op
			mutex.Unlock()
		}(operation, i)
	}

	wg.Wait()

	// Update final progress
	pb.Update(int64(len(job.Operations)))

	return nil
}

func (bm *BatchManager) executeOperation(job *BatchJob, operation *BatchOperation, index int, pb *progress.ProgressBar) {
	operation.Status = "running"
	operation.StartTime = time.Now()

	bm.triggerEvent(BatchEvent{
		Type:      "operation_started",
		JobID:     job.ID,
		Operation: operation,
		Message:   fmt.Sprintf("Started %s: %s", operation.Type, operation.Source),
		Timestamp: time.Now(),
	})

	var err error

	if job.Config.DryRun {
		// Dry run - just simulate
		time.Sleep(100 * time.Millisecond) // Simulate work
		operation.Status = "completed"
	} else {
		// Execute actual operation
		switch operation.Type {
		case "delete":
			err = bm.executeDelete(operation)
		case "copy":
			err = bm.executeCopy(operation, job.Config)
		case "move":
			err = bm.executeMove(operation, job.Config)
		default:
			err = fmt.Errorf("unknown operation type: %s", operation.Type)
		}
	}

	operation.EndTime = time.Now()
	operation.Duration = operation.EndTime.Sub(operation.StartTime)

	if err != nil {
		operation.Status = "failed"
		operation.Error = err.Error()
		job.ErrorCount++
	} else {
		operation.Status = "completed"
		job.SuccessCount++
	}

	// Update progress
	job.ProcessedSize += operation.Size
	job.Progress = float64(job.ProcessedSize) / float64(job.TotalSize)
	pb.Update(int64(job.SuccessCount + job.ErrorCount))

	bm.triggerEvent(BatchEvent{
		Type:      "operation_completed",
		JobID:     job.ID,
		Operation: operation,
		Message:   fmt.Sprintf("Completed %s: %s", operation.Type, operation.Source),
		Timestamp: time.Now(),
	})
}

func (bm *BatchManager) executeDelete(operation *BatchOperation) error {
	return os.RemoveAll(operation.Source)
}

func (bm *BatchManager) executeCopy(operation *BatchOperation, config BatchConfig) error {
	sourceInfo := operation.Metadata["is_directory"].(bool)

	if sourceInfo {
		// Copy directory
		return bm.copyDirectory(operation.Source, operation.Destination, config)
	} else {
		// Copy file
		return bm.copyFile(operation.Source, operation.Destination, config)
	}
}

func (bm *BatchManager) executeMove(operation *BatchOperation, config BatchConfig) error {
	// Try rename first (fastest for same filesystem)
	err := os.Rename(operation.Source, operation.Destination)
	if err == nil {
		return nil
	}

	// Fallback to copy + delete
	err = bm.executeCopy(operation, config)
	if err != nil {
		return err
	}

	return os.RemoveAll(operation.Source)
}

func (bm *BatchManager) copyFile(src, dst string, config BatchConfig) error {
	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Preserve permissions and timestamps
	if config.PreservePermissions {
		sourceInfo, err := os.Stat(src)
		if err == nil {
			os.Chmod(dst, sourceInfo.Mode())
		}
	}

	if config.PreserveTimestamps {
		sourceInfo, err := os.Stat(src)
		if err == nil {
			os.Chtimes(dst, sourceInfo.ModTime(), sourceInfo.ModTime())
		}
	}

	return nil
}

func (bm *BatchManager) copyDirectory(src, dst string, config BatchConfig) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			return bm.copyFile(path, dstPath, config)
		}
	})
}

func (bm *BatchManager) calculateTotalSize(operations []BatchOperation) int64 {
	var total int64
	for _, op := range operations {
		total += op.Size
	}
	return total
}

func (bm *BatchManager) shouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}

func (bm *BatchManager) triggerEvent(event BatchEvent) {
	bm.mutex.RLock()
	callbacks := bm.eventCallbacks[event.Type]
	bm.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(event) // Run callbacks asynchronously
	}
}

// AddEventCallback adds a callback for batch events
func (bm *BatchManager) AddEventCallback(eventType string, callback BatchEventCallback) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	bm.eventCallbacks[eventType] = append(bm.eventCallbacks[eventType], callback)
}
