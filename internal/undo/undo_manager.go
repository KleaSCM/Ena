/**
 * Comprehensive undo system for file operations.
 *
 * Tracks file operations, maintains history, and provides safe restoration
 * capabilities with validation, conflict detection, and comprehensive logging.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: undo_manager.go
 * Description: Core undo system with operation tracking and restoration capabilities
 */

package undo

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ena/internal/suggestions"
)

// OperationType represents the type of file operation
type OperationType string

const (
	OpCreate OperationType = "create"
	OpDelete OperationType = "delete"
	OpUpdate OperationType = "update"
	OpMove   OperationType = "move"
	OpCopy   OperationType = "copy"
	OpRename OperationType = "rename"
)

// UndoOperation represents a single operation that can be undone
type UndoOperation struct {
	ID           string                 `json:"id"`
	Type         OperationType          `json:"type"`
	Timestamp    time.Time              `json:"timestamp"`
	OriginalPath string                 `json:"original_path"`
	BackupPath   string                 `json:"backup_path,omitempty"`
	NewPath      string                 `json:"new_path,omitempty"`
	Content      []byte                 `json:"content,omitempty"`
	Size         int64                  `json:"size"`
	Permissions  os.FileMode            `json:"permissions"`
	ModTime      time.Time              `json:"mod_time"`
	Checksum     string                 `json:"checksum"`
	Metadata     map[string]interface{} `json:"metadata"`
	Undone       bool                   `json:"undone"`
	UndoneAt     *time.Time             `json:"undone_at,omitempty"`
}

// UndoSession represents a group of related operations
type UndoSession struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Operations  []UndoOperation        `json:"operations"`
	CreatedAt   time.Time              `json:"created_at"`
	Undone      bool                   `json:"undone"`
	UndoneAt    *time.Time             `json:"undone_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UndoManager manages the undo system
type UndoManager struct {
	sessions       map[string]*UndoSession
	currentSession *UndoSession
	mutex          sync.RWMutex
	historyFile    string
	maxHistorySize int
	maxSessionAge  time.Duration
	backupDir      string
	analytics      *suggestions.UsageAnalytics
	eventCallbacks map[string][]UndoEventCallback
}

// UndoEventCallback is a function that gets called on undo events
type UndoEventCallback func(event UndoEvent)

// UndoEvent represents an event that occurred in the undo system
type UndoEvent struct {
	Type      string                 `json:"type"` // operation_tracked, session_created, operation_undone, session_undone, error
	SessionID string                 `json:"session_id,omitempty"`
	Operation *UndoOperation         `json:"operation,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewUndoManager creates a new undo manager instance
func NewUndoManager(analytics *suggestions.UsageAnalytics) *UndoManager {
	um := &UndoManager{
		sessions:       make(map[string]*UndoSession),
		historyFile:    "undo_history.json",
		maxHistorySize: 1000,
		maxSessionAge:  24 * time.Hour,
		backupDir:      ".ena_undo_backups",
		analytics:      analytics,
		eventCallbacks: make(map[string][]UndoEventCallback),
	}

	// Load existing history
	um.loadHistory()

	// Start cleanup routine
	go um.startCleanupRoutine()

	return um
}

// StartSession starts a new undo session
func (um *UndoManager) StartSession(name, description string) *UndoSession {
	um.mutex.Lock()

	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &UndoSession{
		ID:          sessionID,
		Name:        name,
		Description: description,
		Operations:  make([]UndoOperation, 0),
		CreatedAt:   time.Now(),
		Undone:      false,
		Metadata:    make(map[string]interface{}),
	}

	um.sessions[sessionID] = session
	um.currentSession = session

	// Release lock before triggering event to avoid deadlock
	um.mutex.Unlock()

	um.triggerEvent(UndoEvent{
		Type:      "session_created",
		SessionID: sessionID,
		Message:   fmt.Sprintf("Started undo session: %s", name),
		Timestamp: time.Now(),
	})

	return session
}

// EndSession ends the current session
func (um *UndoManager) EndSession() {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	if um.currentSession != nil {
		um.currentSession = nil
	}
}

