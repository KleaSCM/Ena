/**
 * Folder Commands Package
 *
 * Provides folder-related command definitions for the Ena virtual assistant,
 * including folder creation, listing, deletion, and information display.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: folder_commands.go
 * Description: Folder operation command definitions and handlers
 */

package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupFolderCommands sets up all folder-related commands
func setupFolderCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Folder operation commands - keeping things organized ✨

	var folderCmd = &cobra.Command{
		Use:   "folder [operation] [args...]",
		Short: "Folder operation commands",
		Long: `Create, list, delete, and display information about folders.

Examples:
  ena folder create /path/to/folder
  ena folder list /path/to/folder
  ena folder delete /path/to/folder
  ena folder info /path/to/folder`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("folder", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Create folder command
	var createCmd = &cobra.Command{
		Use:   "create <path>",
		Short: "Create a folder",
		Long:  "Create a new folder at the specified path.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("folder", append([]string{"create"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// List folder command
	var listCmd = &cobra.Command{
		Use:   "list <path>",
		Short: "List folder contents",
		Long:  "Display the contents of the specified folder.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("folder", append([]string{"list"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgCyan).Println(result)
			}
		},
	}

	// Delete folder command
	var deleteCmd = &cobra.Command{
		Use:   "delete <path>",
		Short: "Delete a folder",
		Long:  "Delete the specified folder and its contents.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("folder", append([]string{"delete"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgYellow).Println(result)
			}
		},
	}

	// Folder info command
	var infoCmd = &cobra.Command{
		Use:   "info <path>",
		Short: "Show folder information",
		Long:  "Display detailed information about the specified folder.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("folder", append([]string{"info"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgBlue).Println(result)
			}
		},
	}

	// Add subcommands
	folderCmd.AddCommand(createCmd)
	folderCmd.AddCommand(listCmd)
	folderCmd.AddCommand(deleteCmd)
	folderCmd.AddCommand(infoCmd)

	rootCmd.AddCommand(folderCmd)
}
