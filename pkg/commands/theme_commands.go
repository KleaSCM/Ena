/**
 * Theme Commands
 *
 * Provides comprehensive theming functionality with dark/light modes and custom color schemes.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: theme_commands.go
 * Description: Theme management command definitions with color scheme support
 */

package commands

import (
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// setupThemeCommands sets up theme management commands
func setupThemeCommands(rootCmd *cobra.Command, assistant *core.Assistant) {
	// Theme command
	var themeCmd = &cobra.Command{
		Use:   "theme <operation> [theme_name]",
		Short: "Manage themes and color schemes with dark/light mode support",
		Long: `Comprehensive theme management system with multiple color schemes and modes:
  • List all available themes with descriptions
  • Set and preview themes with live color samples
  • Toggle between light and dark modes
  • Export theme configurations
  • Demonstrate all themes with color samples

Operations:
  list                    - List all available themes
  current                 - Show current theme information
  set <theme_name>        - Set active theme
  preview <theme_name>    - Preview theme with color samples
  info <theme_name>       - Show detailed theme information
  export <theme_name>     - Export theme configuration
  demo                    - Demonstrate all themes
  toggle                  - Toggle between light and dark modes
  create <name> <desc> <mode> [colors...] - Create custom theme
  delete <theme_name>     - Delete custom theme
  save <theme_name>       - Save theme to disk
  load <file_path>        - Load theme from disk
  setcolor <theme> <element> <hex> - Set specific color
  validate <theme_name>   - Validate theme
  cache <clear|stats>    - Manage color cache

Available Themes:
  default                 - Clean light theme with gentle colors
  dark                    - Modern dark theme with vibrant accents
  solarized-light         - Classic solarized light theme
  solarized-dark          - Classic solarized dark theme
  monokai                 - Popular Monokai color scheme
  dracula                 - Dark Dracula theme with vibrant colors
  nord                    - Arctic-inspired Nord theme

Examples:
  ena theme list                    # List all themes
  ena theme set dark                # Switch to dark theme
  ena theme preview monokai         # Preview Monokai theme
  ena theme toggle                  # Toggle light/dark mode
  ena theme demo                    # Show all themes`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			result, err := assistant.ProcessCommand("theme", args)
			if err != nil {
				cmd.PrintErrln("❌ Error:", err)
				return
			}

			cmd.Println(result)
		},
	}

	rootCmd.AddCommand(themeCmd)
}
