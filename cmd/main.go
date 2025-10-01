/**
 * Ena Virtual Assistant - Main Entry Point
 *
 * A powerful virtual assistant that provides system hooks for file operations,
 * terminal control, application management, and system health monitoring.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: main.go
 * Description: Main entry point for the Ena virtual assistant application
 */

package main

import (
	"os"

	"ena/internal/core"
	"ena/pkg/commands"
)

func main() {
	// Initialize Ena's core engine - the heart of our virtual assistant
	Assistant := core.NewAssistant()

	// Set up the command-line interface for user interaction
	RootCmd := commands.SetupRootCommand(Assistant)

	// Execute the root command with proper error handling
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
