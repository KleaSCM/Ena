/**
 * System Hooks Module
 *
 * Provides comprehensive system integration capabilities including file operations,
 * terminal control, application management, and system utilities.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: system_hooks.go
 * Description: Core system hooks for file, folder, terminal, and application operations
 */

package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"ena/internal/browser"
	"ena/pkg/system"
)

// Operation constants to prevent typos and improve code clarity
const (
	OpCreate  = "create"
	OpRead    = "read"
	OpWrite   = "write"
	OpCopy    = "copy"
	OpMove    = "move"
	OpDelete  = "delete"
	OpInfo    = "info"
	OpList    = "list"
	OpOpen    = "open"
	OpClose   = "close"
	OpExecute = "execute"
	OpCd      = "cd"
	OpStart   = "start"
	OpStop    = "stop"
	OpRestart = "restart"
	OpSleep   = "sleep"
)

// requireArgs validates that the required number of arguments are present
func requireArgs(args []string, n int, context string) error {
	// Validate that we have enough arguments for the operation
	if len(args) < n {
		return fmt.Errorf("%s requires at least %d arguments", context, n)
	}
	return nil
}

// SystemHooks handles all system-level operations
type SystemHooks struct {
	FileManager     *system.FileManager
	TerminalManager *system.TerminalManager
	AppManager      *system.AppManager
}

// NewSystemHooks creates a new instance of system hooks
func NewSystemHooks() *SystemHooks {
	// Initialize all system operation handlers
	return &SystemHooks{
		FileManager:     system.NewFileManager(),
		TerminalManager: system.NewTerminalManager(),
		AppManager:      system.NewAppManager(),
	}
}

// HandleFileOperation processes file-related commands
func (sh *SystemHooks) HandleFileOperation(args []string) (string, error) {
	if err := requireArgs(args, 2, "File operation"); err != nil {
		return "", err
	}

	operation := args[0]
	path := args[1]

	switch operation {
	case OpCreate:
		return sh.FileManager.CreateFile(path)
	case OpRead:
		return sh.FileManager.ReadFile(path)
	case OpWrite:
		if err := requireArgs(args, 3, "File write"); err != nil {
			return "", err
		}
		content := strings.Join(args[2:], " ")
		return sh.FileManager.WriteFile(path, content)
	case OpCopy:
		if err := requireArgs(args, 3, "File copy"); err != nil {
			return "", err
		}
		return sh.FileManager.CopyFile(path, args[2])
	case OpMove:
		if err := requireArgs(args, 3, "File move"); err != nil {
			return "", err
		}
		return sh.FileManager.MoveFile(path, args[2])
	case OpInfo:
		return sh.FileManager.GetFileInfo(path)
	default:
		return "", fmt.Errorf("Unknown file operation: \"%s\" - I don't understand that! 😅", operation)
	}
}

// HandleFolderOperation processes folder-related commands
func (sh *SystemHooks) HandleFolderOperation(args []string) (string, error) {
	if err := requireArgs(args, 2, "Folder operation"); err != nil {
		return "", err
	}

	operation := args[0]
	path := args[1]

	switch operation {
	case OpCreate:
		return sh.FileManager.CreateFolder(path)
	case OpList:
		return sh.FileManager.ListFolder(path)
	case OpDelete:
		// Require --force flag for safety when deleting folders
		if len(args) < 3 || args[2] != "--force" {
			return "", fmt.Errorf("Folder deletion requires --force flag for safety! 😅")
		}
		return sh.FileManager.DeleteFolder(path)
	case OpInfo:
		return sh.FileManager.GetFolderInfo(path)
	default:
		return "", fmt.Errorf("Unknown folder operation: \"%s\" - I don't understand that! 😅", operation)
	}
}

// HandleTerminalOperation processes terminal-related commands
func (sh *SystemHooks) HandleTerminalOperation(args []string) (string, error) {
	if err := requireArgs(args, 1, "Terminal operation"); err != nil {
		return "", err
	}

	operation := args[0]

	switch operation {
	case OpOpen:
		return sh.TerminalManager.OpenTerminal()
	case OpClose:
		return sh.TerminalManager.CloseTerminal()
	case OpExecute:
		if err := requireArgs(args, 2, "Command execution"); err != nil {
			return "", err
		}
		command := strings.Join(args[1:], " ")
		return sh.TerminalManager.ExecuteCommand(command)
	case OpCd:
		if err := requireArgs(args, 2, "Directory change"); err != nil {
			return "", err
		}
		return sh.TerminalManager.ChangeDirectory(args[1])
	default:
		return "", fmt.Errorf("Unknown terminal operation: \"%s\" - I don't understand that! 😅", operation)
	}
}

// HandleApplicationOperation processes application-related commands
func (sh *SystemHooks) HandleApplicationOperation(args []string) (string, error) {
	if err := requireArgs(args, 1, "Application operation"); err != nil {
		return "", err
	}

	operation := args[0]

	switch operation {
	case OpList:
		return sh.AppManager.ListApplications()
	case OpStart:
		if err := requireArgs(args, 2, "Application start"); err != nil {
			return "", err
		}
		return sh.AppManager.StartApplication(args[1])
	case OpStop:
		if err := requireArgs(args, 2, "Application stop"); err != nil {
			return "", err
		}
		return sh.AppManager.StopApplication(args[1])
	case OpRestart:
		if err := requireArgs(args, 2, "Application restart"); err != nil {
			return "", err
		}
		return sh.AppManager.RestartApplication(args[1])
	case OpInfo:
		if err := requireArgs(args, 2, "Application info"); err != nil {
			return "", err
		}
		return sh.AppManager.GetApplicationInfo(args[1])
	default:
		return "", fmt.Errorf("Unknown application operation: \"%s\" - I don't understand that! 😅", operation)
	}
}

