/**
 * CLI commands for the undo system.
 *
 * Provides commands for managing undo history, undoing operations,
 * and restoring files with comprehensive safety checks.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: undo_commands.go
 * Description: Cobra command definitions for undo operations management
 */

package commands

import (
	"fmt"
	"strings"
	"time"

	"ena/internal/undo"

	"github.com/spf13/cobra"
)

// Global undo manager instance
var globalUndoManager *undo.UndoManager

// getGlobalUndoManager returns the global undo manager instance
func getGlobalUndoManager() *undo.UndoManager {
	if globalUndoManager == nil {
		analytics := getGlobalAnalytics()
		globalUndoManager = undo.NewUndoManager(analytics)
	}
	return globalUndoManager
}

// setupUndoCommands adds undo-related commands to the root command
func setupUndoCommands(rootCmd *cobra.Command) {
	// Undo history command
	undoHistoryCmd := &cobra.Command{
		Use:   "undo-history",
		Short: "Show undo history and available operations",
		Long: `Display the undo history showing all tracked operations and sessions.
Shows operations that can be undone with detailed information.

Examples:
  ena undo-history                    # Show all history
  ena undo-history --limit 10         # Show last 10 sessions
  ena undo-history --session <id>     # Show specific session details`,
		Run: func(cmd *cobra.Command, args []string) {
			undoManager := getGlobalUndoManager()

			limit, _ := cmd.Flags().GetInt("limit")
			sessionID, _ := cmd.Flags().GetString("session")

			if sessionID != "" {
				// Show specific session
				session, err := undoManager.GetSession(sessionID)
				if err != nil {
					fmt.Printf("❌ Error getting session: %v\n", err)
					return
				}
				showSessionDetails(session)
			} else {
				// Show all history
				sessions := undoManager.GetHistory()

				if len(sessions) == 0 {
					fmt.Println("🌸 No undo history available")
					return
				}

				fmt.Println("🌸 Undo History (╹◡╹)♡")
				fmt.Println("========================")

				displayCount := len(sessions)
				if limit > 0 && limit < len(sessions) {
					displayCount = limit
				}

				for i := 0; i < displayCount; i++ {
					session := sessions[i]
					fmt.Printf("%d. %s (%s)\n", i+1, session.Name, session.ID)
					fmt.Printf("   📅 Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
					fmt.Printf("   📊 Operations: %d | Undone: %t\n", len(session.Operations), session.Undone)
					if session.Description != "" {
						fmt.Printf("   📝 Description: %s\n", session.Description)
					}
					fmt.Println()
				}
			}
		},
	}

	// Add flags
	undoHistoryCmd.Flags().Int("limit", 0, "Limit number of sessions to show")
	undoHistoryCmd.Flags().String("session", "", "Show specific session details")

	// Undo operation command
	undoOpCmd := &cobra.Command{
		Use:   "undo-operation <operation-id>",
		Short: "Undo a specific operation",
		Long: `Undo a specific operation by its ID. This will restore the file
to its previous state before the operation was performed.

Examples:
  ena undo-operation op_1234567890
  ena undo-operation op_1234567890 --dry-run`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			undoManager := getGlobalUndoManager()

			operationID := args[0]
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			if dryRun {
				fmt.Printf("🔍 Dry run: Would undo operation %s\n", operationID)
				return
			}

			err := undoManager.UndoOperation(operationID)
			if err != nil {
				fmt.Printf("❌ Error undoing operation: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully undone operation: %s\n", operationID)
		},
	}

	undoOpCmd.Flags().Bool("dry-run", false, "Preview what would be undone without actually undoing")

	// Undo session command
	undoSessionCmd := &cobra.Command{
		Use:   "undo-session <session-id>",
		Short: "Undo all operations in a session",
		Long: `Undo all operations in a session. This will restore all files
in the session to their previous states.

Examples:
  ena undo-session session_1234567890
  ena undo-session session_1234567890 --dry-run`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			undoManager := getGlobalUndoManager()

			sessionID := args[0]
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			if dryRun {
				fmt.Printf("🔍 Dry run: Would undo session %s\n", sessionID)
				return
			}

			err := undoManager.UndoSession(sessionID)
			if err != nil {
				fmt.Printf("❌ Error undoing session: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully undone session: %s\n", sessionID)
		},
	}

	undoSessionCmd.Flags().Bool("dry-run", false, "Preview what would be undone without actually undoing")

	// Start session command
	startSessionCmd := &cobra.Command{
		Use:   "start-session <name> [description]",
		Short: "Start a new undo session",
		Long: `Start a new undo session to group related operations.
All subsequent operations will be tracked in this session.

Examples:
  ena start-session "File cleanup"
  ena start-session "Backup creation" "Creating backup of important files"`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			undoManager := getGlobalUndoManager()

			name := args[0]
			description := ""
			if len(args) > 1 {
				description = strings.Join(args[1:], " ")
			}

			session := undoManager.StartSession(name, description)

			fmt.Printf("🌸 Started undo session: %s\n", session.Name)
			fmt.Printf("🆔 Session ID: %s\n", session.ID)
			if description != "" {
				fmt.Printf("📝 Description: %s\n", description)
			}
		},
	}

	// End session command
	endSessionCmd := &cobra.Command{
		Use:   "end-session",
		Short: "End the current undo session",
		Long: `End the current undo session. This will stop tracking
operations in the current session.

Examples:
  ena end-session`,
		Run: func(cmd *cobra.Command, args []string) {
			undoManager := getGlobalUndoManager()

			undoManager.EndSession()
			fmt.Println("🌸 Ended current undo session")
		},
	}

	// Clear history command
	clearHistoryCmd := &cobra.Command{
		Use:   "clear-undo-history [older-than]",
		Short: "Clear old undo history",
		Long: `Clear undo history older than the specified duration.
This will permanently remove old undo data and backup files.

Examples:
  ena clear-undo-history 24h          # Clear history older than 24 hours
  ena clear-undo-history 7d           # Clear history older than 7 days
  ena clear-undo-history --all       # Clear all history`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			undoManager := getGlobalUndoManager()

			clearAll, _ := cmd.Flags().GetBool("all")

			var duration time.Duration
			if clearAll {
				duration = 0 // Clear everything
			} else if len(args) > 0 {
				var err error
				duration, err = time.ParseDuration(args[0])
				if err != nil {
					fmt.Printf("❌ Invalid duration: %v\n", err)
					return
				}
			} else {
				duration = 24 * time.Hour // Default to 24 hours
			}

			err := undoManager.ClearHistory(duration)
			if err != nil {
				fmt.Printf("❌ Error clearing history: %v\n", err)
				return
			}

			if clearAll {
				fmt.Println("✅ Cleared all undo history")
			} else {
				fmt.Printf("✅ Cleared undo history older than %s\n", duration.String())
			}
		},
	}

	clearHistoryCmd.Flags().Bool("all", false, "Clear all undo history")

	// Restore file command
	restoreFileCmd := &cobra.Command{
		Use:   "restore-file <file-path>",
		Short: "Restore a file from undo history",
		Long: `Restore a file from the most recent backup in undo history.
This will find the most recent operation that affected the file
and restore it from backup.

Examples:
  ena restore-file /path/to/file.txt
  ena restore-file /path/to/file.txt --dry-run`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			undoManager := getGlobalUndoManager()

			filePath := args[0]
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			// Find the most recent operation for this file
			sessions := undoManager.GetHistory()
			var latestOperation *undo.UndoOperation

			for _, session := range sessions {
				for _, operation := range session.Operations {
					if operation.OriginalPath == filePath || operation.NewPath == filePath {
						if latestOperation == nil || operation.Timestamp.After(latestOperation.Timestamp) {
							latestOperation = &operation
						}
					}
				}
			}

			if latestOperation == nil {
				fmt.Printf("❌ No undo history found for file: %s\n", filePath)
				return
			}

			if dryRun {
				fmt.Printf("🔍 Dry run: Would restore %s from operation %s\n", filePath, latestOperation.ID)
				return
			}

			err := undoManager.UndoOperation(latestOperation.ID)
			if err != nil {
				fmt.Printf("❌ Error restoring file: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully restored file: %s\n", filePath)
		},
	}

	restoreFileCmd.Flags().Bool("dry-run", false, "Preview what would be restored without actually restoring")

	// Add all commands to root
	rootCmd.AddCommand(undoHistoryCmd)
	rootCmd.AddCommand(undoOpCmd)
	rootCmd.AddCommand(undoSessionCmd)
	rootCmd.AddCommand(startSessionCmd)
	rootCmd.AddCommand(endSessionCmd)
	rootCmd.AddCommand(clearHistoryCmd)
	rootCmd.AddCommand(restoreFileCmd)
}

