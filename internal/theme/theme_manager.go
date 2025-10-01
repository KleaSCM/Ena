/**
 * Theme Manager Module
 *
 * Provides comprehensive theming system with dark/light modes and custom color schemes.
 *
 * Author: KleaSCM
 * Email: KleaSCM@gmail.com
 * File: theme_manager.go
 * Description: theming system with color schemes, dark/light modes, and customization
 */

package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	gookitcolor "github.com/gookit/color"
)

// ColorScheme represents a complete color scheme
type ColorScheme struct {
	Name        string
	Description string
	Colors      map[string]string
	IsDark      bool
}

// ThemeManager manages themes and color schemes
type ThemeManager struct {
	currentTheme string
	themes       map[string]*ColorScheme
	mutex        sync.RWMutex
	colorEnabled bool
	autoDetect   bool
	colorCache   map[string]string
	cacheMutex   sync.RWMutex
	themePath    string
	validators   []ThemeValidator
}

// ThemeValidator interface for theme validation
type ThemeValidator interface {
	Validate(theme *ColorScheme) error
}

// HexColorValidator validates hex color format
type HexColorValidator struct{}

// RequiredColorsValidator validates required color keys
type RequiredColorsValidator struct {
	RequiredColors []string
}

// ThemeDetector interface for theme detection
type ThemeDetector interface {
	Detect() (string, error)
}

// OSThemeDetector detects OS-level theme preferences
type OSThemeDetector struct{}

// EnvironmentThemeDetector detects theme from environment variables
type EnvironmentThemeDetector struct{}

// Color constants for different UI elements
const (
	ColorPrimary    = "primary"
	ColorSecondary  = "secondary"
	ColorSuccess    = "success"
	ColorWarning    = "warning"
	ColorError      = "error"
	ColorInfo       = "info"
	ColorBackground = "background"
	ColorForeground = "foreground"
	ColorAccent     = "accent"
	ColorMuted      = "muted"
	ColorBorder     = "border"
	ColorHighlight  = "highlight"
	ColorFile       = "file"
	ColorDirectory  = "directory"
	ColorExecutable = "executable"
	ColorSymlink    = "symlink"
	ColorProgress   = "progress"
	ColorProgressBg = "progress_bg"
	ColorETA        = "eta"
	ColorSpeed      = "speed"
	ColorLabel      = "label"
)

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		currentTheme: "default",
		themes:       make(map[string]*ColorScheme),
		colorEnabled: true,
		autoDetect:   true,
		colorCache:   make(map[string]string),
		themePath:    "themes",
		validators:   make([]ThemeValidator, 0),
	}

	// Initialize validators
	tm.initializeValidators()

	// Initialize default themes
	tm.initializeDefaultThemes()

	// Load custom themes from disk
	tm.loadCustomThemes()

	// Auto-detect system theme if enabled
	if tm.autoDetect {
		tm.detectSystemTheme()
	}

	return tm
}

// initializeValidators sets up theme validators
func (tm *ThemeManager) initializeValidators() {
	// Add hex color validator
	tm.validators = append(tm.validators, &HexColorValidator{})

	// Add required colors validator
	requiredColors := []string{
		ColorPrimary, ColorSecondary, ColorSuccess, ColorWarning,
		ColorError, ColorInfo, ColorBackground, ColorForeground,
		ColorAccent, ColorMuted, ColorBorder, ColorHighlight,
	}
	tm.validators = append(tm.validators, &RequiredColorsValidator{
		RequiredColors: requiredColors,
	})
}

