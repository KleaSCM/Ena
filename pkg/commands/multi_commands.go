/**
 * Multi-Progress Commands
 *
 * Provides multi-progress bar functionality for processing multiple files simultaneously.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: multi_commands.go
 * Description: Multi-progress command definitions with enhanced progress tracking
 */

package commands

import (
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupMultiCommands sets up multi-progress commands
func setupMultiCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Multi-progress command
	var multiCmd = &cobra.Command{
		Use:   "multi <operation> [files...]",
		Short: "Process multiple files with multiple progress bars",
		Long: `Process multiple files simultaneously with individual progress bars showing:
  • Multiple concurrent progress bars
  • Individual file processing status
  • Real-time updates for each file
  • Enhanced error handling and recovery
  • Color-coded progress indicators

Example:
  ena multi "Processing" file1.txt file2.txt file3.txt
  ena multi "Converting" *.jpg`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("multi", args)
			if err != nil {
				cmd.PrintErrln("❌ Error:", err)
				return
			}

			cmd.Println(result)
		},
	}

	rootCmd.AddCommand(multiCmd)
}
