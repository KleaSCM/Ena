/**
 * Terminal Input Manager
 *
 * Provides advanced terminal input capabilities including tab completion,
 * command history, and interactive command line interface.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: terminal_input.go
 * Description: Advanced terminal input handling with completion and history
 */

package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

// TerminalInput handles advanced terminal input with completion and history
type TerminalInput struct {
	rl       *readline.Instance
	history  []string
	commands []string
}

// NewTerminalInput creates a new terminal input manager
func NewTerminalInput() (*TerminalInput, error) {
	// Initialize terminal input with completion support
	ti := &TerminalInput{
		history: make([]string, 0),
		commands: []string{
			// File operations
			"file", "file create", "file read", "file write", "file copy", "file move", "file delete", "file info",
			// Folder operations
			"folder", "folder create", "folder list", "folder delete", "folder info",
			// Terminal operations
			"terminal", "terminal open", "terminal close", "terminal execute", "terminal cd",
			// Application operations
			"app", "app start", "app stop", "app restart", "app list", "app info",
			// System operations
			"system", "system restart", "system shutdown", "system sleep", "system info",
			// Health and search
			"health", "search", "delete",
			// Built-in commands
			"help", "status", "exit",
		},
	}

	// Configure readline instance
	config := &readline.Config{
		Prompt:            color.New(color.FgYellow, color.Bold).Sprint("Ena> "),
		HistoryFile:       "/tmp/ena_history",
		AutoComplete:      ti,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			switch r {
			case readline.CharCtrlZ:
				return r, false
			}
			return r, true
		},
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize readline: %v", err)
	}

	ti.rl = rl
	return ti, nil
}

// ReadLine reads a line from the terminal with completion support
func (ti *TerminalInput) ReadLine() (string, error) {
	line, err := ti.rl.Readline()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// Close closes the terminal input
func (ti *TerminalInput) Close() error {
	return ti.rl.Close()
}

// Do implements the readline.AutoCompleter interface
func (ti *TerminalInput) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// Get the current word being typed
	lineStr := string(line[:pos])
	words := strings.Fields(lineStr)

	if len(words) == 0 {
		// No words yet, suggest all commands
		return ti.suggestCommands(""), 0
	}

	currentWord := words[len(words)-1]
	prefix := strings.ToLower(currentWord)

	// If we're at the start of a new word or command
	if len(words) == 1 && pos == len(line) {
		// Suggest commands
		return ti.suggestCommands(prefix), len(currentWord)
	}

	// Check if current word looks like a file path
	if strings.HasPrefix(currentWord, "/") || strings.HasPrefix(currentWord, "./") || strings.HasPrefix(currentWord, "~") {
		// Suggest file paths
		return ti.suggestFilePaths(currentWord), len(currentWord)
	}

	// Check if we're in a file operation context
	if len(words) >= 2 {
		operation := words[0]
		if operation == "file" || operation == "folder" || operation == "search" || operation == "delete" {
			// Suggest file paths for file operations
			return ti.suggestFilePaths(currentWord), len(currentWord)
		}
	}

	// Default: suggest commands
	return ti.suggestCommands(prefix), len(currentWord)
}

// suggestCommands suggests commands based on prefix and fuzzy matching
func (ti *TerminalInput) suggestCommands(prefix string) [][]rune {
	var suggestions [][]rune
	var fuzzySuggestions [][]rune

	// First pass: exact prefix matches
	for _, cmd := range ti.commands {
		if prefix == "" || strings.HasPrefix(strings.ToLower(cmd), strings.ToLower(prefix)) {
			suggestions = append(suggestions, []rune(cmd))
		}
	}

	// Second pass: fuzzy matches if we have few exact matches
	if len(suggestions) < 5 && len(prefix) > 0 {
		for _, cmd := range ti.commands {
			cmdLower := strings.ToLower(cmd)
			prefixLower := strings.ToLower(prefix)

			// Skip if already in exact matches
			alreadyExact := false
			for _, exact := range suggestions {
				if string(exact) == cmd {
					alreadyExact = true
					break
				}
			}

			if alreadyExact {
				continue
			}

			// Check fuzzy match (contains all characters in order)
			if ti.fuzzyMatch(prefixLower, cmdLower) {
				fuzzySuggestions = append(fuzzySuggestions, []rune(cmd))
			}
		}
	}

	// Combine exact matches first, then fuzzy matches
	allSuggestions := append(suggestions, fuzzySuggestions...)

	// Limit suggestions to avoid overwhelming
	if len(allSuggestions) > 15 {
		allSuggestions = allSuggestions[:15]
	}

	return allSuggestions
}

// fuzzyMatch checks if query characters appear in order in the text
func (ti *TerminalInput) fuzzyMatch(query, text string) bool {
	if len(query) == 0 {
		return true
	}
	if len(query) > len(text) {
		return false
	}

	queryIndex := 0
	for _, char := range text {
		if queryIndex < len(query) && char == rune(query[queryIndex]) {
			queryIndex++
		}
	}

	return queryIndex == len(query)
}

// suggestFilePaths suggests file paths based on current input
func (ti *TerminalInput) suggestFilePaths(prefix string) [][]rune {
	var suggestions [][]rune

	// Expand tilde to home directory
	expandedPrefix := prefix
	if strings.HasPrefix(prefix, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expandedPrefix = strings.Replace(prefix, "~", homeDir, 1)
		}
	}

	// Get directory and pattern
	dir := filepath.Dir(expandedPrefix)
	pattern := filepath.Base(expandedPrefix)

	// If directory doesn't exist, return empty
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return suggestions
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return suggestions
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless explicitly looking for them
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(pattern, ".") {
			continue
		}

		// Check if name matches pattern
		if pattern == "" || strings.HasPrefix(name, pattern) {
			fullPath := filepath.Join(dir, name)

			// Add trailing slash for directories
			if entry.IsDir() {
				fullPath += "/"
			}

			// Convert back to original prefix format if needed
			if strings.HasPrefix(prefix, "~") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					fullPath = strings.Replace(fullPath, homeDir, "~", 1)
				}
			}

			suggestions = append(suggestions, []rune(fullPath))
		}
	}

	// Limit suggestions
	if len(suggestions) > 20 {
		suggestions = suggestions[:20]
	}

	return suggestions
}

// AddToHistory adds a command to history
func (ti *TerminalInput) AddToHistory(command string) {
	if command != "" && (len(ti.history) == 0 || ti.history[len(ti.history)-1] != command) {
		ti.history = append(ti.history, command)
		// Keep history size reasonable
		if len(ti.history) > 1000 {
			ti.history = ti.history[len(ti.history)-1000:]
		}
	}
}

// GetHistory returns command history
func (ti *TerminalInput) GetHistory() []string {
	return ti.history
}

// ClearHistory clears command history
func (ti *TerminalInput) ClearHistory() {
	ti.history = make([]string, 0)
}

// SetPrompt sets the terminal prompt
func (ti *TerminalInput) SetPrompt(prompt string) {
	ti.rl.SetPrompt(prompt)
}

// GetPrompt returns the current prompt
func (ti *TerminalInput) GetPrompt() string {
	return ti.rl.Config.Prompt
}
