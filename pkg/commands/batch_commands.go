/**
 * CLI commands for batch operations system.
 *
 * Provides commands for batch delete, copy, move operations with
 * progress tracking, error handling, and comprehensive logging.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: batch_commands.go
 * Description: Cobra command definitions for batch operations management
 */

package commands

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"ena/internal/batch"

	"github.com/spf13/cobra"
)

// Global batch manager instance
var globalBatchManager *batch.BatchManager

// getGlobalBatchManager returns the global batch manager instance
func getGlobalBatchManager() *batch.BatchManager {
	if globalBatchManager == nil {
		analytics := getGlobalAnalytics()
		globalBatchManager = batch.NewBatchManager(analytics)
	}
	return globalBatchManager
}

// setupBatchCommands adds batch operation commands to the root command
func setupBatchCommands(rootCmd *cobra.Command) {
	// Batch delete command
	batchDeleteCmd := &cobra.Command{
		Use:   "batch-delete <path1> [path2] [path3] ...",
		Short: "Delete multiple files and folders in batch",
		Long: `Delete multiple files and folders efficiently with progress tracking.
Supports wildcards and recursive deletion with comprehensive error handling.

Examples:
  ena batch-delete file1.txt file2.txt folder1/
  ena batch-delete *.tmp --dry-run
  ena batch-delete /tmp/old_files --confirm-each
  ena batch-delete folder1 folder2 --max-concurrency 8`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			batchManager := getGlobalBatchManager()

			// Parse flags
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			confirmEach, _ := cmd.Flags().GetBool("confirm-each")
			maxConcurrency, _ := cmd.Flags().GetInt("max-concurrency")
			skipErrors, _ := cmd.Flags().GetBool("skip-errors")

			// Expand wildcards
			var expandedPaths []string
			for _, arg := range args {
				if strings.Contains(arg, "*") || strings.Contains(arg, "?") {
					matches, err := filepath.Glob(arg)
					if err != nil {
						fmt.Printf("❌ Error expanding pattern %s: %v\n", arg, err)
						continue
					}
					expandedPaths = append(expandedPaths, matches...)
				} else {
					expandedPaths = append(expandedPaths, arg)
				}
			}

			if len(expandedPaths) == 0 {
				fmt.Println("❌ No files or folders found to delete")
				return
			}

			// Create batch config
			config := batch.BatchConfig{
				MaxConcurrency:   maxConcurrency,
				SkipErrors:       skipErrors,
				DryRun:           dryRun,
				ConfirmEach:      confirmEach,
				ProgressInterval: 500 * time.Millisecond,
				RetryCount:       2,
				RetryDelay:       1 * time.Second,
			}

			// Create batch job
			job, err := batchManager.BatchDelete(expandedPaths, config)
			if err != nil {
				fmt.Printf("❌ Error creating batch delete job: %v\n", err)
				return
			}

			fmt.Printf("🌸 Created batch delete job: %s\n", job.Name)
			fmt.Printf("📊 Total items: %d | Total size: %s\n",
				len(job.Operations), formatBytes(job.TotalSize))

			if dryRun {
				fmt.Println("🔍 Dry run mode - no files will be deleted")
			}

			// Execute job
			fmt.Println("🚀 Starting batch delete operation...")
			err = batchManager.ExecuteBatchJob(job.ID)
			if err != nil {
				fmt.Printf("❌ Error executing batch delete: %v\n", err)
				return
			}

			// Show results
			finalJob, _ := batchManager.GetJobStatus(job.ID)
			fmt.Printf("✅ Batch delete completed!\n")
			fmt.Printf("📊 Success: %d | Errors: %d | Skipped: %d\n",
				finalJob.SuccessCount, finalJob.ErrorCount, finalJob.SkippedCount)
			fmt.Printf("⏱️  Duration: %s\n", finalJob.Duration.String())
		},
	}

	// Add flags
	batchDeleteCmd.Flags().Bool("dry-run", false, "Preview what would be deleted without actually deleting")
	batchDeleteCmd.Flags().Bool("confirm-each", false, "Confirm each deletion individually")
	batchDeleteCmd.Flags().Int("max-concurrency", 4, "Maximum concurrent operations")
	batchDeleteCmd.Flags().Bool("skip-errors", true, "Continue on errors")

	// Batch copy command
	batchCopyCmd := &cobra.Command{
		Use:   "batch-copy <source1> [source2] ... <destination>",
		Short: "Copy multiple files and folders recursively",
		Long: `Copy multiple files and folders recursively with progress tracking.
Supports wildcards and preserves file permissions and timestamps.

Examples:
  ena batch-copy file1.txt file2.txt /backup/
  ena batch-copy folder1/ folder2/ /backup/ --recursive
  ena batch-copy *.txt /backup/ --preserve-permissions
  ena batch-copy source/ /dest/ --exclude "*.tmp" --exclude "*.log"`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			batchManager := getGlobalBatchManager()

			// Parse arguments
			sources := args[:len(args)-1]
			destination := args[len(args)-1]

			// Parse flags
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			maxConcurrency, _ := cmd.Flags().GetInt("max-concurrency")
			preservePermissions, _ := cmd.Flags().GetBool("preserve-permissions")
			preserveTimestamps, _ := cmd.Flags().GetBool("preserve-timestamps")
			excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")
			includePatterns, _ := cmd.Flags().GetStringSlice("include")

			// Expand wildcards in sources
			var expandedSources []string
			for _, source := range sources {
				if strings.Contains(source, "*") || strings.Contains(source, "?") {
					matches, err := filepath.Glob(source)
					if err != nil {
						fmt.Printf("❌ Error expanding pattern %s: %v\n", source, err)
						continue
					}
					expandedSources = append(expandedSources, matches...)
				} else {
					expandedSources = append(expandedSources, source)
				}
			}

			if len(expandedSources) == 0 {
				fmt.Println("❌ No source files or folders found")
				return
			}

			// Create batch config
			config := batch.BatchConfig{
				MaxConcurrency:      maxConcurrency,
				SkipErrors:          true,
				DryRun:              dryRun,
				ConfirmEach:         false,
				ProgressInterval:    500 * time.Millisecond,
				RetryCount:          2,
				RetryDelay:          1 * time.Second,
				PreserveTimestamps:  preserveTimestamps,
				PreservePermissions: preservePermissions,
				FollowSymlinks:      false,
				ExcludePatterns:     excludePatterns,
				IncludePatterns:     includePatterns,
			}

			// Create batch job
			job, err := batchManager.BatchCopy(expandedSources, destination, config)
			if err != nil {
				fmt.Printf("❌ Error creating batch copy job: %v\n", err)
				return
			}

			fmt.Printf("🌸 Created batch copy job: %s\n", job.Name)
			fmt.Printf("📊 Total items: %d | Total size: %s\n",
				len(job.Operations), formatBytes(job.TotalSize))
			fmt.Printf("📁 Destination: %s\n", destination)

			if dryRun {
				fmt.Println("🔍 Dry run mode - no files will be copied")
			}

			// Execute job
			fmt.Println("🚀 Starting batch copy operation...")
			err = batchManager.ExecuteBatchJob(job.ID)
			if err != nil {
				fmt.Printf("❌ Error executing batch copy: %v\n", err)
				return
			}

			// Show results
			finalJob, _ := batchManager.GetJobStatus(job.ID)
			fmt.Printf("✅ Batch copy completed!\n")
			fmt.Printf("📊 Success: %d | Errors: %d | Skipped: %d\n",
				finalJob.SuccessCount, finalJob.ErrorCount, finalJob.SkippedCount)
			fmt.Printf("⏱️  Duration: %s\n", finalJob.Duration.String())
		},
	}

	// Add flags
	batchCopyCmd.Flags().Bool("dry-run", false, "Preview what would be copied without actually copying")
	batchCopyCmd.Flags().Int("max-concurrency", 4, "Maximum concurrent operations")
	batchCopyCmd.Flags().Bool("preserve-permissions", true, "Preserve file permissions")
	batchCopyCmd.Flags().Bool("preserve-timestamps", true, "Preserve file timestamps")
	batchCopyCmd.Flags().StringSlice("exclude", []string{}, "Exclude patterns (e.g., *.tmp, *.log)")
	batchCopyCmd.Flags().StringSlice("include", []string{}, "Include patterns (e.g., *.txt, *.go)")

	// Batch move command
	batchMoveCmd := &cobra.Command{
		Use:   "batch-move <source1> [source2] ... <destination>",
		Short: "Move multiple files and folders",
		Long: `Move multiple files and folders efficiently with progress tracking.
Uses rename when possible, falls back to copy+delete for cross-filesystem moves.

Examples:
  ena batch-move file1.txt file2.txt /new_location/
  ena batch-move folder1/ folder2/ /new_location/
  ena batch-move *.tmp /trash/ --dry-run`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			batchManager := getGlobalBatchManager()

			// Parse arguments
			sources := args[:len(args)-1]
			destination := args[len(args)-1]

			// Parse flags
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			maxConcurrency, _ := cmd.Flags().GetInt("max-concurrency")
			preservePermissions, _ := cmd.Flags().GetBool("preserve-permissions")
			preserveTimestamps, _ := cmd.Flags().GetBool("preserve-timestamps")

			// Expand wildcards in sources
			var expandedSources []string
			for _, source := range sources {
				if strings.Contains(source, "*") || strings.Contains(source, "?") {
					matches, err := filepath.Glob(source)
					if err != nil {
						fmt.Printf("❌ Error expanding pattern %s: %v\n", source, err)
						continue
					}
					expandedSources = append(expandedSources, matches...)
				} else {
					expandedSources = append(expandedSources, source)
				}
			}

			if len(expandedSources) == 0 {
				fmt.Println("❌ No source files or folders found")
				return
			}

			// Create batch config
			config := batch.BatchConfig{
				MaxConcurrency:      maxConcurrency,
				SkipErrors:          true,
				DryRun:              dryRun,
				ConfirmEach:         false,
				ProgressInterval:    500 * time.Millisecond,
				RetryCount:          2,
				RetryDelay:          1 * time.Second,
				PreserveTimestamps:  preserveTimestamps,
				PreservePermissions: preservePermissions,
				FollowSymlinks:      false,
				ExcludePatterns:     []string{},
				IncludePatterns:     []string{},
			}

			// Create batch job
			job, err := batchManager.BatchMove(expandedSources, destination, config)
			if err != nil {
				fmt.Printf("❌ Error creating batch move job: %v\n", err)
				return
			}

			fmt.Printf("🌸 Created batch move job: %s\n", job.Name)
			fmt.Printf("📊 Total items: %d | Total size: %s\n",
				len(job.Operations), formatBytes(job.TotalSize))
			fmt.Printf("📁 Destination: %s\n", destination)

			if dryRun {
				fmt.Println("🔍 Dry run mode - no files will be moved")
			}

			// Execute job
			fmt.Println("🚀 Starting batch move operation...")
			err = batchManager.ExecuteBatchJob(job.ID)
			if err != nil {
				fmt.Printf("❌ Error executing batch move: %v\n", err)
				return
			}

			// Show results
			finalJob, _ := batchManager.GetJobStatus(job.ID)
			fmt.Printf("✅ Batch move completed!\n")
			fmt.Printf("📊 Success: %d | Errors: %d | Skipped: %d\n",
				finalJob.SuccessCount, finalJob.ErrorCount, finalJob.SkippedCount)
			fmt.Printf("⏱️  Duration: %s\n", finalJob.Duration.String())
		},
	}

	// Add flags
	batchMoveCmd.Flags().Bool("dry-run", false, "Preview what would be moved without actually moving")
	batchMoveCmd.Flags().Int("max-concurrency", 4, "Maximum concurrent operations")
	batchMoveCmd.Flags().Bool("preserve-permissions", true, "Preserve file permissions")
	batchMoveCmd.Flags().Bool("preserve-timestamps", true, "Preserve file timestamps")

	// Batch status command
	batchStatusCmd := &cobra.Command{
		Use:   "batch-status [job-id]",
		Short: "Show status of batch operations",
		Long: `Show status of batch operations. If job-id is provided,
shows detailed status of that specific job. Otherwise shows all jobs.

Examples:
  ena batch-status
  ena batch-status batch_1234567890`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			batchManager := getGlobalBatchManager()

			if len(args) == 1 {
				// Show specific job
				jobID := args[0]
				job, err := batchManager.GetJobStatus(jobID)
				if err != nil {
					fmt.Printf("❌ Error getting job status: %v\n", err)
					return
				}

				showJobDetails(job)
			} else {
				// Show all jobs
				jobs := batchManager.ListJobs()
				if len(jobs) == 0 {
					fmt.Println("🌸 No batch jobs found")
					return
				}

				fmt.Println("🌸 Batch Jobs Status (╹◡╹)♡")
				fmt.Println("================================")

				for i, job := range jobs {
					fmt.Printf("%d. %s (%s)\n", i+1, job.Name, job.Status)
					fmt.Printf("   📊 Progress: %.1f%% | Items: %d/%d\n",
						job.Progress*100, job.SuccessCount+job.ErrorCount, len(job.Operations))
					fmt.Printf("   ⏱️  Duration: %s\n", job.Duration.String())
					fmt.Println()
				}
			}
		},
	}

	// Batch cancel command
	batchCancelCmd := &cobra.Command{
		Use:   "batch-cancel <job-id>",
		Short: "Cancel a running batch operation",
		Long: `Cancel a running batch operation. This will stop the operation
and mark it as cancelled.

Examples:
  ena batch-cancel batch_1234567890`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			batchManager := getGlobalBatchManager()

			jobID := args[0]
			err := batchManager.CancelJob(jobID)
			if err != nil {
				fmt.Printf("❌ Error cancelling job: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully cancelled batch job: %s\n", jobID)
		},
	}

	// Add all commands to root
	rootCmd.AddCommand(batchDeleteCmd)
	rootCmd.AddCommand(batchCopyCmd)
	rootCmd.AddCommand(batchMoveCmd)
	rootCmd.AddCommand(batchStatusCmd)
	rootCmd.AddCommand(batchCancelCmd)
}

