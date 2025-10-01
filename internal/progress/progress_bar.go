/**
 * Progress Bar System
 *
 * Provides animated progress bars for long-running operations including
 * file copying, downloads, and other time-consuming tasks.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: progress_bar.go
 * Description: Animated progress bars with real-time updates
 */

package progress

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
)

// ProgressBar represents an animated progress bar with enhanced features
type ProgressBar struct {
	total         int64
	current       int64
	width         int
	showPercent   bool
	showSpeed     bool
	showETA       bool
	startTime     time.Time
	lastUpdate    time.Time
	lastCurrent   int64
	mutex         sync.RWMutex
	done          bool
	customLabel   string
	refreshRate   time.Duration
	lastDisplay   time.Time
	colorEnabled  bool
	multiBarIndex int
	errorMessage  string
	errorOccurred bool

	// Pause/Resume support
	paused         int32 // atomic
	pauseTime      time.Time
	totalPauseTime time.Duration

	// Terminal compatibility
	terminalCaps *TerminalCapabilities

	// Persistent state
	stateFile  string
	persistent bool
}

// TerminalCapabilities represents terminal capabilities
type TerminalCapabilities struct {
	SupportsColor  bool
	SupportsCursor bool
	SupportsClear  bool
	Width          int
	Height         int
	IsDumb         bool
}

// ProgressBarConfig holds configuration for progress bars
type ProgressBarConfig struct {
	Width         int
	ShowPercent   bool
	ShowSpeed     bool
	ShowETA       bool
	CustomLabel   string
	RefreshRate   time.Duration
	ColorEnabled  bool
	MultiBarIndex int
	StateFile     string
	Persistent    bool
}

// MultiProgressManager manages multiple progress bars with persistent state
type MultiProgressManager struct {
	bars         []*ProgressBar
	mutex        sync.RWMutex
	running      int32 // atomic
	stateFile    string
	persistent   bool
	terminalCaps *TerminalCapabilities
}

// DetectTerminalCapabilities detects terminal capabilities
func DetectTerminalCapabilities() *TerminalCapabilities {
	caps := &TerminalCapabilities{
		SupportsColor:  true,
		SupportsCursor: true,
		SupportsClear:  true,
		Width:          80,
		Height:         24,
		IsDumb:         false,
	}

	// Check TERM environment variable
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		caps.IsDumb = true
		caps.SupportsColor = false
		caps.SupportsCursor = false
		caps.SupportsClear = false
	}

	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		caps.SupportsColor = false
	}

	// Check if stdout is a terminal
	if !isTerminal(os.Stdout) {
		caps.SupportsColor = false
		caps.SupportsCursor = false
		caps.SupportsClear = false
	}

	// Try to get terminal size
	if width, height, err := getTerminalSize(); err == nil {
		caps.Width = width
		caps.Height = height
	}

	return caps
}

// isTerminal checks if the file is a terminal
func isTerminal(file *os.File) bool {
	// Simple check - in a real implementation, you'd use syscalls
	// For now, we'll assume it's a terminal if it's stdout/stderr
	return file == os.Stdout || file == os.Stderr
}

// getTerminalSize gets terminal dimensions
func getTerminalSize() (width, height int, err error) {
	// This is a simplified implementation
	// In a real implementation, you'd use syscalls or a library
	return 80, 24, nil
}

// NewProgressBar creates a new progress bar with enhanced features
func NewProgressBar(total int64, config *ProgressBarConfig) *ProgressBar {
	if config == nil {
		config = &ProgressBarConfig{
			Width:         50,
			ShowPercent:   true,
			ShowSpeed:     true,
			ShowETA:       true,
			RefreshRate:   100 * time.Millisecond,
			ColorEnabled:  true,
			MultiBarIndex: 0,
			Persistent:    false,
		}
	}

	// Detect terminal capabilities
	terminalCaps := DetectTerminalCapabilities()

	// Override color setting based on terminal capabilities
	if !terminalCaps.SupportsColor {
		config.ColorEnabled = false
	}

	// Set default refresh rate if not specified
	if config.RefreshRate == 0 {
		config.RefreshRate = 100 * time.Millisecond
	}

	// Adjust width based on terminal capabilities
	if config.Width > terminalCaps.Width-20 {
		config.Width = terminalCaps.Width - 20
	}

	pb := &ProgressBar{
		total:          total,
		current:        0,
		width:          config.Width,
		showPercent:    config.ShowPercent,
		showSpeed:      config.ShowSpeed,
		showETA:        config.ShowETA,
		startTime:      time.Now(),
		lastUpdate:     time.Now(),
		lastCurrent:    0,
		customLabel:    config.CustomLabel,
		refreshRate:    config.RefreshRate,
		lastDisplay:    time.Now(),
		colorEnabled:   config.ColorEnabled,
		multiBarIndex:  config.MultiBarIndex,
		errorOccurred:  false,
		paused:         0,
		totalPauseTime: 0,
		terminalCaps:   terminalCaps,
		stateFile:      config.StateFile,
		persistent:     config.Persistent,
	}

	// Load persistent state if enabled
	if config.Persistent && config.StateFile != "" {
		pb.loadState()
	}

	return pb
}

