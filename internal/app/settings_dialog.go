package app

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/models"
)

// SettingsDialog represents the application settings dialog
type SettingsDialog struct {
	window     fyne.Window
	dialog     *dialog.CustomDialog
	config     *models.AppConfig
	onSave     func(*models.AppConfig) error
	
	// Form widgets
	windowWidthEntry  *widget.Entry
	windowHeightEntry *widget.Entry
	themeSelect       *widget.Select
	autoSaveCheck     *widget.Check
	previewPortEntry  *widget.Entry
	
	// UI customization
	editorFontSizeEntry *widget.Entry
	previewRefreshCheck *widget.Check
	showLineNumbersCheck *widget.Check
	
	// Hugo settings
	hugoPathEntry     *widget.Entry
	defaultThemeEntry *widget.Entry
}

// NewSettingsDialog creates a new settings dialog
func NewSettingsDialog(window fyne.Window, config *models.AppConfig, onSave func(*models.AppConfig) error) *SettingsDialog {
	sd := &SettingsDialog{
		window: window,
		config: config,
		onSave: onSave,
	}
	
	sd.createDialog()
	return sd
}

// createDialog creates the settings dialog UI
func (sd *SettingsDialog) createDialog() {
	// Create form widgets
	sd.createFormWidgets()
	
	// Load current values
	sd.loadCurrentValues()
	
	// Create tabs for different settings categories
	tabs := container.NewAppTabs()
	
	// General settings tab
	generalTab := sd.createGeneralTab()
	tabs.Append(container.NewTabItem("General", generalTab))
	
	// UI settings tab
	uiTab := sd.createUITab()
	tabs.Append(container.NewTabItem("Interface", uiTab))
	
	// Hugo settings tab
	hugoTab := sd.createHugoTab()
	tabs.Append(container.NewTabItem("Hugo", hugoTab))
	
	// Create buttons
	saveButton := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), sd.onSaveClicked)
	saveButton.Importance = widget.HighImportance
	
	cancelButton := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), sd.onCancelClicked)
	
	resetButton := widget.NewButtonWithIcon("Reset to Defaults", theme.ViewRefreshIcon(), sd.onResetClicked)
	resetButton.Importance = widget.MediumImportance
	
	buttons := container.NewHBox(
		resetButton,
		widget.NewSeparator(),
		cancelButton,
		saveButton,
	)
	
	// Create main content
	content := container.NewBorder(
		nil,     // top
		buttons, // bottom
		nil,     // left
		nil,     // right
		tabs,    // center
	)
	
	// Create dialog
	sd.dialog = dialog.NewCustom("Application Settings", "", content, sd.window)
	sd.dialog.Resize(fyne.NewSize(600, 500))
}

// createFormWidgets creates all form input widgets
func (sd *SettingsDialog) createFormWidgets() {
	// Window settings
	sd.windowWidthEntry = widget.NewEntry()
	sd.windowWidthEntry.SetPlaceHolder("1200")
	sd.windowWidthEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < 800 || val > 3840 {
			return fmt.Errorf("must be between 800 and 3840")
		}
		return nil
	}
	
	sd.windowHeightEntry = widget.NewEntry()
	sd.windowHeightEntry.SetPlaceHolder("800")
	sd.windowHeightEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < 600 || val > 2160 {
			return fmt.Errorf("must be between 600 and 2160")
		}
		return nil
	}
	
	// Theme selection
	sd.themeSelect = widget.NewSelect([]string{"default", "dark", "light"}, nil)
	
	// Auto-save setting
	sd.autoSaveCheck = widget.NewCheck("Automatically save changes", nil)
	
	// Preview port
	sd.previewPortEntry = widget.NewEntry()
	sd.previewPortEntry.SetPlaceHolder("1313")
	sd.previewPortEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < 1024 || val > 65535 {
			return fmt.Errorf("must be between 1024 and 65535")
		}
		return nil
	}
	
	// UI customization
	sd.editorFontSizeEntry = widget.NewEntry()
	sd.editorFontSizeEntry.SetPlaceHolder("12")
	sd.editorFontSizeEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < 8 || val > 24 {
			return fmt.Errorf("must be between 8 and 24")
		}
		return nil
	}
	
	sd.previewRefreshCheck = widget.NewCheck("Auto-refresh preview on changes", nil)
	sd.showLineNumbersCheck = widget.NewCheck("Show line numbers in editor", nil)
	
	// Hugo settings
	sd.hugoPathEntry = widget.NewEntry()
	sd.hugoPathEntry.SetPlaceHolder("hugo (auto-detect)")
	
	sd.defaultThemeEntry = widget.NewEntry()
	sd.defaultThemeEntry.SetPlaceHolder("Default theme for new projects")
}

