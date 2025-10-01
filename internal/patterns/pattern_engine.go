/**
 * Pattern-based file operations engine for file matching and filtering.
 *
 * Provides intelligent pattern matching, file filtering, and batch operations
 * based on complex criteria including age, size, content, and metadata.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: pattern_engine.go
 * Description: pattern matching and file operation engine
 */

package patterns

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ena/internal/suggestions"
)

// PatternType defines the type of pattern matching
type PatternType string

const (
	PatternFileExtension PatternType = "extension"
	PatternFileName      PatternType = "filename"
	PatternFilePath      PatternType = "filepath"
	PatternContent       PatternType = "content"
	PatternSize          PatternType = "size"
	PatternAge           PatternType = "age"
	PatternPermissions   PatternType = "permissions"
	PatternOwner         PatternType = "owner"
	PatternGroup         PatternType = "group"
	PatternRegex         PatternType = "regex"
	PatternMimeType      PatternType = "mimetype"
)

// ComparisonOperator defines how to compare values
type ComparisonOperator string

const (
	OpEquals       ComparisonOperator = "equals"
	OpNotEquals    ComparisonOperator = "not_equals"
	OpGreaterThan  ComparisonOperator = "greater_than"
	OpLessThan     ComparisonOperator = "less_than"
	OpGreaterEqual ComparisonOperator = "greater_equal"
	OpLessEqual    ComparisonOperator = "less_equal"
	OpContains     ComparisonOperator = "contains"
	OpNotContains  ComparisonOperator = "not_contains"
	OpStartsWith   ComparisonOperator = "starts_with"
	OpEndsWith     ComparisonOperator = "ends_with"
	OpMatches      ComparisonOperator = "matches"
	OpNotMatches   ComparisonOperator = "not_matches"
	OpIn           ComparisonOperator = "in"
	OpNotIn        ComparisonOperator = "not_in"
)

// FileFilter represents a single filter condition
type FileFilter struct {
	Type          PatternType        `json:"type"`
	Operator      ComparisonOperator `json:"operator"`
	Value         interface{}        `json:"value"`
	CaseSensitive bool               `json:"case_sensitive"`
	Negate        bool               `json:"negate"`
}

// PatternOperation represents a complete pattern-based operation
type PatternOperation struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Enabled     bool         `json:"enabled"`
	Priority    int          `json:"priority"`
	Filters     []FileFilter `json:"filters"`
	Paths       []string     `json:"paths"`
	Recursive   bool         `json:"recursive"`
	MaxDepth    int          `json:"max_depth"`
	Actions     []Action     `json:"actions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	LastRun     *time.Time   `json:"last_run,omitempty"`
}

// Action defines what to do with matched files
type Action struct {
	Type        string            `json:"type"`
	Destination string            `json:"destination,omitempty"`
	Template    string            `json:"template,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

// PatternResult represents the result of a pattern operation
type PatternResult struct {
	OperationID    string                `json:"operation_id"`
	FilesMatched   int                   `json:"files_matched"`
	FilesProcessed int                   `json:"files_processed"`
	FilesSkipped   int                   `json:"files_skipped"`
	FilesFailed    int                   `json:"files_failed"`
	Errors         []string              `json:"errors"`
	Duration       time.Duration         `json:"duration"`
	Details        []FileOperationDetail `json:"details"`
	Summary        PatternSummary        `json:"summary"`
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
	MatchedRules []string  `json:"matched_rules"`
}

// PatternSummary provides a summary of the pattern operation
type PatternSummary struct {
	TotalSize        int64            `json:"total_size"`
	AverageSize      int64            `json:"average_size"`
	OldestFile       string           `json:"oldest_file"`
	NewestFile       string           `json:"newest_file"`
	FileTypes        map[string]int   `json:"file_types"`
	SizeDistribution map[string]int64 `json:"size_distribution"`
}