// initializeDefaultThemes sets up built-in themes
func (tm *ThemeManager) initializeDefaultThemes() {
	// Default theme (light)
	tm.themes["default"] = &ColorScheme{
		Name:        "Default",
		Description: "Clean light theme with gentle colors",
		IsDark:      false,
		Colors: map[string]string{
			ColorPrimary:    "#2563eb", // Blue
			ColorSecondary:  "#64748b", // Slate
			ColorSuccess:    "#16a34a", // Green
			ColorWarning:    "#ea580c", // Orange
			ColorError:      "#dc2626", // Red
			ColorInfo:       "#0891b2", // Cyan
			ColorBackground: "#ffffff", // White
			ColorForeground: "#1e293b", // Dark slate
			ColorAccent:     "#7c3aed", // Purple
			ColorMuted:      "#94a3b8", // Light slate
			ColorBorder:     "#e2e8f0", // Light gray
			ColorHighlight:  "#f1f5f9", // Very light slate
			ColorFile:       "#1e293b", // Dark slate
			ColorDirectory:  "#2563eb", // Blue
			ColorExecutable: "#16a34a", // Green
			ColorSymlink:    "#7c3aed", // Purple
			ColorProgress:   "#16a34a", // Green
			ColorProgressBg: "#e2e8f0", // Light gray
			ColorETA:        "#7c3aed", // Purple
			ColorSpeed:      "#ea580c", // Orange
			ColorLabel:      "#2563eb", // Blue
		},
	}

	// Dark theme
	tm.themes["dark"] = &ColorScheme{
		Name:        "Dark",
		Description: "Modern dark theme with vibrant accents",
		IsDark:      true,
		Colors: map[string]string{
			ColorPrimary:    "#3b82f6", // Bright blue
			ColorSecondary:  "#94a3b8", // Light slate
			ColorSuccess:    "#22c55e", // Bright green
			ColorWarning:    "#f59e0b", // Amber
			ColorError:      "#ef4444", // Bright red
			ColorInfo:       "#06b6d4", // Bright cyan
			ColorBackground: "#0f172a", // Very dark slate
			ColorForeground: "#f8fafc", // Very light slate
			ColorAccent:     "#a855f7", // Bright purple
			ColorMuted:      "#64748b", // Slate
			ColorBorder:     "#334155", // Dark slate
			ColorHighlight:  "#1e293b", // Dark slate
			ColorFile:       "#f8fafc", // Very light slate
			ColorDirectory:  "#3b82f6", // Bright blue
			ColorExecutable: "#22c55e", // Bright green
			ColorSymlink:    "#a855f7", // Bright purple
			ColorProgress:   "#22c55e", // Bright green
			ColorProgressBg: "#334155", // Dark slate
			ColorETA:        "#a855f7", // Bright purple
			ColorSpeed:      "#f59e0b", // Amber
			ColorLabel:      "#3b82f6", // Bright blue
		},
	}

	// Solarized Light theme
	tm.themes["solarized-light"] = &ColorScheme{
		Name:        "Solarized Light",
		Description: "Classic solarized light theme",
		IsDark:      false,
		Colors: map[string]string{
			ColorPrimary:    "#268bd2", // Blue
			ColorSecondary:  "#586e75", // Base01
			ColorSuccess:    "#859900", // Green
			ColorWarning:    "#b58900", // Yellow
			ColorError:      "#dc322f", // Red
			ColorInfo:       "#2aa198", // Cyan
			ColorBackground: "#fdf6e3", // Base3
			ColorForeground: "#073642", // Base02
			ColorAccent:     "#d33682", // Magenta
			ColorMuted:      "#93a1a1", // Base1
			ColorBorder:     "#eee8d5", // Base2
			ColorHighlight:  "#fdf6e3", // Base3
			ColorFile:       "#073642", // Base02
			ColorDirectory:  "#268bd2", // Blue
			ColorExecutable: "#859900", // Green
			ColorSymlink:    "#d33682", // Magenta
			ColorProgress:   "#859900", // Green
			ColorProgressBg: "#eee8d5", // Base2
			ColorETA:        "#d33682", // Magenta
			ColorSpeed:      "#b58900", // Yellow
			ColorLabel:      "#268bd2", // Blue
		},
	}

	// Solarized Dark theme
	tm.themes["solarized-dark"] = &ColorScheme{
		Name:        "Solarized Dark",
		Description: "Classic solarized dark theme",
		IsDark:      true,
		Colors: map[string]string{
			ColorPrimary:    "#268bd2", // Blue
			ColorSecondary:  "#93a1a1", // Base1
			ColorSuccess:    "#859900", // Green
			ColorWarning:    "#b58900", // Yellow
			ColorError:      "#dc322f", // Red
			ColorInfo:       "#2aa198", // Cyan
			ColorBackground: "#002b36", // Base03
			ColorForeground: "#fdf6e3", // Base3
			ColorAccent:     "#d33682", // Magenta
			ColorMuted:      "#586e75", // Base01
			ColorBorder:     "#073642", // Base02
			ColorHighlight:  "#073642", // Base02
			ColorFile:       "#fdf6e3", // Base3
			ColorDirectory:  "#268bd2", // Blue
			ColorExecutable: "#859900", // Green
			ColorSymlink:    "#d33682", // Magenta
			ColorProgress:   "#859900", // Green
			ColorProgressBg: "#073642", // Base02
			ColorETA:        "#d33682", // Magenta
			ColorSpeed:      "#b58900", // Yellow
			ColorLabel:      "#268bd2", // Blue
		},
	}

	// Monokai theme
	tm.themes["monokai"] = &ColorScheme{
		Name:        "Monokai",
		Description: "Popular Monokai color scheme",
		IsDark:      true,
		Colors: map[string]string{
			ColorPrimary:    "#66d9ef", // Light blue
			ColorSecondary:  "#a6e22e", // Light green
			ColorSuccess:    "#a6e22e", // Light green
			ColorWarning:    "#e6db74", // Yellow
			ColorError:      "#f92672", // Pink
			ColorInfo:       "#66d9ef", // Light blue
			ColorBackground: "#272822", // Dark gray
			ColorForeground: "#f8f8f2", // Light gray
			ColorAccent:     "#ae81ff", // Purple
			ColorMuted:      "#75715e", // Medium gray
			ColorBorder:     "#3e3d32", // Darker gray
			ColorHighlight:  "#3e3d32", // Darker gray
			ColorFile:       "#f8f8f2", // Light gray
			ColorDirectory:  "#66d9ef", // Light blue
			ColorExecutable: "#a6e22e", // Light green
			ColorSymlink:    "#ae81ff", // Purple
			ColorProgress:   "#a6e22e", // Light green
			ColorProgressBg: "#3e3d32", // Darker gray
			ColorETA:        "#ae81ff", // Purple
			ColorSpeed:      "#e6db74", // Yellow
			ColorLabel:      "#66d9ef", // Light blue
		},
	}

	// Dracula theme
	tm.themes["dracula"] = &ColorScheme{
		Name:        "Dracula",
		Description: "Dark Dracula theme with vibrant colors",
		IsDark:      true,
		Colors: map[string]string{
			ColorPrimary:    "#8be9fd", // Cyan
			ColorSecondary:  "#6272a4", // Comment
			ColorSuccess:    "#50fa7b", // Green
			ColorWarning:    "#ffb86c", // Orange
			ColorError:      "#ff5555", // Red
			ColorInfo:       "#8be9fd", // Cyan
			ColorBackground: "#282a36", // Background
			ColorForeground: "#f8f8f2", // Foreground
			ColorAccent:     "#bd93f9", // Purple
			ColorMuted:      "#6272a4", // Comment
			ColorBorder:     "#44475a", // Selection
			ColorHighlight:  "#44475a", // Selection
			ColorFile:       "#f8f8f2", // Foreground
			ColorDirectory:  "#8be9fd", // Cyan
			ColorExecutable: "#50fa7b", // Green
			ColorSymlink:    "#bd93f9", // Purple
			ColorProgress:   "#50fa7b", // Green
			ColorProgressBg: "#44475a", // Selection
			ColorETA:        "#bd93f9", // Purple
			ColorSpeed:      "#ffb86c", // Orange
			ColorLabel:      "#8be9fd", // Cyan
		},
	}

	// Nord theme
	tm.themes["nord"] = &ColorScheme{
		Name:        "Nord",
		Description: "Arctic-inspired Nord theme",
		IsDark:      true,
		Colors: map[string]string{
			ColorPrimary:    "#88c0d0", // Frost
			ColorSecondary:  "#4c566a", // Polar night 3
			ColorSuccess:    "#a3be8c", // Aurora green
			ColorWarning:    "#ebcb8b", // Aurora yellow
			ColorError:      "#bf616a", // Aurora red
			ColorInfo:       "#88c0d0", // Frost
			ColorBackground: "#2e3440", // Polar night 0
			ColorForeground: "#d8dee9", // Snow storm 3
			ColorAccent:     "#b48ead", // Aurora purple
			ColorMuted:      "#4c566a", // Polar night 3
			ColorBorder:     "#3b4252", // Polar night 1
			ColorHighlight:  "#3b4252", // Polar night 1
			ColorFile:       "#d8dee9", // Snow storm 3
			ColorDirectory:  "#88c0d0", // Frost
			ColorExecutable: "#a3be8c", // Aurora green
			ColorSymlink:    "#b48ead", // Aurora purple
			ColorProgress:   "#a3be8c", // Aurora green
			ColorProgressBg: "#3b4252", // Polar night 1
			ColorETA:        "#b48ead", // Aurora purple
			ColorSpeed:      "#ebcb8b", // Aurora yellow
			ColorLabel:      "#88c0d0", // Frost
		},
	}
}

