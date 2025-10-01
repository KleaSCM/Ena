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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// ProgressBarTheme defines visual styling for progress bars
type ProgressBarTheme struct {
	FilledChar   string
	EmptyChar    string
	FilledColor  color.Attribute
	EmptyColor   color.Attribute
	PercentColor color.Attribute
	SpeedColor   color.Attribute
	ETAColor     color.Attribute
	LabelColor   color.Attribute
	ErrorColor   color.Attribute
	PauseColor   color.Attribute
}

// EventType represents different progress bar events
type EventType int

const (
	EventStart EventType = iota
	EventUpdate
	EventPause
	EventResume
	EventComplete
	EventError
	EventFinish
)

// EventCallback is a function that gets called on progress bar events
type EventCallback func(event EventType, pb *ProgressBar, data interface{})

// Predefined themes
var (
	DefaultTheme = ProgressBarTheme{
		FilledChar:   "█",
		EmptyChar:    "░",
		FilledColor:  color.FgGreen,
		EmptyColor:   color.FgWhite,
		PercentColor: color.FgCyan,
		SpeedColor:   color.FgYellow,
		ETAColor:     color.FgMagenta,
		LabelColor:   color.FgBlue,
		ErrorColor:   color.FgRed,
		PauseColor:   color.FgYellow,
	}

	RainbowTheme = ProgressBarTheme{
		FilledChar:   "█",
		EmptyChar:    "░",
		FilledColor:  color.FgMagenta,
		EmptyColor:   color.FgWhite,
		PercentColor: color.FgCyan,
		SpeedColor:   color.FgYellow,
		ETAColor:     color.FgGreen,
		LabelColor:   color.FgBlue,
		ErrorColor:   color.FgRed,
		PauseColor:   color.FgYellow,
	}

	MinimalTheme = ProgressBarTheme{
		FilledChar:   "=",
		EmptyChar:    "-",
		FilledColor:  color.FgBlue,
		EmptyColor:   color.FgWhite,
		PercentColor: color.FgWhite,
		SpeedColor:   color.FgWhite,
		ETAColor:     color.FgWhite,
		LabelColor:   color.FgWhite,
		ErrorColor:   color.FgRed,
		PauseColor:   color.FgYellow,
	}
)

// ProgressBarInterface defines the interface for progress bars (for testing)
type ProgressBarInterface interface {
	Update(current int64)
	Add(increment int64)
	SetTotal(total int64)
	Finish()
	Pause()
	Resume()
	IsPaused() bool
	SetError(message string)
	ClearError()
	IsError() bool
	Display()
	SaveState() error
	loadState() error
}

// MultiProgressManagerInterface defines the interface for multi-progress managers (for testing)
type MultiProgressManagerInterface interface {
	AddBar(total int64, config *ProgressBarConfig) *ProgressBar
	DisplayAll()
	Start()
	Stop()
	SaveAllStates() error
}

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

	// Adaptive refresh
	adaptiveRefresh bool
	lastSpeed       int64
	refreshHistory  []time.Duration

	// Custom themes and event hooks
	theme          ProgressBarTheme
	eventCallbacks map[EventType][]EventCallback

	// Concurrency optimizations
	updateChannel  chan int64
	displayChannel chan bool
	stopChannel    chan bool

	// Dynamic resizing
	resizeChannel chan struct{}
	originalWidth int
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
	Width           int
	ShowPercent     bool
	ShowSpeed       bool
	ShowETA         bool
	CustomLabel     string
	RefreshRate     time.Duration
	ColorEnabled    bool
	MultiBarIndex   int
	StateFile       string
	Persistent      bool
	AdaptiveRefresh bool
	Theme           *ProgressBarTheme
	EventCallbacks  map[EventType][]EventCallback
	EnableChannels  bool
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

// getTerminalSize gets terminal dimensions using golang.org/x/term
func getTerminalSize() (width, height int, err error) {
	// Try to get terminal size from stdout
	if width, height, err = term.GetSize(int(os.Stdout.Fd())); err == nil {
		return width, height, nil
	}

	// Fallback: try stderr
	if width, height, err = term.GetSize(int(os.Stderr.Fd())); err == nil {
		return width, height, nil
	}

	// Final fallback: environment variables
	if term := os.Getenv("COLUMNS"); term != "" {
		if w, err := fmt.Sscanf(term, "%d", &width); err == nil && w == 1 {
			// Width found, try to get height
			if term := os.Getenv("LINES"); term != "" {
				if h, err := fmt.Sscanf(term, "%d", &height); err == nil && h == 1 {
					return width, height, nil
				}
			}
			return width, 24, nil // Default height
		}
	}

	// Ultimate fallback
	return 80, 24, fmt.Errorf("unable to determine terminal size")
}

