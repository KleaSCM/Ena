/**
 * System Health Monitoring Module
 *
 * Provides comprehensive system health monitoring including CPU, memory,
 * disk usage, and system performance metrics.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: system_health.go
 * Description: System health monitoring and reporting functionality
 */

package health

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemHealth handles system health monitoring
type SystemHealth struct {
	LastCheck time.Time
}

// NewSystemHealth creates a new system health monitor
func NewSystemHealth() *SystemHealth {
	// Initialize system health monitoring with care ✨
	return &SystemHealth{
		LastCheck: time.Now(),
	}
}

// GetHealthReport returns a comprehensive system health report
func (sh *SystemHealth) GetHealthReport() (string, error) {
	// Comprehensive health check - examine all system components
	report := []string{
		"🏥 System Health Report",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"",
	}

	// CPU information
	cpuInfo, err := sh.GetCPUInfo()
	if err == nil {
		report = append(report, cpuInfo)
		report = append(report, "")
	}

	// Memory information
	memInfo, err := sh.GetMemoryInfo()
	if err == nil {
		report = append(report, memInfo)
		report = append(report, "")
	}

	// Disk information
	diskInfo, err := sh.GetDiskInfo()
	if err == nil {
		report = append(report, diskInfo)
		report = append(report, "")
	}

	// Go runtime information
	goInfo := sh.GetGoRuntimeInfo()
	report = append(report, goInfo)
	report = append(report, "")

	// Overall health status assessment
	overallHealth := sh.GetOverallHealthStatus()
	report = append(report, overallHealth)

	sh.LastCheck = time.Now()
	return strings.Join(report, "\n"), nil
}

// GetCPUInfo returns CPU usage and information
func (sh *SystemHealth) GetCPUInfo() (string, error) {
	// Check CPU status thoroughly - monitoring load levels
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return "", fmt.Errorf("Failed to get CPU information: %v", err)
	}

	cpuUsage := percentages[0]

	// Get CPU core count
	numCores := runtime.NumCPU()

	// Get detailed CPU information if available
	info, err := cpu.Info()
	if err != nil {
		return "", fmt.Errorf("Failed to get detailed CPU information: %v", err)
	}

	cpuModel := "Unknown"
	if len(info) > 0 {
		cpuModel = info[0].ModelName
	}

	status := "🟢 Normal"
	if cpuUsage > 80 {
		status = "🔴 High Load"
	} else if cpuUsage > 60 {
		status = "🟡 Warning"
	}

	return fmt.Sprintf(`💻 CPU Information:
   Model: %s
   Cores: %d
   Usage: %.1f%%
   Status: %s`, cpuModel, numCores, cpuUsage, status), nil
}

// GetMemoryInfo returns memory usage information
func (sh *SystemHealth) GetMemoryInfo() (string, error) {
	// Check memory status - monitoring for insufficient resources
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", fmt.Errorf("Failed to get memory information: %v", err)
	}

	totalGB := float64(vmStat.Total) / (1024 * 1024 * 1024)
	usedGB := float64(vmStat.Used) / (1024 * 1024 * 1024)
	freeGB := float64(vmStat.Free) / (1024 * 1024 * 1024)
	usagePercent := vmStat.UsedPercent

	status := "🟢 Normal"
	if usagePercent > 90 {
		status = "🔴 Critical"
	} else if usagePercent > 80 {
		status = "🟡 Warning"
	}

	return fmt.Sprintf(`🧠 Memory Information:
   Total: %.1f GB
   Used: %.1f GB
   Free: %.1f GB
   Usage: %.1f%%
   Status: %s`, totalGB, usedGB, freeGB, usagePercent, status), nil
}

// GetDiskInfo returns disk usage information
func (sh *SystemHealth) GetDiskInfo() (string, error) {
	// Check disk status - monitoring available space
	diskStat, err := disk.Usage("/")
	if err != nil {
		return "", fmt.Errorf("Failed to get disk information: %v", err)
	}

	totalGB := float64(diskStat.Total) / (1024 * 1024 * 1024)
	usedGB := float64(diskStat.Used) / (1024 * 1024 * 1024)
	freeGB := float64(diskStat.Free) / (1024 * 1024 * 1024)
	usagePercent := (float64(diskStat.Used) / float64(diskStat.Total)) * 100

	status := "🟢 Normal"
	if usagePercent > 95 {
		status = "🔴 Critical"
	} else if usagePercent > 85 {
		status = "🟡 Warning"
	}

	return fmt.Sprintf(`💾 Disk Information:
   Total: %.1f GB
   Used: %.1f GB
   Free: %.1f GB
   Usage: %.1f%%
   Status: %s`, totalGB, usedGB, freeGB, usagePercent, status), nil
}

// GetGoRuntimeInfo returns Go runtime information
func (sh *SystemHealth) GetGoRuntimeInfo() string {
	// Check Go runtime status - monitoring execution environment
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := float64(m.Alloc) / (1024 * 1024)
	totalAllocMB := float64(m.TotalAlloc) / (1024 * 1024)
	sysMB := float64(m.Sys) / (1024 * 1024)
	numGC := m.NumGC

	return fmt.Sprintf(`🐹 Go Runtime Information:
   Current Allocation: %.1f MB
   Total Allocation: %.1f MB
   System Memory: %.1f MB
   Garbage Collections: %d
   Goroutines: %d`, allocMB, totalAllocMB, sysMB, numGC, runtime.NumGoroutine())
}

// GetOverallHealthStatus returns an overall health assessment
func (sh *SystemHealth) GetOverallHealthStatus() string {
	// Evaluate overall health status - providing diagnostic results
	report := []string{
		"📊 Overall Health Status:",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	}

	// Simple health check
	healthy := true
	issues := []string{}

	// CPU check
	cpuPercentages, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercentages) > 0 {
		if cpuPercentages[0] > 90 {
			healthy = false
			issues = append(issues, "CPU usage is very high")
		}
	}

	// Memory check
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		if vmStat.UsedPercent > 95 {
			healthy = false
			issues = append(issues, "Going to Crash soon! Memory usage is at critical level")
		}
	}

	// Disk check
	diskStat, err := disk.Usage("/")
	if err == nil {
		usagePercent := (float64(diskStat.Used) / float64(diskStat.Total)) * 100
		if usagePercent > 95 {
			healthy = false
			issues = append(issues, "At 95% of Disk space! You should delete some files!")
		}
	}

	if healthy {
		report = append(report, "🟢 System is healthy! ✨")
		report = append(report, "   All major components are operating normally.")
	} else {
		report = append(report, "🔴 System is borked! 😅")
		report = append(report, "   The following issues were detected:")
		for _, issue := range issues {
			report = append(report, fmt.Sprintf("   • %s", issue))
		}
	}

	report = append(report, "")
	report = append(report, fmt.Sprintf("Last Check: %s", sh.LastCheck.Format("2006-01-02 15:04:05")))

	return strings.Join(report, "\n")
}