// HandleSystemOperation processes system-level commands
func (sh *SystemHooks) HandleSystemOperation(args []string) (string, error) {
	if err := requireArgs(args, 1, "System operation"); err != nil {
		return "", err
	}

	operation := args[0]

	switch operation {
	case OpRestart:
		return sh.RestartSystem()
	case "shutdown":
		return sh.ShutdownSystem()
	case OpSleep:
		return sh.SleepSystem()
	case OpInfo:
		return sh.GetSystemInfo()
	default:
		return "", fmt.Errorf("Unknown system operation: \"%s\" - I don't understand that! 😅", operation)
	}
}

// HandleFileSearch performs file search operations
func (sh *SystemHooks) HandleFileSearch(args []string) (string, error) {
	if err := requireArgs(args, 2, "File search"); err != nil {
		return "", err
	}

	pattern := args[0]
	directory := args[1]

	return sh.FileManager.SearchFiles(pattern, directory)
}

// HandleFileBrowser handles interactive file browsing
func (sh *SystemHooks) HandleFileBrowser(args []string) (string, error) {
	// Interactive file browser - navigate and select files
	startPath := "."
	if len(args) > 0 {
		startPath = args[0]
	}

	// Create and start file browser
	browser, err := browser.NewFileBrowser(startPath)
	if err != nil {
		return "", fmt.Errorf("Failed to start file browser: %v", err)
	}
	defer browser.Close()

	selectedPath, err := browser.Start()
	if err != nil {
		return "", fmt.Errorf("File browser error: %v", err)
	}

	return fmt.Sprintf("Selected file: \"%s\" ✨", selectedPath), nil
}

// HandleFileDeletion handles file deletion with safety checks
func (sh *SystemHooks) HandleFileDeletion(args []string) (string, error) {
	if err := requireArgs(args, 1, "File deletion"); err != nil {
		return "", err
	}

	path := args[0]
	force := false

	// Check for force flag for safety
	if len(args) > 1 && args[1] == "--force" {
		force = true
	} else {
		// Require --force flag for safety
		return "", fmt.Errorf("File deletion requires --force flag for safety! 😅")
	}

	return sh.FileManager.DeleteFile(path, force)
}

// RestartSystem restarts the system
func (sh *SystemHooks) RestartSystem() (string, error) {
	// Restart the system with proper safety warnings
	cmd := exec.Command("sudo", "reboot")
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Restart failed: %v", err)
	}

	return "⚠️  System restarting... Are you sure? Thank you for using Ena! ✨ (╹◡╹)", nil
}

// ShutdownSystem shuts down the system
func (sh *SystemHooks) ShutdownSystem() (string, error) {
	// Shutdown the system with proper safety warnings
	cmd := exec.Command("sudo", "shutdown", "now")
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Shutdown failed: %v", err)
	}

	return "⚠️  System shutting down... Are you sure? Thank you for using Ena! ✨ (╹◡╹)", nil
}

// SleepSystem puts the system to sleep
func (sh *SystemHooks) SleepSystem() (string, error) {
	// Put the system to sleep gently
	cmd := exec.Command("systemctl", "suspend")
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Sleep failed: %v", err)
	}

	return "System is going to sleep... Good night! ✨ (╹◡╹)♡", nil
}

// GetSystemInfo returns comprehensive system information
func (sh *SystemHooks) GetSystemInfo() (string, error) {
	// Provide detailed system information
	info := []string{
		"🖥️  System Information:",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"",
	}

	// OS Information
	info = append(info, fmt.Sprintf("OS: %s/%s", runtime.GOOS, runtime.GOARCH))
	info = append(info, fmt.Sprintf("CPU Cores: %d", runtime.NumCPU()))
	info = append(info, fmt.Sprintf("Go Version: %s", runtime.Version()))
	info = append(info, "")

	// Hostname
	hostname, err := os.Hostname()
	if err == nil {
		info = append(info, fmt.Sprintf("Hostname: %s", hostname))
	}

	// Current directory
	wd, err := os.Getwd()
	if err == nil {
		info = append(info, fmt.Sprintf("Current Directory: %s", wd))
	}

	// Current time
	info = append(info, fmt.Sprintf("Current Time: %s", time.Now().Format("2006-01-02 15:04:05")))

	// Environment variables
	info = append(info, "")
	info = append(info, "Environment Variables:")
	info = append(info, fmt.Sprintf("  User: %s", os.Getenv("USER")))
	info = append(info, fmt.Sprintf("  Home: %s", os.Getenv("HOME")))
	info = append(info, fmt.Sprintf("  Shell: %s", os.Getenv("SHELL")))
	info = append(info, fmt.Sprintf("  PATH: %s", os.Getenv("PATH")))

	return strings.Join(info, "\n"), nil
}