// NewProgressBar creates a new progress bar with enhanced features
func NewProgressBar(total int64, config *ProgressBarConfig) *ProgressBar {
	if config == nil {
		config = &ProgressBarConfig{
			Width:           50,
			ShowPercent:     true,
			ShowSpeed:       true,
			ShowETA:         true,
			RefreshRate:     100 * time.Millisecond,
			ColorEnabled:    true,
			MultiBarIndex:   0,
			Persistent:      false,
			AdaptiveRefresh: true,
			Theme:           &DefaultTheme,
			EventCallbacks:  make(map[EventType][]EventCallback),
			EnableChannels:  false,
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
		total:           total,
		current:         0,
		width:           config.Width,
		showPercent:     config.ShowPercent,
		showSpeed:       config.ShowSpeed,
		showETA:         config.ShowETA,
		startTime:       time.Now(),
		lastUpdate:      time.Now(),
		lastCurrent:     0,
		customLabel:     config.CustomLabel,
		refreshRate:     config.RefreshRate,
		lastDisplay:     time.Now(),
		colorEnabled:    config.ColorEnabled,
		multiBarIndex:   config.MultiBarIndex,
		errorOccurred:   false,
		paused:          0,
		totalPauseTime:  0,
		terminalCaps:    terminalCaps,
		stateFile:       config.StateFile,
		persistent:      config.Persistent,
		adaptiveRefresh: config.AdaptiveRefresh,
		lastSpeed:       0,
		refreshHistory:  make([]time.Duration, 0, 10),
		theme:           DefaultTheme,
		eventCallbacks:  make(map[EventType][]EventCallback),
		originalWidth:   config.Width,
	}

	// Initialize channels for concurrency optimization
	if config.EnableChannels {
		pb.updateChannel = make(chan int64, 100)
		pb.displayChannel = make(chan bool, 10)
		pb.stopChannel = make(chan bool, 1)
		pb.resizeChannel = make(chan struct{}, 1)

		// Start background goroutines
		go pb.updateWorker()
		go pb.displayWorker()
		go pb.resizeWorker()
	}

	// Set theme if provided
	if config.Theme != nil {
		pb.theme = *config.Theme
	}

	// Copy event callbacks
	if config.EventCallbacks != nil {
		for eventType, callbacks := range config.EventCallbacks {
			pb.eventCallbacks[eventType] = callbacks
		}
	}

	// Load persistent state if enabled
	if config.Persistent && config.StateFile != "" {
		pb.loadState()
	}

	// Trigger start event
	pb.triggerEvent(EventStart, nil)

	return pb
}

// Update updates the progress bar with new current value
func (pb *ProgressBar) Update(current int64) {
	if pb.updateChannel != nil {
		// Use channel for concurrency optimization
		select {
		case pb.updateChannel <- current:
		default:
			// Channel full, fall back to direct update
			pb.mutex.Lock()
			pb.current = current
			pb.lastUpdate = time.Now()
			pb.mutex.Unlock()
		}
	} else {
		// Direct update for non-channel mode
		pb.mutex.Lock()
		pb.current = current
		pb.lastUpdate = time.Now()
		pb.mutex.Unlock()
	}
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
	pb.triggerEvent(EventComplete, nil)
	pb.triggerEvent(EventFinish, nil)
}

// SetError sets an error state for the progress bar
func (pb *ProgressBar) SetError(message string) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.errorOccurred = true
	pb.errorMessage = message
	pb.done = false
	pb.triggerEvent(EventError, message)
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
		pb.triggerEvent(EventPause, nil)
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
		pb.triggerEvent(EventResume, nil)
	}
}

// IsPaused returns true if the progress bar is paused
func (pb *ProgressBar) IsPaused() bool {
	return atomic.LoadInt32(&pb.paused) == 1
}

// ProgressBarState represents the serializable state of a progress bar
type ProgressBarState struct {
	Current        int64           `json:"current"`
	Total          int64           `json:"total"`
	StartTime      time.Time       `json:"startTime"`
	TotalPauseTime time.Duration   `json:"totalPauseTime"`
	Done           bool            `json:"done"`
	ErrorOccurred  bool            `json:"errorOccurred"`
	ErrorMessage   string          `json:"errorMessage"`
	Paused         bool            `json:"paused"`
	CustomLabel    string          `json:"customLabel"`
	LastSpeed      int64           `json:"lastSpeed"`
	RefreshHistory []time.Duration `json:"refreshHistory"`
	Version        string          `json:"version"`
}

