/**
 * CLI commands for automated backup management.
 *
 * Provides commands for creating, managing, and restoring backups
 * with comprehensive backup lifecycle management.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: backup_commands.go
 * Description: Cobra command definitions for backup operations
 */

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ena/internal/backup"

	"github.com/spf13/cobra"
)

// Global backup engine instance
var globalBackupEngine *backup.BackupEngine

// getGlobalBackupEngine returns the global backup engine instance
func getGlobalBackupEngine() *backup.BackupEngine {
	if globalBackupEngine == nil {
		analytics := getGlobalAnalytics()
		globalBackupEngine = backup.NewBackupEngine(analytics)
	}
	return globalBackupEngine
}

// setupBackupCommands adds backup management commands to the root command
func setupBackupCommands(rootCmd *cobra.Command) {
	// Get backup engine
	engine := getGlobalBackupEngine()

	// Create backup command
	createBackupCmd := &cobra.Command{
		Use:   "create-backup <path>",
		Short: "Create a backup of a file or directory",
		Long: `Create a backup of the specified file or directory.
The backup will be stored with metadata and checksums for integrity verification.

Examples:
  ena create-backup ~/Documents/important.txt
  ena create-backup ~/Projects/my-app
  ena create-backup /etc/config --description "System config backup"`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sourcePath := args[0]
			description, _ := cmd.Flags().GetString("description")
			tags, _ := cmd.Flags().GetStringSlice("tags")
			operationID := fmt.Sprintf("manual_%d", time.Now().UnixNano())

			// Validate source path
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				fmt.Printf("❌ Source path does not exist: %s\n", sourcePath)
				return
			}

			if description == "" {
				description = fmt.Sprintf("Manual backup of %s", filepath.Base(sourcePath))
			}

			fmt.Printf("🌸 Creating backup of: %s\n", sourcePath)
			if len(tags) > 0 {
				fmt.Printf("🏷️  Tags: %s\n", strings.Join(tags, ", "))
			}

			// Create backup
			metadata, err := engine.CreateBackup(sourcePath, operationID, description, tags)
			if err != nil {
				fmt.Printf("❌ Error creating backup: %v\n", err)
				return
			}

			fmt.Printf("✅ Backup created successfully!\n")
			fmt.Printf("🆔 Backup ID: %s\n", filepath.Base(metadata.BackupPath))
			fmt.Printf("📂 Backup Path: %s\n", metadata.BackupPath)
			fmt.Printf("📊 Size: %s\n", formatBytesBackup(metadata.Size))
			fmt.Printf("🔍 Checksum: %s\n", metadata.Checksum)
			fmt.Printf("📅 Created: %s\n", metadata.CreatedAt.Format("2006-01-02 15:04:05"))
			if metadata.ExpiresAt != nil {
				fmt.Printf("⏰ Expires: %s\n", metadata.ExpiresAt.Format("2006-01-02 15:04:05"))
			}
		},
	}

	createBackupCmd.Flags().String("description", "", "Description for the backup")
	createBackupCmd.Flags().StringSlice("tags", []string{}, "Tags for the backup")

	// List backups command
	listBackupsCmd := &cobra.Command{
		Use:   "list-backups",
		Short: "List all backups",
		Long: `List all backups with their details.
Supports filtering by operation, type, status, and tags.

Examples:
  ena list-backups
  ena list-backups --operation-id manual_1234567890
  ena list-backups --type file --status verified`,
		Run: func(cmd *cobra.Command, args []string) {
			operationID, _ := cmd.Flags().GetString("operation-id")
			backupType, _ := cmd.Flags().GetString("type")
			status, _ := cmd.Flags().GetString("status")
			tag, _ := cmd.Flags().GetString("tag")
			limit, _ := cmd.Flags().GetInt("limit")

			// Build filter
			filter := make(map[string]interface{})
			if operationID != "" {
				filter["operation_id"] = operationID
			}
			if backupType != "" {
				filter["type"] = backupType
			}
			if status != "" {
				filter["status"] = status
			}
			if tag != "" {
				filter["tag"] = tag
			}

			backups := engine.ListBackups(filter)
			if len(backups) == 0 {
				fmt.Println("🌸 No backups found matching the criteria")
				return
			}

			// Limit results if specified
			if limit > 0 && limit < len(backups) {
				backups = backups[:limit]
			}

			fmt.Printf("🌸 Found %d backups (╹◡╹)♡\n", len(backups))
			fmt.Println("=====================================")

			for i, backup := range backups {
				fmt.Printf("%d. %s\n", i+1, filepath.Base(backup.BackupPath))
				fmt.Printf("   📂 Original: %s\n", backup.OriginalPath)
				fmt.Printf("   💾 Backup: %s\n", backup.BackupPath)
				fmt.Printf("   📊 Size: %s\n", formatBytesBackup(backup.Size))
				fmt.Printf("   🏷️  Type: %s | Status: %s\n", backup.Type, backup.Status)
				fmt.Printf("   📅 Created: %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
				if backup.ExpiresAt != nil {
					fmt.Printf("   ⏰ Expires: %s\n", backup.ExpiresAt.Format("2006-01-02 15:04:05"))
				}
				if len(backup.Tags) > 0 {
					fmt.Printf("   🏷️  Tags: %s\n", strings.Join(backup.Tags, ", "))
				}
				if backup.Description != "" {
					fmt.Printf("   📝 Description: %s\n", backup.Description)
				}
				fmt.Printf("   🔍 Checksum: %s\n", backup.Checksum)
				fmt.Println()
			}
		},
	}

	listBackupsCmd.Flags().String("operation-id", "", "Filter by operation ID")
	listBackupsCmd.Flags().String("type", "", "Filter by backup type (file, directory)")
	listBackupsCmd.Flags().String("status", "", "Filter by status (created, verified, corrupted)")
	listBackupsCmd.Flags().String("tag", "", "Filter by tag")
	listBackupsCmd.Flags().Int("limit", 0, "Limit number of results")

	// Restore backup command
	restoreBackupCmd := &cobra.Command{
		Use:   "restore-backup <backup-id> [destination]",
		Short: "Restore a backup to its original location or a new location",
		Long: `Restore a backup to its original location or a specified destination.
The backup will be verified before restoration.

Examples:
  ena restore-backup backup_1234567890
  ena restore-backup backup_1234567890 ~/restored-file.txt
  ena restore-backup backup_1234567890 --overwrite`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			backupID := args[0]
			var destinationPath string
			if len(args) > 1 {
				destinationPath = args[1]
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")

			fmt.Printf("🌸 Restoring backup: %s\n", backupID)
			if destinationPath != "" {
				fmt.Printf("📂 Destination: %s\n", destinationPath)
			}
			if overwrite {
				fmt.Println("⚠️  Overwrite mode enabled")
			}

			// Restore backup
			err := engine.RestoreBackup(backupID, destinationPath, overwrite)
			if err != nil {
				fmt.Printf("❌ Error restoring backup: %v\n", err)
				return
			}

			fmt.Printf("✅ Backup restored successfully!\n")
		},
	}

	restoreBackupCmd.Flags().Bool("overwrite", false, "Overwrite existing files")

	// Delete backup command
	deleteBackupCmd := &cobra.Command{
		Use:   "delete-backup <backup-id>",
		Short: "Delete a backup and its associated files",
		Long: `Delete a backup and remove all associated files.
This action cannot be undone.

Examples:
  ena delete-backup backup_1234567890`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			backupID := args[0]
			confirm, _ := cmd.Flags().GetBool("confirm")

			if !confirm {
				fmt.Printf("⚠️  This will permanently delete backup %s\n", backupID)
				fmt.Print("Type 'yes' to confirm: ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "yes" {
					fmt.Println("❌ Operation cancelled")
					return
				}
			}

			fmt.Printf("🌸 Deleting backup: %s\n", backupID)

			// Delete backup
			err := engine.DeleteBackup(backupID)
			if err != nil {
				fmt.Printf("❌ Error deleting backup: %v\n", err)
				return
			}

			fmt.Printf("✅ Backup deleted successfully!\n")
		},
	}

	deleteBackupCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	// Backup stats command
	backupStatsCmd := &cobra.Command{
		Use:   "backup-stats",
		Short: "Show backup statistics and system information",
		Long: `Show comprehensive backup statistics including total backups,
storage usage, status distribution, and system health.

Examples:
  ena backup-stats`,
		Run: func(cmd *cobra.Command, args []string) {
			stats := engine.GetBackupStats()

			fmt.Println("🌸 Backup System Statistics (╹◡╹)♡")
			fmt.Println("=====================================")

			fmt.Printf("📊 Total Backups: %v\n", stats["total_backups"])
			fmt.Printf("⚙️  Total Operations: %v\n", stats["total_operations"])
			fmt.Printf("💾 Total Size: %s\n", formatBytesBackup(stats["total_size"].(int64)))

			// Status distribution
			if statusCounts, ok := stats["status_counts"].(map[string]int); ok {
				fmt.Println("\n📈 Status Distribution:")
				for status, count := range statusCounts {
					fmt.Printf("   %s: %d\n", status, count)
				}
			}

			// Type distribution
			if typeCounts, ok := stats["type_counts"].(map[string]int); ok {
				fmt.Println("\n📁 Type Distribution:")
				for backupType, count := range typeCounts {
					fmt.Printf("   %s: %d\n", backupType, count)
				}
			}

			// Time information
			if oldestBackup, ok := stats["oldest_backup"].(time.Time); ok && !oldestBackup.IsZero() {
				fmt.Printf("\n⏰ Oldest Backup: %s\n", oldestBackup.Format("2006-01-02 15:04:05"))
			}
			if newestBackup, ok := stats["newest_backup"].(time.Time); ok && !newestBackup.IsZero() {
				fmt.Printf("⏰ Newest Backup: %s\n", newestBackup.Format("2006-01-02 15:04:05"))
			}

			// Configuration
			config := engine.GetConfig()
			fmt.Println("\n⚙️  Configuration:")
			fmt.Printf("   Enabled: %v\n", config.Enabled)
			fmt.Printf("   Max Backups: %d\n", config.MaxBackups)
			fmt.Printf("   Retention Days: %d\n", config.RetentionDays)
			fmt.Printf("   Compression: %v\n", config.Compression)
			fmt.Printf("   Encryption: %v\n", config.Encryption)
			fmt.Printf("   Backup Directory: %s\n", config.BackupDirectory)
			fmt.Printf("   Auto Cleanup: %v\n", config.AutoCleanup)
		},
	}

	// Cleanup command
	cleanupCmd := &cobra.Command{
		Use:   "backup-cleanup",
		Short: "Clean up expired backups",
		Long: `Remove expired backups based on the retention policy.
This helps free up disk space and maintain system performance.

Examples:
  ena backup-cleanup`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🌸 Cleaning up expired backups...")

			cleanedCount, err := engine.CleanupExpiredBackups()
			if err != nil {
				fmt.Printf("❌ Error during cleanup: %v\n", err)
				return
			}

			if cleanedCount == 0 {
				fmt.Println("✅ No expired backups found - system is clean!")
			} else {
				fmt.Printf("✅ Cleaned up %d expired backups\n", cleanedCount)
			}
		},
	}

	// Add all commands to root
	rootCmd.AddCommand(createBackupCmd)
	rootCmd.AddCommand(listBackupsCmd)
	rootCmd.AddCommand(restoreBackupCmd)
	rootCmd.AddCommand(deleteBackupCmd)
	rootCmd.AddCommand(backupStatsCmd)
	rootCmd.AddCommand(cleanupCmd)
}

// Helper function to format bytes (backup-specific)
func formatBytesBackup(bytes int64) string {
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