// Update updates the progress bar with new current value
func (pb *ProgressBar) Update(current int64) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.current = current
	pb.lastUpdate = time.Now()
}

// Add adds to the current progress
func (pb *ProgressBar) Add(increment int64) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.current += increment
	pb.lastUpdate = time.Now()
}

// SetTotal sets the total value
func (pb *ProgressBar) SetTotal(total int64) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.total = total
}

// Finish marks the progress bar as complete
func (pb *ProgressBar) Finish() {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.current = pb.total
	pb.done = true
}

// SetError sets an error state for the progress bar
func (pb *ProgressBar) SetError(message string) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.errorOccurred = true
	pb.errorMessage = message
	pb.done = false
}

// ClearError clears any error state
func (pb *ProgressBar) ClearError() {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.errorOccurred = false
	pb.errorMessage = ""
}

// IsError returns true if the progress bar is in an error state
func (pb *ProgressBar) IsError() bool {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	return pb.errorOccurred
}

// Pause pauses the progress bar
func (pb *ProgressBar) Pause() {
	if atomic.CompareAndSwapInt32(&pb.paused, 0, 1) {
		pb.mutex.Lock()
		pb.pauseTime = time.Now()
		pb.mutex.Unlock()
	}
}

// Resume resumes the progress bar
func (pb *ProgressBar) Resume() {
	if atomic.CompareAndSwapInt32(&pb.paused, 1, 0) {
		pb.mutex.Lock()
		if !pb.pauseTime.IsZero() {
			pb.totalPauseTime += time.Since(pb.pauseTime)
			pb.pauseTime = time.Time{}
		}
		pb.mutex.Unlock()
	}
}

// IsPaused returns true if the progress bar is paused
func (pb *ProgressBar) IsPaused() bool {
	return atomic.LoadInt32(&pb.paused) == 1
}

// SaveState saves the current state to file
func (pb *ProgressBar) SaveState() error {
	if !pb.persistent || pb.stateFile == "" {
		return nil
	}

	pb.mutex.RLock()
	state := map[string]interface{}{
		"current":        pb.current,
		"total":          pb.total,
		"startTime":      pb.startTime,
		"totalPauseTime": pb.totalPauseTime,
		"done":           pb.done,
		"errorOccurred":  pb.errorOccurred,
		"errorMessage":   pb.errorMessage,
		"paused":         pb.IsPaused(),
	}
	pb.mutex.RUnlock()

	// In a real implementation, you'd serialize this to JSON/YAML
	// For now, we'll just create a simple text file
	content := fmt.Sprintf("current=%d\ntotal=%d\ndone=%t\nerror=%t\npaused=%t\n",
		state["current"], state["total"], state["done"], state["errorOccurred"], state["paused"])

	return os.WriteFile(pb.stateFile, []byte(content), 0644)
}

// loadState loads state from file
func (pb *ProgressBar) loadState() error {
	if !pb.persistent || pb.stateFile == "" {
		return nil
	}

	data, err := os.ReadFile(pb.stateFile)
	if err != nil {
		return err // File doesn't exist or can't be read
	}

	// Simple parsing - in a real implementation, you'd use JSON/YAML
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "current":
			if val, err := fmt.Sscanf(parts[1], "%d", &pb.current); err == nil && val == 1 {
				// Successfully loaded current value
			}
		case "done":
			if parts[1] == "true" {
				pb.done = true
			}
		case "error":
			if parts[1] == "true" {
				pb.errorOccurred = true
			}
		}
	}

	return nil
}

