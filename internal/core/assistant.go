/**
 * Ena Virtual Assistant Core Engine
 *
 * The heart and soul of Ena - handles all command processing, system hooks,
 * and coordination between different assistant modules.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: assistant.go
 * Description: Core assistant engine that orchestrates all system operations
 */

package core

import (
	"fmt"
	"time"

	"ena/internal/health"
	"ena/internal/hooks"
)

// Assistant represents the main virtual assistant instance
type Assistant struct {
	Name        string
	Version     string
	SystemHooks *hooks.SystemHooks
	Health      *health.SystemHealth
	IsRunning   bool
	StartTime   time.Time
}

// NewAssistant creates a new instance of the Ena virtual assistant
func NewAssistant() *Assistant {
	// Initialize Ena's core components with love and care ✨
	assistant := &Assistant{
		Name:      "Ena",
		Version:   "1.0.0",
		IsRunning: true,
		StartTime: time.Now(),
	}

	// Initialize system hooks for comprehensive system operations
	assistant.SystemHooks = hooks.NewSystemHooks()

	// Add system health monitoring capabilities
	assistant.Health = health.NewSystemHealth()

	return assistant
}

// Greet returns a warm greeting message from Ena
func (a *Assistant) Greet() string {
	// A warm, loving greeting to welcome users ♡
	return fmt.Sprintf("Hello! I'm %s ✨ VA!", a.Name)
}

// GetStatus returns the current status of the assistant
func (a *Assistant) GetStatus() map[string]interface{} {
	// Provide detailed status information for transparency
	uptime := time.Since(a.StartTime)

	return map[string]interface{}{
		"name":      a.Name,
		"version":   a.Version,
		"running":   a.IsRunning,
		"uptime":    uptime.String(),
		"startTime": a.StartTime.Format("2006-01-02 15:04:05"),
	}
}

// ProcessCommand handles incoming commands and delegates to appropriate handlers
func (a *Assistant) ProcessCommand(command string, args []string) (string, error) {
	// Command processing logic - route commands to appropriate handlers
	switch command {
	case "file":
		return a.SystemHooks.HandleFileOperation(args)
	case "folder":
		return a.SystemHooks.HandleFolderOperation(args)
	case "terminal":
		return a.SystemHooks.HandleTerminalOperation(args)
	case "app":
		return a.SystemHooks.HandleApplicationOperation(args)
	case "system":
		return a.SystemHooks.HandleSystemOperation(args)
	case "health":
		return a.Health.GetHealthReport()
	case "search":
		return a.SystemHooks.HandleFileSearch(args)
	case "delete":
		return a.SystemHooks.HandleFileDeletion(args)
	default:
		return "", fmt.Errorf("Unknown command: \"%s\" - I don't understand that! 😅", command)
	}
}

// Shutdown gracefully shuts down the assistant
func (a *Assistant) Shutdown() {
	// Gracefully shut down the assistant
	a.IsRunning = false
	fmt.Println("Ena says goodbye! ✨ (╹◡╹)♡")
}
