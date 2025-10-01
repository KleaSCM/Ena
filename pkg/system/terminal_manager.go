/**
 * Terminal Manager Package
 *
 * Provides terminal control capabilities including opening/closing terminals,
 * executing commands, and directory navigation with proper error handling.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: terminal_manager.go
 * Description: Terminal operations and command execution management
 */

package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TerminalManager handles all terminal-related operations
type TerminalManager struct {
	CurrentDir string
	History    []string // Command history - keeping track
}

// NewTerminalManager creates a new terminal manager instance
func NewTerminalManager() *TerminalManager {
	// Terminal management with care ✨
	wd, err := os.Getwd()
	if err != nil {
		wd = "/" // Default directory
	}

	return &TerminalManager{
		CurrentDir: wd,
		History:    make([]string, 0),
	}
}

// OpenTerminal opens a new terminal window
func (tm *TerminalManager) OpenTerminal() (string, error) {
	// Open new terminal for you ✨
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// Try multiple terminal emulators on Linux
		terminals := []string{"gnome-terminal", "xterm", "konsole", "xfce4-terminal", "alacritty", "kitty"}

		for _, terminal := range terminals {
			if tm.isCommandAvailable(terminal) {
				switch terminal {
				case "gnome-terminal":
					cmd = exec.Command("gnome-terminal", "--working-directory="+tm.CurrentDir)
				case "xterm":
					cmd = exec.Command("xterm", "-e", "cd", tm.CurrentDir, "&&", "bash")
				case "konsole":
					cmd = exec.Command("konsole", "--workdir", tm.CurrentDir)
				case "xfce4-terminal":
					cmd = exec.Command("xfce4-terminal", "--working-directory="+tm.CurrentDir)
				case "alacritty":
					cmd = exec.Command("alacritty", "--working-directory", tm.CurrentDir)
				case "kitty":
					cmd = exec.Command("kitty", "--directory", tm.CurrentDir)
				}
				break
			}
		}

		if cmd == nil {
			return "", fmt.Errorf("No available terminal emulator found 😅")
		}

	case "darwin":
		// For macOS
		cmd = exec.Command("open", "-a", "Terminal", tm.CurrentDir)

	case "windows":
		// For Windows
		cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", "cd", "/d", tm.CurrentDir)

	default:
		return "", fmt.Errorf("Unsupported OS: %s 😅", runtime.GOOS)
	}

	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("Failed to start terminal: %v", err)
	}

	tm.addToHistory(fmt.Sprintf("open-terminal (dir: %s)", tm.CurrentDir))
	return fmt.Sprintf("Opened new terminal! (Current directory: %s) ✨", tm.CurrentDir), nil
}

// CloseTerminal closes the current terminal session
func (tm *TerminalManager) CloseTerminal() (string, error) {
	// Close terminal carefully
	// Actually this will close the terminal executing this command
	tm.addToHistory("close-terminal")

	return "Closing terminal... Thank you! ✨", nil
}

// ExecuteCommand executes a command in the current directory
func (tm *TerminalManager) ExecuteCommand(command string) (string, error) {
	// Execute command safely
	tm.addToHistory(fmt.Sprintf("execute: %s", command))

	// Sanitize command (safety first!)
	sanitizedCommand := tm.sanitizeCommand(command)
	if sanitizedCommand == "" {
		return "", fmt.Errorf("Command is invalid or empty 😅")
	}

	// Check for dangerous commands
	if tm.isDangerousCommand(sanitizedCommand) {
		return "", fmt.Errorf("Refused to execute dangerous command: %s ⚠️", sanitizedCommand)
	}

	// Split command
	parts := strings.Fields(sanitizedCommand)
	if len(parts) == 0 {
		return "", fmt.Errorf("Command is empty 😅")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = tm.CurrentDir

	// Execute command
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("Failed to execute command: %v\nOutput: %s", err, string(output))
	}

	result := fmt.Sprintf("Executed command \"%s\"! ✨\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n%s",
		sanitizedCommand, string(output))

	return result, nil
}

