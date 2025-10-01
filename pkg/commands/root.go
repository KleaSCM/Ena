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
	"ena/internal/input"
)

// HelpEntry represents a single help entry
type HelpEntry struct {
	Category string
	Command  string
	Desc     string
}

// GetHelpEntries returns all available help entries
func GetHelpEntries() []HelpEntry {
	// Centralized help entries - easy to maintain
	return []HelpEntry{
		{"📁 File Operations", "file create <path>", "Create a file"},
		{"📁 File Operations", "file read <path>", "Read a file"},
		{"📁 File Operations", "file write <path> <content>", "Write to a file"},
		{"📁 File Operations", "file copy <src> <dest>", "Copy a file"},
		{"📁 File Operations", "file move <src> <dest>", "Move a file"},
		{"📁 File Operations", "file delete <path> [--force]", "Delete a file"},
		{"📁 File Operations", "file info <path>", "Show file information"},
		{"📂 Folder Operations", "folder create <path>", "Create a folder"},
		{"📂 Folder Operations", "folder list <path>", "List folder contents"},
		{"📂 Folder Operations", "folder delete <path>", "Delete a folder"},
		{"📂 Folder Operations", "folder info <path>", "Show folder information"},
		{"🖥️  Terminal Operations", "terminal open", "Open a new terminal"},
		{"🖥️  Terminal Operations", "terminal close", "Close terminal"},
		{"🖥️  Terminal Operations", "terminal execute <command>", "Execute a command"},
		{"🖥️  Terminal Operations", "terminal cd <directory>", "Change directory"},
		{"📱 Application Operations", "app start <app_name>", "Start an application"},
		{"📱 Application Operations", "app stop <app_name>", "Stop an application"},
		{"📱 Application Operations", "app restart <app_name>", "Restart an application"},
		{"📱 Application Operations", "app list", "List running applications"},
		{"📱 Application Operations", "app info <app_name>", "Show application information"},
		{"⚡ System Operations", "system restart", "Restart system"},
		{"⚡ System Operations", "system shutdown", "Shutdown system"},
		{"⚡ System Operations", "system sleep", "Put system to sleep"},
		{"⚡ System Operations", "system info", "Show system information"},
		{"🏥 System Health Check", "health", "Check system health"},
		{"🔍 Search & Delete", "search <pattern> <directory>", "Search for files"},
		{"🔍 Search & Delete", "delete <path> [--force]", "Delete a file"},
		{"🗂️ File Browser", "browse [path]", "Interactive file browser"},
		{"📥 Download", "download <url> <filename>", "Download file with progress bar"},
		{"📊 Multi-Progress", "multi <operation> [files...]", "Process multiple files with multiple progress bars"},
		{"⏸️ Pause/Resume", "pause demo", "Demonstrate pause/resume functionality"},
		{"⏸️ Pause/Resume", "pause test", "Test terminal compatibility"},
		{"⏸️ Pause/Resume", "pause state", "Test JSON state persistence"},
		{"⏸️ Pause/Resume", "pause adaptive", "Test adaptive refresh functionality"},
		{"⏸️ Pause/Resume", "pause theme", "Test custom progress bar themes"},
		{"⏸️ Pause/Resume", "pause events", "Test event hooks and callbacks"},
		{"⏸️ Pause/Resume", "pause http", "Test real HTTP download functionality"},
		{"👀 File Watching", "watch start [paths...]", "Start real-time file system monitoring"},
		{"👀 File Watching", "watch stop", "Stop file watching session"},
		{"👀 File Watching", "watch status", "Show file watching status"},
		{"👀 File Watching", "watch demo", "Demonstrate file watching capabilities"},
		{"👀 File Watching", "watch debug", "Test enhanced debug features with detailed logging"},
		{"👀 File Watching", "watch advanced", "Test enterprise-grade features: batching, prioritization, metrics, error recovery"},
		{"👀 File Watching", "watch add <path>", "Add path dynamically to running watcher"},
		{"👀 File Watching", "watch remove <path>", "Remove path dynamically from running watcher"},
		{"👀 File Watching", "watch metrics", "Show detailed performance metrics"},
		{"👀 File Watching", "watch reload", "Reload configuration without restart"},
		{"🎨 Theme Management", "theme list", "List all available themes"},
		{"🎨 Theme Management", "theme current", "Show current theme information"},
		{"🎨 Theme Management", "theme set <name>", "Set active theme"},
		{"🎨 Theme Management", "theme preview <name>", "Preview theme with color samples"},
		{"🎨 Theme Management", "theme toggle", "Toggle between light and dark modes"},
		{"🎨 Theme Management", "theme demo", "Demonstrate all themes"},
		{"🎨 Theme Management", "theme create <name> <desc> <mode>", "Create custom theme"},
		{"🎨 Theme Management", "theme delete <name>", "Delete custom theme"},
		{"🎨 Theme Management", "theme save <name>", "Save theme to disk"},
		{"🎨 Theme Management", "theme load <file>", "Load theme from disk"},
		{"🎨 Theme Management", "theme setcolor <theme> <element> <hex>", "Set specific color"},
		{"🎨 Theme Management", "theme validate <name>", "Validate theme"},
		{"🎨 Theme Management", "theme cache <clear|stats>", "Manage color cache"},
		{"🔔 Desktop Notifications", "notify test", "Send test notification"},
		{"🔔 Desktop Notifications", "notify send <type> <title> <message>", "Send custom notification"},
		{"🔔 Desktop Notifications", "notify status", "Show notification system status"},
		{"🔔 Desktop Notifications", "notify history", "Show notification history"},
		{"🔔 Desktop Notifications", "notify demo", "Demonstrate all notification types"},
		{"🧠 Smart Suggestions", "suggest", "Get intelligent suggestions based on usage patterns"},
		{"🧠 Smart Suggestions", "stats", "Show usage statistics and analytics"},
		{"🧠 Smart Suggestions", "feedback <id> <type>", "Provide feedback on suggestions"},
		{"🧠 Smart Suggestions", "workflow", "Show workflow optimization suggestions"},
		{"🧠 Smart Suggestions", "optimize", "Show system optimization suggestions"},
		{"📦 Batch Operations", "batch-delete <paths...>", "Delete multiple files/folders with progress tracking"},
		{"📦 Batch Operations", "batch-copy <sources...> <dest>", "Copy multiple files/folders recursively"},
		{"📦 Batch Operations", "batch-move <sources...> <dest>", "Move multiple files/folders efficiently"},
		{"📦 Batch Operations", "batch-status [job-id]", "Show status of batch operations"},
		{"📦 Batch Operations", "batch-cancel <job-id>", "Cancel a running batch operation"},
		{"↩️ Undo Operations", "undo-history", "Show undo history and available operations"},
		{"↩️ Undo Operations", "undo-operation <id>", "Undo a specific operation"},
		{"↩️ Undo Operations", "undo-session <id>", "Undo all operations in a session"},
		{"↩️ Undo Operations", "restore-file <path>", "Restore a file from undo history"},
		{"↩️ Undo Operations", "start-session <name>", "Start a new undo session"},
		{"↩️ Undo Operations", "end-session", "End the current undo session"},
		{"💡 Other", "help", "Show this help"},
		{"💡 Other", "status", "Show Ena's status"},
		{"💡 Other", "exit", "Say goodbye to Ena"},
	}
}

