/**
 * System Commands Package
 *
 * Provides system-related command definitions for the Ena virtual assistant,
 * including system restart, shutdown, sleep, and information display.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: system_commands.go
 * Description: System operation command definitions and handlers
 */

package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupSystemCommands sets up all system-related commands
func setupSystemCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// System operation commands - managing system with care ✨

	var systemCmd = &cobra.Command{
		Use:   "system [operation]",
		Short: "System operation commands",
		Long: `Restart, shutdown, sleep, and display information about the system.

⚠️  Warning: These commands affect the entire system. Use with caution.

Examples:
  ena system restart
  ena system shutdown
  ena system sleep
  ena system info`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("system", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// System restart command
	var restartCmd = &cobra.Command{
		Use:   "restart",
		Short: "Restart the system",
		Long:  "⚠️  Restart the entire system. Save your work before executing.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			color.New(color.FgRed, color.Bold).Println("⚠️  Warning: System will restart!")
			color.New(color.FgRed, color.Bold).Println("⚠️  Unsaved work will be lost!")

			result, err := assistant.ProcessCommand("system", []string{"restart"})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgYellow).Println(result)
			}
		},
	}

	// System shutdown command
	var shutdownCmd = &cobra.Command{
		Use:   "shutdown",
		Short: "Shutdown the system",
		Long:  "⚠️  Shutdown the entire system. Save your work before executing.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			color.New(color.FgRed, color.Bold).Println("⚠️  Warning: System will shutdown!")
			color.New(color.FgRed, color.Bold).Println("⚠️  Unsaved work will be lost!")

			result, err := assistant.ProcessCommand("system", []string{"shutdown"})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgYellow).Println(result)
			}
		},
	}

	// System sleep command
	var sleepCmd = &cobra.Command{
		Use:   "sleep",
		Short: "Put system to sleep",
		Long:  "Put the system into sleep mode.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("system", []string{"sleep"})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgBlue).Println(result)
			}
		},
	}

	// System info command
	var infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Show system information",
		Long:  "Display detailed information about the current system.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("system", []string{"info"})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgCyan).Println(result)
			}
		},
	}

	// Add subcommands
	systemCmd.AddCommand(restartCmd)
	systemCmd.AddCommand(shutdownCmd)
	systemCmd.AddCommand(sleepCmd)
	systemCmd.AddCommand(infoCmd)

	rootCmd.AddCommand(systemCmd)
}
