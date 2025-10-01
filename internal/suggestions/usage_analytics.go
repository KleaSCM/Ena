/**
 * Usage analytics and pattern tracking system.
 *
 * Tracks user command patterns, file operations, and system interactions
 * to provide intelligent suggestions and improve user experience.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: usage_analytics.go
 * Description: Core analytics engine for tracking and analyzing user behavior patterns
 */

package suggestions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// CommandUsage represents a single command execution
type CommandUsage struct {
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	Timestamp  time.Time         `json:"timestamp"`
	Duration   time.Duration     `json:"duration"`
	Success    bool              `json:"success"`
	Context    map[string]string `json:"context"`
	WorkingDir string            `json:"working_dir"`
	FileCount  int               `json:"file_count"`
	ErrorMsg   string            `json:"error_msg,omitempty"`
}

// FileOperation represents file system operations
type FileOperation struct {
	Operation string    `json:"operation"` // create, read, update, delete, copy, move
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	FileType  string    `json:"file_type"`
	Directory string    `json:"directory"`
}

// UsagePattern represents discovered patterns in user behavior
type UsagePattern struct {
	PatternType string                 `json:"pattern_type"` // command_sequence, file_operation, time_based, frequency
	Description string                 `json:"description"`
	Frequency   int                    `json:"frequency"`
	Confidence  float64                `json:"confidence"` // 0.0 to 1.0
	LastSeen    time.Time              `json:"last_seen"`
	Data        map[string]interface{} `json:"data"`
	Suggestions []string               `json:"suggestions"`
}

// SmartSuggestion represents an intelligent suggestion for the user
type SmartSuggestion struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // command, file_op, workflow, optimization
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Command      string                 `json:"command,omitempty"`
	Confidence   float64                `json:"confidence"`
	Priority     int                    `json:"priority"` // 1-10, higher is more important
	Category     string                 `json:"category"` // productivity, safety, optimization, discovery
	Context      map[string]interface{} `json:"context"`
	CreatedAt    time.Time              `json:"created_at"`
	LastShown    time.Time              `json:"last_shown,omitempty"`
	TimesShown   int                    `json:"times_shown"`
	UserFeedback string                 `json:"user_feedback,omitempty"` // helpful, not_helpful, dismiss
}

// UsageAnalytics manages all usage tracking and analytics
type UsageAnalytics struct {
	commandHistory   []CommandUsage
	fileOperations   []FileOperation
	patterns         []UsagePattern
	suggestions      []SmartSuggestion
	mutex            sync.RWMutex
	dataFile         string
	maxHistorySize   int
	patternThreshold float64
	suggestionEngine *SuggestionEngine
}

// NewUsageAnalytics creates a new analytics instance
func NewUsageAnalytics() *UsageAnalytics {
	ua := &UsageAnalytics{
		commandHistory:   make([]CommandUsage, 0),
		fileOperations:   make([]FileOperation, 0),
		patterns:         make([]UsagePattern, 0),
		suggestions:      make([]SmartSuggestion, 0),
		dataFile:         "usage_analytics.json",
		maxHistorySize:   10000,
		patternThreshold: 0.7,
		suggestionEngine: NewSuggestionEngine(),
	}

	// Load existing data
	ua.loadData()

	// Start background analysis
	go ua.startBackgroundAnalysis()

	return ua
}

// TrackCommand records a command execution
func (ua *UsageAnalytics) TrackCommand(command string, args []string, duration time.Duration, success bool, workingDir string, context map[string]string, errorMsg string) {
	ua.mutex.Lock()
	defer ua.mutex.Unlock()

	usage := CommandUsage{
		Command:    command,
		Args:       args,
		Timestamp:  time.Now(),
		Duration:   duration,
		Success:    success,
		Context:    context,
		WorkingDir: workingDir,
		FileCount:  len(args),
		ErrorMsg:   errorMsg,
	}

	ua.commandHistory = append(ua.commandHistory, usage)

	// Keep history size manageable
	if len(ua.commandHistory) > ua.maxHistorySize {
		ua.commandHistory = ua.commandHistory[len(ua.commandHistory)-ua.maxHistorySize:]
	}

	// Trigger real-time analysis for immediate suggestions
	go ua.analyzeRecentCommand(usage)
}

