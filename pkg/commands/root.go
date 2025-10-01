/**
 * Root Command Interface
 *
 * Provides the main command-line interface for the Ena virtual assistant,
 * handling command parsing, help display, and command delegation.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: root.go
 * Description: Main CLI interface and command routing for the virtual assistant
 */

package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"ena/internal/core"
)

// SetupRootCommand creates and configures the root command
func SetupRootCommand(assistant *core.Assistant) *cobra.Command {
	// あたしのコマンドラインインターフェース...美しく使いやすくするのよ〜 (๑˃̵ᴗ˂̵)
	var rootCmd = &cobra.Command{
		Use:   "ena",
		Short: "Ena - Your gentle virtual assistant ✨",
		Long: `Ena is your gentle virtual assistant that manages your system with care! 💕

What I can do:
  📁 File and folder operations
  🖥️  Terminal control
  📱 Application management
  🏥 System health monitoring
  🔍 File search and deletion
  ⚡ System restart and shutdown

Let's make your computer life fun and easy together! (╹◡╹)♡`,
		Run: func(cmd *cobra.Command, args []string) {
			// Default behavior: start interactive mode
			startInteractiveMode(assistant)
		},
	}

	// Add subcommands
	setupFileCommands(rootCmd, assistant)
	setupFolderCommands(rootCmd, assistant)
	setupTerminalCommands(rootCmd, assistant)
	setupAppCommands(rootCmd, assistant)
	setupSystemCommands(rootCmd, assistant)
	setupHealthCommands(rootCmd, assistant)
	setupSearchCommands(rootCmd, assistant)

	return rootCmd
}

// startInteractiveMode starts the interactive command mode
func startInteractiveMode(assistant *core.Assistant) {
	// Interactive mode for user interaction
	color.New(color.FgMagenta, color.Bold).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.New(color.FgMagenta, color.Bold).Println("🌸 Ena - Your Gentle Virtual Assistant 🌸")
	color.New(color.FgMagenta, color.Bold).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println(assistant.Greet())
	fmt.Println()

	color.New(color.FgCyan).Println("💡 Tip: Type 'help' to see what I can do!")
	color.New(color.FgCyan).Println("💡 Tip: Type 'exit' to say goodbye...")
	fmt.Println()

	for {
		// Display prompt
		color.New(color.FgYellow, color.Bold).Print("Ena> ")

		var input string
		fmt.Scanln(&input)

		if strings.ToLower(input) == "exit" {
			assistant.Shutdown()
			break
		}

		if strings.ToLower(input) == "help" {
			showHelp()
			continue
		}

		if strings.ToLower(input) == "status" {
			showStatus(assistant)
			continue
		}

		// Parse and execute command
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		result, err := assistant.ProcessCommand(command, args)
		if err != nil {
			color.New(color.FgRed).Printf("❌ Error: %v\n", err)
		} else {
			color.New(color.FgGreen).Println(result)
		}

		fmt.Println()
	}
}

// showHelp displays the help information
func showHelp() {
	// Display comprehensive help information
	helpText := []string{
		"🌸 Ena's Command List 🌸",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"",
		"📁 File Operations:",
		"  file create <path>              - Create a file",
		"  file read <path>                - Read a file",
		"  file write <path> <content>     - Write to a file",
		"  file copy <src> <dest>          - Copy a file",
		"  file move <src> <dest>          - Move a file",
		"  file delete <path> [--force]    - Delete a file",
		"  file info <path>                - Show file information",
		"",
		"📂 Folder Operations:",
		"  folder create <path>            - Create a folder",
		"  folder list <path>              - List folder contents",
		"  folder delete <path>            - Delete a folder",
		"  folder info <path>              - Show folder information",
		"",
		"🖥️  Terminal Operations:",
		"  terminal open                  - Open a new terminal",
		"  terminal close                 - Close terminal",
		"  terminal execute <command>     - Execute a command",
		"  terminal cd <directory>        - Change directory",
		"",
		"📱 Application Operations:",
		"  app start <app_name>           - Start an application",
		"  app stop <app_name>            - Stop an application",
		"  app restart <app_name>         - Restart an application",
		"  app list                       - List running applications",
		"  app info <app_name>            - Show application information",
		"",
		"⚡ System Operations:",
		"  system restart                 - Restart system",
		"  system shutdown                - Shutdown system",
		"  system sleep                   - Put system to sleep",
		"  system info                    - Show system information",
		"",
		"🏥 System Health Check:",
		"  health                         - Check system health",
		"",
		"🔍 Search & Delete:",
		"  search <pattern> <directory>   - Search for files",
		"  delete <path> [--force]        - Delete a file",
		"",
		"💡 Other:",
		"  help                          - Show this help",
		"  status                        - Show Ena's status",
		"  exit                          - Say goodbye to Ena",
		"",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"Let's have fun managing your system together! (╹◡╹)♡",
	}

	color.New(color.FgCyan).Println(strings.Join(helpText, "\n"))
}

// showStatus displays the assistant status
func showStatus(assistant *core.Assistant) {
	// Show Ena's current status
	status := assistant.GetStatus()

	statusText := []string{
		"🌸 Ena's Status 🌸",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		fmt.Sprintf("Name: %s", status["name"]),
		fmt.Sprintf("Version: %s", status["version"]),
		fmt.Sprintf("Running: %v", status["running"]),
		fmt.Sprintf("Uptime: %s", status["uptime"]),
		fmt.Sprintf("Start Time: %s", status["startTime"]),
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"I'm doing great! (๑˃̵ᴗ˂̵)",
	}

	color.New(color.FgGreen).Println(strings.Join(statusText, "\n"))
}