// createGeneralTab creates the general settings tab
func (sd *SettingsDialog) createGeneralTab() *fyne.Container {
	return container.NewVBox(
		widget.NewCard("Window Settings", "",
			container.NewVBox(
				container.NewGridWithColumns(2,
					widget.NewLabel("Default Width:"),
					sd.windowWidthEntry,
					widget.NewLabel("Default Height:"),
					sd.windowHeightEntry,
				),
			),
		),
		
		widget.NewCard("Application Settings", "",
			container.NewVBox(
				container.NewGridWithColumns(2,
					widget.NewLabel("Theme:"),
					sd.themeSelect,
				),
				sd.autoSaveCheck,
			),
		),
		
		widget.NewCard("Preview Settings", "",
			container.NewVBox(
				container.NewGridWithColumns(2,
					widget.NewLabel("Preview Port:"),
					sd.previewPortEntry,
				),
			),
		),
	)
}

// createUITab creates the UI customization tab
func (sd *SettingsDialog) createUITab() *fyne.Container {
	return container.NewVBox(
		widget.NewCard("Editor Settings", "",
			container.NewVBox(
				container.NewGridWithColumns(2,
					widget.NewLabel("Font Size:"),
					sd.editorFontSizeEntry,
				),
				sd.showLineNumbersCheck,
			),
		),
		
		widget.NewCard("Preview Settings", "",
			container.NewVBox(
				sd.previewRefreshCheck,
			),
		),
		
		widget.NewLabel("Note: Some UI changes may require restarting the application."),
	)
}

// createHugoTab creates the Hugo settings tab
func (sd *SettingsDialog) createHugoTab() *fyne.Container {
	return container.NewVBox(
		widget.NewCard("Hugo Configuration", "",
			container.NewVBox(
				container.NewGridWithColumns(2,
					widget.NewLabel("Hugo Path:"),
					sd.hugoPathEntry,
					widget.NewLabel("Default Theme:"),
					sd.defaultThemeEntry,
				),
			),
		),
		
		widget.NewLabel("Hugo Path: Leave empty to auto-detect Hugo installation"),
		widget.NewLabel("Default Theme: Theme to use when creating new projects"),
	)
}

// loadCurrentValues loads current configuration values into the form
func (sd *SettingsDialog) loadCurrentValues() {
	if sd.config == nil {
		return
	}
	
	// Window settings
	sd.windowWidthEntry.SetText(strconv.Itoa(sd.config.WindowWidth))
	sd.windowHeightEntry.SetText(strconv.Itoa(sd.config.WindowHeight))
	
	// Theme
	sd.themeSelect.SetSelected(sd.config.Theme)
	
	// Auto-save
	sd.autoSaveCheck.SetChecked(sd.config.AutoSave)
	
	// Preview port
	sd.previewPortEntry.SetText(strconv.Itoa(sd.config.PreviewPort))
	
	// UI settings (using extended config if available)
	if extConfig, ok := sd.config.Custom["editor_font_size"]; ok {
		if fontSize, ok := extConfig.(float64); ok {
			sd.editorFontSizeEntry.SetText(strconv.Itoa(int(fontSize)))
		}
	} else {
		sd.editorFontSizeEntry.SetText("12") // default
	}
	
	if extConfig, ok := sd.config.Custom["preview_auto_refresh"]; ok {
		if autoRefresh, ok := extConfig.(bool); ok {
			sd.previewRefreshCheck.SetChecked(autoRefresh)
		}
	} else {
		sd.previewRefreshCheck.SetChecked(true) // default
	}
	
	if extConfig, ok := sd.config.Custom["show_line_numbers"]; ok {
		if showLines, ok := extConfig.(bool); ok {
			sd.showLineNumbersCheck.SetChecked(showLines)
		}
	} else {
		sd.showLineNumbersCheck.SetChecked(true) // default
	}
	
	// Hugo settings
	if extConfig, ok := sd.config.Custom["hugo_path"]; ok {
		if hugoPath, ok := extConfig.(string); ok {
			sd.hugoPathEntry.SetText(hugoPath)
		}
	}
	
	if extConfig, ok := sd.config.Custom["default_theme"]; ok {
		if defaultTheme, ok := extConfig.(string); ok {
			sd.defaultThemeEntry.SetText(defaultTheme)
		}
	}
}