// TrackFileOperation records a file system operation
func (ua *UsageAnalytics) TrackFileOperation(operation, path string, size int64, success bool) {
	ua.mutex.Lock()
	defer ua.mutex.Unlock()

	fileOp := FileOperation{
		Operation: operation,
		Path:      path,
		Size:      size,
		Timestamp: time.Now(),
		Success:   success,
		FileType:  ua.getFileType(path),
		Directory: filepath.Dir(path),
	}

	ua.fileOperations = append(ua.fileOperations, fileOp)

	// Keep history size manageable
	if len(ua.fileOperations) > ua.maxHistorySize {
		ua.fileOperations = ua.fileOperations[len(ua.fileOperations)-ua.maxHistorySize:]
	}
}

// GetSuggestions returns intelligent suggestions based on usage patterns
func (ua *UsageAnalytics) GetSuggestions(limit int) []SmartSuggestion {
	ua.mutex.RLock()
	defer ua.mutex.RUnlock()

	// Generate fresh suggestions
	suggestions := ua.suggestionEngine.GenerateSuggestions(ua.commandHistory, ua.fileOperations, ua.patterns)

	// Sort by priority and confidence
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Priority != suggestions[j].Priority {
			return suggestions[i].Priority > suggestions[j].Priority
		}
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	// Limit results
	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions
}

// GetCommandSuggestions returns suggestions for command completion
func (ua *UsageAnalytics) GetCommandSuggestions(partialCommand string, context map[string]string) []string {
	ua.mutex.RLock()
	defer ua.mutex.RUnlock()

	return ua.suggestionEngine.GetCommandSuggestions(partialCommand, ua.commandHistory, context)
}

// GetWorkflowSuggestions returns suggestions for common workflows
func (ua *UsageAnalytics) GetWorkflowSuggestions() []SmartSuggestion {
	ua.mutex.RLock()
	defer ua.mutex.RUnlock()

	return ua.suggestionEngine.GetWorkflowSuggestions(ua.commandHistory, ua.fileOperations)
}

// GetOptimizationSuggestions returns system optimization suggestions
func (ua *UsageAnalytics) GetOptimizationSuggestions() []SmartSuggestion {
	ua.mutex.RLock()
	defer ua.mutex.RUnlock()

	return ua.suggestionEngine.GetOptimizationSuggestions(ua.commandHistory, ua.fileOperations)
}

// ProvideFeedback allows users to provide feedback on suggestions
func (ua *UsageAnalytics) ProvideFeedback(suggestionID, feedback string) error {
	ua.mutex.Lock()
	defer ua.mutex.Unlock()

	for i, suggestion := range ua.suggestions {
		if suggestion.ID == suggestionID {
			ua.suggestions[i].UserFeedback = feedback
			ua.suggestions[i].LastShown = time.Now()
			ua.suggestions[i].TimesShown++
			break
		}
	}

	return ua.saveData()
}

// GetUsageStats returns comprehensive usage statistics
func (ua *UsageAnalytics) GetUsageStats() map[string]interface{} {
	ua.mutex.RLock()
	defer ua.mutex.RUnlock()

	stats := make(map[string]interface{})

	// Command statistics
	commandCounts := make(map[string]int)
	totalDuration := time.Duration(0)
	successCount := 0

	for _, cmd := range ua.commandHistory {
		commandCounts[cmd.Command]++
		totalDuration += cmd.Duration
		if cmd.Success {
			successCount++
		}
	}

	// File operation statistics
	fileOpCounts := make(map[string]int)
	totalFileSize := int64(0)

	for _, fileOp := range ua.fileOperations {
		fileOpCounts[fileOp.Operation]++
		totalFileSize += fileOp.Size
	}

	stats["total_commands"] = len(ua.commandHistory)
	stats["total_file_operations"] = len(ua.fileOperations)
	stats["most_used_commands"] = ua.getTopItems(commandCounts, 5)
	stats["most_common_file_operations"] = ua.getTopItems(fileOpCounts, 5)
	if len(ua.commandHistory) > 0 {
		stats["average_command_duration"] = ua.formatDuration(totalDuration / time.Duration(len(ua.commandHistory)))
		stats["success_rate"] = float64(successCount) / float64(len(ua.commandHistory)) * 100
	} else {
		stats["average_command_duration"] = "0s"
		stats["success_rate"] = 0.0
	}
	stats["total_file_size_processed"] = ua.formatBytes(totalFileSize)
	stats["patterns_discovered"] = len(ua.patterns)
	stats["suggestions_generated"] = len(ua.suggestions)
	stats["analysis_period"] = ua.getAnalysisPeriod()

	return stats
}