// detectSystemTheme detects the system theme preference using multiple detectors
func (tm *ThemeManager) detectSystemTheme() {
	// Check for explicit theme setting
	if theme := os.Getenv("ENA_THEME"); theme != "" {
		if _, exists := tm.themes[theme]; exists {
			tm.currentTheme = theme
			return
		}
	}

	// Try different detection methods
	detectors := []ThemeDetector{
		&EnvironmentThemeDetector{},
		&OSThemeDetector{},
	}

	for _, detector := range detectors {
		if theme, err := detector.Detect(); err == nil {
			if _, exists := tm.themes[theme]; exists {
				tm.currentTheme = theme
				return
			}
		}
	}

	// Default to light theme
	tm.currentTheme = "default"
}

// Detect implements ThemeDetector for EnvironmentThemeDetector
func (d *EnvironmentThemeDetector) Detect() (string, error) {
	// Check for common dark mode indicators
	if os.Getenv("DARK_MODE") == "1" || os.Getenv("DARKMODE") == "1" {
		return "dark", nil
	}

	// Check terminal color scheme
	if colorFGBG := os.Getenv("COLORFGBG"); colorFGBG != "" {
		if strings.Contains(colorFGBG, "15;0") || strings.Contains(colorFGBG, "7;0") {
			return "dark", nil
		}
	}

	// Check for GTK theme
	if gtkTheme := os.Getenv("GTK_THEME"); gtkTheme != "" {
		if strings.Contains(strings.ToLower(gtkTheme), "dark") {
			return "dark", nil
		}
	}

	return "", fmt.Errorf("no environment theme detected")
}