// validateForm validates all form inputs
func (sd *SettingsDialog) validateForm() error {
	// Validate window width
	if err := sd.windowWidthEntry.Validate(); err != nil {
		return fmt.Errorf("window width: %w", err)
	}
	
	// Validate window height
	if err := sd.windowHeightEntry.Validate(); err != nil {
		return fmt.Errorf("window height: %w", err)
	}
	
	// Validate preview port
	if err := sd.previewPortEntry.Validate(); err != nil {
		return fmt.Errorf("preview port: %w", err)
	}
	
	// Validate editor font size
	if err := sd.editorFontSizeEntry.Validate(); err != nil {
		return fmt.Errorf("editor font size: %w", err)
	}
	
	return nil
}

// saveSettings saves the current form values to the configuration
func (sd *SettingsDialog) saveSettings() error {
	// Validate form first
	if err := sd.validateForm(); err != nil {
		return err
	}
	
	// Create new config based on current values
	newConfig := *sd.config // Copy current config
	
	// Initialize Custom map if nil
	if newConfig.Custom == nil {
		newConfig.Custom = make(map[string]interface{})
	}
	
	// Window settings
	if width, err := strconv.Atoi(sd.windowWidthEntry.Text); err == nil {
		newConfig.WindowWidth = width
	}
	if height, err := strconv.Atoi(sd.windowHeightEntry.Text); err == nil {
		newConfig.WindowHeight = height
	}
	
	// Theme
	newConfig.Theme = sd.themeSelect.Selected
	
	// Auto-save
	newConfig.AutoSave = sd.autoSaveCheck.Checked
	
	// Preview port
	if port, err := strconv.Atoi(sd.previewPortEntry.Text); err == nil {
		newConfig.PreviewPort = port
	}
	
	// UI settings
	if fontSize, err := strconv.Atoi(sd.editorFontSizeEntry.Text); err == nil {
		newConfig.Custom["editor_font_size"] = fontSize
	}
	newConfig.Custom["preview_auto_refresh"] = sd.previewRefreshCheck.Checked
	newConfig.Custom["show_line_numbers"] = sd.showLineNumbersCheck.Checked
	
	// Hugo settings
	newConfig.Custom["hugo_path"] = sd.hugoPathEntry.Text
	newConfig.Custom["default_theme"] = sd.defaultThemeEntry.Text
	
	// Validate the new configuration
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Save using the callback
	if sd.onSave != nil {
		return sd.onSave(&newConfig)
	}
	
	return nil
}

// Event handlers

func (sd *SettingsDialog) onSaveClicked() {
	err := sd.saveSettings()
	if err != nil {
		// Show error dialog
		dialog.ShowError(err, sd.window)
		return
	}
	
	// Show success message
	dialog.ShowInformation("Settings Saved", "Application settings have been saved successfully.", sd.window)
	
	// Close dialog
	sd.dialog.Hide()
}

func (sd *SettingsDialog) onCancelClicked() {
	sd.dialog.Hide()
}

func (sd *SettingsDialog) onResetClicked() {
	// Show confirmation dialog
	confirm := dialog.NewConfirm(
		"Reset Settings",
		"Are you sure you want to reset all settings to their default values?",
		func(confirmed bool) {
			if confirmed {
				sd.resetToDefaults()
			}
		},
		sd.window,
	)
	confirm.Show()
}

// resetToDefaults resets all form values to their defaults
func (sd *SettingsDialog) resetToDefaults() {
	// Create default config
	defaultConfig := &models.AppConfig{
		RecentProjects: []string{},
		WindowWidth:    1200,
		WindowHeight:   800,
		Theme:          "default",
		AutoSave:       true,
		PreviewPort:    1313,
		Custom:         make(map[string]interface{}),
	}
	
	// Set default custom values
	defaultConfig.Custom["editor_font_size"] = 12
	defaultConfig.Custom["preview_auto_refresh"] = true
	defaultConfig.Custom["show_line_numbers"] = true
	defaultConfig.Custom["hugo_path"] = ""
	defaultConfig.Custom["default_theme"] = ""
	
	// Update the config reference
	sd.config = defaultConfig
	
	// Reload form values
	sd.loadCurrentValues()
}

// Show displays the settings dialog
func (sd *SettingsDialog) Show() {
	sd.dialog.Show()
}

// Hide hides the settings dialog
func (sd *SettingsDialog) Hide() {
	sd.dialog.Hide()
}