/**
 * Intelligent suggestion engine for analyzing user patterns and generating smart suggestions.
 *
 * Analyzes command history, file operations, and usage patterns to provide
 * intelligent recommendations for improved productivity and workflow optimization.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: suggestion_engine.go
 * Description: Core suggestion engine with pattern analysis and recommendation algorithms
 */

package suggestions

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// SuggestionEngine analyzes patterns and generates intelligent suggestions
type SuggestionEngine struct {
	commandPatterns     map[string]*CommandPattern
	workflowPatterns    map[string]*WorkflowPattern
	optimizationRules   []OptimizationRule
	suggestionTemplates map[string]*SuggestionTemplate
}

// CommandPattern represents a discovered command usage pattern
type CommandPattern struct {
	Command     string            `json:"command"`
	Frequency   int               `json:"frequency"`
	AvgDuration time.Duration     `json:"avg_duration"`
	SuccessRate float64           `json:"success_rate"`
	CommonArgs  []string          `json:"common_args"`
	Context     map[string]string `json:"context"`
	LastUsed    time.Time         `json:"last_used"`
}

// WorkflowPattern represents a sequence of related operations
type WorkflowPattern struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Steps       []WorkflowStep    `json:"steps"`
	Frequency   int               `json:"frequency"`
	SuccessRate float64           `json:"success_rate"`
	AvgDuration time.Duration     `json:"avg_duration"`
	Context     map[string]string `json:"context"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Command  string        `json:"command"`
	Args     []string      `json:"args"`
	Duration time.Duration `json:"duration"`
	Success  bool          `json:"success"`
	Order    int           `json:"order"`
}

// OptimizationRule represents a rule for system optimization
type OptimizationRule struct {
	ID          string                            `json:"id"`
	Name        string                            `json:"name"`
	Description string                            `json:"description"`
	Condition   func(map[string]interface{}) bool `json:"-"`
	Suggestion  string                            `json:"suggestion"`
	Priority    int                               `json:"priority"`
	Category    string                            `json:"category"`
}

// SuggestionTemplate provides templates for generating suggestions
type SuggestionTemplate struct {
	Type        string                                        `json:"type"`
	Title       string                                        `json:"title"`
	Description string                                        `json:"description"`
	Category    string                                        `json:"category"`
	Priority    int                                           `json:"priority"`
	Generate    func(map[string]interface{}) *SmartSuggestion `json:"-"`
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine() *SuggestionEngine {
	se := &SuggestionEngine{
		commandPatterns:     make(map[string]*CommandPattern),
		workflowPatterns:    make(map[string]*WorkflowPattern),
		optimizationRules:   make([]OptimizationRule, 0),
		suggestionTemplates: make(map[string]*SuggestionTemplate),
	}

	se.initializeOptimizationRules()
	se.initializeSuggestionTemplates()

	return se
}

// AnalyzePatterns analyzes command history and file operations to discover patterns
func (se *SuggestionEngine) AnalyzePatterns(commandHistory []CommandUsage, fileOperations []FileOperation) []UsagePattern {
	var patterns []UsagePattern

	// Analyze command frequency patterns
	patterns = append(patterns, se.analyzeCommandFrequency(commandHistory)...)

	// Analyze workflow patterns
	patterns = append(patterns, se.analyzeWorkflowPatterns(commandHistory)...)

	// Analyze file operation patterns
	patterns = append(patterns, se.analyzeFileOperationPatterns(fileOperations)...)

	// Analyze time-based patterns
	patterns = append(patterns, se.analyzeTimeBasedPatterns(commandHistory)...)

	// Analyze error patterns
	patterns = append(patterns, se.analyzeErrorPatterns(commandHistory)...)

	return patterns
}

// GenerateSuggestions creates intelligent suggestions based on patterns
func (se *SuggestionEngine) GenerateSuggestions(commandHistory []CommandUsage, fileOperations []FileOperation, patterns []UsagePattern) []SmartSuggestion {
	var suggestions []SmartSuggestion

	// Generate command suggestions
	suggestions = append(suggestions, se.generateCommandSuggestions(commandHistory)...)

	// Generate workflow suggestions
	suggestions = append(suggestions, se.generateWorkflowSuggestions(commandHistory)...)

	// Generate optimization suggestions
	suggestions = append(suggestions, se.generateOptimizationSuggestions(commandHistory, fileOperations)...)

	// Generate productivity suggestions
	suggestions = append(suggestions, se.generateProductivitySuggestions(commandHistory, fileOperations)...)

	// Generate safety suggestions
	suggestions = append(suggestions, se.generateSafetySuggestions(commandHistory, fileOperations)...)

	// Filter and rank suggestions
	suggestions = se.filterAndRankSuggestions(suggestions)

	return suggestions
}

// GetCommandSuggestions returns command completion suggestions
func (se *SuggestionEngine) GetCommandSuggestions(partialCommand string, commandHistory []CommandUsage, context map[string]string) []string {
	var suggestions []string

	// Get frequently used commands that match the partial
	commandCounts := make(map[string]int)
	for _, cmd := range commandHistory {
		if strings.HasPrefix(cmd.Command, partialCommand) {
			commandCounts[cmd.Command]++
		}
	}

	// Sort by frequency
	type cmdFreq struct {
		command string
		count   int
	}
	var sortedCmds []cmdFreq
	for cmd, count := range commandCounts {
		sortedCmds = append(sortedCmds, cmdFreq{cmd, count})
	}

	sort.Slice(sortedCmds, func(i, j int) bool {
		return sortedCmds[i].count > sortedCmds[j].count
	})

	// Return top suggestions
	for _, cmd := range sortedCmds {
		suggestions = append(suggestions, cmd.command)
		if len(suggestions) >= 10 {
			break
		}
	}

	return suggestions
}

// GetWorkflowSuggestions returns workflow optimization suggestions
func (se *SuggestionEngine) GetWorkflowSuggestions(commandHistory []CommandUsage, fileOperations []FileOperation) []SmartSuggestion {
	var suggestions []SmartSuggestion

	// Analyze common command sequences
	sequences := se.findCommonSequences(commandHistory)

	for _, seq := range sequences {
		if len(seq) >= 2 {
			suggestion := &SmartSuggestion{
				ID:          fmt.Sprintf("workflow_%d", len(suggestions)),
				Type:        "workflow",
				Title:       "Common Command Sequence Detected",
				Description: fmt.Sprintf("You often run: %s", strings.Join(seq, " → ")),
				Command:     se.generateWorkflowCommand(seq),
				Confidence:  0.8,
				Priority:    7,
				Category:    "productivity",
				Context: map[string]interface{}{
					"sequence":  seq,
					"frequency": se.getSequenceFrequency(seq, commandHistory),
				},
				CreatedAt: time.Now(),
			}
			suggestions = append(suggestions, *suggestion)
		}
	}

	return suggestions
}

// GetOptimizationSuggestions returns system optimization suggestions
func (se *SuggestionEngine) GetOptimizationSuggestions(commandHistory []CommandUsage, fileOperations []FileOperation) []SmartSuggestion {
	var suggestions []SmartSuggestion

	// Check optimization rules
	for _, rule := range se.optimizationRules {
		context := map[string]interface{}{
			"command_history": commandHistory,
			"file_operations": fileOperations,
		}

		if rule.Condition(context) {
			suggestion := &SmartSuggestion{
				ID:          fmt.Sprintf("optimization_%s", rule.ID),
				Type:        "optimization",
				Title:       rule.Name,
				Description: rule.Description,
				Command:     rule.Suggestion,
				Confidence:  0.9,
				Priority:    rule.Priority,
				Category:    rule.Category,
				Context:     context,
				CreatedAt:   time.Now(),
			}
			suggestions = append(suggestions, *suggestion)
		}
	}

	return suggestions
}

// Private analysis methods

func (se *SuggestionEngine) analyzeCommandFrequency(commandHistory []CommandUsage) []UsagePattern {
	commandCounts := make(map[string]int)
	totalDuration := make(map[string]time.Duration)
	successCounts := make(map[string]int)

	for _, cmd := range commandHistory {
		commandCounts[cmd.Command]++
		totalDuration[cmd.Command] += cmd.Duration
		if cmd.Success {
			successCounts[cmd.Command]++
		}
	}

	var patterns []UsagePattern
	for command, count := range commandCounts {
		if count >= 3 { // Only patterns with 3+ occurrences
			avgDuration := totalDuration[command] / time.Duration(count)
			successRate := float64(successCounts[command]) / float64(count)

			pattern := UsagePattern{
				PatternType: "command_frequency",
				Description: fmt.Sprintf("Frequently used command: %s", command),
				Frequency:   count,
				Confidence:  math.Min(float64(count)/10.0, 1.0),
				LastSeen:    time.Now(),
				Data: map[string]interface{}{
					"command":      command,
					"frequency":    count,
					"avg_duration": avgDuration.String(),
					"success_rate": successRate,
				},
				Suggestions: []string{
					fmt.Sprintf("Consider creating an alias for '%s'", command),
					fmt.Sprintf("This command is used %d times", count),
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

func (se *SuggestionEngine) analyzeWorkflowPatterns(commandHistory []CommandUsage) []UsagePattern {
	var patterns []UsagePattern

	// Look for command sequences that occur together
	sequences := se.findCommonSequences(commandHistory)

	for _, seq := range sequences {
		if len(seq) >= 2 {
			frequency := se.getSequenceFrequency(seq, commandHistory)
			if frequency >= 2 {
				pattern := UsagePattern{
					PatternType: "workflow_sequence",
					Description: fmt.Sprintf("Common workflow: %s", strings.Join(seq, " → ")),
					Frequency:   frequency,
					Confidence:  math.Min(float64(frequency)/5.0, 1.0),
					LastSeen:    time.Now(),
					Data: map[string]interface{}{
						"sequence":  seq,
						"frequency": frequency,
					},
					Suggestions: []string{
						"Consider creating a script for this workflow",
						"This sequence could be automated",
					},
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns
}

func (se *SuggestionEngine) analyzeFileOperationPatterns(fileOperations []FileOperation) []UsagePattern {
	var patterns []UsagePattern

	// Analyze file type patterns
	fileTypeCounts := make(map[string]int)
	operationCounts := make(map[string]int)

	for _, op := range fileOperations {
		fileTypeCounts[op.FileType]++
		operationCounts[op.Operation]++
	}

	// File type patterns
	for fileType, count := range fileTypeCounts {
		if count >= 5 {
			pattern := UsagePattern{
				PatternType: "file_type_usage",
				Description: fmt.Sprintf("Frequently working with .%s files", fileType),
				Frequency:   count,
				Confidence:  math.Min(float64(count)/20.0, 1.0),
				LastSeen:    time.Now(),
				Data: map[string]interface{}{
					"file_type": fileType,
					"count":     count,
				},
				Suggestions: []string{
					fmt.Sprintf("Consider organizing .%s files in a dedicated folder", fileType),
					fmt.Sprintf("You work with .%s files %d times", fileType, count),
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	// Operation patterns
	for operation, count := range operationCounts {
		if count >= 3 {
			pattern := UsagePattern{
				PatternType: "file_operation",
				Description: fmt.Sprintf("Frequent %s operations", operation),
				Frequency:   count,
				Confidence:  math.Min(float64(count)/10.0, 1.0),
				LastSeen:    time.Now(),
				Data: map[string]interface{}{
					"operation": operation,
					"count":     count,
				},
				Suggestions: []string{
					fmt.Sprintf("Consider batch %s operations", operation),
					fmt.Sprintf("You perform %s operations %d times", operation, count),
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

func (se *SuggestionEngine) analyzeTimeBasedPatterns(commandHistory []CommandUsage) []UsagePattern {
	var patterns []UsagePattern

	// Analyze usage by hour of day
	hourCounts := make(map[int]int)
	for _, cmd := range commandHistory {
		hourCounts[cmd.Timestamp.Hour()]++
	}

	// Find peak usage hours
	var peakHours []int
	maxCount := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHours = []int{hour}
		} else if count == maxCount {
			peakHours = append(peakHours, hour)
		}
	}

	if len(peakHours) > 0 && maxCount >= 3 {
		pattern := UsagePattern{
			PatternType: "time_based",
			Description: fmt.Sprintf("Peak usage around %d:00", peakHours[0]),
			Frequency:   maxCount,
			Confidence:  0.7,
			LastSeen:    time.Now(),
			Data: map[string]interface{}{
				"peak_hours": peakHours,
				"max_count":  maxCount,
			},
			Suggestions: []string{
				"Consider scheduling automated tasks during off-peak hours",
				"You're most active during these hours",
			},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

func (se *SuggestionEngine) analyzeErrorPatterns(commandHistory []CommandUsage) []UsagePattern {
	var patterns []UsagePattern

	// Analyze failed commands
	failedCommands := make(map[string]int)
	for _, cmd := range commandHistory {
		if !cmd.Success {
			failedCommands[cmd.Command]++
		}
	}

	for command, count := range failedCommands {
		if count >= 2 {
			pattern := UsagePattern{
				PatternType: "error_pattern",
				Description: fmt.Sprintf("Command '%s' fails frequently", command),
				Frequency:   count,
				Confidence:  0.9,
				LastSeen:    time.Now(),
				Data: map[string]interface{}{
					"command":  command,
					"failures": count,
				},
				Suggestions: []string{
					fmt.Sprintf("Check the syntax for '%s' command", command),
					fmt.Sprintf("Consider using 'help %s' for assistance", command),
					"This command has failed multiple times",
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// Private suggestion generation methods

func (se *SuggestionEngine) generateCommandSuggestions(commandHistory []CommandUsage) []SmartSuggestion {
	var suggestions []SmartSuggestion

	// Suggest aliases for frequently used commands
	commandCounts := make(map[string]int)
	for _, cmd := range commandHistory {
		commandCounts[cmd.Command]++
	}

	for command, count := range commandCounts {
		if count >= 5 {
			suggestion := &SmartSuggestion{
				ID:          fmt.Sprintf("alias_%s", command),
				Type:        "command",
				Title:       fmt.Sprintf("Create alias for '%s'", command),
				Description: fmt.Sprintf("You use '%s' %d times. Consider creating an alias for faster access.", command, count),
				Command:     fmt.Sprintf("alias %s='%s'", command, command),
				Confidence:  0.8,
				Priority:    6,
				Category:    "productivity",
				Context: map[string]interface{}{
					"command": command,
					"count":   count,
				},
				CreatedAt: time.Now(),
			}
			suggestions = append(suggestions, *suggestion)
		}
	}

	return suggestions
}

func (se *SuggestionEngine) generateWorkflowSuggestions(commandHistory []CommandUsage) []SmartSuggestion {
	return se.GetWorkflowSuggestions(commandHistory, []FileOperation{})
}

func (se *SuggestionEngine) generateOptimizationSuggestions(commandHistory []CommandUsage, fileOperations []FileOperation) []SmartSuggestion {
	return se.GetOptimizationSuggestions(commandHistory, fileOperations)
}

func (se *SuggestionEngine) generateProductivitySuggestions(commandHistory []CommandUsage, fileOperations []FileOperation) []SmartSuggestion {
	var suggestions []SmartSuggestion

	// Suggest batch operations for file operations
	fileOpCounts := make(map[string]int)
	for _, op := range fileOperations {
		fileOpCounts[op.Operation]++
	}

	for operation, count := range fileOpCounts {
		if count >= 3 {
			suggestion := &SmartSuggestion{
				ID:          fmt.Sprintf("batch_%s", operation),
				Type:        "productivity",
				Title:       fmt.Sprintf("Batch %s operations", operation),
				Description: fmt.Sprintf("You perform %s operations %d times. Consider batching them for efficiency.", operation, count),
				Command:     fmt.Sprintf("multi %s [files...]", operation),
				Confidence:  0.7,
				Priority:    5,
				Category:    "productivity",
				Context: map[string]interface{}{
					"operation": operation,
					"count":     count,
				},
				CreatedAt: time.Now(),
			}
			suggestions = append(suggestions, *suggestion)
		}
	}

	return suggestions
}

func (se *SuggestionEngine) generateSafetySuggestions(commandHistory []CommandUsage, fileOperations []FileOperation) []SmartSuggestion {
	var suggestions []SmartSuggestion

	// Check for destructive operations without backups
	deleteCount := 0
	backupCount := 0

	for _, cmd := range commandHistory {
		if strings.Contains(cmd.Command, "delete") {
			deleteCount++
		}
		if strings.Contains(cmd.Command, "backup") || strings.Contains(cmd.Command, "copy") {
			backupCount++
		}
	}

	if deleteCount > backupCount && deleteCount >= 2 {
		suggestion := SmartSuggestion{
			ID:          "safety_backup",
			Type:        "safety",
			Title:       "Enable automatic backups",
			Description: "You perform delete operations frequently. Consider enabling automatic backups for safety.",
			Command:     "system backup enable",
			Confidence:  0.9,
			Priority:    8,
			Category:    "safety",
			Context: map[string]interface{}{
				"delete_count": deleteCount,
				"backup_count": backupCount,
			},
			CreatedAt: time.Now(),
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

// Private helper methods

func (se *SuggestionEngine) findCommonSequences(commandHistory []CommandUsage) [][]string {
	sequences := make(map[string]int)

	// Look for sequences of 2-4 commands
	for i := 0; i < len(commandHistory)-1; i++ {
		for length := 2; length <= 4 && i+length <= len(commandHistory); length++ {
			var seq []string
			for j := i; j < i+length; j++ {
				seq = append(seq, commandHistory[j].Command)
			}
			seqKey := strings.Join(seq, "|")
			sequences[seqKey]++
		}
	}

	// Convert back to slices and filter by frequency
	var result [][]string
	for seqKey, count := range sequences {
		if count >= 2 {
			seq := strings.Split(seqKey, "|")
			result = append(result, seq)
		}
	}

	return result
}

func (se *SuggestionEngine) getSequenceFrequency(sequence []string, commandHistory []CommandUsage) int {
	count := 0

	for i := 0; i <= len(commandHistory)-len(sequence); i++ {
		match := true
		for j, cmd := range sequence {
			if commandHistory[i+j].Command != cmd {
				match = false
				break
			}
		}
		if match {
			count++
		}
	}

	return count
}

func (se *SuggestionEngine) generateWorkflowCommand(sequence []string) string {
	// Generate a simple script command for the sequence
	return fmt.Sprintf("script create workflow_%d --commands '%s'", len(sequence), strings.Join(sequence, ","))
}

func (se *SuggestionEngine) filterAndRankSuggestions(suggestions []SmartSuggestion) []SmartSuggestion {
	// Remove duplicates and rank by priority and confidence
	seen := make(map[string]bool)
	var filtered []SmartSuggestion

	for _, suggestion := range suggestions {
		key := suggestion.Type + ":" + suggestion.Title
		if !seen[key] {
			seen[key] = true
			filtered = append(filtered, suggestion)
		}
	}

	// Sort by priority (descending) then confidence (descending)
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Priority != filtered[j].Priority {
			return filtered[i].Priority > filtered[j].Priority
		}
		return filtered[i].Confidence > filtered[j].Confidence
	})

	return filtered
}

func (se *SuggestionEngine) initializeOptimizationRules() {
	se.optimizationRules = []OptimizationRule{
		{
			ID:          "frequent_errors",
			Name:        "Frequent Command Errors",
			Description: "You're experiencing frequent command errors. Consider using the help system more often.",
			Condition: func(context map[string]interface{}) bool {
				commandHistory := context["command_history"].([]CommandUsage)
				errorCount := 0
				for _, cmd := range commandHistory {
					if !cmd.Success {
						errorCount++
					}
				}
				return len(commandHistory) > 10 && float64(errorCount)/float64(len(commandHistory)) > 0.2
			},
			Suggestion: "help",
			Priority:   8,
			Category:   "optimization",
		},
		{
			ID:          "slow_commands",
			Name:        "Slow Command Execution",
			Description: "Some commands are taking longer than expected. Consider optimizing your workflow.",
			Condition: func(context map[string]interface{}) bool {
				commandHistory := context["command_history"].([]CommandUsage)
				slowCount := 0
				for _, cmd := range commandHistory {
					if cmd.Duration > 5*time.Second {
						slowCount++
					}
				}
				return len(commandHistory) > 5 && slowCount >= 3
			},
			Suggestion: "system optimize",
			Priority:   6,
			Category:   "optimization",
		},
	}
}

func (se *SuggestionEngine) initializeSuggestionTemplates() {
	se.suggestionTemplates = map[string]*SuggestionTemplate{
		"command_alias": {
			Type:        "command",
			Title:       "Command Alias Suggestion",
			Description: "Frequently used command detected",
			Category:    "productivity",
			Priority:    6,
		},
		"workflow_automation": {
			Type:        "workflow",
			Title:       "Workflow Automation",
			Description: "Common command sequence detected",
			Category:    "productivity",
			Priority:    7,
		},
		"safety_backup": {
			Type:        "safety",
			Title:       "Safety Recommendation",
			Description: "Backup recommendation for destructive operations",
			Category:    "safety",
			Priority:    8,
		},
	}
}
