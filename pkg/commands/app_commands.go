/**
 * Application Commands Package
 *
 * Provides application-related command definitions for the Ena virtual assistant,
 * including application starting, stopping, restarting, and listing.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: app_commands.go
 * Description: Application operation command definitions and handlers
 */

package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupAppCommands sets up all application-related commands
func setupAppCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Application management commands - handling apps with care ✨

	var appCmd = &cobra.Command{
		Use:   "app [operation] [args...]",
		Short: "Application operation commands",
		Long: `Start, stop, restart, list, and display information about applications.

Examples:
  ena app start firefox
  ena app stop firefox
  ena app restart firefox
  ena app list
  ena app info firefox`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("app", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Start application command
	var startCmd = &cobra.Command{
		Use:   "start <app_name>",
		Short: "Start an application",
		Long:  "Start the specified application.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("app", append([]string{"start"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// Stop application command
	var stopCmd = &cobra.Command{
		Use:   "stop <app_name>",
		Short: "Stop an application",
		Long:  "Stop the specified application.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("app", append([]string{"stop"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgYellow).Println(result)
			}
		},
	}

	// Restart application command
	var restartCmd = &cobra.Command{
		Use:   "restart <app_name>",
		Short: "Restart an application",
		Long:  "Restart the specified application.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("app", append([]string{"restart"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	// List applications command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List running applications",
		Long:  "Display a list of currently running applications.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("app", []string{"list"})
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgCyan).Println(result)
			}
		},
	}

	// Application info command
	var infoCmd = &cobra.Command{
		Use:   "info <app_name>",
		Short: "Show application information",
		Long:  "Display detailed information about the specified application.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("app", append([]string{"info"}, args...))
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgBlue).Println(result)
			}
		},
	}

	// Add subcommands
	appCmd.AddCommand(startCmd)
	appCmd.AddCommand(stopCmd)
	appCmd.AddCommand(restartCmd)
	appCmd.AddCommand(listCmd)
	appCmd.AddCommand(infoCmd)

	rootCmd.AddCommand(appCmd)
}