// Helper functions

func showJobDetails(job *batch.BatchJob) {
	fmt.Printf("🌸 Batch Job Details: %s (╹◡╹)♡\n", job.Name)
	fmt.Println("=====================================")
	fmt.Printf("📋 Description: %s\n", job.Description)
	fmt.Printf("🆔 Job ID: %s\n", job.ID)
	fmt.Printf("📊 Status: %s\n", job.Status)
	fmt.Printf("📈 Progress: %.1f%%\n", job.Progress*100)
	fmt.Printf("📁 Total Items: %d\n", len(job.Operations))
	fmt.Printf("✅ Success: %d\n", job.SuccessCount)
	fmt.Printf("❌ Errors: %d\n", job.ErrorCount)
	fmt.Printf("⏭️  Skipped: %d\n", job.SkippedCount)
	fmt.Printf("💾 Total Size: %s\n", formatBytes(job.TotalSize))
	fmt.Printf("💾 Processed: %s\n", formatBytes(job.ProcessedSize))
	fmt.Printf("⏱️  Duration: %s\n", job.Duration.String())

	if !job.StartTime.IsZero() {
		fmt.Printf("🚀 Started: %s\n", job.StartTime.Format("2006-01-02 15:04:05"))
	}
	if !job.EndTime.IsZero() {
		fmt.Printf("🏁 Ended: %s\n", job.EndTime.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n📋 Operations:")
	for i, op := range job.Operations {
		if i >= 10 { // Limit display
			fmt.Printf("   ... and %d more operations\n", len(job.Operations)-10)
			break
		}
		fmt.Printf("   %d. %s %s (%s)\n", i+1, op.Type, op.Source, op.Status)
		if op.Error != "" {
			fmt.Printf("      ❌ Error: %s\n", op.Error)
		}
	}
}

func formatBytes(bytes int64) string {
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
