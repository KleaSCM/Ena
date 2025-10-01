/**
 * Notification Commands
 *
 * Provides desktop notification functionality with cross-platform support.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: notification_commands.go
 * Description: Desktop notification command definitions with platform-specific support
 */

package commands

import (
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupNotificationCommands sets up notification management commands
func setupNotificationCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Notification command
	var notifyCmd = &cobra.Command{
		Use:   "notify <operation> [args...]",
		Short: "Manage desktop notifications with cross-platform support",
		Long: `Cross-platform desktop notification system with comprehensive features:
  • Send notifications for completed tasks and system events
  • Support for different notification types (success, error, warning, info, task)
  • Platform-specific implementations (macOS, Linux, Windows)
  • Notification history and management
  • Configurable settings and preferences

Operations:
  test                    - Send a test notification
  send <type> <title> <message> - Send custom notification
  status                  - Show notification system status
  history                 - Show notification history
  clear                   - Clear notification history
  enable                  - Enable notifications
  disable                 - Disable notifications
  config                  - Show notification configuration
  demo                    - Demonstrate different notification types

Notification Types:
  success                 - Success notifications (green)
  error                   - Error notifications (red)
  warning                 - Warning notifications (yellow)
  info                    - Info notifications (blue)
  task                    - Task completion notifications (purple)

Platform Support:
  macOS                   - Uses osascript for native notifications
  Linux                   - Uses notify-send for desktop notifications
  Windows                 - Uses PowerShell for toast notifications

Examples:
  ena notify test                           # Send test notification
  ena notify send success "Done!" "Task completed" # Send success notification
  ena notify send error "Error" "Something went wrong" # Send error notification
  ena notify status                         # Check notification system status
  ena notify demo                           # Demonstrate all notification types
  ena notify history                        # View notification history`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("notify", args)
			if err != nil {
				cmd.PrintErrln("❌ Error:", err)
				return
			}

			cmd.Println(result)
		},
	}

	rootCmd.AddCommand(notifyCmd)
}
