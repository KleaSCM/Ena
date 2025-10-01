/**
 * CLI commands for the smart file organization system.
 *
 * Provides commands for managing organization rules, organizing files,
 * and configuring automatic file sorting and management.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: organizer_commands.go
 * Description: Cobra command definitions for file organization management
 */

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ena/internal/organizer"

	"github.com/spf13/cobra"
)

// Type aliases to avoid import issues
type OrganizationRule = organizer.OrganizationRule
type OrganizationResult = organizer.OrganizationResult
type RuleAction = organizer.RuleAction

// Global file organizer instance
var globalFileOrganizer *organizer.FileOrganizer

// getGlobalFileOrganizer returns the global file organizer instance
func getGlobalFileOrganizer() *organizer.FileOrganizer {
	if globalFileOrganizer == nil {
		analytics := getGlobalAnalytics()
		globalFileOrganizer = organizer.NewFileOrganizer(analytics)
	}
	return globalFileOrganizer
}

// setupOrganizerCommands adds file organization commands to the root command
func setupOrganizerCommands(rootCmd *cobra.Command) {
	// Get file organizer from system hooks
	organizer := getGlobalFileOrganizer()

	// Organize command
	organizeCmd := &cobra.Command{
		Use:   "organize <path> [paths...]",
		Short: "Organize files using smart organization rules",
		Long: `Organize files in specified directories using configured organization rules.
Files will be automatically sorted, moved, copied, or renamed based on their type and rules.

Examples:
  ena organize ~/Downloads                    # Organize downloads folder
  ena organize ~/Desktop ~/Documents          # Organize multiple folders
  ena organize ~/Downloads --dry-run          # Preview organization without making changes`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Validate paths
			var validPaths []string
			for _, path := range args {
				if _, err := os.Stat(path); err != nil {
					fmt.Printf("⚠️ Warning: Path %s does not exist, skipping\n", path)
					continue
				}
				validPaths = append(validPaths, path)
			}

			if len(validPaths) == 0 {
				fmt.Println("❌ No valid paths provided")
				return
			}

			fmt.Printf("🌸 Starting file organization for %d path(s)...\n", len(validPaths))
			if dryRun {
				fmt.Println("🔍 Dry run mode - no files will be moved")
			}

			// Organize files
			results, err := organizer.OrganizeFiles(validPaths, dryRun)
			if err != nil {
				fmt.Printf("❌ Error organizing files: %v\n", err)
				return
			}

			// Display results
			showOrganizationResults(results, verbose)
		},
	}

	organizeCmd.Flags().Bool("dry-run", false, "Preview organization without making changes")
	organizeCmd.Flags().Bool("verbose", false, "Show detailed information about each file operation")

	// Add rule command
	addRuleCmd := &cobra.Command{
		Use:   "add-rule <name>",
		Short: "Add a new organization rule",
		Long: `Add a new organization rule for automatic file sorting.
This command will prompt for rule configuration details.

Examples:
  ena add-rule "Documents Organization"
  ena add-rule "Image Sorting"`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ruleName := args[0]

			// Create a basic rule
			rule := &OrganizationRule{
				Name:        ruleName,
				Description: fmt.Sprintf("Organization rule for %s", ruleName),
				Enabled:     true,
				Priority:    10,
				SourcePaths: []string{"~/Downloads"},
				DestPath:    "~/Documents/{category}/{date}",
				FileTypes:   []string{"documents"},
				Actions: []RuleAction{
					{
						Type:        "move",
						Destination: "~/Documents/{category}/{date}",
					},
				},
			}

			err := organizer.AddRule(rule)
			if err != nil {
				fmt.Printf("❌ Error adding rule: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully added organization rule: %s\n", rule.Name)
			fmt.Printf("🆔 Rule ID: %s\n", rule.ID)
		},
	}

	// List rules command
	listRulesCmd := &cobra.Command{
		Use:   "list-rules",
		Short: "List all organization rules",
		Long: `List all configured organization rules with their details.
Shows rule names, priorities, file types, and status.

Examples:
  ena list-rules
  ena list-rules --enabled-only`,
		Run: func(cmd *cobra.Command, args []string) {
			enabledOnly, _ := cmd.Flags().GetBool("enabled-only")

			rules := organizer.GetRules()
			if len(rules) == 0 {
				fmt.Println("🌸 No organization rules configured")
				return
			}

			fmt.Println("🌸 Organization Rules (╹◡╹)♡")
			fmt.Println("================================")

			for i, rule := range rules {
				if enabledOnly && !rule.Enabled {
					continue
				}

				status := "❌ Disabled"
				if rule.Enabled {
					status = "✅ Enabled"
				}

				fmt.Printf("%d. %s (%s)\n", i+1, rule.Name, status)
				fmt.Printf("   📝 Description: %s\n", rule.Description)
				fmt.Printf("   🆔 ID: %s\n", rule.ID)
				fmt.Printf("   📊 Priority: %d\n", rule.Priority)
				fmt.Printf("   📁 Source Paths: %s\n", strings.Join(rule.SourcePaths, ", "))
				fmt.Printf("   📂 Destination: %s\n", rule.DestPath)
				fmt.Printf("   🏷️  File Types: %s\n", strings.Join(rule.FileTypes, ", "))
				fmt.Printf("   📅 Created: %s\n", rule.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Println()
			}
		},
	}

	listRulesCmd.Flags().Bool("enabled-only", false, "Show only enabled rules")

	// Get watched paths command
	watchedPathsCmd := &cobra.Command{
		Use:   "watched-paths",
		Short: "Show paths being watched by organization rules",
		Run: func(cmd *cobra.Command, args []string) {
			paths := organizer.GetWatchedPaths()
			if len(paths) == 0 {
				fmt.Println("🌸 No paths being watched")
				return
			}

			fmt.Println("🌸 Watched Paths (╹◡╹)♡")
			fmt.Println("========================")
			for i, path := range paths {
				fmt.Printf("%d. %s\n", i+1, path)
			}
		},
	}

	// Get file extensions command
	extensionsCmd := &cobra.Command{
		Use:   "file-extensions",
		Short: "Show all supported file extensions",
		Run: func(cmd *cobra.Command, args []string) {
			extensions := organizer.GetAllFileExtensions()
			if len(extensions) == 0 {
				fmt.Println("🌸 No file extensions configured")
				return
			}

			fmt.Println("🌸 Supported File Extensions (╹◡╹)♡")
			fmt.Println("====================================")
			for i, ext := range extensions {
				fmt.Printf("%d. %s\n", i+1, ext)
			}
		},
	}

	// Add all commands to root
	rootCmd.AddCommand(organizeCmd)
	rootCmd.AddCommand(addRuleCmd)
	rootCmd.AddCommand(listRulesCmd)
	rootCmd.AddCommand(watchedPathsCmd)
	rootCmd.AddCommand(extensionsCmd)
}

