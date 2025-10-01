/**
 * CLI commands for smart application detection and management.
 *
 * Provides commands for scanning, managing, and analyzing installed
 * applications with comprehensive metadata and categorization.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: app_detection_commands.go
 * Description: Cobra command definitions for application detection operations
 */

package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"ena/internal/appdetect"

	"github.com/spf13/cobra"
)

// Global app scanner instance
var globalAppScanner *appdetect.AppScanner

// getGlobalAppScanner returns the global app scanner instance
func getGlobalAppScanner() *appdetect.AppScanner {
	if globalAppScanner == nil {
		analytics := getGlobalAnalytics()
		globalAppScanner = appdetect.NewAppScanner(analytics)
	}
	return globalAppScanner
}

// setupAppDetectionCommands adds application detection commands to the root command
func setupAppDetectionCommands(rootCmd *cobra.Command) {
	// Get app scanner
	scanner := getGlobalAppScanner()

	// Scan apps command
	scanAppsCmd := &cobra.Command{
		Use:   "scan-apps",
		Short: "Scan for installed applications",
		Long: `Scan the system for installed applications and update the application database.
Performs comprehensive detection across different platforms and application types.

Examples:
  ena scan-apps
  ena scan-apps --deep
  ena scan-apps --update`,
		Run: func(cmd *cobra.Command, args []string) {
			deepScan, _ := cmd.Flags().GetBool("deep")
			update, _ := cmd.Flags().GetBool("update")

			fmt.Println("🌸 Scanning for installed applications...")
			if deepScan {
				fmt.Println("🔍 Deep scan enabled - this may take longer")
			}

			// Perform scan
			result, err := scanner.ScanForApps(deepScan)
			if err != nil {
				fmt.Printf("❌ Error scanning applications: %v\n", err)
				return
			}

			// Display results
			fmt.Printf("✅ Scan completed in %s!\n", result.ScanDuration.String())
			fmt.Printf("📊 Found %d applications\n", result.AppsFound)
			fmt.Printf("🔄 Updated %d applications\n", result.AppsUpdated)
			fmt.Printf("🗑️  Removed %d applications\n", result.AppsRemoved)
			fmt.Printf("🏷️  Categories found: %d\n", len(result.CategoriesFound))

			if len(result.Errors) > 0 {
				fmt.Printf("⚠️  Errors encountered: %d\n", len(result.Errors))
				if update {
					for _, err := range result.Errors {
						fmt.Printf("   - %s\n", err)
					}
				}
			}

			if update && len(result.Apps) > 0 {
				fmt.Println("\n🌸 Recently detected applications:")
				// Show first 10 apps
				limit := 10
				if len(result.Apps) < limit {
					limit = len(result.Apps)
				}

				for i := 0; i < limit; i++ {
					app := result.Apps[i]
					fmt.Printf("%d. %s (%s)\n", i+1, app.DisplayName, app.Category)
					if app.Version != "" {
						fmt.Printf("   📦 Version: %s\n", app.Version)
					}
					if app.Status == appdetect.StatusRunning {
						fmt.Printf("   🟢 Status: Running\n")
					}
				}

				if len(result.Apps) > 10 {
					fmt.Printf("... and %d more applications\n", len(result.Apps)-10)
				}
			}
		},
	}

	scanAppsCmd.Flags().Bool("deep", false, "Perform deep scan (slower but more comprehensive)")
	scanAppsCmd.Flags().Bool("update", false, "Show detailed update information")

	// List apps command
	listAppsCmd := &cobra.Command{
		Use:   "list-apps",
		Short: "List detected applications",
		Long: `List all detected applications with filtering options.
Supports filtering by category, status, and search terms.

Examples:
  ena list-apps
  ena list-apps --category development
  ena list-apps --running
  ena list-apps --search "code"`,
		Run: func(cmd *cobra.Command, args []string) {
			category, _ := cmd.Flags().GetString("category")
			status, _ := cmd.Flags().GetString("status")
			search, _ := cmd.Flags().GetString("search")
			running, _ := cmd.Flags().GetBool("running")
			limit, _ := cmd.Flags().GetInt("limit")

			// Build filter
			filter := make(map[string]interface{})
			if category != "" {
				filter["category"] = category
			}
			if status != "" {
				filter["status"] = status
			}
			if running {
				filter["status"] = appdetect.StatusRunning
			}

			apps := scanner.GetApps(filter)

			// Apply search filter if specified
			if search != "" {
				var filteredApps []appdetect.AppInfo
				for _, app := range apps {
					if strings.Contains(strings.ToLower(app.Name), strings.ToLower(search)) ||
						strings.Contains(strings.ToLower(app.DisplayName), strings.ToLower(search)) ||
						strings.Contains(strings.ToLower(app.Description), strings.ToLower(search)) {
						filteredApps = append(filteredApps, app)
					}
				}
				apps = filteredApps
			}

			if len(apps) == 0 {
				fmt.Println("🌸 No applications found matching the criteria")
				return
			}

			// Limit results if specified
			if limit > 0 && limit < len(apps) {
				apps = apps[:limit]
			}

			fmt.Printf("🌸 Found %d applications (╹◡╹)♡\n", len(apps))
			fmt.Println("=====================================")

			for i, app := range apps {
				fmt.Printf("%d. %s\n", i+1, app.DisplayName)
				fmt.Printf("   🏷️  Category: %s\n", app.Category)
				fmt.Printf("   📦 Version: %s\n", app.Version)
				fmt.Printf("   📂 Path: %s\n", app.ExecutablePath)

				statusIcon := "⚪"
				switch app.Status {
				case appdetect.StatusRunning:
					statusIcon = "🟢"
				case appdetect.StatusInstalled:
					statusIcon = "📦"
				case appdetect.StatusOutdated:
					statusIcon = "🟡"
				case appdetect.StatusCorrupted:
					statusIcon = "🔴"
				}
				fmt.Printf("   %s Status: %s\n", statusIcon, app.Status)

				if app.Description != "" {
					fmt.Printf("   📝 Description: %s\n", app.Description)
				}
				if app.IsDefaultApp {
					fmt.Printf("   ⭐ Default application\n")
				}
				fmt.Println()
			}
		},
	}

	listAppsCmd.Flags().String("category", "", "Filter by category (development, productivity, media, etc.)")
	listAppsCmd.Flags().String("status", "", "Filter by status (installed, running, outdated, etc.)")
	listAppsCmd.Flags().String("search", "", "Search applications by name or description")
	listAppsCmd.Flags().Bool("running", false, "Show only running applications")
	listAppsCmd.Flags().Int("limit", 0, "Limit number of results")

	// App info command
	appInfoCmd := &cobra.Command{
		Use:   "app-info <app-id>",
		Short: "Show detailed information about an application",
		Long: `Show comprehensive information about a specific application
including metadata, file associations, and system integration details.

Examples:
  ena app-info app_firefox_1234567890
  ena app-info app_vscode_1234567890`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appID := args[0]

			app, err := scanner.GetAppByID(appID)
			if err != nil {
				fmt.Printf("❌ Error finding application: %v\n", err)
				return
			}

			fmt.Printf("🌸 Application Information (╹◡╹)♡\n")
			fmt.Printf("===============================\n")
			fmt.Printf("🆔 ID: %s\n", app.ID)
			fmt.Printf("📱 Name: %s\n", app.DisplayName)
			fmt.Printf("🏷️  Category: %s\n", app.Category)
			fmt.Printf("📦 Version: %s\n", app.Version)

			statusIcon := "⚪"
			switch app.Status {
			case appdetect.StatusRunning:
				statusIcon = "🟢"
			case appdetect.StatusInstalled:
				statusIcon = "📦"
			case appdetect.StatusOutdated:
				statusIcon = "🟡"
			case appdetect.StatusCorrupted:
				statusIcon = "🔴"
			}
			fmt.Printf("%s Status: %s\n", statusIcon, app.Status)

			fmt.Printf("📂 Executable: %s\n", app.ExecutablePath)
			fmt.Printf("📁 Install Path: %s\n", app.InstallPath)

			if app.IconPath != "" {
				fmt.Printf("🎨 Icon: %s\n", app.IconPath)
			}
			if app.Description != "" {
				fmt.Printf("📝 Description: %s\n", app.Description)
			}
			if app.Author != "" {
				fmt.Printf("👤 Author: %s\n", app.Author)
			}
			if app.Website != "" {
				fmt.Printf("🌐 Website: %s\n", app.Website)
			}
			if app.License != "" {
				fmt.Printf("📄 License: %s\n", app.License)
			}

			fmt.Printf("📊 Size: %s\n", formatBytesAppDetection(app.Size))
			fmt.Printf("📅 Detected: %s\n", app.DetectedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("🔄 Updated: %s\n", app.UpdatedAt.Format("2006-01-02 15:04:05"))

			if app.InstallDate != nil {
				fmt.Printf("📦 Installed: %s\n", app.InstallDate.Format("2006-01-02 15:04:05"))
			}
			if app.LastUsed != nil {
				fmt.Printf("🕒 Last Used: %s\n", app.LastUsed.Format("2006-01-02 15:04:05"))
			}

			if app.IsDefaultApp {
				fmt.Printf("⭐ Default Application\n")
			}

			if len(app.FileAssociations) > 0 {
				fmt.Printf("📎 File Associations: %s\n", strings.Join(app.FileAssociations, ", "))
			}

			if len(app.CommandLineArgs) > 0 {
				fmt.Printf("⚙️  Command Line Args: %s\n", strings.Join(app.CommandLineArgs, ", "))
			}
		},
	}

	// App stats command
	appStatsCmd := &cobra.Command{
		Use:   "app-stats",
		Short: "Show application detection statistics",
		Long: `Show comprehensive statistics about detected applications
including category distribution, platform information, and system health.

Examples:
  ena app-stats`,
		Run: func(cmd *cobra.Command, args []string) {
			stats := scanner.GetAppStats()

			fmt.Println("🌸 Application Detection Statistics (╹◡╹)♡")
			fmt.Println("=========================================")

			fmt.Printf("📊 Total Applications: %v\n", stats["total_apps"])
			fmt.Printf("🟢 Running Applications: %v\n", stats["running_apps"])
			fmt.Printf("⭐ Default Applications: %v\n", stats["default_apps"])
			fmt.Printf("💾 Total Size: %s\n", formatBytesAppDetection(stats["total_size"].(int64)))
			fmt.Printf("🖥️  Platform: %s\n", stats["platform"])

			if lastScan, ok := stats["last_scan"].(time.Time); ok && !lastScan.IsZero() {
				fmt.Printf("🔄 Last Scan: %s\n", lastScan.Format("2006-01-02 15:04:05"))
			}

			// Category distribution
			if categories, ok := stats["categories"].(map[string]int); ok && len(categories) > 0 {
				fmt.Println("\n📈 Category Distribution:")
				// Sort categories by count
				type categoryCount struct {
					name  string
					count int
				}
				var sortedCategories []categoryCount
				for name, count := range categories {
					sortedCategories = append(sortedCategories, categoryCount{name, count})
				}
				sort.Slice(sortedCategories, func(i, j int) bool {
					return sortedCategories[i].count > sortedCategories[j].count
				})

				for _, cat := range sortedCategories {
					fmt.Printf("   %s: %d\n", cat.name, cat.count)
				}
			}

			// Status distribution
			if statusCounts, ok := stats["status_counts"].(map[string]int); ok && len(statusCounts) > 0 {
				fmt.Println("\n📊 Status Distribution:")
				for status, count := range statusCounts {
					fmt.Printf("   %s: %d\n", status, count)
				}
			}

			// Time information
			if oldestInstall, ok := stats["oldest_install"].(time.Time); ok && !oldestInstall.IsZero() {
				fmt.Printf("\n⏰ Oldest Installation: %s\n", oldestInstall.Format("2006-01-02"))
			}
			if newestInstall, ok := stats["newest_install"].(time.Time); ok && !newestInstall.IsZero() {
				fmt.Printf("⏰ Newest Installation: %s\n", newestInstall.Format("2006-01-02"))
			}
		},
	}

	// Running apps command
	runningAppsCmd := &cobra.Command{
		Use:   "running-apps",
		Short: "Show currently running applications",
		Long: `Show all applications that are currently running
with their process information and resource usage.

Examples:
  ena running-apps`,
		Run: func(cmd *cobra.Command, args []string) {
			apps := scanner.GetRunningApps()

			if len(apps) == 0 {
				fmt.Println("🌸 No running applications detected")
				return
			}

			fmt.Printf("🌸 Running Applications (╹◡╹)♡\n")
			fmt.Printf("===============================\n")

			for i, app := range apps {
				fmt.Printf("%d. %s\n", i+1, app.DisplayName)
				fmt.Printf("   🏷️  Category: %s\n", app.Category)
				fmt.Printf("   📦 Version: %s\n", app.Version)
				fmt.Printf("   📂 Path: %s\n", app.ExecutablePath)
				fmt.Printf("   🕒 Last Updated: %s\n", app.UpdatedAt.Format("2006-01-02 15:04:05"))
				fmt.Println()
			}
		},
	}

	// Default apps command
	defaultAppsCmd := &cobra.Command{
		Use:   "default-apps",
		Short: "Show default applications for file types",
		Long: `Show applications that are set as defaults for various file types
and system operations.

Examples:
  ena default-apps`,
		Run: func(cmd *cobra.Command, args []string) {
			apps := scanner.GetDefaultApps()

			if len(apps) == 0 {
				fmt.Println("🌸 No default applications configured")
				return
			}

			fmt.Printf("🌸 Default Applications (╹◡╹)♡\n")
			fmt.Printf("===============================\n")

			for i, app := range apps {
				fmt.Printf("%d. %s\n", i+1, app.DisplayName)
				fmt.Printf("   🏷️  Category: %s\n", app.Category)
				fmt.Printf("   📂 Path: %s\n", app.ExecutablePath)

				if len(app.FileAssociations) > 0 {
					fmt.Printf("   📎 File Types: %s\n", strings.Join(app.FileAssociations, ", "))
				}
				fmt.Println()
			}
		},
	}

	// Add all commands to root
	rootCmd.AddCommand(scanAppsCmd)
	rootCmd.AddCommand(listAppsCmd)
	rootCmd.AddCommand(appInfoCmd)
	rootCmd.AddCommand(appStatsCmd)
	rootCmd.AddCommand(runningAppsCmd)
	rootCmd.AddCommand(defaultAppsCmd)
}

// Helper function to format bytes (app detection specific)
func formatBytesAppDetection(bytes int64) string {
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