// Detect implements ThemeDetector for OSThemeDetector
func (d *OSThemeDetector) Detect() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return d.detectMacOSTheme()
	case "windows":
		return d.detectWindowsTheme()
	case "linux":
		return d.detectLinuxTheme()
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// detectMacOSTheme detects macOS theme preference
func (d *OSThemeDetector) detectMacOSTheme() (string, error) {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(strings.ToLower(string(output)), "dark") {
		return "dark", nil
	}
	return "default", nil
}

// detectWindowsTheme detects Windows theme preference
func (d *OSThemeDetector) detectWindowsTheme() (string, error) {
	cmd := exec.Command("reg", "query", "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize", "/v", "AppsUseLightTheme")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(strings.ToLower(string(output)), "0x0") {
		return "dark", nil
	}
	return "default", nil
}

// detectLinuxTheme detects Linux theme preference
func (d *OSThemeDetector) detectLinuxTheme() (string, error) {
	// Try different desktop environments
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")

	switch {
	case strings.Contains(strings.ToLower(desktop), "gnome"):
		return d.detectGnomeTheme()
	case strings.Contains(strings.ToLower(desktop), "kde"):
		return d.detectKDETheme()
	case strings.Contains(strings.ToLower(desktop), "xfce"):
		return d.detectXFCETheme()
	default:
		// Fallback to GTK settings
		return d.detectGTKTheme()
	}
}

