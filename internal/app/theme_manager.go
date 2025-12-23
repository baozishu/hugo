package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/models"
	"hugo-visual-client/internal/repository"
)

// ThemeInfo represents information about a Hugo theme
type ThemeInfo struct {
	Name        string
	Path        string
	Description string
	Version     string
	Author      string
	License     string
	IsActive    bool
}

// ThemeManager provides a visual interface for managing Hugo themes
type ThemeManager struct {
	widget.BaseWidget
	
	// Configuration
	projectPath string
	configRepo  *repository.ConfigRepository
	configPath  string
	currentConfig *models.SiteConfig
	
	// UI components
	container     *container.Split
	themesList    *widget.List
	themeDetails  *fyne.Container
	previewArea   *container.Scroll
	
	// Data
	themes        []ThemeInfo
	selectedTheme *ThemeInfo
	
	// Callbacks
	onThemeChanged func(themeName string) error
	onRefresh      func()
}

// ThemePreview provides a preview interface for themes
type ThemePreview struct {
	widget.BaseWidget
	
	theme     *ThemeInfo
	container *fyne.Container
}

// NewThemeManager creates a new theme manager
func NewThemeManager(projectPath string, configRepo *repository.ConfigRepository, configPath string) *ThemeManager {
	manager := &ThemeManager{
		projectPath: projectPath,
		configRepo:  configRepo,
		configPath:  configPath,
		themes:      []ThemeInfo{},
	}
	
	manager.ExtendBaseWidget(manager)
	return manager
}

// LoadThemes scans for available themes in the project
func (tm *ThemeManager) LoadThemes() error {
	tm.themes = []ThemeInfo{}
	
	// Load current configuration to know active theme
	config, err := tm.configRepo.LoadSiteConfig(tm.configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	tm.currentConfig = config
	
	// Scan themes directory
	themesDir := filepath.Join(tm.projectPath, "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		// No themes directory, create empty list
		return nil
	}
	
	err = filepath.WalkDir(themesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Look for theme directories (one level deep)
		if d.IsDir() && path != themesDir {
			relPath, _ := filepath.Rel(themesDir, path)
			if !strings.Contains(relPath, string(filepath.Separator)) {
				// This is a theme directory
				themeInfo := tm.scanThemeDirectory(path, relPath)
				themeInfo.IsActive = (themeInfo.Name == config.Theme)
				tm.themes = append(tm.themes, themeInfo)
			}
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to scan themes directory: %w", err)
	}
	
	return nil
}

// scanThemeDirectory scans a theme directory for information
func (tm *ThemeManager) scanThemeDirectory(path, name string) ThemeInfo {
	info := ThemeInfo{
		Name: name,
		Path: path,
		Description: "Hugo theme",
		Version: "Unknown",
		Author: "Unknown",
		License: "Unknown",
	}
	
	// Try to read theme.toml or theme.yaml for metadata
	themeConfigPaths := []string{
		filepath.Join(path, "theme.toml"),
		filepath.Join(path, "theme.yaml"),
		filepath.Join(path, "theme.yml"),
	}
	
	for _, configPath := range themeConfigPaths {
		if _, err := os.Stat(configPath); err == nil {
			// Found theme config, try to parse basic info
			if data, err := os.ReadFile(configPath); err == nil {
				content := string(data)
				info.Description = tm.extractConfigValue(content, "description")
				info.Version = tm.extractConfigValue(content, "version")
				info.Author = tm.extractConfigValue(content, "author")
				info.License = tm.extractConfigValue(content, "license")
			}
			break
		}
	}
	
	// If no description found, try README
	readmePaths := []string{
		filepath.Join(path, "README.md"),
		filepath.Join(path, "readme.md"),
		filepath.Join(path, "README.txt"),
	}
	
	if info.Description == "Hugo theme" {
		for _, readmePath := range readmePaths {
			if data, err := os.ReadFile(readmePath); err == nil {
				lines := strings.Split(string(data), "\n")
				if len(lines) > 0 {
					// Use first non-empty line as description
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" && !strings.HasPrefix(line, "#") {
							info.Description = line
							if len(info.Description) > 100 {
								info.Description = info.Description[:100] + "..."
							}
							break
						}
					}
				}
				break
			}
		}
	}
	
	return info
}

// extractConfigValue extracts a value from theme config content
func (tm *ThemeManager) extractConfigValue(content, key string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key) {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"'`)
				return value
			}
			// Try YAML format
			parts = strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"'`)
				return value
			}
		}
	}
	return "Unknown"
}

