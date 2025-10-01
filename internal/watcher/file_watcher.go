/**
 * File Watcher Module
 *
 * Provides real-time file system monitoring with live updates and notifications.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: file_watcher.go
 * Description: Real-time file system monitoring with event handling and live updates
 */

package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	EventType EventType
	Timestamp time.Time
	Size      int64
	IsDir     bool
}

// EventType represents the type of file system event
type EventType int

const (
	EventCreate EventType = iota
	EventModify
	EventDelete
	EventRename
	EventMove
)

// String returns the string representation of EventType
func (et EventType) String() string {
	switch et {
	case EventCreate:
		return "CREATE"
	case EventModify:
		return "MODIFY"
	case EventDelete:
		return "DELETE"
	case EventRename:
		return "RENAME"
	case EventMove:
		return "MOVE"
	default:
		return "UNKNOWN"
	}
}

// EventCallback is a function that gets called when file events occur
type EventCallback func(event FileEvent)

// WatchConfig holds configuration for file watching
type WatchConfig struct {
	Paths            []string
	Recursive        bool
	IncludeHidden    bool
	FileExtensions   []string
	ExcludePatterns  []string
	DebounceTime     time.Duration
	EventCallbacks   map[EventType][]EventCallback
	DebugMode        bool
	LogIgnoredEvents bool
	BatchEvents      bool
	BatchSize        int
	BatchTimeout     time.Duration
	EventPriority    map[EventType]int
	MetricsEnabled   bool
	ConfigFile       string
	HotReload        bool
	ErrorRecovery    bool
	MaxRetries       int
	RetryDelay       time.Duration
}

// WatcherMetrics holds performance and usage metrics
type WatcherMetrics struct {
	EventsProcessed   int64
	EventsBatched     int64
	EventsDropped     int64
	EventsDebounced   int64
	EventsIgnored     int64
	PathsWatched      int64
	ErrorsEncountered int64
	RetriesAttempted  int64
	StartTime         time.Time
	LastEventTime     time.Time
	AverageLatency    time.Duration
	PeakLatency       time.Duration
}

// EventBatch represents a batch of events for processing
type EventBatch struct {
	Events    []FileEvent
	Timestamp time.Time
	Priority  int
}

// FileWatcher manages file system monitoring
type FileWatcher struct {
	watcher        *fsnotify.Watcher
	config         *WatchConfig
	callbacks      map[EventType][]EventCallback
	mutex          sync.RWMutex
	running        bool
	stopChan       chan bool
	eventChan      chan FileEvent
	batchChan      chan EventBatch
	debounce       map[string]time.Time
	debounceMu     sync.Mutex
	watchedPaths   []string
	pathsMutex     sync.RWMutex
	moveTracker    map[string]time.Time
	moveMutex      sync.Mutex
	metrics        *WatcherMetrics
	metricsMutex   sync.RWMutex
	configMutex    sync.RWMutex
	reloadChan     chan bool
	errorChan      chan error
	retryCount     int64
	lastConfigHash string
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(config *WatchConfig) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	if config == nil {
		config = &WatchConfig{
			Recursive:        true,
			IncludeHidden:    false,
			DebounceTime:     100 * time.Millisecond,
			EventCallbacks:   make(map[EventType][]EventCallback),
			DebugMode:        false,
			LogIgnoredEvents: false,
			BatchEvents:      false,
			BatchSize:        10,
			BatchTimeout:     1 * time.Second,
			EventPriority:    make(map[EventType]int),
			MetricsEnabled:   false,
			HotReload:        false,
			ErrorRecovery:    true,
			MaxRetries:       3,
			RetryDelay:       5 * time.Second,
		}
	}

	// Set default event priorities
	if config.EventPriority == nil {
		config.EventPriority = map[EventType]int{
			EventDelete: 1, // Highest priority
			EventCreate: 2,
			EventModify: 3,
			EventMove:   4,
			EventRename: 5, // Lowest priority
		}
	}

	fw := &FileWatcher{
		watcher:      watcher,
		config:       config,
		callbacks:    make(map[EventType][]EventCallback),
		running:      false,
		stopChan:     make(chan bool, 1),
		eventChan:    make(chan FileEvent, 100),
		batchChan:    make(chan EventBatch, 10),
		debounce:     make(map[string]time.Time),
		watchedPaths: make([]string, 0),
		moveTracker:  make(map[string]time.Time),
		metrics: &WatcherMetrics{
			StartTime: time.Now(),
		},
		reloadChan: make(chan bool, 1),
		errorChan:  make(chan error, 10),
	}

	// Copy event callbacks
	if config.EventCallbacks != nil {
		for eventType, callbacks := range config.EventCallbacks {
			fw.callbacks[eventType] = callbacks
		}
	}

	return fw, nil
}