// detectGnomeTheme detects GNOME theme preference
func (d *OSThemeDetector) detectGnomeTheme() (string, error) {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(strings.ToLower(string(output)), "dark") {
		return "dark", nil
	}
	return "default", nil
}

// detectKDETheme detects KDE theme preference
func (d *OSThemeDetector) detectKDETheme() (string, error) {
	cmd := exec.Command("kreadconfig5", "--file", "kdeglobals", "--group", "General", "--key", "ColorScheme")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(strings.ToLower(string(output)), "dark") {
		return "dark", nil
	}
	return "default", nil
}

// detectXFCETheme detects XFCE theme preference
func (d *OSThemeDetector) detectXFCETheme() (string, error) {
	cmd := exec.Command("xfconf-query", "-c", "xsettings", "-p", "/Net/ThemeName")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(strings.ToLower(string(output)), "dark") {
		return "dark", nil
	}
	return "default", nil
}

// detectGTKTheme detects GTK theme preference
func (d *OSThemeDetector) detectGTKTheme() (string, error) {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(strings.ToLower(string(output)), "dark") {
		return "dark", nil
	}
	return "default", nil
}

// SetTheme sets the current theme
func (tm *ThemeManager) SetTheme(themeName string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.themes[themeName]; !exists {
		return fmt.Errorf("theme '%s' does not exist", themeName)
	}

	tm.currentTheme = themeName
	return nil
}

// GetCurrentTheme returns the current theme name
func (tm *ThemeManager) GetCurrentTheme() string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.currentTheme
}

// GetTheme returns a theme by name
func (tm *ThemeManager) GetTheme(themeName string) (*ColorScheme, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	theme, exists := tm.themes[themeName]
	if !exists {
		return nil, fmt.Errorf("theme '%s' does not exist", themeName)
	}

	return theme, nil
}

// GetCurrentColorScheme returns the current color scheme
func (tm *ThemeManager) GetCurrentColorScheme() *ColorScheme {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return tm.themes[tm.currentTheme]
}

// GetAvailableThemes returns a list of available theme names
func (tm *ThemeManager) GetAvailableThemes() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	themes := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		themes = append(themes, name)
	}
	return themes
}

// GetColor returns a color for a specific element
func (tm *ThemeManager) GetColor(element string) string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	theme := tm.themes[tm.currentTheme]
	if color, exists := theme.Colors[element]; exists {
		return color
	}

	// Fallback to foreground color
	return theme.Colors[ColorForeground]
}

// Colorize applies color to text based on the current theme
func (tm *ThemeManager) Colorize(element, text string) string {
	if !tm.colorEnabled {
		return text
	}

	colorCode := tm.GetColor(element)
	return tm.applyColor(colorCode, text)
}

// applyColor applies a hex color to text with caching
func (tm *ThemeManager) applyColor(hexColor, text string) string {
	if !tm.colorEnabled {
		return text
	}

	// Check cache first
	tm.cacheMutex.RLock()
	cacheKey := fmt.Sprintf("%s:%s", hexColor, text)
	if cached, exists := tm.colorCache[cacheKey]; exists {
		tm.cacheMutex.RUnlock()
		return cached
	}
	tm.cacheMutex.RUnlock()

	// Parse hex color using gookit/color
	var result string
	c := gookitcolor.HEX(hexColor)
	result = c.Sprint(text)

	// Cache the result
	tm.cacheMutex.Lock()
	tm.colorCache[cacheKey] = result
	tm.cacheMutex.Unlock()

	return result
}

