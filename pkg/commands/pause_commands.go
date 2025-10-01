/**
 * Pause/Resume Commands
 *
 * Provides pause/resume functionality for progress bars with terminal compatibility testing.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: pause_commands.go
 * Description: Pause/resume command definitions with enhanced progress tracking
 */

package commands

import (
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupPauseCommands sets up pause/resume commands
func setupPauseCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Pause command
	var pauseCmd = &cobra.Command{
		Use:   "pause <operation>",
		Short: "Pause/resume progress bars and test terminal compatibility",
		Long: `Control progress bar pause/resume functionality and test terminal capabilities:
  • Demo pause/resume functionality with visual feedback
  • Test terminal compatibility (colors, cursor control, etc.)
  • Demonstrate persistent state management
  • Show graceful degradation for different terminal types

Operations:
  demo - Demonstrate pause/resume functionality
  test - Test terminal compatibility and capabilities

Example:
  ena pause demo
  ena pause test`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("pause", args)
			if err != nil {
				cmd.PrintErrln("❌ Error:", err)
				return
			}

			cmd.Println(result)
		},
	}

	rootCmd.AddCommand(pauseCmd)
}