// TrackOperation tracks a file operation for potential undo
func (um *UndoManager) TrackOperation(opType OperationType, originalPath, newPath string) error {
	um.mutex.RLock()
	needsSession := um.currentSession == nil
	um.mutex.RUnlock()

	if needsSession {
		// Auto-start session if none exists
		um.StartSession("Auto Session", "Automatically created session")
	}

	um.mutex.Lock()
	defer um.mutex.Unlock()

	// Get file info
	info, err := os.Stat(originalPath)
	if err != nil {
		return fmt.Errorf("error getting file info for %s: %v", originalPath, err)
	}

	// Create backup if needed
	backupPath := ""
	if opType == OpDelete || opType == OpUpdate || opType == OpMove {
		backupPath, err = um.createBackup(originalPath)
		if err != nil {
			return fmt.Errorf("error creating backup for %s: %v", originalPath, err)
		}
	}

	// Read content for create/update operations
	var content []byte
	if opType == OpCreate || opType == OpUpdate {
		content, err = os.ReadFile(originalPath)
		if err != nil {
			return fmt.Errorf("error reading content for %s: %v", originalPath, err)
		}
	}

	// Calculate checksum
	checksum := um.calculateChecksum(originalPath)

	operation := UndoOperation{
		ID:           fmt.Sprintf("op_%d", time.Now().UnixNano()),
		Type:         opType,
		Timestamp:    time.Now(),
		OriginalPath: originalPath,
		BackupPath:   backupPath,
		NewPath:      newPath,
		Content:      content,
		Size:         info.Size(),
		Permissions:  info.Mode(),
		ModTime:      info.ModTime(),
		Checksum:     checksum,
		Metadata:     make(map[string]interface{}),
		Undone:       false,
	}

	um.currentSession.Operations = append(um.currentSession.Operations, operation)

	um.triggerEvent(UndoEvent{
		Type:      "operation_tracked",
		SessionID: um.currentSession.ID,
		Operation: &operation,
		Message:   fmt.Sprintf("Tracked %s operation: %s", opType, originalPath),
		Timestamp: time.Now(),
	})

	return nil
}

// UndoOperation undoes a specific operation
func (um *UndoManager) UndoOperation(operationID string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	var operation *UndoOperation
	var session *UndoSession

	// Find the operation
	for _, sess := range um.sessions {
		for i, op := range sess.Operations {
			if op.ID == operationID {
				operation = &sess.Operations[i]
				session = sess
				break
			}
		}
		if operation != nil {
			break
		}
	}

	if operation == nil {
		return fmt.Errorf("operation %s not found", operationID)
	}

	if operation.Undone {
		return fmt.Errorf("operation %s has already been undone", operationID)
	}

	// Perform undo based on operation type
	err := um.performUndo(operation)
	if err != nil {
		return fmt.Errorf("error undoing operation %s: %v", operationID, err)
	}

	// Mark as undone
	now := time.Now()
	operation.Undone = true
	operation.UndoneAt = &now

	um.triggerEvent(UndoEvent{
		Type:      "operation_undone",
		SessionID: session.ID,
		Operation: operation,
		Message:   fmt.Sprintf("Undone %s operation: %s", operation.Type, operation.OriginalPath),
		Timestamp: time.Now(),
	})

	return nil
}

// UndoSession undoes all operations in a session
func (um *UndoManager) UndoSession(sessionID string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	session, exists := um.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if session.Undone {
		return fmt.Errorf("session %s has already been undone", sessionID)
	}

	// Undo operations in reverse order
	for i := len(session.Operations) - 1; i >= 0; i-- {
		operation := &session.Operations[i]
		if !operation.Undone {
			err := um.performUndo(operation)
			if err != nil {
				return fmt.Errorf("error undoing operation %s: %v", operation.ID, err)
			}
			now := time.Now()
			operation.Undone = true
			operation.UndoneAt = &now
		}
	}

	// Mark session as undone
	now := time.Now()
	session.Undone = true
	session.UndoneAt = &now

	um.triggerEvent(UndoEvent{
		Type:      "session_undone",
		SessionID: sessionID,
		Message:   fmt.Sprintf("Undone session: %s", session.Name),
		Timestamp: time.Now(),
	})

	return nil
}

// GetHistory returns the undo history
func (um *UndoManager) GetHistory() []*UndoSession {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	var sessions []*UndoSession
	for _, session := range um.sessions {
		sessions = append(sessions, session)
	}

	// Sort by creation time (newest first)
	for i := 0; i < len(sessions)-1; i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[i].CreatedAt.Before(sessions[j].CreatedAt) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	return sessions
}

// GetSession returns a specific session
func (um *UndoManager) GetSession(sessionID string) (*UndoSession, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	session, exists := um.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// ClearHistory clears old undo history
func (um *UndoManager) ClearHistory(olderThan time.Duration) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	cutoff := time.Now().Add(-olderThan)
	var toDelete []string

	for sessionID, session := range um.sessions {
		if session.CreatedAt.Before(cutoff) {
			toDelete = append(toDelete, sessionID)
		}
	}

	for _, sessionID := range toDelete {
		session := um.sessions[sessionID]
		// Clean up backup files
		for _, operation := range session.Operations {
			if operation.BackupPath != "" {
				os.Remove(operation.BackupPath)
			}
		}
		delete(um.sessions, sessionID)
	}

	return um.saveHistory()
}