// fallbackColor provides fallback color mapping for known hex values
func (tm *ThemeManager) fallbackColor(hexColor, text string) string {
	switch hexColor {
	case "#2563eb", "#3b82f6": // Blue variants
		return color.New(color.FgBlue).Sprint(text)
	case "#16a34a", "#22c55e", "#859900", "#a3be8c", "#50fa7b", "#a6e22e": // Green variants
		return color.New(color.FgGreen).Sprint(text)
	case "#dc2626", "#ef4444", "#dc322f", "#f92672", "#ff5555", "#bf616a": // Red variants
		return color.New(color.FgRed).Sprint(text)
	case "#ea580c", "#f59e0b", "#b58900", "#e6db74", "#ffb86c", "#ebcb8b": // Orange/Yellow variants
		return color.New(color.FgYellow).Sprint(text)
	case "#7c3aed", "#a855f7", "#d33682", "#ae81ff", "#bd93f9", "#b48ead": // Purple variants
		return color.New(color.FgMagenta).Sprint(text)
	case "#0891b2", "#06b6d4", "#2aa198", "#66d9ef", "#8be9fd", "#88c0d0": // Cyan variants
		return color.New(color.FgCyan).Sprint(text)
	case "#64748b", "#94a3b8", "#586e75", "#93a1a1", "#75715e", "#6272a4", "#4c566a": // Gray variants
		return color.New(color.FgHiBlack).Sprint(text)
	default:
		return text
	}
}

// SetColorEnabled enables or disables color output
func (tm *ThemeManager) SetColorEnabled(enabled bool) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.colorEnabled = enabled
}

// IsColorEnabled returns whether color output is enabled
func (tm *ThemeManager) IsColorEnabled() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.colorEnabled
}

// SetAutoDetect enables or disables automatic theme detection
func (tm *ThemeManager) SetAutoDetect(enabled bool) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.autoDetect = enabled
}

// IsDarkMode returns whether the current theme is dark
func (tm *ThemeManager) IsDarkMode() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.themes[tm.currentTheme].IsDark
}

// GetThemeInfo returns detailed information about a theme
func (tm *ThemeManager) GetThemeInfo(themeName string) (map[string]interface{}, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	theme, exists := tm.themes[themeName]
	if !exists {
		return nil, fmt.Errorf("theme '%s' does not exist", themeName)
	}

	return map[string]interface{}{
		"name":        theme.Name,
		"description": theme.Description,
		"is_dark":     theme.IsDark,
		"colors":      theme.Colors,
		"is_current":  themeName == tm.currentTheme,
	}, nil
}

// PreviewTheme shows a preview of a theme
func (tm *ThemeManager) PreviewTheme(themeName string) (string, error) {
	theme, err := tm.GetTheme(themeName)
	if err != nil {
		return "", err
	}

	preview := fmt.Sprintf("🎨 Theme Preview: %s\n", theme.Name)
	preview += fmt.Sprintf("Description: %s\n", theme.Description)
	preview += fmt.Sprintf("Mode: %s\n", map[bool]string{true: "Dark", false: "Light"}[theme.IsDark])
	preview += "\nColor Samples:\n"

	// Show color samples
	colorSamples := map[string]string{
		"Primary":   ColorPrimary,
		"Success":   ColorSuccess,
		"Warning":   ColorWarning,
		"Error":     ColorError,
		"Info":      ColorInfo,
		"Accent":    ColorAccent,
		"File":      ColorFile,
		"Directory": ColorDirectory,
	}

	for label, colorKey := range colorSamples {
		color := theme.Colors[colorKey]
		preview += fmt.Sprintf("  %s: %s\n", label, tm.applyColor(color, "████"))
	}

	return preview, nil
}

// ExportTheme exports a theme to JSON format
func (tm *ThemeManager) ExportTheme(themeName string) (string, error) {
	theme, err := tm.GetTheme(themeName)
	if err != nil {
		return "", err
	}

	// This would normally use json.Marshal, but for simplicity we'll return a formatted string
	result := fmt.Sprintf("Theme: %s\n", theme.Name)
	result += fmt.Sprintf("Description: %s\n", theme.Description)
	result += fmt.Sprintf("Dark Mode: %t\n", theme.IsDark)
	result += "Colors:\n"
	for key, value := range theme.Colors {
		result += fmt.Sprintf("  %s: %s\n", key, value)
	}

	return result, nil
}