// Helper functions

func showSessionDetails(session *undo.UndoSession) {
	fmt.Printf("🌸 Session Details: %s (╹◡╹)♡\n", session.Name)
	fmt.Println("=====================================")
	fmt.Printf("🆔 Session ID: %s\n", session.ID)
	fmt.Printf("📝 Description: %s\n", session.Description)
	fmt.Printf("📅 Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("📊 Total Operations: %d\n", len(session.Operations))
	fmt.Printf("✅ Undone: %t\n", session.Undone)

	if session.UndoneAt != nil {
		fmt.Printf("🔄 Undone At: %s\n", session.UndoneAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n📋 Operations:")
	for i, op := range session.Operations {
		fmt.Printf("   %d. %s %s\n", i+1, op.Type, op.OriginalPath)
		fmt.Printf("      🆔 ID: %s\n", op.ID)
		fmt.Printf("      📅 Time: %s\n", op.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("      📊 Size: %s\n", formatBytes(op.Size))
		fmt.Printf("      ✅ Undone: %t\n", op.Undone)
		if op.NewPath != "" {
			fmt.Printf("      📁 New Path: %s\n", op.NewPath)
		}
		if op.BackupPath != "" {
			fmt.Printf("      💾 Backup: %s\n", op.BackupPath)
		}
		fmt.Println()
	}
}
