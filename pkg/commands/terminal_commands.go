/**
 * Terminal Commands Package
 *
 * Provides terminal-related command definitions for the Ena virtual assistant,
 * including terminal opening, closing, command execution, and directory navigation.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: terminal_commands.go
 * Description: Terminal operation command definitions and handlers
 */

package commands

import (
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupTerminalCommands sets up all terminal-related commands
func setupTerminalCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Terminal operation commands - handling command line with care ✨

	var terminalCmd = &cobra.Command{
		Use:   "terminal [operation] [args...]",
		Short: "Terminal operation commands",
		Long: `Open, close terminals, execute commands, and change directories.

Examples:
  ena terminal open
  ena terminal close
  ena terminal execute "ls -la"
  ena terminal cd /home/user`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("terminal", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Open terminal command
	var openCmd = &cobra.Command{
		Use:   "open",
		Short: "Open a new terminal",
		Long:  "Open a new terminal window.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("terminal", []string{"open"})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Close terminal command
	var closeCmd = &cobra.Command{
		Use:   "close",
		Short: "Close the terminal",
		Long:  "Close the current terminal session.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("terminal", []string{"close"})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgYellow).Println(result)
			}
		},
	}

	// Execute command
	var executeCmd = &cobra.Command{
		Use:   "execute <command>",
		Short: "Execute a command",
		Long:  "Execute the specified command in the terminal.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			command := strings.Join(args, " ")
			result, err := assistant.ProcessCommand("terminal", []string{"execute", command})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgCyan).Println(result)
			}
		},
	}

	// Change directory command
	var cdCmd = &cobra.Command{
		Use:   "cd <directory>",
		Short: "Change directory",
		Long:  "Change to the specified directory.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("terminal", append([]string{"cd"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Add subcommands
	terminalCmd.AddCommand(openCmd)
	terminalCmd.AddCommand(closeCmd)
	terminalCmd.AddCommand(executeCmd)
	terminalCmd.AddCommand(cdCmd)

	rootCmd.AddCommand(terminalCmd)
}