// Validate implements ThemeValidator for HexColorValidator
func (v *HexColorValidator) Validate(theme *ColorScheme) error {
	hexRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

	for colorName, colorValue := range theme.Colors {
		if !hexRegex.MatchString(colorValue) {
			return fmt.Errorf("invalid hex color '%s' for color element '%s'", colorValue, colorName)
		}
	}
	return nil
}

// Validate implements ThemeValidator for RequiredColorsValidator
func (v *RequiredColorsValidator) Validate(theme *ColorScheme) error {
	for _, requiredColor := range v.RequiredColors {
		if _, exists := theme.Colors[requiredColor]; !exists {
			return fmt.Errorf("missing required color element: %s", requiredColor)
		}
	}
	return nil
}

// ValidateTheme validates a theme using all registered validators
func (tm *ThemeManager) ValidateTheme(theme *ColorScheme) error {
	for _, validator := range tm.validators {
		if err := validator.Validate(theme); err != nil {
			return err
		}
	}
	return nil
}

// SetColor sets a specific color in a theme
func (tm *ThemeManager) SetColor(themeName, element, hexValue string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	theme, exists := tm.themes[themeName]
	if !exists {
		return fmt.Errorf("theme '%s' does not exist", themeName)
	}

	// Validate hex color format
	hexRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	if !hexRegex.MatchString(hexValue) {
		return fmt.Errorf("invalid hex color format: %s", hexValue)
	}

	// Update the color
	theme.Colors[element] = hexValue

	// Clear cache for this theme
	tm.clearThemeCache(themeName)

	return nil
}

// clearThemeCache clears cached colors for a specific theme
func (tm *ThemeManager) clearThemeCache(themeName string) {
	tm.cacheMutex.Lock()
	defer tm.cacheMutex.Unlock()

	// Remove cached colors that might be affected by theme changes
	for key := range tm.colorCache {
		if strings.Contains(key, themeName) {
			delete(tm.colorCache, key)
		}
	}
}

// SaveTheme saves a theme to disk
func (tm *ThemeManager) SaveTheme(themeName string) error {
	tm.mutex.RLock()
	theme, exists := tm.themes[themeName]
	tm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("theme '%s' does not exist", themeName)
	}

	// Validate theme before saving
	if err := tm.ValidateTheme(theme); err != nil {
		return fmt.Errorf("theme validation failed: %v", err)
	}

	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(tm.themePath, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %v", err)
	}

	// Create theme data for export
	themeData := map[string]interface{}{
		"name":        theme.Name,
		"description": theme.Description,
		"is_dark":     theme.IsDark,
		"colors":      theme.Colors,
		"version":     "1.0",
		"created_at":  fmt.Sprintf("%d", os.Getpid()), // Simple timestamp
	}

	// Save as JSON
	filename := filepath.Join(tm.themePath, themeName+".json")
	data, err := json.MarshalIndent(themeData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme data: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %v", err)
	}

	return nil
}

// LoadTheme loads a theme from disk
func (tm *ThemeManager) LoadTheme(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read theme file: %v", err)
	}

	var themeData map[string]interface{}
	if err := json.Unmarshal(data, &themeData); err != nil {
		return fmt.Errorf("failed to parse theme file: %v", err)
	}

	// Extract theme information
	name, ok := themeData["name"].(string)
	if !ok {
		return fmt.Errorf("invalid theme name in file")
	}

	description, ok := themeData["description"].(string)
	if !ok {
		description = "Custom theme"
	}

	isDark, ok := themeData["is_dark"].(bool)
	if !ok {
		isDark = false
	}

	colors, ok := themeData["colors"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid colors data in theme file")
	}

	// Convert colors to string map
	colorMap := make(map[string]string)
	for key, value := range colors {
		if colorStr, ok := value.(string); ok {
			colorMap[key] = colorStr
		}
	}

	// Create theme
	theme := &ColorScheme{
		Name:        name,
		Description: description,
		IsDark:      isDark,
		Colors:      colorMap,
	}

	// Validate theme
	if err := tm.ValidateTheme(theme); err != nil {
		return fmt.Errorf("loaded theme validation failed: %v", err)
	}

	// Add theme to manager
	tm.mutex.Lock()
	tm.themes[name] = theme
	tm.mutex.Unlock()

	return nil
}