// Helper function to display organization results
func showOrganizationResults(results []OrganizationResult, verbose bool) {
	if len(results) == 0 {
		fmt.Println("🌸 No organization rules applied")
		return
	}

	totalFiles := 0
	totalMoved := 0
	totalCopied := 0
	totalRenamed := 0
	totalDeleted := 0
	totalErrors := 0

	fmt.Println("🌸 Organization Results (╹◡╹)♡")
	fmt.Println("================================")

	for i, result := range results {
		totalFiles += result.FilesProcessed
		totalMoved += result.FilesMoved
		totalCopied += result.FilesCopied
		totalRenamed += result.FilesRenamed
		totalDeleted += result.FilesDeleted
		totalErrors += len(result.Errors)

		fmt.Printf("Rule %d (%s):\n", i+1, result.RuleID)
		fmt.Printf("  📊 Files Processed: %d\n", result.FilesProcessed)
		fmt.Printf("  📁 Files Moved: %d\n", result.FilesMoved)
		fmt.Printf("  📋 Files Copied: %d\n", result.FilesCopied)
		fmt.Printf("  ✏️  Files Renamed: %d\n", result.FilesRenamed)
		fmt.Printf("  🗑️  Files Deleted: %d\n", result.FilesDeleted)
		fmt.Printf("  ⏱️  Duration: %s\n", result.Duration.String())

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
			for _, detail := range result.Details {
				status := "✅"
				if !detail.Success {
					status = "❌"
				}
				fmt.Printf("    %s %s -> %s\n", status, detail.Action, filepath.Base(detail.FilePath))
				if detail.Destination != "" {
					fmt.Printf("      📂 %s\n", detail.Destination)
				}
				if detail.Error != "" {
					fmt.Printf("      ❌ %s\n", detail.Error)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("📊 Summary:")
	fmt.Printf("  📁 Total Files Processed: %d\n", totalFiles)
	fmt.Printf("  📁 Total Files Moved: %d\n", totalMoved)
	fmt.Printf("  📋 Total Files Copied: %d\n", totalCopied)
	fmt.Printf("  ✏️  Total Files Renamed: %d\n", totalRenamed)
	fmt.Printf("  🗑️  Total Files Deleted: %d\n", totalDeleted)
	if totalErrors > 0 {
		fmt.Printf("  ❌ Total Errors: %d\n", totalErrors)
	}
}
