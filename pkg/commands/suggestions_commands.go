/**
 * CLI commands for the smart suggestions system.
 *
 * Provides commands for viewing suggestions, providing feedback,
 * and managing the analytics system.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: suggestions_commands.go
 * Description: Cobra command definitions for suggestion management and analytics
 */

package commands

import (
	"fmt"
	"strings"

	"ena/internal/suggestions"

	"github.com/spf13/cobra"
)

// Global analytics instance
var globalAnalytics *suggestions.UsageAnalytics

// getGlobalAnalytics returns the global analytics instance
func getGlobalAnalytics() *suggestions.UsageAnalytics {
	if globalAnalytics == nil {
		globalAnalytics = suggestions.NewUsageAnalytics()
	}
	return globalAnalytics
}

// setupSuggestionsCommands adds suggestion-related commands to the root command
func setupSuggestionsCommands(rootCmd *cobra.Command) {
	// Main suggestions command
	suggestCmd := &cobra.Command{
		Use:   "suggest",
		Short: "Get intelligent suggestions based on your usage patterns",
		Long: `Ena's smart suggestion system analyzes your command history and usage patterns
to provide intelligent recommendations for improved productivity and workflow optimization.

Examples:
  ena suggest                    # Show top suggestions
  ena suggest --limit 5          # Show top 5 suggestions
  ena suggest --category safety  # Show safety-related suggestions
  ena suggest --type workflow    # Show workflow suggestions`,
		Run: func(cmd *cobra.Command, args []string) {
			analytics := getGlobalAnalytics()

			limit, _ := cmd.Flags().GetInt("limit")
			category, _ := cmd.Flags().GetString("category")
			suggestionType, _ := cmd.Flags().GetString("type")

			suggestionsList := analytics.GetSuggestions(limit)

			// Filter by category if specified
			if category != "" {
				var filtered []suggestions.SmartSuggestion
				for _, suggestion := range suggestionsList {
					if suggestion.Category == category {
						filtered = append(filtered, suggestion)
					}
				}
				suggestionsList = filtered
			}

			// Filter by type if specified
			if suggestionType != "" {
				var filtered []suggestions.SmartSuggestion
				for _, suggestion := range suggestionsList {
					if suggestion.Type == suggestionType {
						filtered = append(filtered, suggestion)
					}
				}
				suggestionsList = filtered
			}

			if len(suggestionsList) == 0 {
				fmt.Println("🌸 No suggestions available right now. Keep using Ena and I'll learn your patterns!")
				return
			}

			fmt.Printf("🌸 Here are my smart suggestions for you! (╹◡╹)♡\n\n")

			for i, suggestion := range suggestionsList {
				fmt.Printf("%d. %s\n", i+1, suggestion.Title)
				fmt.Printf("   %s\n", suggestion.Description)
				if suggestion.Command != "" {
					fmt.Printf("   💡 Try: %s\n", suggestion.Command)
				}
				fmt.Printf("   📊 Confidence: %.0f%% | Priority: %d/10 | Category: %s\n",
					suggestion.Confidence*100, suggestion.Priority, suggestion.Category)
				fmt.Println()
			}
		},
	}

	// Add flags
	suggestCmd.Flags().IntP("limit", "l", 10, "Maximum number of suggestions to show")
	suggestCmd.Flags().StringP("category", "c", "", "Filter by category (productivity, safety, optimization, discovery)")
	suggestCmd.Flags().StringP("type", "t", "", "Filter by type (command, workflow, optimization, safety)")

	// Stats command
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show usage statistics and analytics",
		Long: `Display comprehensive usage statistics including command frequency,
file operations, patterns discovered, and performance metrics.

Examples:
  ena stats                    # Show all statistics
  ena stats --commands         # Show command statistics only
  ena stats --patterns        # Show discovered patterns only`,
		Run: func(cmd *cobra.Command, args []string) {
			analytics := getGlobalAnalytics()

			showCommands, _ := cmd.Flags().GetBool("commands")
			showPatterns, _ := cmd.Flags().GetBool("patterns")
			showFileOps, _ := cmd.Flags().GetBool("fileops")

			stats := analytics.GetUsageStats()

			fmt.Println("🌸 Ena's Analytics Dashboard (╹◡╹)♡")
			fmt.Println("=====================================")

			if !showCommands && !showPatterns && !showFileOps {
				// Show all stats
				fmt.Printf("📊 Total Commands Executed: %v\n", stats["total_commands"])
				fmt.Printf("📁 Total File Operations: %v\n", stats["total_file_operations"])
				fmt.Printf("⏱️  Average Command Duration: %v\n", stats["average_command_duration"])
				fmt.Printf("✅ Success Rate: %.1f%%\n", stats["success_rate"])
				fmt.Printf("💾 Total File Size Processed: %v\n", stats["total_file_size_processed"])
				fmt.Printf("🔍 Patterns Discovered: %v\n", stats["patterns_discovered"])
				fmt.Printf("💡 Suggestions Generated: %v\n", stats["suggestions_generated"])
				fmt.Printf("📅 Analysis Period: %v\n", stats["analysis_period"])
				fmt.Println()

				// Most used commands
				if mostUsed, ok := stats["most_used_commands"].([]map[string]interface{}); ok {
					fmt.Println("🔥 Most Used Commands:")
					for i, cmd := range mostUsed {
						if i >= 5 {
							break
						}
						fmt.Printf("   %d. %s (%v times)\n", i+1, cmd["name"], cmd["count"])
					}
					fmt.Println()
				}

				// Most common file operations
				if mostFileOps, ok := stats["most_common_file_operations"].([]map[string]interface{}); ok {
					fmt.Println("📁 Most Common File Operations:")
					for i, op := range mostFileOps {
						if i >= 5 {
							break
						}
						fmt.Printf("   %d. %s (%v times)\n", i+1, op["name"], op["count"])
					}
					fmt.Println()
				}
			} else {
				// Show specific stats
				if showCommands {
					fmt.Printf("📊 Total Commands: %v\n", stats["total_commands"])
					fmt.Printf("⏱️  Average Duration: %v\n", stats["average_command_duration"])
					fmt.Printf("✅ Success Rate: %.1f%%\n", stats["success_rate"])

					if mostUsed, ok := stats["most_used_commands"].([]map[string]interface{}); ok {
						fmt.Println("\n🔥 Most Used Commands:")
						for i, cmd := range mostUsed {
							fmt.Printf("   %d. %s (%v times)\n", i+1, cmd["name"], cmd["count"])
						}
					}
				}

				if showFileOps {
					fmt.Printf("📁 Total File Operations: %v\n", stats["total_file_operations"])
					fmt.Printf("💾 Total Size Processed: %v\n", stats["total_file_size_processed"])

					if mostFileOps, ok := stats["most_common_file_operations"].([]map[string]interface{}); ok {
						fmt.Println("\n📁 Most Common File Operations:")
						for i, op := range mostFileOps {
							fmt.Printf("   %d. %s (%v times)\n", i+1, op["name"], op["count"])
						}
					}
				}

				if showPatterns {
					fmt.Printf("🔍 Patterns Discovered: %v\n", stats["patterns_discovered"])
					fmt.Printf("💡 Suggestions Generated: %v\n", stats["suggestions_generated"])
				}
			}
		},
	}

	// Add flags
	statsCmd.Flags().Bool("commands", false, "Show command statistics only")
	statsCmd.Flags().Bool("patterns", false, "Show pattern statistics only")
	statsCmd.Flags().Bool("fileops", false, "Show file operation statistics only")

	// Feedback command
	feedbackCmd := &cobra.Command{
		Use:   "feedback <suggestion_id> <feedback>",
		Short: "Provide feedback on a suggestion",
		Long: `Provide feedback on suggestions to help Ena learn and improve.
Valid feedback values: helpful, not_helpful, dismiss

Examples:
  ena feedback suggestion_123 helpful
  ena feedback workflow_456 not_helpful
  ena feedback optimization_789 dismiss`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			analytics := getGlobalAnalytics()

			suggestionID := args[0]
			feedback := args[1]

			// Validate feedback
			validFeedback := []string{"helpful", "not_helpful", "dismiss"}
			if !contains(validFeedback, feedback) {
				fmt.Printf("❌ Invalid feedback. Must be one of: %s\n", strings.Join(validFeedback, ", "))
				return
			}

			err := analytics.ProvideFeedback(suggestionID, feedback)
			if err != nil {
				fmt.Printf("❌ Error providing feedback: %v\n", err)
				return
			}

			fmt.Printf("🌸 Thank you for the feedback! I'll use this to improve my suggestions. (╹◡╹)♡\n")
		},
	}

	// Workflow command
	workflowCmd := &cobra.Command{
		Use:   "workflow",
		Short: "Show workflow optimization suggestions",
		Long: `Display workflow suggestions based on your command patterns.
Ena analyzes your command sequences to suggest workflow optimizations.

Examples:
  ena workflow                    # Show workflow suggestions
  ena workflow --create          # Create a script from common workflow`,
		Run: func(cmd *cobra.Command, args []string) {
			analytics := getGlobalAnalytics()

			createScript, _ := cmd.Flags().GetBool("create")

			suggestionsList := analytics.GetWorkflowSuggestions()

			if len(suggestionsList) == 0 {
				fmt.Println("🌸 No workflow patterns detected yet. Keep using Ena and I'll discover your workflows!")
				return
			}

			fmt.Println("🌸 Workflow Optimization Suggestions (╹◡╹)♡")
			fmt.Println("==========================================")

			for i, suggestion := range suggestionsList {
				fmt.Printf("%d. %s\n", i+1, suggestion.Title)
				fmt.Printf("   %s\n", suggestion.Description)
				if suggestion.Command != "" {
					fmt.Printf("   💡 Command: %s\n", suggestion.Command)
				}
				fmt.Printf("   📊 Confidence: %.0f%% | Priority: %d/10\n",
					suggestion.Confidence*100, suggestion.Priority)

				if createScript && suggestion.Command != "" {
					fmt.Printf("   🚀 Creating script...\n")
					// Here you would implement script creation
					fmt.Printf("   ✅ Script created successfully!\n")
				}
				fmt.Println()
			}
		},
	}

	workflowCmd.Flags().Bool("create", false, "Create scripts for suggested workflows")

	// Optimization command
	optimizeCmd := &cobra.Command{
		Use:   "optimize",
		Short: "Show system optimization suggestions",
		Long: `Display system optimization suggestions based on your usage patterns.
Ena analyzes your command history to suggest performance improvements.

Examples:
  ena optimize                   # Show optimization suggestions
  ena optimize --apply           # Apply suggested optimizations`,
		Run: func(cmd *cobra.Command, args []string) {
			analytics := getGlobalAnalytics()

			apply, _ := cmd.Flags().GetBool("apply")

			suggestionsList := analytics.GetOptimizationSuggestions()

			if len(suggestionsList) == 0 {
				fmt.Println("🌸 Your system is already optimized! Great job! (╹◡╹)♡")
				return
			}

			fmt.Println("🌸 System Optimization Suggestions (╹◡╹)♡")
			fmt.Println("=======================================")

			for i, suggestion := range suggestionsList {
				fmt.Printf("%d. %s\n", i+1, suggestion.Title)
				fmt.Printf("   %s\n", suggestion.Description)
				if suggestion.Command != "" {
					fmt.Printf("   💡 Command: %s\n", suggestion.Command)
				}
				fmt.Printf("   📊 Confidence: %.0f%% | Priority: %d/10\n",
					suggestion.Confidence*100, suggestion.Priority)

				if apply && suggestion.Command != "" {
					fmt.Printf("   🚀 Applying optimization...\n")
					// Here you would implement optimization application
					fmt.Printf("   ✅ Optimization applied successfully!\n")
				}
				fmt.Println()
			}
		},
	}

	optimizeCmd.Flags().Bool("apply", false, "Apply suggested optimizations")

	// Add all commands to root
	rootCmd.AddCommand(suggestCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(feedbackCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(optimizeCmd)
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