// AddPath adds a path to watch
func (fw *FileWatcher) AddPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Add to watcher
	err := fw.watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to add path to watcher: %v", err)
	}

	// Track watched path
	fw.pathsMutex.Lock()
	fw.watchedPaths = append(fw.watchedPaths, path)
	fw.pathsMutex.Unlock()

	if fw.config.DebugMode {
		log.Printf("Added path to watcher: %s", path)
	}

	// Add recursive paths if enabled
	if fw.config.Recursive {
		err = fw.addRecursivePaths(path)
		if err != nil {
			return fmt.Errorf("failed to add recursive paths: %v", err)
		}
	}

	return nil
}

// addRecursivePaths adds all subdirectories to the watcher
func (fw *FileWatcher) addRecursivePaths(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if fw.config.DebugMode {
				log.Printf("Error walking path %s: %v", path, err)
			}
			return err
		}

		// Skip hidden files if not included
		if !fw.config.IncludeHidden && strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip excluded patterns
		if fw.isExcluded(path) {
			if fw.config.LogIgnoredEvents {
				log.Printf("Ignored excluded path: %s", path)
			}
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Add directory to watcher
		if info.IsDir() {
			err = fw.watcher.Add(path)
			if err != nil {
				if fw.config.DebugMode {
					log.Printf("Failed to add directory %s: %v", path, err)
				}
				return err
			}

			// Track watched path
			fw.pathsMutex.Lock()
			fw.watchedPaths = append(fw.watchedPaths, path)
			fw.pathsMutex.Unlock()

			if fw.config.DebugMode {
				log.Printf("Added recursive directory: %s", path)
			}
		}

		return nil
	})
}

// isExcluded checks if a path matches any exclude patterns
func (fw *FileWatcher) isExcluded(path string) bool {
	for _, pattern := range fw.config.ExcludePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}

// AddCallback adds an event callback
func (fw *FileWatcher) AddCallback(eventType EventType, callback EventCallback) {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	fw.callbacks[eventType] = append(fw.callbacks[eventType], callback)
}

// Start starts the file watcher
func (fw *FileWatcher) Start() error {
	if fw.running {
		return fmt.Errorf("file watcher is already running")
	}

	fw.running = true

	// Start event processing goroutine
	go fw.processEvents()

	// Start file system event processing
	go fw.watchEvents()

	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() error {
	if !fw.running {
		return fmt.Errorf("file watcher is not running")
	}

	fw.running = false
	fw.stopChan <- true

	return fw.watcher.Close()
}

// watchEvents processes file system events
func (fw *FileWatcher) watchEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)

		case <-fw.stopChan:
			return
		}
	}
}

// handleEvent processes a file system event
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	if fw.config.DebugMode {
		log.Printf("Received event: %s %s", event.Op.String(), event.Name)
	}

	// Check if file extension is filtered
	if len(fw.config.FileExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(event.Name))
		found := false
		for _, allowedExt := range fw.config.FileExtensions {
			if ext == strings.ToLower(allowedExt) {
				found = true
				break
			}
		}
		if !found {
			if fw.config.LogIgnoredEvents {
				log.Printf("Ignored file extension: %s (ext: %s)", event.Name, ext)
			}
			return
		}
	}

	// Enhanced debounce with event type consideration
	if fw.config.DebounceTime > 0 {
		fw.debounceMu.Lock()
		debounceKey := fmt.Sprintf("%s:%s", event.Name, event.Op.String())
		lastTime, exists := fw.debounce[debounceKey]
		now := time.Now()

		if exists && now.Sub(lastTime) < fw.config.DebounceTime {
			if fw.config.LogIgnoredEvents {
				log.Printf("Debounced event: %s %s", event.Op.String(), event.Name)
			}
			fw.debounceMu.Unlock()
			return
		}

		fw.debounce[debounceKey] = now
		fw.debounceMu.Unlock()
	}

	// Get file info
	info, err := os.Stat(event.Name)
	size := int64(0)
	isDir := false
	if err == nil {
		size = info.Size()
		isDir = info.IsDir()
	} else if fw.config.DebugMode {
		log.Printf("Could not stat file %s: %v", event.Name, err)
	}

	// Determine event type with enhanced move detection
	var eventType EventType
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		// Check if this might be a move operation
		if fw.isMoveOperation(event.Name) {
			eventType = EventMove
		} else {
			eventType = EventCreate
		}
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = EventModify
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		// Track potential move source
		fw.trackPotentialMove(event.Name)
		eventType = EventDelete
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		// Track potential move source
		fw.trackPotentialMove(event.Name)
		eventType = EventRename
	default:
		eventType = EventModify
	}

	// Create file event
	fileEvent := FileEvent{
		Path:      event.Name,
		EventType: eventType,
		Timestamp: time.Now(),
		Size:      size,
		IsDir:     isDir,
	}

	if fw.config.DebugMode {
		log.Printf("Processed event: %s %s (size: %d, isDir: %t)",
			eventType.String(), event.Name, size, isDir)
	}

	// Send to event channel
	select {
	case fw.eventChan <- fileEvent:
	default:
		if fw.config.DebugMode {
			log.Printf("Event channel full, dropping event: %s", event.Name)
		}
	}
}

