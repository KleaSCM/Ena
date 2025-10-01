/**
 * CLI commands for pattern-based file operations.
 *
 * Provides commands for creating, managing, and executing pattern-based
 * file operations with advanced filtering and matching capabilities.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: pattern_commands.go
 * Description: Cobra command definitions for pattern-based operations
 */

package commands

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ena/internal/patterns"

	"github.com/spf13/cobra"
)

// Global pattern engine instance
var globalPatternEngine *patterns.PatternEngine

// getGlobalPatternEngine returns the global pattern engine instance
func getGlobalPatternEngine() *patterns.PatternEngine {
	if globalPatternEngine == nil {
		analytics := getGlobalAnalytics()
		globalPatternEngine = patterns.NewPatternEngine(analytics)
	}
	return globalPatternEngine
}

// setupPatternCommands adds pattern-based operation commands to the root command
func setupPatternCommands(rootCmd *cobra.Command) {
	// Get pattern engine
	engine := getGlobalPatternEngine()

	// Find command - execute pattern-based file finding
	findCmd := &cobra.Command{
		Use:   "find <pattern> [paths...]",
		Short: "Find files matching pattern criteria",
		Long: `Find files using advanced pattern matching and filtering.
Supports filtering by extension, age, size, content, and more.

Examples:
  ena find "*.txt older than 30d" ~/Documents
  ena find "files > 100MB" ~/Downloads
  ena find "*.jpg created today" ~/Pictures
  ena find "files containing 'TODO'" ~/Projects`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pattern := args[0]
			var searchPaths []string

			if len(args) > 1 {
				searchPaths = args[1:]
			} else {
				searchPaths = []string{"."}
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			verbose, _ := cmd.Flags().GetBool("verbose")
			limit, _ := cmd.Flags().GetInt("limit")

			// Create temporary operation for this search
			operation := createOperationFromPattern(pattern, searchPaths)
			if operation == nil {
				fmt.Println("❌ Invalid pattern format")
				return
			}

			fmt.Printf("🌸 Searching for: %s\n", operation.Description)
			if dryRun {
				fmt.Println("🔍 Dry run mode - no files will be modified")
			}

			// Execute the search
			result, err := engine.ExecuteOperation(operation.ID, dryRun)
			if err != nil {
				fmt.Printf("❌ Error executing search: %v\n", err)
				return
			}

			// Display results
			showPatternResults([]patterns.PatternResult{*result}, verbose, limit)
		},
	}

	findCmd.Flags().Bool("dry-run", false, "Preview results without making changes")
	findCmd.Flags().Bool("verbose", false, "Show detailed information about each file")
	findCmd.Flags().Int("limit", 0, "Limit number of results (0 = no limit)")

	// Create operation command
	createCmd := &cobra.Command{
		Use:   "create-operation <name>",
		Short: "Create a new pattern operation",
		Long: `Create a new pattern operation with custom filters and actions.
This will prompt for operation configuration details.

Examples:
  ena create-operation "Clean Downloads"
  ena create-operation "Archive Old Files"`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			// Create a sample operation
			operation := engine.CreateSampleOperation()
			operation.Name = name
			operation.Description = fmt.Sprintf("Pattern operation: %s", name)

			err := engine.AddOperation(operation)
			if err != nil {
				fmt.Printf("❌ Error creating operation: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully created pattern operation: %s\n", operation.Name)
			fmt.Printf("🆔 Operation ID: %s\n", operation.ID)
			fmt.Printf("📋 Filters: %d\n", len(operation.Filters))
			fmt.Printf("📁 Paths: %s\n", strings.Join(operation.Paths, ", "))
		},
	}

	// List operations command
	listCmd := &cobra.Command{
		Use:   "list-operations",
		Short: "List all pattern operations",
		Long: `List all configured pattern operations with their details.
Shows operation names, filters, paths, and status.

Examples:
  ena list-operations
  ena list-operations --enabled-only`,
		Run: func(cmd *cobra.Command, args []string) {
			enabledOnly, _ := cmd.Flags().GetBool("enabled-only")

			operations := engine.GetOperations()
			if len(operations) == 0 {
				fmt.Println("🌸 No pattern operations configured")
				return
			}

			fmt.Println("🌸 Pattern Operations (╹◡╹)♡")
			fmt.Println("================================")

			for i, operation := range operations {
				if enabledOnly && !operation.Enabled {
					continue
				}

				status := "❌ Disabled"
				if operation.Enabled {
					status = "✅ Enabled"
				}

				fmt.Printf("%d. %s (%s)\n", i+1, operation.Name, status)
				fmt.Printf("   📝 Description: %s\n", operation.Description)
				fmt.Printf("   🆔 ID: %s\n", operation.ID)
				fmt.Printf("   📊 Priority: %d\n", operation.Priority)
				fmt.Printf("   📁 Paths: %s\n", strings.Join(operation.Paths, ", "))
				fmt.Printf("   🔍 Filters: %d\n", len(operation.Filters))
				fmt.Printf("   ⚡ Actions: %d\n", len(operation.Actions))
				fmt.Printf("   📅 Created: %s\n", operation.CreatedAt.Format("2006-01-02 15:04:05"))
				if operation.LastRun != nil {
					fmt.Printf("   🏃 Last Run: %s\n", operation.LastRun.Format("2006-01-02 15:04:05"))
				}
				fmt.Println()
			}
		},
	}

	listCmd.Flags().Bool("enabled-only", false, "Show only enabled operations")

	// Execute operation command
	executeCmd := &cobra.Command{
		Use:   "execute-operation <id>",
		Short: "Execute a specific pattern operation",
		Long: `Execute a specific pattern operation by ID.
Applies all filters and actions defined in the operation.

Examples:
  ena execute-operation pattern_1234567890
  ena execute-operation pattern_1234567890 --dry-run`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			operationID := args[0]
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Get operation details
			operation, err := engine.GetOperationByID(operationID)
			if err != nil {
				fmt.Printf("❌ Error getting operation: %v\n", err)
				return
			}

			fmt.Printf("🌸 Executing operation: %s\n", operation.Name)
			if dryRun {
				fmt.Println("🔍 Dry run mode - no files will be modified")
			}

			// Execute the operation
			result, err := engine.ExecuteOperation(operationID, dryRun)
			if err != nil {
				fmt.Printf("❌ Error executing operation: %v\n", err)
				return
			}

			// Display results
			showPatternResults([]patterns.PatternResult{*result}, verbose, 0)
		},
	}

	executeCmd.Flags().Bool("dry-run", false, "Preview operation without making changes")
	executeCmd.Flags().Bool("verbose", false, "Show detailed information about each file")

	// Execute all operations command
	executeAllCmd := &cobra.Command{
		Use:   "execute-all",
		Short: "Execute all enabled pattern operations",
		Long: `Execute all enabled pattern operations in priority order.
Each operation will be executed independently.

Examples:
  ena execute-all
  ena execute-all --dry-run`,
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			verbose, _ := cmd.Flags().GetBool("verbose")

			fmt.Println("🌸 Executing all enabled pattern operations...")
			if dryRun {
				fmt.Println("🔍 Dry run mode - no files will be modified")
			}

			// Execute all operations
			results, err := engine.ExecuteAllOperations(dryRun)
			if err != nil {
				fmt.Printf("❌ Error executing operations: %v\n", err)
				return
			}

			// Display results
			showPatternResults(results, verbose, 0)
		},
	}

	executeAllCmd.Flags().Bool("dry-run", false, "Preview operations without making changes")
	executeAllCmd.Flags().Bool("verbose", false, "Show detailed information about each file")

	// Remove operation command
	removeCmd := &cobra.Command{
		Use:   "remove-operation <id>",
		Short: "Remove a pattern operation",
		Long: `Remove a pattern operation by ID.
This will permanently delete the operation and its configuration.

Examples:
  ena remove-operation pattern_1234567890`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			operationID := args[0]
			confirm, _ := cmd.Flags().GetBool("confirm")

			if !confirm {
				fmt.Printf("⚠️  This will permanently remove operation %s\n", operationID)
				fmt.Print("Type 'yes' to confirm: ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "yes" {
					fmt.Println("❌ Operation cancelled")
					return
				}
			}

			err := engine.RemoveOperation(operationID)
			if err != nil {
				fmt.Printf("❌ Error removing operation: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully removed operation: %s\n", operationID)
		},
	}

	removeCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	// Add all commands to root
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(executeCmd)
	rootCmd.AddCommand(executeAllCmd)
	rootCmd.AddCommand(removeCmd)
}

