/**
 * Search Commands Package
 *
 * Provides search and deletion command definitions for the Ena virtual assistant,
 * including file searching and safe file deletion with confirmation.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: search_commands.go
 * Description: Search and deletion command definitions and handlers
 */

package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupSearchCommands sets up all search and deletion related commands
func setupSearchCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Search and deletion commands - finding and removing files with care ✨

	// File search command
	var searchCmd = &cobra.Command{
		Use:   "search <pattern> <directory>",
		Short: "Search for files",
		Long: `Search for files matching the specified pattern.

Examples:
  ena search "*.txt" /home/user
  ena search "*.go" /home/user/projects
  ena search "config" /etc`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("search", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgCyan).Println(result)
			}
		},
	}

	// File deletion command
	var deleteCmd = &cobra.Command{
		Use:   "delete <path> [--force]",
		Short: "Delete files",
		Long: `Delete the specified file.

⚠️  Note: Confirmation prompt will be displayed unless --force flag is used.

Examples:
  ena delete /path/to/file.txt
  ena delete /path/to/file.txt --force`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("delete", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgYellow).Println(result)
			}
		},
	}

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(deleteCmd)
}