// SaveState saves the current state to file in JSON format
func (pb *ProgressBar) SaveState() error {
	if !pb.persistent || pb.stateFile == "" {
		return nil
	}

	pb.mutex.RLock()
	state := ProgressBarState{
		Current:        pb.current,
		Total:          pb.total,
		StartTime:      pb.startTime,
		TotalPauseTime: pb.totalPauseTime,
		Done:           pb.done,
		ErrorOccurred:  pb.errorOccurred,
		ErrorMessage:   pb.errorMessage,
		Paused:         pb.IsPaused(),
		CustomLabel:    pb.customLabel,
		LastSpeed:      pb.lastSpeed,
		RefreshHistory: make([]time.Duration, len(pb.refreshHistory)),
		Version:        "1.0",
	}
	copy(state.RefreshHistory, pb.refreshHistory)
	pb.mutex.RUnlock()

	// Serialize to JSON with proper indentation
	jsonData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state to JSON: %v", err)
	}

	return os.WriteFile(pb.stateFile, jsonData, 0644)
}

// loadState loads state from file with robust JSON parsing and error handling
func (pb *ProgressBar) loadState() error {
	if !pb.persistent || pb.stateFile == "" {
		return nil
	}

	data, err := os.ReadFile(pb.stateFile)
	if err != nil {
		return err // File doesn't exist or can't be read
	}

	// Try JSON parsing first
	var state ProgressBarState
	if err := json.Unmarshal(data, &state); err == nil {
		// Successfully parsed JSON
		pb.mutex.Lock()
		pb.current = state.Current
		pb.total = state.Total
		pb.startTime = state.StartTime
		pb.totalPauseTime = state.TotalPauseTime
		pb.done = state.Done
		pb.errorOccurred = state.ErrorOccurred
		pb.errorMessage = state.ErrorMessage
		pb.customLabel = state.CustomLabel
		pb.lastSpeed = state.LastSpeed
		pb.refreshHistory = make([]time.Duration, len(state.RefreshHistory))
		copy(pb.refreshHistory, state.RefreshHistory)

		// Set pause state if needed
		if state.Paused {
			atomic.StoreInt32(&pb.paused, 1)
		}
		pb.mutex.Unlock()
		return nil
	}

	// Fallback: try legacy key=value format
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "current":
			if val, err := fmt.Sscanf(value, "%d", &pb.current); err != nil || val != 1 {
				return fmt.Errorf("failed to parse current value: %v", err)
			}
		case "total":
			if val, err := fmt.Sscanf(value, "%d", &pb.total); err != nil || val != 1 {
				return fmt.Errorf("failed to parse total value: %v", err)
			}
		case "done":
			pb.done = (value == "true")
		case "error":
			pb.errorOccurred = (value == "true")
		case "paused":
			if value == "true" {
				atomic.StoreInt32(&pb.paused, 1)
			}
		}
	}

	return nil
}

// shouldUpdateAdaptive determines if the progress bar should update based on adaptive refresh logic
func (pb *ProgressBar) shouldUpdateAdaptive() bool {
	now := time.Now()
	timeSinceLastUpdate := now.Sub(pb.lastDisplay)

	// Always update if enough time has passed (minimum refresh rate)
	if timeSinceLastUpdate >= pb.refreshRate {
		return true
	}

	// Calculate current speed
	currentSpeed := pb.calculateSpeed()

	// If speed has changed significantly, update more frequently
	if pb.lastSpeed > 0 {
		speedChange := float64(currentSpeed-pb.lastSpeed) / float64(pb.lastSpeed)
		if speedChange > 0.1 || speedChange < -0.1 { // 10% change threshold
			pb.lastSpeed = currentSpeed
			return true
		}
	}

	// Update more frequently for faster progress
	if currentSpeed > pb.lastSpeed*2 {
		pb.lastSpeed = currentSpeed
		return true
	}

	// Update less frequently for slower progress
	if currentSpeed < pb.lastSpeed/2 && pb.lastSpeed > 0 {
		pb.lastSpeed = currentSpeed
		return timeSinceLastUpdate >= pb.refreshRate*2
	}

	// Record refresh timing for analysis
	pb.refreshHistory = append(pb.refreshHistory, timeSinceLastUpdate)
	if len(pb.refreshHistory) > 10 {
		pb.refreshHistory = pb.refreshHistory[1:] // Keep only last 10
	}

	return false
}

// triggerEvent calls all registered callbacks for an event
func (pb *ProgressBar) triggerEvent(event EventType, data interface{}) {
	if pb.eventCallbacks == nil {
		return
	}

	callbacks, exists := pb.eventCallbacks[event]
	if !exists {
		return
	}

	for _, callback := range callbacks {
		go func(cb EventCallback) {
			defer func() {
				if r := recover(); r != nil {
					// Ignore panics in event callbacks
				}
			}()
			cb(event, pb, data)
		}(callback)
	}
}