// NewMultiProgressManager creates a new multi-progress manager with enhanced features
func NewMultiProgressManager() *MultiProgressManager {
	return &MultiProgressManager{
		bars:         make([]*ProgressBar, 0),
		running:      0,
		persistent:   false,
		terminalCaps: DetectTerminalCapabilities(),
	}
}

// AddBar adds a progress bar to the manager
func (mpm *MultiProgressManager) AddBar(total int64, config *ProgressBarConfig) *ProgressBar {
	mpm.mutex.Lock()
	defer mpm.mutex.Unlock()

	if config == nil {
		config = &ProgressBarConfig{}
	}
	config.MultiBarIndex = len(mpm.bars)

	pb := NewProgressBar(total, config)
	mpm.bars = append(mpm.bars, pb)
	return pb
}

// DisplayAll displays all progress bars with improved concurrency safety
func (mpm *MultiProgressManager) DisplayAll() {
	mpm.mutex.RLock()
	bars := make([]*ProgressBar, len(mpm.bars))
	copy(bars, mpm.bars)
	terminalCaps := mpm.terminalCaps
	mpm.mutex.RUnlock()

	if len(bars) == 0 {
		return
	}

	// Clear screen and move cursor to top only if terminal supports it
	if terminalCaps.SupportsClear && terminalCaps.SupportsCursor {
		fmt.Print("\033[2J\033[H")
	} else {
		// Fallback: just print newlines
		fmt.Print("\n\n")
	}

	// Display each bar with individual mutex protection
	for i, bar := range bars {
		fmt.Printf("[%d] ", i+1)
		bar.Display()
		fmt.Println()
	}
}

// Start starts the multi-progress display loop with atomic operations
func (mpm *MultiProgressManager) Start() {
	if !atomic.CompareAndSwapInt32(&mpm.running, 0, 1) {
		return // Already running
	}

	go func() {
		for {
			if atomic.LoadInt32(&mpm.running) == 0 {
				break
			}

			mpm.DisplayAll()
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// Stop stops the multi-progress display with atomic operations
func (mpm *MultiProgressManager) Stop() {
	atomic.StoreInt32(&mpm.running, 0)
}

// SaveAllStates saves state for all progress bars
func (mpm *MultiProgressManager) SaveAllStates() error {
	mpm.mutex.RLock()
	bars := make([]*ProgressBar, len(mpm.bars))
	copy(bars, mpm.bars)
	mpm.mutex.RUnlock()

	for _, bar := range bars {
		if err := bar.SaveState(); err != nil {
			return err
		}
	}

	return nil
}

// Display renders the progress bar to stdout with colors and dynamic refresh
func (pb *ProgressBar) Display() {
	pb.mutex.RLock()
	current := pb.current
	total := pb.total
	done := pb.done
	errorOccurred := pb.errorOccurred
	errorMessage := pb.errorMessage
	paused := pb.IsPaused()
	pb.mutex.RUnlock()

	// Check if we should update based on refresh rate
	if time.Since(pb.lastDisplay) < pb.refreshRate && !done && !errorOccurred && !paused {
		return
	}
	pb.lastDisplay = time.Now()

	// Handle error display
	if errorOccurred {
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			color.New(color.FgRed, color.Bold).Printf("\r\033[K❌ Error: %s", errorMessage)
		} else {
			fmt.Printf("\r\033[K❌ Error: %s", errorMessage)
		}
		return
	}

	// Handle paused display
	if paused {
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			color.New(color.FgYellow, color.Bold).Printf("\r\033[K⏸️ Paused: %s", pb.customLabel)
		} else {
			fmt.Printf("\r\033[K⏸️ Paused: %s", pb.customLabel)
		}
		return
	}

	// Calculate percentage
	percent := float64(0)
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}

	// Calculate filled width
	filledWidth := int(float64(pb.width) * percent / 100)
	if filledWidth > pb.width {
		filledWidth = pb.width
	}

	// Build progress bar with colors
	var bar string
	if pb.colorEnabled && pb.terminalCaps.SupportsColor {
		// Green for progress, gray for remaining
		greenBar := color.New(color.FgGreen).Sprint(strings.Repeat("█", filledWidth))
		grayBar := color.New(color.FgWhite).Sprint(strings.Repeat("░", pb.width-filledWidth))
		bar = greenBar + grayBar
	} else {
		bar = strings.Repeat("█", filledWidth) + strings.Repeat("░", pb.width-filledWidth)
	}

	// Build status line
	status := fmt.Sprintf("[%s] ", bar)

	// Add percentage with color
	if pb.showPercent {
		var percentStr string
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			if done {
				percentStr = color.New(color.FgGreen, color.Bold).Sprintf("%.1f%%", percent)
			} else {
				percentStr = color.New(color.FgCyan).Sprintf("%.1f%%", percent)
			}
		} else {
			percentStr = fmt.Sprintf("%.1f%%", percent)
		}
		status += percentStr + " "
	}

	// Add current/total with color
	if total > 0 {
		var sizeStr string
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			sizeStr = color.New(color.FgWhite).Sprintf("(%s/%s)", formatBytes(current), formatBytes(total))
		} else {
			sizeStr = fmt.Sprintf("(%s/%s)", formatBytes(current), formatBytes(total))
		}
		status += sizeStr + " "
	} else {
		var sizeStr string
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			sizeStr = color.New(color.FgWhite).Sprintf("(%s)", formatBytes(current))
		} else {
			sizeStr = fmt.Sprintf("(%s)", formatBytes(current))
		}
		status += sizeStr + " "
	}

	// Add speed with color
	if pb.showSpeed && !done {
		speed := pb.calculateSpeed()
		if speed > 0 {
			var speedStr string
			if pb.colorEnabled && pb.terminalCaps.SupportsColor {
				speedStr = color.New(color.FgYellow).Sprintf("%s/s", formatBytes(speed))
			} else {
				speedStr = fmt.Sprintf("%s/s", formatBytes(speed))
			}
			status += speedStr + " "
		}
	}

	// Add ETA with color
	if pb.showETA && !done && percent > 0 {
		eta := pb.calculateETA()
		if eta > 0 {
			var etaStr string
			if pb.colorEnabled && pb.terminalCaps.SupportsColor {
				etaStr = color.New(color.FgMagenta).Sprintf("ETA: %s", formatDuration(eta))
			} else {
				etaStr = fmt.Sprintf("ETA: %s", formatDuration(eta))
			}
			status += etaStr + " "
		}
	}

	// Add custom label with color
	if pb.customLabel != "" {
		var labelStr string
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			labelStr = color.New(color.FgBlue).Sprint(pb.customLabel)
		} else {
			labelStr = pb.customLabel
		}
		status += labelStr + " "
	}

	// Clear line and print status
	fmt.Print("\r\033[K" + status)
}