// Helper function to create operation from pattern string
func createOperationFromPattern(pattern string, paths []string) *patterns.PatternOperation {
	// Expand paths (handle ~)
	expandedPaths := make([]string, len(paths))
	for i, path := range paths {
		expandedPaths[i] = expandPath(path)
	}

	operation := &patterns.PatternOperation{
		ID:          fmt.Sprintf("temp_%d", time.Now().UnixNano()),
		Name:        "Temporary Pattern Search",
		Description: pattern,
		Enabled:     true,
		Priority:    1,
		Paths:       expandedPaths,
		Recursive:   true,
		MaxDepth:    10,
		Actions: []patterns.Action{
			{
				Type: "list", // Just list files, don't modify
			},
		},
	}

	// Parse pattern string for common formats
	pattern = strings.ToLower(pattern)

	// Check for file extension patterns
	if strings.Contains(pattern, "*.txt") {
		operation.Filters = append(operation.Filters, patterns.FileFilter{
			Type:     patterns.PatternFileExtension,
			Operator: patterns.OpEquals,
			Value:    ".txt",
		})
	}
	if strings.Contains(pattern, "*.jpg") || strings.Contains(pattern, "*.jpeg") {
		operation.Filters = append(operation.Filters, patterns.FileFilter{
			Type:     patterns.PatternFileExtension,
			Operator: patterns.OpEquals,
			Value:    ".jpg",
		})
	}
	if strings.Contains(pattern, "*.pdf") {
		operation.Filters = append(operation.Filters, patterns.FileFilter{
			Type:     patterns.PatternFileExtension,
			Operator: patterns.OpEquals,
			Value:    ".pdf",
		})
	}
	if strings.Contains(pattern, "*.go") {
		operation.Filters = append(operation.Filters, patterns.FileFilter{
			Type:     patterns.PatternFileExtension,
			Operator: patterns.OpEquals,
			Value:    ".go",
		})
	}

	// Generic wildcard pattern matching
	if strings.HasPrefix(pattern, "*.") {
		ext := strings.TrimPrefix(pattern, "*")
		operation.Filters = append(operation.Filters, patterns.FileFilter{
			Type:     patterns.PatternFileExtension,
			Operator: patterns.OpEquals,
			Value:    ext,
		})
	}

	// Check for age patterns
	if strings.Contains(pattern, "older than") {
		parts := strings.Split(pattern, "older than")
		if len(parts) > 1 {
			ageStr := strings.TrimSpace(parts[1])
			operation.Filters = append(operation.Filters, patterns.FileFilter{
				Type:     patterns.PatternAge,
				Operator: patterns.OpGreaterThan,
				Value:    ageStr,
			})
		}
	}
	if strings.Contains(pattern, "newer than") {
		parts := strings.Split(pattern, "newer than")
		if len(parts) > 1 {
			ageStr := strings.TrimSpace(parts[1])
			operation.Filters = append(operation.Filters, patterns.FileFilter{
				Type:     patterns.PatternAge,
				Operator: patterns.OpLessThan,
				Value:    ageStr,
			})
		}
	}

	// Check for size patterns
	if strings.Contains(pattern, ">") && strings.Contains(pattern, "mb") {
		// Extract size value
		parts := strings.Split(pattern, ">")
		if len(parts) > 1 {
			sizeStr := strings.TrimSpace(parts[1])
			sizeStr = strings.ReplaceAll(sizeStr, "mb", "")
			sizeStr = strings.ReplaceAll(sizeStr, " ", "")
			if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
				operation.Filters = append(operation.Filters, patterns.FileFilter{
					Type:     patterns.PatternSize,
					Operator: patterns.OpGreaterThan,
					Value:    size * 1024 * 1024, // Convert MB to bytes
				})
			}
		}
	}

	// Check for content patterns
	if strings.Contains(pattern, "containing") {
		parts := strings.Split(pattern, "containing")
		if len(parts) > 1 {
			content := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
			operation.Filters = append(operation.Filters, patterns.FileFilter{
				Type:     patterns.PatternContent,
				Operator: patterns.OpContains,
				Value:    content,
			})
		}
	}

	// If no filters were created, return nil
	if len(operation.Filters) == 0 {
		return nil
	}

	return operation
}