// SetupRootCommand creates and configures the root command
func SetupRootCommand(assistant *core.Assistant) *cobra.Command {
	// Command line interface configuration
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
	setupDownloadCommands(rootCmd, assistant)
	setupMultiCommands(rootCmd, assistant)
	setupPauseCommands(rootCmd, assistant)
	setupWatchCommands(rootCmd, assistant)
	setupThemeCommands(rootCmd, assistant)
	setupNotificationCommands(rootCmd, assistant)
	setupSuggestionsCommands(rootCmd)
	setupBatchCommands(rootCmd)
	setupUndoCommands(rootCmd)

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
	color.New(color.FgCyan).Println("💡 Tip: Press TAB for command completion!")
	color.New(color.FgCyan).Println("💡 Tip: Try typing 'fil' and press TAB for fuzzy search!")
	fmt.Println()

	// Initialize terminal input with completion support
	terminalInput, err := input.NewTerminalInput()
	if err != nil {
		color.New(color.FgRed).Printf("❌ Failed to initialize terminal input: %v\n", err)
		color.New(color.FgYellow).Println("Falling back to basic input mode...")
		startBasicInteractiveMode(assistant)
		return
	}
	defer terminalInput.Close()

	for {
		// Read line with completion support
		inputStr, err := terminalInput.ReadLine()
		if err != nil {
			if err.Error() == "EOF" || err.Error() == "interrupt" {
				break
			}
			color.New(color.FgRed).Printf("❌ Error reading input: %v\n", err)
			continue
		}

		if inputStr == "" {
			continue
		}

		// Add to history
		terminalInput.AddToHistory(inputStr)

		if strings.ToLower(inputStr) == "exit" {
			assistant.Shutdown()
			break
		}

		if strings.ToLower(inputStr) == "help" {
			showHelp()
			continue
		}

		if strings.ToLower(inputStr) == "status" {
			showStatus(assistant)
			continue
		}

		// Parse and execute command
		parts := strings.Fields(inputStr)
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

// startBasicInteractiveMode starts basic interactive mode without completion
func startBasicInteractiveMode(assistant *core.Assistant) {
	// Basic interactive mode fallback
	for {
		fmt.Print("Ena> ")

		var inputStr string
		fmt.Scanln(&inputStr)
		inputStr = strings.TrimSpace(inputStr)

		if inputStr == "" {
			continue
		}

		if strings.ToLower(inputStr) == "exit" {
			assistant.Shutdown()
			break
		}

		if strings.ToLower(inputStr) == "help" {
			showHelp()
			continue
		}

		if strings.ToLower(inputStr) == "status" {
			showStatus(assistant)
			continue
		}

		// Parse and execute command
		parts := strings.Fields(inputStr)
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
	color.New(color.FgCyan, color.Bold).Println("🌸 Ena's Command List 🌸")
	color.New(color.FgCyan, color.Bold).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	entries := GetHelpEntries()
	currentCategory := ""

	for _, entry := range entries {
		// Print category header when it changes
		if entry.Category != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			color.New(color.FgYellow, color.Bold).Printf("%s:\n", entry.Category)
			currentCategory = entry.Category
		}

		// Print command with proper formatting
		color.New(color.FgWhite).Printf("  %-35s", entry.Command)
		color.New(color.FgCyan).Printf(" - %s\n", entry.Desc)
	}

	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.New(color.FgMagenta).Println("Let's have fun managing your system together! (╹◡╹)♡")
}

// showStatus displays the assistant status
func showStatus(assistant *core.Assistant) {
	// Show Ena's current status
	status := assistant.GetStatus()

	color.New(color.FgMagenta, color.Bold).Println("🌸 Ena's Status 🌸")
	color.New(color.FgMagenta, color.Bold).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Name
	color.New(color.FgWhite).Print("Name: ")
	color.New(color.FgCyan, color.Bold).Println(status["name"])

	// Version
	color.New(color.FgWhite).Print("Version: ")
	color.New(color.FgGreen, color.Bold).Println(status["version"])

	// Running status
	color.New(color.FgWhite).Print("Running: ")
	if status["running"] == "true" {
		color.New(color.FgGreen, color.Bold).Println("✓ Yes")
	} else {
		color.New(color.FgRed).Println("✗ No")
	}

	// Uptime
	color.New(color.FgWhite).Print("Uptime: ")
	color.New(color.FgYellow).Println(status["uptime"])

	// Start Time
	color.New(color.FgWhite).Print("Start Time: ")
	color.New(color.FgCyan).Println(status["startTime"])

	color.New(color.FgMagenta, color.Bold).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.New(color.FgGreen).Println("I'm doing great! (๑˃̵ᴗ˂̵)")
}