// processEvents processes events from the event channel
func (fw *FileWatcher) processEvents() {
	for {
		select {
		case event := <-fw.eventChan:
			fw.triggerCallbacks(event)

		case <-fw.stopChan:
			return
		}
	}
}

// trackPotentialMove tracks a file that might be moved
func (fw *FileWatcher) trackPotentialMove(path string) {
	fw.moveMutex.Lock()
	defer fw.moveMutex.Unlock()
	fw.moveTracker[path] = time.Now()
}

// isMoveOperation checks if a create event might be part of a move operation
func (fw *FileWatcher) isMoveOperation(path string) bool {
	fw.moveMutex.Lock()
	defer fw.moveMutex.Unlock()

	// Check if any tracked move source exists within a reasonable time window
	now := time.Now()
	for trackedPath, timestamp := range fw.moveTracker {
		if now.Sub(timestamp) < 5*time.Second {
			// Simple heuristic: if the base names match, it might be a move
			if filepath.Base(trackedPath) == filepath.Base(path) {
				delete(fw.moveTracker, trackedPath)
				return true
			}
		} else {
			// Clean up old entries
			delete(fw.moveTracker, trackedPath)
		}
	}
	return false
}

// triggerCallbacks calls all registered callbacks for an event
func (fw *FileWatcher) triggerCallbacks(event FileEvent) {
	fw.mutex.RLock()
	callbacks, exists := fw.callbacks[event.EventType]
	fw.mutex.RUnlock()

	if !exists {
		return
	}

	for _, callback := range callbacks {
		go func(cb EventCallback) {
			defer func() {
				if r := recover(); r != nil {
					if fw.config.DebugMode {
						log.Printf("Callback panic recovered: %v", r)
					}
				}
			}()
			cb(event)
		}(callback)
	}
}

// GetWatchedPaths returns all currently watched paths
func (fw *FileWatcher) GetWatchedPaths() []string {
	fw.pathsMutex.RLock()
	defer fw.pathsMutex.RUnlock()

	// Return a copy to prevent external modification
	result := make([]string, len(fw.watchedPaths))
	copy(result, fw.watchedPaths)
	return result
}

// IsRunning returns true if the watcher is running
func (fw *FileWatcher) IsRunning() bool {
	return fw.running
}

// GetStats returns statistics about the file watcher
func (fw *FileWatcher) GetStats() map[string]interface{} {
	fw.mutex.RLock()
	defer fw.mutex.RUnlock()

	fw.metricsMutex.RLock()
	defer fw.metricsMutex.RUnlock()

	return map[string]interface{}{
		"running":        fw.running,
		"watched_paths":  len(fw.watchedPaths),
		"callbacks":      len(fw.callbacks),
		"debounce_time":  fw.config.DebounceTime,
		"recursive":      fw.config.Recursive,
		"include_hidden": fw.config.IncludeHidden,
		"metrics":        fw.metrics,
		"retry_count":    atomic.LoadInt64(&fw.retryCount),
	}
}