// calculateSpeed calculates the current speed in bytes per second
func (pb *ProgressBar) calculateSpeed() int64 {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	elapsed := time.Since(pb.startTime).Seconds()
	if elapsed <= 0 {
		return 0
	}

	return int64(float64(pb.current) / elapsed)
}

// calculateETA calculates the estimated time remaining
func (pb *ProgressBar) calculateETA() time.Duration {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	if pb.current <= 0 || pb.total <= 0 {
		return 0
	}

	elapsed := time.Since(pb.startTime)
	remaining := pb.total - pb.current
	speed := float64(pb.current) / elapsed.Seconds()

	if speed <= 0 {
		return 0
	}

	etaSeconds := float64(remaining) / speed
	return time.Duration(etaSeconds) * time.Second
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration into human readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// ProgressWriter wraps an io.Writer to show progress
type ProgressWriter struct {
	writer io.Writer
	pb     *ProgressBar
}

// NewProgressWriter creates a new progress writer
func NewProgressWriter(writer io.Writer, pb *ProgressBar) *ProgressWriter {
	return &ProgressWriter{
		writer: writer,
		pb:     pb,
	}
}

// Write implements io.Writer interface
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	if pw.pb != nil {
		pw.pb.Add(int64(n))
		pw.pb.Display()
	}
	return n, err
}

// ProgressReader wraps an io.Reader to show progress
type ProgressReader struct {
	reader io.Reader
	pb     *ProgressBar
}

// NewProgressReader creates a new progress reader
func NewProgressReader(reader io.Reader, pb *ProgressBar) *ProgressReader {
	return &ProgressReader{
		reader: reader,
		pb:     pb,
	}
}

// Read implements io.Reader interface
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if pr.pb != nil {
		pr.pb.Add(int64(n))
		pr.pb.Display()
	}
	return n, err
}

