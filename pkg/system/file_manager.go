/**
 * File Manager Package
 *
 * Provides comprehensive file and directory operations including creation,
 * reading, writing, copying, moving, searching, and deletion with safety checks.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: file_manager.go
 * Description: File and directory management operations with safety and error handling
 */

package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ena/internal/progress"
)

// FileManager handles all file and directory operations
type FileManager struct {
	SafeMode bool // Safe mode - protecting important files
}

// NewFileManager creates a new file manager instance
func NewFileManager() *FileManager {
	// File management with care and attention ✨
	return &FileManager{
		SafeMode: true, // Enable safe mode by default
	}
}

// CreateFile creates a new file with the given path and content
func (fm *FileManager) CreateFile(path string) (string, error) {
	// Create new file gently
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("Failed to create directory: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	return fmt.Sprintf("Created file \"%s\"! ✨", path), nil
}

// ReadFile reads and returns the contents of a file
func (fm *FileManager) ReadFile(path string) (string, error) {
	// Read file contents gently
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Failed to read file: %v", err)
	}

	return fmt.Sprintf("File \"%s\" contents:\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n%s",
		path, string(content)), nil
}

// WriteFile writes content to a file
func (fm *FileManager) WriteFile(path string, content string) (string, error) {
	// Write to file gently
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("Failed to create directory: %v", err)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("Failed to write to file: %v", err)
	}

	return fmt.Sprintf("Wrote to file \"%s\"! ✨", path), nil
}

// CopyFile copies a file from source to destination
func (fm *FileManager) CopyFile(src, dest string) (string, error) {
	// Copy file gently with progress bar
	err := progress.CopyFileWithProgress(src, dest)
	if err != nil {
		return "", fmt.Errorf("Failed to copy file: %v", err)
	}

	return fmt.Sprintf("Copied file \"%s\" to \"%s\"! ✨", src, dest), nil
}

// MoveFile moves a file from source to destination
func (fm *FileManager) MoveFile(src, dest string) (string, error) {
	// Move file gently to new location
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("移動先Failed to create directory: %v", err)
	}

	err := os.Rename(src, dest)
	if err != nil {
		return "", fmt.Errorf("Failed to move file: %v", err)
	}

	return fmt.Sprintf("Moved file \"%s\" to \"%s\"! ✨", src, dest), nil
}

// DeleteFile deletes a file with optional force flag
func (fm *FileManager) DeleteFile(path string, force bool) (string, error) {
	// Delete file safely with care
	if fm.SafeMode && !force {
		// Request confirmation in safe mode
		fmt.Printf("⚠️  Delete file \"%s\"? (y/N): ", path)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			return "Deletion cancelled 😅", nil
		}
	}

	err := os.Remove(path)
	if err != nil {
		return "", fmt.Errorf("Failed to delete file: %v", err)
	}

	return fmt.Sprintf("Deleted file \"%s\" 🗑️", path), nil
}

// CreateFolder creates a new directory
func (fm *FileManager) CreateFolder(path string) (string, error) {
	// Create new folder - organization is important
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create folder: %v", err)
	}

	return fmt.Sprintf("Created folder \"%s\"! ✨", path), nil
}

// ListFolder lists the contents of a directory
func (fm *FileManager) ListFolder(path string) (string, error) {
	// Look inside folder gently - showing contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("Failed to read folder: %v", err)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("Folder \"%s\" is empty 😅", path), nil
	}

	result := []string{
		fmt.Sprintf("Folder \"%s\" contents:", path),
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
		}

		size := "Folder"
		if !entry.IsDir() {
			size = formatFileSize(info.Size())
		}

		result = append(result, fmt.Sprintf("%s %s (%s) - %s",
			icon, entry.Name(), size, info.ModTime().Format("2006-01-02 15:04")))
	}

	return strings.Join(result, "\n"), nil
}

// DeleteFolder deletes a directory and all its contents
func (fm *FileManager) DeleteFolder(path string) (string, error) {
	// Delete folder with caution - all contents will be removed
	if fm.SafeMode {
		fmt.Printf("⚠️  Delete folder \"%s\" and all its contents? (y/N): ", path)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			return "Deletion cancelled 😅", nil
		}
	}

	err := os.RemoveAll(path)
	if err != nil {
		return "", fmt.Errorf("Failed to delete folder: %v", err)
	}

	return fmt.Sprintf("Deleted folder \"%s\" 🗑️", path), nil
}

// SearchFiles searches for files matching a pattern in a directory
func (fm *FileManager) SearchFiles(pattern, directory string) (string, error) {
	// Search for files - finding what you need ✨
	var matches []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue even with errors
		}

		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return nil
		}

		if matched {
			matches = append(matches, path)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("Error occurred during search: %v", err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No files matching pattern \"%s\" found 😅", pattern), nil
	}

	result := []string{
		fmt.Sprintf("Search results for pattern \"%s\" (%d files):", pattern, len(matches)),
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	}

	for _, match := range matches {
		result = append(result, fmt.Sprintf("📄 %s", match))
	}

	return strings.Join(result, "\n"), nil
}

// GetFileInfo returns detailed information about a file
func (fm *FileManager) GetFileInfo(path string) (string, error) {
	// Get detailed file information
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("Failed to get file information: %v", err)
	}

	icon := "📄"
	if info.IsDir() {
		icon = "📁"
	}

	result := []string{
		fmt.Sprintf("%s File Information: %s", icon, path),
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		fmt.Sprintf("Name: %s", info.Name()),
		fmt.Sprintf("size: %s", formatFileSize(info.Size())),
		fmt.Sprintf("Permissions: %s", info.Mode().String()),
		fmt.Sprintf("Modified: %s", info.ModTime().Format("2006-01-02 15:04:05")),
	}

	if !info.IsDir() {
		result = append(result, fmt.Sprintf("Extension: %s", filepath.Ext(path)))
	}

	return strings.Join(result, "\n"), nil
}

// GetFolderInfo returns information about a directory
func (fm *FileManager) GetFolderInfo(path string) (string, error) {
	// あたし、Folderの詳細情報を調べてあげるの
	return fm.GetFileInfo(path) // Can use the same function
}

// formatFileSize formats file size in human-readable format
func formatFileSize(size int64) string {
	// あたし、ファイルsizeを見やすくフォーマットするの
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
