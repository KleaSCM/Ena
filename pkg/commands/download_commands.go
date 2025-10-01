/**
 * Download Commands
 *
 * Provides download functionality with progress bars for file downloads.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: download_commands.go
 * Description: Download command definitions with progress tracking
 */

package commands

import (
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupDownloadCommands sets up download-related commands
func setupDownloadCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Download command
	var downloadCmd = &cobra.Command{
		Use:   "download <url> <filename>",
		Short: "Download a file with progress bar",
		Long: `Download a file from a URL with a beautiful progress bar showing:
  • Download progress percentage
  • Transfer speed
  • Estimated time remaining
  • File size information

Example:
  ena download "https://example.com/file.zip" "downloaded_file.zip"`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("download", args)
			if err != nil {
				cmd.PrintErrln("❌ Error:", err)
				return
			}

			cmd.Println(result)
		},
	}

	rootCmd.AddCommand(downloadCmd)
}