// CopyFileWithProgress copies a file with enhanced progress bar and error handling
func CopyFileWithProgress(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Get file info for size
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %v", err)
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	// Create enhanced progress bar with colors
	pb := NewProgressBar(srcInfo.Size(), &ProgressBarConfig{
		Width:        50,
		ShowPercent:  true,
		ShowSpeed:    true,
		ShowETA:      true,
		CustomLabel:  fmt.Sprintf("Copying %s", filepath.Base(src)),
		RefreshRate:  50 * time.Millisecond, // More frequent updates
		ColorEnabled: true,
	})

	// Wrap writer with progress
	progressWriter := NewProgressWriter(dstFile, pb)

	// Copy with progress and error handling
	fmt.Printf("📁 Copying file: %s → %s\n", src, dst)
	_, err = io.Copy(progressWriter, srcFile)
	if err != nil {
		pb.SetError(fmt.Sprintf("Copy failed: %v", err))
		pb.Display()
		fmt.Println()
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Finish progress bar
	pb.Finish()
	pb.Display()
	fmt.Println() // New line after progress bar

	return nil
}

// DownloadWithProgress downloads a file with enhanced progress bar
func DownloadWithProgress(url, filename string) error {
	// This would integrate with HTTP client for actual downloads
	// For now, we'll simulate a download
	fmt.Printf("🌐 Downloading: %s → %s\n", url, filename)

	// Simulate download progress
	totalSize := int64(1024 * 1024) // 1MB simulation
	pb := NewProgressBar(totalSize, &ProgressBarConfig{
		Width:        50,
		ShowPercent:  true,
		ShowSpeed:    true,
		ShowETA:      true,
		CustomLabel:  fmt.Sprintf("Downloading %s", filepath.Base(filename)),
		RefreshRate:  50 * time.Millisecond,
		ColorEnabled: true,
	})

	// Simulate download with error handling
	for i := int64(0); i < totalSize; i += 1024 {
		// Simulate occasional network errors
		if i > totalSize/2 && i%10000 == 0 {
			pb.SetError("Network timeout - retrying...")
			pb.Display()
			time.Sleep(100 * time.Millisecond)
			pb.ClearError()
		}

		pb.Update(i)
		pb.Display()
		time.Sleep(5 * time.Millisecond) // Simulate network delay
	}

	pb.Finish()
	pb.Display()
	fmt.Println()

	return nil
}

// ProcessMultipleFilesWithProgress processes multiple files with multiple progress bars
func ProcessMultipleFilesWithProgress(files []string, operation string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to process")
	}

	// Create multi-progress manager
	mpm := NewMultiProgressManager()
	mpm.Start()
	defer mpm.Stop()

	// Add progress bars for each file
	bars := make([]*ProgressBar, len(files))
	for i, file := range files {
		// Get file size
		info, err := os.Stat(file)
		if err != nil {
			mpm.Stop()
			return fmt.Errorf("failed to get file info for %s: %v", file, err)
		}

		// Create progress bar for this file
		bars[i] = mpm.AddBar(info.Size(), &ProgressBarConfig{
			Width:        40,
			ShowPercent:  true,
			ShowSpeed:    true,
			ShowETA:      true,
			CustomLabel:  fmt.Sprintf("%s %s", operation, filepath.Base(file)),
			RefreshRate:  100 * time.Millisecond,
			ColorEnabled: true,
		})
	}

	// Process each file
	for i, file := range files {
		// Simulate file processing
		info, _ := os.Stat(file)
		totalSize := info.Size()

		for processed := int64(0); processed < totalSize; processed += 1024 {
			bars[i].Update(processed)
			time.Sleep(2 * time.Millisecond) // Simulate processing time
		}

		bars[i].Finish()
	}

	// Wait a moment to show final state
	time.Sleep(500 * time.Millisecond)

	return nil
}

// ProcessWithProgress runs a process with enhanced progress bar and error handling
func ProcessWithProgress(total int64, processName string, processFunc func(pb *ProgressBar) error) error {
	pb := NewProgressBar(total, &ProgressBarConfig{
		Width:        50,
		ShowPercent:  true,
		ShowSpeed:    true,
		ShowETA:      true,
		CustomLabel:  processName,
		RefreshRate:  100 * time.Millisecond,
		ColorEnabled: true,
	})

	fmt.Printf("⚙️ Processing: %s\n", processName)

	err := processFunc(pb)
	if err != nil {
		pb.SetError(fmt.Sprintf("Process failed: %v", err))
		pb.Display()
		fmt.Println()
		return err
	}

	pb.Finish()
	pb.Display()
	fmt.Println()

	return nil
}
