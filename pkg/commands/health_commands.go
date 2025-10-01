/**
 * Health Commands Package
 *
 * Provides system health monitoring command definitions for the Ena virtual assistant,
 * including comprehensive system health reports and monitoring.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: health_commands.go
 * Description: System health monitoring command definitions and handlers
 */

package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupHealthCommands sets up all health-related commands
func setupHealthCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Health check commands - monitoring system health with care ✨

	var healthCmd = &cobra.Command{
		Use:   "health",
		Short: "Check system health status",
		Long: `Generate a comprehensive system health status report.

Includes the following information:
  💻 CPU usage and information
  🧠 Memory usage
  💾 Disk usage
  🐹 Go runtime information
  📊 Overall health assessment

Example:
  ena health`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("health", args)
			if err != nil {
				color.New(color.FgRed).Printf("❌ Error: %v\n", err)
			} else {
				color.New(color.FgGreen).Println(result)
			}
		},
	}

	rootCmd.AddCommand(healthCmd)
}
