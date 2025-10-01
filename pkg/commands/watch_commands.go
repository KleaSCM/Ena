/**
 * Watch Commands
 *
 * Provides file watching functionality with real-time monitoring and live updates.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: watch_commands.go
 * Description: File watching command definitions with real-time file system monitoring
 */

package commands

import (
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupWatchCommands sets up file watching commands
func setupWatchCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Watch command
	var watchCmd = &cobra.Command{
		Use:   "watch <operation> [paths...]",
		Short: "Real-time file system monitoring with live updates",
		Long: `Monitor file system changes in real-time with live updates and notifications:
  • Start monitoring specific paths or directories
  • Stop current file watching session
  • Check status of active file watchers
  • Demonstrate file watching capabilities

Operations:
  start [paths...] - Start watching specified paths (defaults to current directory)
  stop            - Stop current file watching session
  status          - Show current file watching status
  demo            - Demonstrate file watching with test files

Examples:
  ena watch start                    # Watch current directory
  ena watch start /tmp /home/user     # Watch multiple paths
  ena watch stop                     # Stop watching
  ena watch status                   # Show status
  ena watch demo                     # Run demonstration`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("watch", args)
			if err != nil {
				cmd.PrintErrln("❌ Error:", err)
				return
			}

			cmd.Println(result)
		},
	}

	rootCmd.AddCommand(watchCmd)
}