// Private helper methods

func (ua *UsageAnalytics) getFileType(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return "directory"
	}
	return ext[1:] // Remove the dot
}

func (ua *UsageAnalytics) getTopItems(counts map[string]int, limit int) []map[string]interface{} {
	type item struct {
		name  string
		count int
	}

	var items []item
	for name, count := range counts {
		items = append(items, item{name, count})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].count > items[j].count
	})

	var result []map[string]interface{}
	for i, item := range items {
		if i >= limit {
			break
		}
		result = append(result, map[string]interface{}{
			"name":  item.name,
			"count": item.count,
		})
	}

	return result
}

func (ua *UsageAnalytics) formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

func (ua *UsageAnalytics) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (ua *UsageAnalytics) getAnalysisPeriod() string {
	if len(ua.commandHistory) == 0 {
		return "No data"
	}

	first := ua.commandHistory[0].Timestamp
	last := ua.commandHistory[len(ua.commandHistory)-1].Timestamp
	duration := last.Sub(first)

	if duration < 24*time.Hour {
		return fmt.Sprintf("%.0f hours", duration.Hours())
	} else {
		return fmt.Sprintf("%.1f days", duration.Hours()/24)
	}
}

func (ua *UsageAnalytics) analyzeRecentCommand(usage CommandUsage) {
	// This will be called in background to analyze patterns
	// Implementation will be in the suggestion engine
}

func (ua *UsageAnalytics) startBackgroundAnalysis() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ua.performBackgroundAnalysis()
	}
}

func (ua *UsageAnalytics) performBackgroundAnalysis() {
	ua.mutex.Lock()
	defer ua.mutex.Unlock()

	// Analyze patterns and generate suggestions
	ua.patterns = ua.suggestionEngine.AnalyzePatterns(ua.commandHistory, ua.fileOperations)
	ua.suggestions = ua.suggestionEngine.GenerateSuggestions(ua.commandHistory, ua.fileOperations, ua.patterns)

	// Save updated data
	ua.saveData()
}

func (ua *UsageAnalytics) loadData() error {
	if _, err := os.Stat(ua.dataFile); os.IsNotExist(err) {
		return nil // No existing data file
	}

	data, err := os.ReadFile(ua.dataFile)
	if err != nil {
		return fmt.Errorf("failed to read analytics data: %v", err)
	}

	var analyticsData struct {
		CommandHistory []CommandUsage    `json:"command_history"`
		FileOperations []FileOperation   `json:"file_operations"`
		Patterns       []UsagePattern    `json:"patterns"`
		Suggestions    []SmartSuggestion `json:"suggestions"`
	}

	if err := json.Unmarshal(data, &analyticsData); err != nil {
		return fmt.Errorf("failed to parse analytics data: %v", err)
	}

	ua.commandHistory = analyticsData.CommandHistory
	ua.fileOperations = analyticsData.FileOperations
	ua.patterns = analyticsData.Patterns
	ua.suggestions = analyticsData.Suggestions

	return nil
}

func (ua *UsageAnalytics) saveData() error {
	analyticsData := struct {
		CommandHistory []CommandUsage    `json:"command_history"`
		FileOperations []FileOperation   `json:"file_operations"`
		Patterns       []UsagePattern    `json:"patterns"`
		Suggestions    []SmartSuggestion `json:"suggestions"`
		LastUpdated    time.Time         `json:"last_updated"`
	}{
		CommandHistory: ua.commandHistory,
		FileOperations: ua.fileOperations,
		Patterns:       ua.patterns,
		Suggestions:    ua.suggestions,
		LastUpdated:    time.Now(),
	}

	data, err := json.MarshalIndent(analyticsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analytics data: %v", err)
	}

	if err := os.WriteFile(ua.dataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write analytics data: %v", err)
	}

	return nil
}