// PatternEngine manages pattern-based file operations
type PatternEngine struct {
	operations     map[string]*PatternOperation
	analytics      *suggestions.UsageAnalytics
	mutex          sync.RWMutex
	configFile     string
	resultsFile    string
	eventCallbacks map[string][]PatternEventCallback
	isRunning      bool
	stopChan       chan struct{}
}

// PatternEventCallback is a function that gets called on pattern events
type PatternEventCallback func(event PatternEvent)

// PatternEvent represents an event that occurred during pattern operations
type PatternEvent struct {
	Type        string                 `json:"type"`
	OperationID string                 `json:"operation_id,omitempty"`
	FilePath    string                 `json:"file_path,omitempty"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewPatternEngine creates a new pattern engine instance
func NewPatternEngine(analytics *suggestions.UsageAnalytics) *PatternEngine {
	pe := &PatternEngine{
		operations:     make(map[string]*PatternOperation),
		analytics:      analytics,
		configFile:     "pattern_operations.json",
		resultsFile:    "pattern_results.json",
		eventCallbacks: make(map[string][]PatternEventCallback),
		stopChan:       make(chan struct{}),
	}

	// Load existing operations
	pe.loadOperations()

	return pe
}

// AddOperation adds a new pattern operation
func (pe *PatternEngine) AddOperation(operation *PatternOperation) error {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	if operation.ID == "" {
		operation.ID = fmt.Sprintf("pattern_%d", time.Now().UnixNano())
	}

	operation.CreatedAt = time.Now()
	operation.UpdatedAt = time.Now()

	pe.operations[operation.ID] = operation

	pe.triggerEvent(PatternEvent{
		Type:        "operation_added",
		OperationID: operation.ID,
		Message:     fmt.Sprintf("Added pattern operation: %s", operation.Name),
		Timestamp:   time.Now(),
	})

	return pe.saveOperations()
}

// RemoveOperation removes a pattern operation
func (pe *PatternEngine) RemoveOperation(operationID string) error {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	if _, exists := pe.operations[operationID]; !exists {
		return fmt.Errorf("operation %s not found", operationID)
	}

	delete(pe.operations, operationID)

	pe.triggerEvent(PatternEvent{
		Type:        "operation_removed",
		OperationID: operationID,
		Message:     fmt.Sprintf("Removed pattern operation: %s", operationID),
		Timestamp:   time.Now(),
	})

	return pe.saveOperations()
}

// GetOperations returns all pattern operations
func (pe *PatternEngine) GetOperations() []*PatternOperation {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	var operations []*PatternOperation
	for _, operation := range pe.operations {
		operations = append(operations, operation)
	}

	// Sort by priority (higher priority first)
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].Priority > operations[j].Priority
	})

	return operations
}

// ExecuteOperation executes a pattern operation
func (pe *PatternEngine) ExecuteOperation(operationID string, dryRun bool) (*PatternResult, error) {
	pe.mutex.RLock()
	operation, exists := pe.operations[operationID]
	pe.mutex.RUnlock()

	// For temporary operations, create a new one
	if !exists && strings.HasPrefix(operationID, "temp_") {
		operation = pe.CreateSampleOperation()
		operation.ID = operationID
	}

	if !exists && !strings.HasPrefix(operationID, "temp_") {
		return nil, fmt.Errorf("operation %s not found", operationID)
	}

	if !operation.Enabled {
		return nil, fmt.Errorf("operation %s is disabled", operationID)
	}

	result := &PatternResult{
		OperationID: operationID,
		Summary: PatternSummary{
			FileTypes:        make(map[string]int),
			SizeDistribution: make(map[string]int64),
		},
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime)
		pe.saveResult(result)
	}()

	pe.triggerEvent(PatternEvent{
		Type:        "operation_started",
		OperationID: operationID,
		Message:     fmt.Sprintf("Started pattern operation: %s", operation.Name),
		Timestamp:   time.Now(),
	})

	// Collect files to process
	var filesToProcess []string
	for _, path := range operation.Paths {
		err := pe.collectFiles(path, operation, &filesToProcess, 0)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("error collecting files from %s: %v", path, err))
		}
	}

	// Filter files based on pattern criteria
	var matchedFiles []string
	for _, filePath := range filesToProcess {
		if pe.fileMatchesFilters(filePath, operation.Filters) {
			matchedFiles = append(matchedFiles, filePath)
		}
	}

	result.FilesMatched = len(matchedFiles)

	// Process matched files
	for _, filePath := range matchedFiles {
		detail, err := pe.processFile(operation, filePath, dryRun)
		if err != nil {
			detail.Error = err.Error()
			result.Errors = append(result.Errors, err.Error())
			result.FilesFailed++
		} else if detail.Success {
			result.FilesProcessed++
			pe.updateSummary(&result.Summary, detail)
		} else {
			result.FilesSkipped++
		}

		result.Details = append(result.Details, detail)
	}

	// Update operation last run time
	pe.mutex.Lock()
	operation.LastRun = &startTime
	pe.mutex.Unlock()

	pe.triggerEvent(PatternEvent{
		Type:        "operation_completed",
		OperationID: operationID,
		Message:     fmt.Sprintf("Completed pattern operation: %s (%d files processed)", operation.Name, result.FilesProcessed),
		Data: map[string]interface{}{
			"files_matched":   result.FilesMatched,
			"files_processed": result.FilesProcessed,
			"files_failed":    result.FilesFailed,
			"duration":        result.Duration.String(),
		},
		Timestamp: time.Now(),
	})

	return result, nil
}

// ExecuteAllOperations executes all enabled pattern operations
func (pe *PatternEngine) ExecuteAllOperations(dryRun bool) ([]PatternResult, error) {
	operations := pe.GetOperations()
	var results []PatternResult

	for _, operation := range operations {
		if operation.Enabled {
			result, err := pe.ExecuteOperation(operation.ID, dryRun)
			if err != nil {
				// Log error but continue with other operations
				results = append(results, PatternResult{
					OperationID: operation.ID,
					Errors:      []string{err.Error()},
				})
				continue
			}
			results = append(results, *result)
		}
	}

	return results, nil
}

// Private helper methods

func (pe *PatternEngine) collectFiles(path string, operation *PatternOperation, files *[]string, depth int) error {
	// Check max depth
	if operation.MaxDepth > 0 && depth >= operation.MaxDepth {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if !operation.Recursive && depth > 0 {
			return nil
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			if entry.IsDir() {
				if operation.Recursive {
					pe.collectFiles(entryPath, operation, files, depth+1)
				}
			} else {
				*files = append(*files, entryPath)
			}
		}
	} else {
		*files = append(*files, path)
	}

	return nil
}

func (pe *PatternEngine) fileMatchesFilters(filePath string, filters []FileFilter) bool {
	for _, filter := range filters {
		matches := pe.evaluateFilter(filePath, filter)
		if filter.Negate {
			matches = !matches
		}
		if !matches {
			return false
		}
	}
	return true
}

func (pe *PatternEngine) evaluateFilter(filePath string, filter FileFilter) bool {
	switch filter.Type {
	case PatternFileExtension:
		return pe.matchExtension(filePath, filter)
	case PatternFileName:
		return pe.matchFileName(filePath, filter)
	case PatternFilePath:
		return pe.matchFilePath(filePath, filter)
	case PatternContent:
		return pe.matchContent(filePath, filter)
	case PatternSize:
		return pe.matchSize(filePath, filter)
	case PatternAge:
		return pe.matchAge(filePath, filter)
	case PatternPermissions:
		return pe.matchPermissions(filePath, filter)
	case PatternRegex:
		return pe.matchRegex(filePath, filter)
	default:
		return false
	}
}

func (pe *PatternEngine) matchExtension(filePath string, filter FileFilter) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	value := strings.ToLower(fmt.Sprintf("%v", filter.Value))

	if !filter.CaseSensitive {
		ext = strings.ToLower(ext)
	}

	switch filter.Operator {
	case OpEquals:
		return ext == value
	case OpNotEquals:
		return ext != value
	case OpIn:
		if values, ok := filter.Value.([]string); ok {
			for _, v := range values {
				if strings.ToLower(v) == ext {
					return true
				}
			}
		}
		return false
	case OpNotIn:
		if values, ok := filter.Value.([]string); ok {
			for _, v := range values {
				if strings.ToLower(v) == ext {
					return false
				}
			}
		}
		return true
	default:
		return false
	}
}

func (pe *PatternEngine) matchFileName(filePath string, filter FileFilter) bool {
	fileName := filepath.Base(filePath)
	value := fmt.Sprintf("%v", filter.Value)

	if !filter.CaseSensitive {
		fileName = strings.ToLower(fileName)
		value = strings.ToLower(value)
	}

	switch filter.Operator {
	case OpEquals:
		return fileName == value
	case OpNotEquals:
		return fileName != value
	case OpContains:
		return strings.Contains(fileName, value)
	case OpNotContains:
		return !strings.Contains(fileName, value)
	case OpStartsWith:
		return strings.HasPrefix(fileName, value)
	case OpEndsWith:
		return strings.HasSuffix(fileName, value)
	default:
		return false
	}
}

func (pe *PatternEngine) matchFilePath(filePath string, filter FileFilter) bool {
	value := fmt.Sprintf("%v", filter.Value)

	if !filter.CaseSensitive {
		filePath = strings.ToLower(filePath)
		value = strings.ToLower(value)
	}

	switch filter.Operator {
	case OpContains:
		return strings.Contains(filePath, value)
	case OpNotContains:
		return !strings.Contains(filePath, value)
	case OpStartsWith:
		return strings.HasPrefix(filePath, value)
	case OpEndsWith:
		return strings.HasSuffix(filePath, value)
	default:
		return false
	}
}

func (pe *PatternEngine) matchContent(filePath string, filter FileFilter) bool {
	value := fmt.Sprintf("%v", filter.Value)

	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 1KB for content matching
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	content := string(buffer[:n])
	if !filter.CaseSensitive {
		content = strings.ToLower(content)
		value = strings.ToLower(value)
	}

	switch filter.Operator {
	case OpContains:
		return strings.Contains(content, value)
	case OpNotContains:
		return !strings.Contains(content, value)
	default:
		return false
	}
}

func (pe *PatternEngine) matchSize(filePath string, filter FileFilter) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	fileSize := info.Size()
	value, ok := filter.Value.(float64)
	if !ok {
		if str, ok := filter.Value.(string); ok {
			if parsed, err := strconv.ParseInt(str, 10, 64); err == nil {
				value = float64(parsed)
			} else {
				return false
			}
		} else {
			return false
		}
	}

	switch filter.Operator {
	case OpEquals:
		return fileSize == int64(value)
	case OpNotEquals:
		return fileSize != int64(value)
	case OpGreaterThan:
		return fileSize > int64(value)
	case OpLessThan:
		return fileSize < int64(value)
	case OpGreaterEqual:
		return fileSize >= int64(value)
	case OpLessEqual:
		return fileSize <= int64(value)
	default:
		return false
	}
}

func (pe *PatternEngine) matchAge(filePath string, filter FileFilter) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	fileAge := time.Since(info.ModTime())
	value := fmt.Sprintf("%v", filter.Value)

	// Parse age string (e.g., "30d", "1h", "7d")
	duration, err := pe.parseAge(value)
	if err != nil {
		return false
	}

	switch filter.Operator {
	case OpGreaterThan:
		return fileAge > duration
	case OpLessThan:
		return fileAge < duration
	case OpGreaterEqual:
		return fileAge >= duration
	case OpLessEqual:
		return fileAge <= duration
	default:
		return false
	}
}

func (pe *PatternEngine) matchPermissions(filePath string, filter FileFilter) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	perm := info.Mode().Perm()
	value := fmt.Sprintf("%v", filter.Value)

	// Parse permission string (e.g., "644", "755")
	var expectedPerm os.FileMode
	if parsed, err := strconv.ParseInt(value, 8, 32); err == nil {
		expectedPerm = os.FileMode(parsed)
	} else {
		return false
	}

	switch filter.Operator {
	case OpEquals:
		return perm == expectedPerm
	case OpNotEquals:
		return perm != expectedPerm
	default:
		return false
	}
}

func (pe *PatternEngine) matchRegex(filePath string, filter FileFilter) bool {
	value := fmt.Sprintf("%v", filter.Value)

	regex, err := regexp.Compile(value)
	if err != nil {
		return false
	}

	fileName := filepath.Base(filePath)
	if !filter.CaseSensitive {
		fileName = strings.ToLower(fileName)
	}

	switch filter.Operator {
	case OpMatches:
		return regex.MatchString(fileName)
	case OpNotMatches:
		return !regex.MatchString(fileName)
	default:
		return false
	}
}

func (pe *PatternEngine) parseAge(ageStr string) (time.Duration, error) {
	ageStr = strings.TrimSpace(ageStr)

	// Parse common age formats
	if strings.HasSuffix(ageStr, "s") {
		if seconds, err := strconv.Atoi(strings.TrimSuffix(ageStr, "s")); err == nil {
			return time.Duration(seconds) * time.Second, nil
		}
	} else if strings.HasSuffix(ageStr, "m") {
		if minutes, err := strconv.Atoi(strings.TrimSuffix(ageStr, "m")); err == nil {
			return time.Duration(minutes) * time.Minute, nil
		}
	} else if strings.HasSuffix(ageStr, "h") {
		if hours, err := strconv.Atoi(strings.TrimSuffix(ageStr, "h")); err == nil {
			return time.Duration(hours) * time.Hour, nil
		}
	} else if strings.HasSuffix(ageStr, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(ageStr, "d")); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	} else if strings.HasSuffix(ageStr, "w") {
		if weeks, err := strconv.Atoi(strings.TrimSuffix(ageStr, "w")); err == nil {
			return time.Duration(weeks) * 7 * 24 * time.Hour, nil
		}
	}

	return 0, fmt.Errorf("invalid age format: %s", ageStr)
}

func (pe *PatternEngine) processFile(operation *PatternOperation, filePath string, dryRun bool) (FileOperationDetail, error) {
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
	for _, action := range operation.Actions {
		switch action.Type {
		case "move":
			if !dryRun {
				destPath := pe.buildDestinationPath(filePath, action.Destination)
				err = pe.moveFile(filePath, destPath)
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
				destPath := pe.buildDestinationPath(filePath, action.Destination)
				err = pe.copyFile(filePath, destPath)
				if err != nil {
					detail.Error = err.Error()
					return detail, err
				}
				detail.Destination = destPath
			}
			detail.Action = "copy"
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

		case "rename":
			if !dryRun {
				newName := pe.buildFileName(filePath, action.Template)
				err = pe.renameFile(filePath, newName)
				if err != nil {
					detail.Error = err.Error()
					return detail, err
				}
				detail.Destination = newName
			}
			detail.Action = "rename"
			detail.Success = true
		}
	}

	return detail, nil
}

func (pe *PatternEngine) buildDestinationPath(filePath, template string) string {
	if template == "" {
		return filePath
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
	destPath = strings.ReplaceAll(destPath, "{day}", time.Now().Format("02"))

	return destPath
}

func (pe *PatternEngine) buildFileName(filePath, template string) string {
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

func (pe *PatternEngine) moveFile(src, dest string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return os.Rename(src, dest)
}

func (pe *PatternEngine) copyFile(src, dest string) error {
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

func (pe *PatternEngine) renameFile(src, newName string) error {
	dest := filepath.Join(filepath.Dir(src), newName)
	return os.Rename(src, dest)
}

func (pe *PatternEngine) updateSummary(summary *PatternSummary, detail FileOperationDetail) {
	summary.TotalSize += detail.Size

	// Update file type distribution
	ext := strings.ToLower(filepath.Ext(detail.FilePath))
	if ext == "" {
		ext = "no_extension"
	}
	summary.FileTypes[ext]++

	// Update size distribution
	if detail.Size < 1024 {
		summary.SizeDistribution["< 1KB"]++
	} else if detail.Size < 1024*1024 {
		summary.SizeDistribution["< 1MB"]++
	} else if detail.Size < 1024*1024*1024 {
		summary.SizeDistribution["< 1GB"]++
	} else {
		summary.SizeDistribution[">= 1GB"]++
	}

	// Update oldest/newest files
	if summary.OldestFile == "" || detail.ModifiedTime.Before(time.Time{}) {
		summary.OldestFile = detail.FilePath
	}
	if summary.NewestFile == "" || detail.ModifiedTime.After(time.Time{}) {
		summary.NewestFile = detail.FilePath
	}
}

func (pe *PatternEngine) loadOperations() error {
	if _, err := os.Stat(pe.configFile); os.IsNotExist(err) {
		return nil // No config file exists yet
	}

	data, err := os.ReadFile(pe.configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	var operations []*PatternOperation
	if err := json.Unmarshal(data, &operations); err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	for _, operation := range operations {
		pe.operations[operation.ID] = operation
	}

	return nil
}

func (pe *PatternEngine) saveOperations() error {
	var operations []*PatternOperation
	for _, operation := range pe.operations {
		operations = append(operations, operation)
	}

	data, err := json.MarshalIndent(operations, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling operations: %v", err)
	}

	if err := os.WriteFile(pe.configFile, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

func (pe *PatternEngine) saveResult(result *PatternResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling result: %v", err)
	}

	resultFile := fmt.Sprintf("pattern_result_%s_%d.json", result.OperationID, time.Now().Unix())
	if err := os.WriteFile(resultFile, data, 0644); err != nil {
		return fmt.Errorf("error writing result file: %v", err)
	}

	return nil
}

func (pe *PatternEngine) triggerEvent(event PatternEvent) {
	pe.mutex.RLock()
	callbacks := pe.eventCallbacks[event.Type]
	pe.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(event) // Run callbacks asynchronously
	}
}

// AddEventCallback adds a callback for pattern events
func (pe *PatternEngine) AddEventCallback(eventType string, callback PatternEventCallback) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	pe.eventCallbacks[eventType] = append(pe.eventCallbacks[eventType], callback)
}

// GetOperationByID returns a specific operation by ID
func (pe *PatternEngine) GetOperationByID(operationID string) (*PatternOperation, error) {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	operation, exists := pe.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", operationID)
	}

	return operation, nil
}

// CreateSampleOperation creates a sample pattern operation for testing
func (pe *PatternEngine) CreateSampleOperation() *PatternOperation {
	return &PatternOperation{
		Name:        "Find Old Text Files",
		Description: "Find all .txt files older than 30 days",
		Enabled:     true,
		Priority:    10,
		Filters: []FileFilter{
			{
				Type:     PatternFileExtension,
				Operator: OpEquals,
				Value:    ".txt",
			},
			{
				Type:     PatternAge,
				Operator: OpGreaterThan,
				Value:    "30d",
			},
		},
		Paths:     []string{"."},
		Recursive: true,
		MaxDepth:  5,
		Actions: []Action{
			{
				Type:        "move",
				Destination: "~/Archive/old_text_files/{date}",
			},
		},
	}
}