// updateWorker processes updates from the channel
func (pb *ProgressBar) updateWorker() {
	for {
		select {
		case current := <-pb.updateChannel:
			pb.mutex.Lock()
			pb.current = current
			pb.lastUpdate = time.Now()
			pb.mutex.Unlock()

			pb.triggerEvent(EventUpdate, current)

			// Trigger display update
			select {
			case pb.displayChannel <- true:
			default:
			}

		case <-pb.stopChannel:
			return
		}
	}
}

// displayWorker handles display updates
func (pb *ProgressBar) displayWorker() {
	ticker := time.NewTicker(pb.refreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-pb.displayChannel:
			pb.Display()
		case <-ticker.C:
			pb.Display()
		case <-pb.stopChannel:
			return
		}
	}
}

// resizeWorker handles terminal resize events
func (pb *ProgressBar) resizeWorker() {
	for {
		select {
		case <-pb.resizeChannel:
			pb.handleResize()
		case <-pb.stopChannel:
			return
		}
	}
}

// handleResize adjusts the progress bar width based on terminal size
func (pb *ProgressBar) handleResize() {
	if pb.terminalCaps == nil {
		return
	}

	newWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return
	}

	// Adjust width, keeping some margin
	adjustedWidth := newWidth - 20
	if adjustedWidth < 10 {
		adjustedWidth = 10
	}

	pb.mutex.Lock()
	pb.width = adjustedWidth
	pb.mutex.Unlock()
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

	// Check if we should update based on refresh rate (adaptive or fixed)
	shouldUpdate := false
	if pb.adaptiveRefresh {
		shouldUpdate = pb.shouldUpdateAdaptive()
	} else {
		shouldUpdate = time.Since(pb.lastDisplay) >= pb.refreshRate
	}

	if !shouldUpdate && !done && !errorOccurred && !paused {
		return
	}
	pb.lastDisplay = time.Now()

	// Handle error display
	if errorOccurred {
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			color.New(pb.theme.ErrorColor, color.Bold).Printf("\r\033[K❌ Error: %s", errorMessage)
		} else {
			fmt.Printf("\r\033[K❌ Error: %s", errorMessage)
		}
		return
	}

	// Handle paused display
	if paused {
		if pb.colorEnabled && pb.terminalCaps.SupportsColor {
			color.New(pb.theme.PauseColor, color.Bold).Printf("\r\033[K⏸️ Paused: %s", pb.customLabel)
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

	// Build progress bar with custom theme
	var bar string
	if pb.colorEnabled && pb.terminalCaps.SupportsColor {
		filledBar := color.New(pb.theme.FilledColor).Sprint(strings.Repeat(pb.theme.FilledChar, filledWidth))
		emptyBar := color.New(pb.theme.EmptyColor).Sprint(strings.Repeat(pb.theme.EmptyChar, pb.width-filledWidth))
		bar = filledBar + emptyBar
	} else {
		bar = strings.Repeat(pb.theme.FilledChar, filledWidth) + strings.Repeat(pb.theme.EmptyChar, pb.width-filledWidth)
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
				percentStr = color.New(pb.theme.PercentColor).Sprintf("%.1f%%", percent)
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
				speedStr = color.New(pb.theme.SpeedColor).Sprintf("%s/s", formatBytes(speed))
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
				etaStr = color.New(pb.theme.ETAColor).Sprintf("ETA: %s", formatDuration(eta))
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
			labelStr = color.New(pb.theme.LabelColor).Sprint(pb.customLabel)
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

// DownloadFileWithProgress downloads a file from URL with real HTTP and progress tracking
func DownloadFileWithProgress(url, filename string, config *ProgressBarConfig) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Start HTTP request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to start download: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Get content length
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return fmt.Errorf("unknown content length")
	}

	// Create progress bar
	if config == nil {
		config = &ProgressBarConfig{
			Width:          50,
			ShowPercent:    true,
			ShowSpeed:      true,
			ShowETA:        true,
			CustomLabel:    filename,
			RefreshRate:    100 * time.Millisecond,
			ColorEnabled:   true,
			Theme:          &DefaultTheme,
			EnableChannels: true,
		}
	}

	pb := NewProgressBar(contentLength, config)

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Download with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer
	totalRead := int64(0)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// Write to file
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				pb.SetError(fmt.Sprintf("write error: %v", writeErr))
				return writeErr
			}

			// Update progress
			totalRead += int64(n)
			pb.Update(totalRead)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			pb.SetError(fmt.Sprintf("read error: %v", err))
			return err
		}
	}

	// Finish progress bar
	pb.Finish()
	pb.Display()
	fmt.Println()

	return nil
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