// SetActiveTheme sets the active theme in the configuration
func (tm *ThemeManager) SetActiveTheme(themeName string) error {
	if tm.currentConfig == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Update configuration
	tm.currentConfig.Theme = themeName
	
	// Save configuration
	err := tm.configRepo.SaveSiteConfig(tm.configPath, tm.currentConfig)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	// Update theme active status
	for i := range tm.themes {
		tm.themes[i].IsActive = (tm.themes[i].Name == themeName)
	}
	
	// Call callback if set
	if tm.onThemeChanged != nil {
		return tm.onThemeChanged(themeName)
	}
	
	return nil
}

// CreateRenderer creates the visual representation of the theme manager
func (tm *ThemeManager) CreateRenderer() fyne.WidgetRenderer {
	tm.setupUI()
	return widget.NewSimpleRenderer(tm.container)
}

// setupUI creates the user interface components
func (tm *ThemeManager) setupUI() {
	// Create themes list
	tm.themesList = widget.NewList(
		func() int {
			return len(tm.themes)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Theme Name"),
				widget.NewLabel("Active"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(tm.themes) {
				return
			}
			
			themeInfo := tm.themes[id]
			if hbox, ok := obj.(*fyne.Container); ok {
				// Update icon based on active status
				icon := hbox.Objects[0].(*widget.Icon)
				if themeInfo.IsActive {
					icon.SetResource(theme.ConfirmIcon())
				} else {
					icon.SetResource(theme.DocumentIcon())
				}
				
				// Update theme name
				nameLabel := hbox.Objects[1].(*widget.Label)
				nameLabel.SetText(themeInfo.Name)
				
				// Update active status
				statusLabel := hbox.Objects[2].(*widget.Label)
				if themeInfo.IsActive {
					statusLabel.SetText("✓ Active")
				} else {
					statusLabel.SetText("")
				}
			}
		},
	)
	
	tm.themesList.OnSelected = func(id widget.ListItemID) {
		if id < len(tm.themes) {
			tm.selectedTheme = &tm.themes[id]
			tm.updateThemeDetails()
		}
	}
	
	// Create theme details panel
	tm.themeDetails = container.NewVBox(
		widget.NewLabel("Select a theme to view details"),
	)
	
	// Create preview area
	tm.previewArea = container.NewScroll(
		widget.NewLabel("Theme preview will be shown here"),
	)
	tm.previewArea.SetMinSize(fyne.NewSize(300, 200))
	
	// Create action buttons
	activateButton := widget.NewButtonWithIcon("Activate Theme", theme.ConfirmIcon(), func() {
		tm.handleActivateTheme()
	})
	activateButton.Importance = widget.HighImportance
	
	refreshButton := widget.NewButtonWithIcon("Refresh Themes", theme.ViewRefreshIcon(), func() {
		tm.handleRefreshThemes()
	})
	
	installButton := widget.NewButtonWithIcon("Install Theme", theme.DownloadIcon(), func() {
		tm.handleInstallTheme()
	})
	
	buttonContainer := container.NewHBox(
		activateButton,
		refreshButton,
		installButton,
	)
	
	// Create right panel with details and preview
	rightPanel := container.NewVBox(
		widget.NewCard("Theme Details", "", tm.themeDetails),
		widget.NewCard("Preview", "", tm.previewArea),
		buttonContainer,
	)
	
	// Create main split container
	leftPanel := container.NewVBox(
		widget.NewLabel("Available Themes"),
		tm.themesList,
	)
	
	tm.container = container.NewHSplit(leftPanel, rightPanel)
	tm.container.SetOffset(0.4) // 40% for themes list
}

// updateThemeDetails updates the theme details panel
func (tm *ThemeManager) updateThemeDetails() {
	if tm.selectedTheme == nil {
		tm.themeDetails = container.NewVBox(
			widget.NewLabel("Select a theme to view details"),
		)
		tm.themeDetails.Refresh()
		return
	}
	
	themeInfo := tm.selectedTheme
	
	// Create details content
	details := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Name:"),
			widget.NewLabel(themeInfo.Name),
		),
		container.NewHBox(
			widget.NewLabel("Status:"),
			func() *widget.Label {
				if themeInfo.IsActive {
					label := widget.NewLabel("Active")
					return label
				}
				return widget.NewLabel("Inactive")
			}(),
		),
		container.NewHBox(
			widget.NewLabel("Version:"),
			widget.NewLabel(themeInfo.Version),
		),
		container.NewHBox(
			widget.NewLabel("Author:"),
			widget.NewLabel(themeInfo.Author),
		),
		container.NewHBox(
			widget.NewLabel("License:"),
			widget.NewLabel(themeInfo.License),
		),
		widget.NewSeparator(),
		widget.NewLabel("Description:"),
		widget.NewRichTextFromMarkdown(themeInfo.Description),
	)
	
	tm.themeDetails = details
	tm.themeDetails.Refresh()
	
	// Update preview
	tm.updateThemePreview()
}