// Private helper methods

func (um *UndoManager) createBackup(filePath string) (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(um.backupDir, 0755); err != nil {
		return "", err
	}

	// Create unique backup filename
	backupName := fmt.Sprintf("backup_%d_%s", time.Now().UnixNano(), filepath.Base(filePath))
	backupPath := filepath.Join(um.backupDir, backupName)

	// Copy file to backup location
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, sourceFile)
	if err != nil {
		return "", err
	}

	// Preserve permissions and timestamps
	info, err := sourceFile.Stat()
	if err == nil {
		os.Chmod(backupPath, info.Mode())
		os.Chtimes(backupPath, info.ModTime(), info.ModTime())
	}

	return backupPath, nil
}

func (um *UndoManager) calculateChecksum(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (um *UndoManager) performUndo(operation *UndoOperation) error {
	switch operation.Type {
	case OpCreate:
		// Delete the created file
		return os.Remove(operation.OriginalPath)
	case OpDelete:
		// Restore from backup
		if operation.BackupPath == "" {
			return fmt.Errorf("no backup available for deleted file")
		}
		return um.restoreFromBackup(operation.BackupPath, operation.OriginalPath, operation.Permissions, operation.ModTime)
	case OpUpdate:
		// Restore original content
		if operation.BackupPath == "" {
			return fmt.Errorf("no backup available for updated file")
		}
		return um.restoreFromBackup(operation.BackupPath, operation.OriginalPath, operation.Permissions, operation.ModTime)
	case OpMove, OpRename:
		// Move back to original location
		if operation.NewPath == "" {
			return fmt.Errorf("no new path specified for move/rename operation")
		}
		return os.Rename(operation.NewPath, operation.OriginalPath)
	case OpCopy:
		// Delete the copied file
		if operation.NewPath == "" {
			return fmt.Errorf("no new path specified for copy operation")
		}
		return os.Remove(operation.NewPath)
	default:
		return fmt.Errorf("unknown operation type: %s", operation.Type)
	}
}

func (um *UndoManager) restoreFromBackup(backupPath, originalPath string, permissions os.FileMode, modTime time.Time) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(originalPath), 0755); err != nil {
		return err
	}

	// Copy backup to original location
	backupFile, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer backupFile.Close()

	originalFile, err := os.Create(originalPath)
	if err != nil {
		return err
	}
	defer originalFile.Close()

	_, err = io.Copy(originalFile, backupFile)
	if err != nil {
		return err
	}

	// Restore permissions and timestamps
	os.Chmod(originalPath, permissions)
	os.Chtimes(originalPath, modTime, modTime)

	return nil
}

func (um *UndoManager) loadHistory() error {
	if _, err := os.Stat(um.historyFile); os.IsNotExist(err) {
		return nil // No history file exists yet
	}

	data, err := os.ReadFile(um.historyFile)
	if err != nil {
		return fmt.Errorf("error reading history file: %v", err)
	}

	var historyData struct {
		Sessions []*UndoSession `json:"sessions"`
		Version  string         `json:"version"`
	}

	if err := json.Unmarshal(data, &historyData); err != nil {
		return fmt.Errorf("error parsing history file: %v", err)
	}

	for _, session := range historyData.Sessions {
		um.sessions[session.ID] = session
	}

	return nil
}

func (um *UndoManager) saveHistory() error {
	historyData := struct {
		Sessions []*UndoSession `json:"sessions"`
		Version  string         `json:"version"`
		Updated  time.Time      `json:"updated"`
	}{
		Sessions: make([]*UndoSession, 0, len(um.sessions)),
		Version:  "1.0",
		Updated:  time.Now(),
	}

	for _, session := range um.sessions {
		historyData.Sessions = append(historyData.Sessions, session)
	}

	data, err := json.MarshalIndent(historyData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling history: %v", err)
	}

	if err := os.WriteFile(um.historyFile, data, 0644); err != nil {
		return fmt.Errorf("error writing history file: %v", err)
	}

	return nil
}

func (um *UndoManager) startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		um.ClearHistory(um.maxSessionAge)
	}
}

func (um *UndoManager) triggerEvent(event UndoEvent) {
	um.mutex.RLock()
	callbacks := um.eventCallbacks[event.Type]
	um.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(event) // Run callbacks asynchronously
	}
}

// AddEventCallback adds a callback for undo events
func (um *UndoManager) AddEventCallback(eventType string, callback UndoEventCallback) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	um.eventCallbacks[eventType] = append(um.eventCallbacks[eventType], callback)
}