// AddPathDynamic adds a path while the watcher is running
func (fw *FileWatcher) AddPathDynamic(path string) error {
	if !fw.running {
		return fw.AddPath(path)
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Add to watcher
	err := fw.watcher.Add(path)
	if err != nil {
		if fw.config.ErrorRecovery {
			return fw.handleErrorWithRetry("add_path", err, func() error {
				return fw.watcher.Add(path)
			})
		}
		return fmt.Errorf("failed to add path to watcher: %v", err)
	}

	// Track watched path
	fw.pathsMutex.Lock()
	fw.watchedPaths = append(fw.watchedPaths, path)
	fw.pathsMutex.Unlock()

	// Update metrics
	if fw.config.MetricsEnabled {
		atomic.AddInt64(&fw.metrics.PathsWatched, 1)
	}

	if fw.config.DebugMode {
		log.Printf("Dynamically added path: %s", path)
	}

	// Add recursive paths if enabled
	if fw.config.Recursive {
		err = fw.addRecursivePaths(path)
		if err != nil {
			return fmt.Errorf("failed to add recursive paths: %v", err)
		}
	}

	return nil
}

// RemovePathDynamic removes a path while the watcher is running
func (fw *FileWatcher) RemovePathDynamic(path string) error {
	if !fw.running {
		return fmt.Errorf("watcher is not running")
	}

	// Remove from watcher
	err := fw.watcher.Remove(path)
	if err != nil {
		if fw.config.ErrorRecovery {
			return fw.handleErrorWithRetry("remove_path", err, func() error {
				return fw.watcher.Remove(path)
			})
		}
		return fmt.Errorf("failed to remove path from watcher: %v", err)
	}

	// Remove from tracked paths
	fw.pathsMutex.Lock()
	for i, p := range fw.watchedPaths {
		if p == path {
			fw.watchedPaths = append(fw.watchedPaths[:i], fw.watchedPaths[i+1:]...)
			break
		}
	}
	fw.pathsMutex.Unlock()

	if fw.config.DebugMode {
		log.Printf("Dynamically removed path: %s", path)
	}

	return nil
}

// ReloadConfig reloads configuration from file
func (fw *FileWatcher) ReloadConfig() error {
	if fw.config.ConfigFile == "" {
		return fmt.Errorf("no config file specified")
	}

	data, err := os.ReadFile(fw.config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var newConfig WatchConfig
	err = json.Unmarshal(data, &newConfig)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// Update configuration
	fw.configMutex.Lock()
	fw.config = &newConfig
	fw.configMutex.Unlock()

	if fw.config.DebugMode {
		log.Printf("Configuration reloaded from: %s", fw.config.ConfigFile)
	}

	return nil
}

// GetMetrics returns detailed metrics
func (fw *FileWatcher) GetMetrics() *WatcherMetrics {
	fw.metricsMutex.RLock()
	defer fw.metricsMutex.RUnlock()

	// Create a copy to prevent external modification
	metrics := *fw.metrics
	return &metrics
}

// ResetMetrics resets all metrics counters
func (fw *FileWatcher) ResetMetrics() {
	fw.metricsMutex.Lock()
	defer fw.metricsMutex.Unlock()

	fw.metrics = &WatcherMetrics{
		StartTime: time.Now(),
	}
}

// handleErrorWithRetry handles errors with retry logic
func (fw *FileWatcher) handleErrorWithRetry(operation string, err error, retryFunc func() error) error {
	retryCount := atomic.LoadInt64(&fw.retryCount)
	if retryCount >= int64(fw.config.MaxRetries) {
		atomic.AddInt64(&fw.metrics.ErrorsEncountered, 1)
		return fmt.Errorf("max retries exceeded for %s: %v", operation, err)
	}

	atomic.AddInt64(&fw.retryCount, 1)
	atomic.AddInt64(&fw.metrics.RetriesAttempted, 1)

	if fw.config.DebugMode {
		log.Printf("Retrying %s (attempt %d): %v", operation, retryCount+1, err)
	}

	time.Sleep(fw.config.RetryDelay)
	return retryFunc()
}

// StartEnhanced starts the watcher with enhanced features
func (fw *FileWatcher) StartEnhanced() error {
	if fw.running {
		return fmt.Errorf("file watcher is already running")
	}

	fw.running = true

	// Start event processing goroutine
	go fw.processEventsEnhanced()

	// Start file system event processing
	go fw.watchEventsEnhanced()

	// Start config hot-reload if enabled
	if fw.config.HotReload && fw.config.ConfigFile != "" {
		go fw.watchConfigFile()
	}

	// Start error recovery goroutine
	if fw.config.ErrorRecovery {
		go fw.handleErrors()
	}

	return nil
}

// processEventsEnhanced processes events with batching and prioritization
func (fw *FileWatcher) processEventsEnhanced() {
	var eventBatch []FileEvent
	var batchTimer *time.Timer
	var batchTimeout time.Duration

	if fw.config.BatchEvents {
		batchTimeout = fw.config.BatchTimeout
		batchTimer = time.NewTimer(batchTimeout)
		defer batchTimer.Stop()
	}

	for {
		select {
		case event := <-fw.eventChan:
			if fw.config.MetricsEnabled {
				atomic.AddInt64(&fw.metrics.EventsProcessed, 1)
				fw.metricsMutex.Lock()
				fw.metrics.LastEventTime = time.Now()
				fw.metricsMutex.Unlock()
			}

			if fw.config.BatchEvents {
				eventBatch = append(eventBatch, event)

				// Process batch if size reached
				if len(eventBatch) >= fw.config.BatchSize {
					fw.processBatch(eventBatch)
					eventBatch = eventBatch[:0]
					batchTimer.Reset(batchTimeout)
				}
			} else {
				fw.triggerCallbacks(event)
			}

		case <-batchTimer.C:
			if len(eventBatch) > 0 {
				fw.processBatch(eventBatch)
				eventBatch = eventBatch[:0]
			}
			batchTimer.Reset(batchTimeout)

		case <-fw.stopChan:
			// Process remaining events in batch
			if len(eventBatch) > 0 {
				fw.processBatch(eventBatch)
			}
			return
		}
	}
}

// processBatch processes a batch of events with prioritization
func (fw *FileWatcher) processBatch(events []FileEvent) {
	if fw.config.MetricsEnabled {
		atomic.AddInt64(&fw.metrics.EventsBatched, 1)
	}

	// Sort events by priority
	fw.configMutex.RLock()
	priority := fw.config.EventPriority
	fw.configMutex.RUnlock()

	// Simple priority sort (higher priority = lower number)
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if priority[events[i].EventType] > priority[events[j].EventType] {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	// Process events in priority order
	for _, event := range events {
		fw.triggerCallbacks(event)
	}
}

// watchEventsEnhanced processes file system events with error recovery
func (fw *FileWatcher) watchEventsEnhanced() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}

			if fw.config.ErrorRecovery {
				fw.errorChan <- err
			} else {
				fmt.Printf("File watcher error: %v\n", err)
			}

		case <-fw.stopChan:
			return
		}
	}
}