// Helper function to display pattern operation results
func showPatternResults(results []patterns.PatternResult, verbose bool, limit int) {
	if len(results) == 0 {
		fmt.Println("🌸 No pattern operations executed")
		return
	}

	totalFiles := 0
	totalProcessed := 0
	totalSkipped := 0
	totalFailed := 0
	totalErrors := 0

	fmt.Println("🌸 Pattern Operation Results (╹◡╹)♡")
	fmt.Println("=====================================")

	for i, result := range results {
		totalFiles += result.FilesMatched
		totalProcessed += result.FilesProcessed
		totalSkipped += result.FilesSkipped
		totalFailed += result.FilesFailed
		totalErrors += len(result.Errors)

		fmt.Printf("Operation %d (%s):\n", i+1, result.OperationID)
		fmt.Printf("  🔍 Files Matched: %d\n", result.FilesMatched)
		fmt.Printf("  ✅ Files Processed: %d\n", result.FilesProcessed)
		fmt.Printf("  ⏭️  Files Skipped: %d\n", result.FilesSkipped)
		fmt.Printf("  ❌ Files Failed: %d\n", result.FilesFailed)
		fmt.Printf("  ⏱️  Duration: %s\n", result.Duration.String())

		// Show summary statistics
		if result.Summary.TotalSize > 0 {
			fmt.Printf("  📊 Total Size: %s\n", formatBytesPattern(result.Summary.TotalSize))
			fmt.Printf("  📊 Average Size: %s\n", formatBytesPattern(result.Summary.TotalSize/int64(result.FilesProcessed)))
		}

		if len(result.Errors) > 0 {
			fmt.Printf("  ❌ Errors: %d\n", len(result.Errors))
			if verbose {
				for _, err := range result.Errors {
					fmt.Printf("    - %s\n", err)
				}
			}
		}

		if verbose && len(result.Details) > 0 {
			fmt.Println("  📋 Details:")
			displayCount := len(result.Details)
			if limit > 0 && limit < displayCount {
				displayCount = limit
			}

			for j := 0; j < displayCount; j++ {
				detail := result.Details[j]
				status := "✅"
				if !detail.Success {
					status = "❌"
				}
				fmt.Printf("    %s %s\n", status, filepath.Base(detail.FilePath))
				if detail.Action != "" {
					fmt.Printf("      Action: %s\n", detail.Action)
				}
				if detail.Destination != "" {
					fmt.Printf("      Destination: %s\n", detail.Destination)
				}
				if detail.Error != "" {
					fmt.Printf("      Error: %s\n", detail.Error)
				}
				fmt.Printf("      Size: %s, Modified: %s\n",
					formatBytesPattern(detail.Size),
					detail.ModifiedTime.Format("2006-01-02 15:04:05"))
			}

			if limit > 0 && len(result.Details) > limit {
				fmt.Printf("    ... and %d more files\n", len(result.Details)-limit)
			}
		}
		fmt.Println()
	}

	fmt.Println("📊 Summary:")
	fmt.Printf("  🔍 Total Files Matched: %d\n", totalFiles)
	fmt.Printf("  ✅ Total Files Processed: %d\n", totalProcessed)
	fmt.Printf("  ⏭️  Total Files Skipped: %d\n", totalSkipped)
	fmt.Printf("  ❌ Total Files Failed: %d\n", totalFailed)
	if totalErrors > 0 {
		fmt.Printf("  ❌ Total Errors: %d\n", totalErrors)
	}
}

// Helper function to format bytes (pattern-specific)
func formatBytesPattern(bytes int64) string {
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

// Helper function to expand paths (handle ~)
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return path // Return original if can't get user home
		}
		return filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}
