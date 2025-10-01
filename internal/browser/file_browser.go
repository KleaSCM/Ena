/**
 * Interactive File Browser
 *
 * Provides an interactive file browser with arrow key navigation,
 * file selection, and preview capabilities.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: file_browser.go
 * Description: Interactive file system browser with keyboard navigation
 */

package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

// FileItem represents a file or directory in the browser
type FileItem struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	IsHidden bool
}

// FileBrowser handles interactive file browsing
type FileBrowser struct {
	currentPath   string
	items         []FileItem
	selectedIndex int
	scrollOffset  int
	maxItems      int
	rl            *readline.Instance
}

// NewFileBrowser creates a new file browser instance
func NewFileBrowser(startPath string) (*FileBrowser, error) {
	// Initialize file browser
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			startPath = "/"
		}
	}

	fb := &FileBrowser{
		currentPath:   startPath,
		selectedIndex: 0,
		scrollOffset:  0,
		maxItems:      20, // Show 20 items at a time
	}

	// Initialize readline for keyboard input
	config := &readline.Config{
		Prompt:          "",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		FuncFilterInputRune: func(r rune) (rune, bool) {
			// Handle special keys
			switch r {
			case '\x1a': // Ctrl+Z
				return r, false
			case '\r': // Enter
				return r, true
			case '\x03': // Ctrl+C
				return r, true
			}
			return r, true
		},
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize readline: %v", err)
	}

	fb.rl = rl
	return fb, nil
}

// Start starts the interactive file browser
func (fb *FileBrowser) Start() (string, error) {
	// Start interactive file browsing
	for {
		if err := fb.loadDirectory(); err != nil {
			return "", fmt.Errorf("failed to load directory: %v", err)
		}

		fb.displayBrowser()

		line, err := fb.rl.Readline()
		if err != nil {
			return "", err
		}

		// Handle special commands
		line = strings.TrimSpace(line)

		switch line {
		case "":
			// Enter pressed - select current item
			// Select current item
			if fb.selectedIndex < len(fb.items) {
				selected := fb.items[fb.selectedIndex]
				if selected.IsDir {
					// Navigate into directory
					fb.currentPath = selected.Path
					fb.selectedIndex = 0
					fb.scrollOffset = 0
				} else {
					// Return selected file path
					fb.rl.Close()
					return selected.Path, nil
				}
			}
		case "q", "Q":
			// Quit browser
			fb.rl.Close()
			return "", fmt.Errorf("browser cancelled")
		case "h", "H":
			// Go to parent directory
			parent := filepath.Dir(fb.currentPath)
			if parent != fb.currentPath {
				fb.currentPath = parent
				fb.selectedIndex = 0
				fb.scrollOffset = 0
			}
		case "k", "up":
			// Move up
			if fb.selectedIndex > 0 {
				fb.selectedIndex--
				if fb.selectedIndex < fb.scrollOffset {
					fb.scrollOffset = fb.selectedIndex
				}
			}
		case "j", "down":
			// Move down
			if fb.selectedIndex < len(fb.items)-1 {
				fb.selectedIndex++
				if fb.selectedIndex >= fb.scrollOffset+fb.maxItems {
					fb.scrollOffset = fb.selectedIndex - fb.maxItems + 1
				}
			}
		case "r", "R":
			// Refresh directory
			fb.loadDirectory()
		case "f", "F":
			// Toggle hidden files
			fb.toggleHiddenFiles()
		}
	}
}

// loadDirectory loads the current directory contents
func (fb *FileBrowser) loadDirectory() error {
	// Load directory contents
	entries, err := os.ReadDir(fb.currentPath)
	if err != nil {
		return err
	}

	fb.items = make([]FileItem, 0)

	// Add parent directory if not at root
	if fb.currentPath != "/" && fb.currentPath != "" {
		parent := filepath.Dir(fb.currentPath)
		fb.items = append(fb.items, FileItem{
			Name:  ".. (parent)",
			Path:  parent,
			IsDir: true,
		})
	}

	// Process directory entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		item := FileItem{
			Name:     entry.Name(),
			Path:     filepath.Join(fb.currentPath, entry.Name()),
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			IsHidden: strings.HasPrefix(entry.Name(), "."),
		}

		// Filter hidden files if needed
		if !item.IsHidden || fb.showHiddenFiles() {
			fb.items = append(fb.items, item)
		}
	}

	// Sort items: directories first, then files, alphabetically
	sort.Slice(fb.items, func(i, j int) bool {
		// Parent directory always first
		if strings.HasPrefix(fb.items[i].Name, "..") {
			return true
		}
		if strings.HasPrefix(fb.items[j].Name, "..") {
			return false
		}

		// Directories before files
		if fb.items[i].IsDir && !fb.items[j].IsDir {
			return true
		}
		if !fb.items[i].IsDir && fb.items[j].IsDir {
			return false
		}

		// Alphabetical within same type
		return strings.ToLower(fb.items[i].Name) < strings.ToLower(fb.items[j].Name)
	})

	// Adjust selected index if out of bounds
	if fb.selectedIndex >= len(fb.items) {
		fb.selectedIndex = len(fb.items) - 1
		if fb.selectedIndex < 0 {
			fb.selectedIndex = 0
		}
	}

	return nil
}

// displayBrowser displays the file browser interface
func (fb *FileBrowser) displayBrowser() {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Header
	color.New(color.FgMagenta, color.Bold).Println("🌸 Ena File Browser 🌸")
	color.New(color.FgCyan).Printf("📁 %s\n", fb.currentPath)
	color.New(color.FgYellow).Printf("Items: %d | Selected: %d\n", len(fb.items), fb.selectedIndex+1)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Display items
	start := fb.scrollOffset
	end := start + fb.maxItems
	if end > len(fb.items) {
		end = len(fb.items)
	}

	for i := start; i < end; i++ {
		item := fb.items[i]
		isSelected := i == fb.selectedIndex

		// Selection indicator
		if isSelected {
			color.New(color.FgYellow, color.Bold).Print("▶ ")
		} else {
			fmt.Print("  ")
		}

		// File type indicator
		if item.IsDir {
			color.New(color.FgBlue).Print("📁 ")
		} else {
			color.New(color.FgGreen).Print("📄 ")
		}

		// Name
		if isSelected {
			color.New(color.FgWhite, color.Bold).Print(item.Name)
		} else {
			color.New(color.FgWhite).Print(item.Name)
		}

		// Size (for files)
		if !item.IsDir && item.Size > 0 {
			fmt.Printf(" (%s)", fb.formatSize(item.Size))
		}

		// Hidden file indicator
		if item.IsHidden {
			color.New(color.FgRed).Print(" (hidden)")
		}

		fmt.Println()
	}

	// Footer
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.New(color.FgCyan).Println("Controls:")
	color.New(color.FgWhite).Print("  ↑/k: Up  ↓/j: Down  Enter: Select  h: Parent  r: Refresh  f: Toggle hidden  q: Quit")
	fmt.Println()
}

// formatSize formats file size in human readable format
func (fb *FileBrowser) formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// showHiddenFiles returns whether hidden files should be shown
func (fb *FileBrowser) showHiddenFiles() bool {
	// Simple toggle - in a real implementation, this could be persistent
	return true // For now, always show hidden files
}

// toggleHiddenFiles toggles the display of hidden files
func (fb *FileBrowser) toggleHiddenFiles() {
	// This would toggle the hidden files display
	// For now, we always show them, but this could be enhanced
}

// Close closes the file browser
func (fb *FileBrowser) Close() error {
	return fb.rl.Close()
}