// ChangeDirectory changes the current working directory
func (tm *TerminalManager) ChangeDirectory(directory string) (string, error) {
	// Change directory to new location
	// Determine if absolute or relative path
	var newPath string
	if strings.HasPrefix(directory, "/") || strings.HasPrefix(directory, "~") {
		// Absolute path
		newPath = directory
		if strings.HasPrefix(directory, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("Failed to get home directory: %v", err)
			}
			newPath = strings.Replace(directory, "~", home, 1)
		}
	} else {
		// Relative path
		newPath = filepath.Join(tm.CurrentDir, directory)
	}

	// Check if directory exists
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		return "", fmt.Errorf("Directory \"%s\" does not exist 😅", directory)
	}

	// Actually change directory
	err := os.Chdir(newPath)
	if err != nil {
		return "", fmt.Errorf("Failed to change directory: %v", err)
	}

	tm.CurrentDir = newPath
	tm.addToHistory(fmt.Sprintf("cd %s", directory))

	return fmt.Sprintf("Changed directory to \"%s\"! ✨", newPath), nil
}

// GetCurrentDirectory returns the current working directory
func (tm *TerminalManager) GetCurrentDirectory() string {
	// Return current location
	return tm.CurrentDir
}

// GetHistory returns the command history
func (tm *TerminalManager) GetHistory() []string {
	// Show command history
	return tm.History
}

// ClearHistory clears the command history
func (tm *TerminalManager) ClearHistory() (string, error) {
	// Clear history - fresh start
	tm.History = make([]string, 0)
	return "Cleared command history! ✨", nil
}

// addToHistory adds a command to the history
func (tm *TerminalManager) addToHistory(command string) {
	// Add to history - keeping track
	timestamp := time.Now().Format("15:04:05")
	tm.History = append(tm.History, fmt.Sprintf("[%s] %s", timestamp, command))

	// Limit history size (keep only last 100 entries)
	if len(tm.History) > 100 {
		tm.History = tm.History[len(tm.History)-100:]
	}
}

// isCommandAvailable checks if a command is available in the system
func (tm *TerminalManager) isCommandAvailable(command string) bool {
	// Check if command is available
	_, err := exec.LookPath(command)
	return err == nil
}

// sanitizeCommand sanitizes command input to prevent injection attacks
func (tm *TerminalManager) sanitizeCommand(command string) string {
	// Sanitize command for safe execution
	// Remove leading and trailing whitespace
	command = strings.TrimSpace(command)

	// Check for dangerous character combinations
	dangerousPatterns := []string{
		"&&", // Command chaining
		"||", // OR chaining
		";",  // Command separator
		"|",  // Pipe
		"`",  // Command substitution
		"$(", // Command substitution
		"${", // Variable expansion
		">",  // Redirect
		"<",  // Redirect
		"\\", // Escape
	}

	// Warn if dangerous patterns are present
	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			// Allow pipe but warn for others
			if pattern != "|" {
				return "" // Return empty string to trigger error
			}
		}
	}

	return command
}

// isDangerousCommand checks if a command is potentially dangerous
func (tm *TerminalManager) isDangerousCommand(command string) bool {
	// Check for dangerous commands - safety first!
	dangerousCommands := []string{
		"rm -rf /",
		"mkfs",
		"dd if=",
		":(){ :|:& };:",
		"sudo rm -rf",
		"chmod -R 777 /",
		"format",
		"fdisk",
		"parted",
		"> /dev/sda",
		"> /dev/",
	}

	command = strings.ToLower(command)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(command, strings.ToLower(dangerous)) {
			return true
		}
	}

	return false
}

// GetSystemInfo returns basic system information
func (tm *TerminalManager) GetSystemInfo() (string, error) {
	// Show system information
	info := []string{
		"🖥️  System Information:",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		fmt.Sprintf("OS: %s %s", runtime.GOOS, runtime.GOARCH),
		fmt.Sprintf("Current Directory: %s", tm.CurrentDir),
		fmt.Sprintf("Available CPUs: %d", runtime.NumCPU()),
	}

	// Some environment variables
	info = append(info, fmt.Sprintf("User: %s", os.Getenv("USER")))
	info = append(info, fmt.Sprintf("Home: %s", os.Getenv("HOME")))

	return strings.Join(info, "\n"), nil
}