// watchConfigFile monitors config file for changes
func (fw *FileWatcher) watchConfigFile() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := os.ReadFile(fw.config.ConfigFile)
			if err != nil {
				continue
			}

			hash := fmt.Sprintf("%x", data)
			if hash != fw.lastConfigHash {
				fw.lastConfigHash = hash
				fw.reloadChan <- true
			}

		case <-fw.stopChan:
			return
		}
	}
}

// handleErrors processes errors with recovery
func (fw *FileWatcher) handleErrors() {
	for {
		select {
		case err := <-fw.errorChan:
			if fw.config.DebugMode {
				log.Printf("Handling error: %v", err)
			}

			// Attempt recovery
			if fw.config.ErrorRecovery {
				// Simple recovery: restart watcher
				fw.restartWatcher()
			}

		case <-fw.stopChan:
			return
		}
	}
}

// restartWatcher restarts the file watcher
func (fw *FileWatcher) restartWatcher() {
	if fw.config.DebugMode {
		log.Printf("Restarting file watcher...")
	}

	// Close current watcher
	fw.watcher.Close()

	// Create new watcher
	newWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create new watcher: %v", err)
		return
	}

	fw.watcher = newWatcher

	// Re-add all watched paths
	fw.pathsMutex.RLock()
	paths := make([]string, len(fw.watchedPaths))
	copy(paths, fw.watchedPaths)
	fw.pathsMutex.RUnlock()

	for _, path := range paths {
		err = fw.watcher.Add(path)
		if err != nil {
			log.Printf("Failed to re-add path %s: %v", path, err)
		}
	}

	if fw.config.DebugMode {
		log.Printf("File watcher restarted successfully")
	}
}