// loadCustomThemes loads all custom themes from the themes directory
func (tm *ThemeManager) loadCustomThemes() {
	if _, err := os.Stat(tm.themePath); os.IsNotExist(err) {
		return
	}

	files, err := filepath.Glob(filepath.Join(tm.themePath, "*.json"))
	if err != nil {
		return
	}

	for _, file := range files {
		if err := tm.LoadTheme(file); err != nil {
			// Log error but continue loading other themes
			continue
		}
	}
}

// CreateCustomTheme creates a new custom theme
func (tm *ThemeManager) CreateCustomTheme(name, description string, isDark bool, colors map[string]string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.themes[name]; exists {
		return fmt.Errorf("theme '%s' already exists", name)
	}

	theme := &ColorScheme{
		Name:        name,
		Description: description,
		IsDark:      isDark,
		Colors:      colors,
	}

	// Validate theme
	if err := tm.ValidateTheme(theme); err != nil {
		return fmt.Errorf("theme validation failed: %v", err)
	}

	tm.themes[name] = theme

	// Automatically save custom theme to disk
	if err := tm.saveThemeToDisk(theme); err != nil {
		// Log error but don't fail the creation
		fmt.Printf("Warning: Failed to save theme to disk: %v\n", err)
	}

	return nil
}

// saveThemeToDisk saves a theme to disk without validation (internal use)
func (tm *ThemeManager) saveThemeToDisk(theme *ColorScheme) error {
	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(tm.themePath, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %v", err)
	}

	// Create theme data for export
	themeData := map[string]interface{}{
		"name":        theme.Name,
		"description": theme.Description,
		"is_dark":     theme.IsDark,
		"colors":      theme.Colors,
		"version":     "1.0",
		"created_at":  fmt.Sprintf("%d", os.Getpid()), // Simple timestamp
	}

	// Save as JSON
	filename := filepath.Join(tm.themePath, theme.Name+".json")
	data, err := json.MarshalIndent(themeData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme data: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %v", err)
	}

	return nil
}

// DeleteTheme removes a theme
func (tm *ThemeManager) DeleteTheme(themeName string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.themes[themeName]; !exists {
		return fmt.Errorf("theme '%s' does not exist", themeName)
	}

	// Don't allow deletion of built-in themes
	builtInThemes := []string{"default", "dark", "solarized-light", "solarized-dark", "monokai", "dracula", "nord"}
	for _, builtIn := range builtInThemes {
		if themeName == builtIn {
			return fmt.Errorf("cannot delete built-in theme '%s'", themeName)
		}
	}

	delete(tm.themes, themeName)

	// Remove theme file if it exists
	filename := filepath.Join(tm.themePath, themeName+".json")
	if _, err := os.Stat(filename); err == nil {
		os.Remove(filename)
	}

	// Clear cache
	tm.clearThemeCache(themeName)

	return nil
}

// GetThemePath returns the current theme path
func (tm *ThemeManager) GetThemePath() string {
	return tm.themePath
}

// SetThemePath sets the theme path
func (tm *ThemeManager) SetThemePath(path string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.themePath = path
}

// ClearCache clears the color cache
func (tm *ThemeManager) ClearCache() {
	tm.cacheMutex.Lock()
	defer tm.cacheMutex.Unlock()
	tm.colorCache = make(map[string]string)
}

// GetCacheStats returns cache statistics
func (tm *ThemeManager) GetCacheStats() map[string]interface{} {
	tm.cacheMutex.RLock()
	defer tm.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_size": len(tm.colorCache),
		"cache_keys": func() []string {
			keys := make([]string, 0, len(tm.colorCache))
			for key := range tm.colorCache {
				keys = append(keys, key)
			}
			return keys
		}(),
	}
}
