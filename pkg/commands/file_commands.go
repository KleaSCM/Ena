/**
 * File Commands Package
 *
 * Provides file-related command definitions for the Ena virtual assistant,
 * including file creation, reading, writing, copying, moving, and deletion.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: file_commands.go
 * Description: File operation command definitions and handlers
 */

package commands

import (
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupFileCommands sets up all file-related commands
func setupFileCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// File operation commands - handling files with care ✨

	var fileCmd = &cobra.Command{
		Use:   "file [operation] [args...]",
		Short: "File operation commands",
		Long: `Create, read, write, copy, move, delete, and display information about files.

Examples:
  ena file create /path/to/file.txt
  ena file read /path/to/file.txt
  ena file write /path/to/file.txt "Hello, World!"
  ena file copy /source.txt /dest.txt
  ena file move /old.txt /new.txt
  ena file delete /path/to/file.txt
  ena file info /path/to/file.txt`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("file", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Create file command
	var createCmd = &cobra.Command{
		Use:   "create <path>",
		Short: "Create a file",
		Long:  "Create a new file at the specified path.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("file", append([]string{"create"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Read file command
	var readCmd = &cobra.Command{
		Use:   "read <path>",
		Short: "Read a file",
		Long:  "Read and display the contents of the specified file.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("file", append([]string{"read"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgCyan).Println(result)
			}
		},
	}

	// Write file command
	var writeCmd = &cobra.Command{
		Use:   "write <path> <content>",
		Short: "Write to a file",
		Long:  "Write content to the specified file.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			content := strings.Join(args[1:], " ")
			result, err := assistant.ProcessCommand("file", []string{"write", path, content})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Copy file command
	var copyCmd = &cobra.Command{
		Use:   "copy <source> <destination>",
		Short: "Copy a file",
		Long:  "Copy the specified file to a new location.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("file", append([]string{"copy"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Move file command
	var moveCmd = &cobra.Command{
		Use:   "move <source> <destination>",
		Short: "Move a file",
		Long:  "Move the specified file to a new location.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("file", append([]string{"move"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Delete file command
	var deleteCmd = &cobra.Command{
		Use:   "delete <path> [--force]",
		Short: "Delete a file",
		Long:  "Delete the specified file. Use --force flag to delete without confirmation.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("file", append([]string{"delete"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgYellow).Println(result)
			}
		},
	}

	// File info command
	var infoCmd = &cobra.Command{
		Use:   "info <path>",
		Short: "Show file information",
		Long:  "Display detailed information about the specified file.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("file", append([]string{"info"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgBlue).Println(result)
			}
		},
	}

	// Add subcommands
	fileCmd.AddCommand(createCmd)
	fileCmd.AddCommand(readCmd)
	fileCmd.AddCommand(writeCmd)
	fileCmd.AddCommand(copyCmd)
	fileCmd.AddCommand(moveCmd)
	fileCmd.AddCommand(deleteCmd)
	fileCmd.AddCommand(infoCmd)

	rootCmd.AddCommand(fileCmd)
}