// updateThemePreview updates the theme preview area
func (tm *ThemeManager) updateThemePreview() {
	if tm.selectedTheme == nil {
		return
	}
	
	// For now, show basic theme information
	// In a real implementation, this could show screenshots or generate a preview
	previewContent := container.NewVBox(
		widget.NewLabel("Theme: " + tm.selectedTheme.Name),
		widget.NewLabel("Path: " + tm.selectedTheme.Path),
		widget.NewSeparator(),
		widget.NewLabel("Preview functionality would show:"),
		widget.NewLabel("• Theme screenshots"),
		widget.NewLabel("• Layout examples"),
		widget.NewLabel("• Color schemes"),
		widget.NewLabel("• Typography samples"),
	)
	
	tm.previewArea.Content = previewContent
	tm.previewArea.Refresh()
}

// Event handlers
func (tm *ThemeManager) handleActivateTheme() {
	if tm.selectedTheme == nil {
		dialog.ShowInformation("No Theme Selected", "Please select a theme to activate.", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	if tm.selectedTheme.IsActive {
		dialog.ShowInformation("Theme Already Active", "The selected theme is already active.", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	dialog.ShowConfirm("Activate Theme", 
		fmt.Sprintf("Are you sure you want to activate the theme '%s'?", tm.selectedTheme.Name),
		func(confirmed bool) {
			if confirmed {
				err := tm.SetActiveTheme(tm.selectedTheme.Name)
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
					return
				}
				
				// Refresh the list to show updated active status
				tm.themesList.Refresh()
				tm.updateThemeDetails()
				
				dialog.ShowInformation("Success", 
					fmt.Sprintf("Theme '%s' has been activated successfully!", tm.selectedTheme.Name), 
					fyne.CurrentApp().Driver().AllWindows()[0])
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (tm *ThemeManager) handleRefreshThemes() {
	err := tm.LoadThemes()
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	tm.themesList.Refresh()
	tm.selectedTheme = nil
	tm.updateThemeDetails()
	
	if tm.onRefresh != nil {
		tm.onRefresh()
	}
	
	dialog.ShowInformation("Refreshed", 
		fmt.Sprintf("Found %d themes in the project.", len(tm.themes)), 
		fyne.CurrentApp().Driver().AllWindows()[0])
}

func (tm *ThemeManager) handleInstallTheme() {
	// Show a dialog for theme installation options
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://github.com/user/theme-name")
	
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("theme-name")
	
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Theme URL:", Widget: urlEntry},
			{Text: "Theme Name:", Widget: nameEntry},
		},
	}
	
	dialog.ShowForm("Install Theme", "Install", "Cancel", form.Items, func(confirmed bool) {
		if confirmed && urlEntry.Text != "" && nameEntry.Text != "" {
			tm.installThemeFromURL(urlEntry.Text, nameEntry.Text)
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// installThemeFromURL installs a theme from a URL (placeholder implementation)
func (tm *ThemeManager) installThemeFromURL(url, name string) {
	// This is a placeholder implementation
	// In a real application, this would:
	// 1. Clone the git repository
	// 2. Extract the theme files
	// 3. Place them in the themes directory
	// 4. Refresh the themes list
	
	dialog.ShowInformation("Theme Installation", 
		fmt.Sprintf("Theme installation from %s is not yet implemented.\n\nTo install themes manually:\n1. Clone the theme repository to themes/%s\n2. Click 'Refresh Themes'", url, name), 
		fyne.CurrentApp().Driver().AllWindows()[0])
}

// Callback setters
func (tm *ThemeManager) SetOnThemeChanged(callback func(themeName string) error) {
	tm.onThemeChanged = callback
}

func (tm *ThemeManager) SetOnRefresh(callback func()) {
	tm.onRefresh = callback
}

// GetWidget returns the widget for embedding in other containers
func (tm *ThemeManager) GetWidget() fyne.CanvasObject {
	return tm
}

// GetActiveTheme returns the currently active theme
func (tm *ThemeManager) GetActiveTheme() *ThemeInfo {
	for _, theme := range tm.themes {
		if theme.IsActive {
			return &theme
		}
	}
	return nil
}

// GetThemes returns all available themes
func (tm *ThemeManager) GetThemes() []ThemeInfo {
	return tm.themes
}