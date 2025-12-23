package services

import (
	"fmt"
	"path/filepath"

	"hugo-visual-client/internal/models"
	"hugo-visual-client/internal/repository"
)

// SettingsManager handles application settings operations
type SettingsManager struct {
	configRepo *repository.ConfigRepository
	configPath string
	config     *models.AppConfig
}

// NewSettingsManager creates a new settings manager
func NewSettingsManager(configDir string) (*SettingsManager, error) {
	configRepo := repository.NewConfigRepository(configDir)
	configPath := filepath.Join(configDir, "config.json")
	
	// Load existing configuration
	config, err := configRepo.LoadAppConfig("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}
	
	return &SettingsManager{
		configRepo: configRepo,
		configPath: configPath,
		config:     config,
	}, nil
}

// GetConfig returns the current application configuration
func (sm *SettingsManager) GetConfig() *models.AppConfig {
	return sm.config
}

// SaveConfig saves the application configuration
func (sm *SettingsManager) SaveConfig(config *models.AppConfig) error {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Save to file
	err := sm.configRepo.SaveAppConfig("config.json", config)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	// Update internal config
	sm.config = config
	
	return nil
}

// ReloadConfig reloads the configuration from disk
func (sm *SettingsManager) ReloadConfig() error {
	config, err := sm.configRepo.LoadAppConfig("config.json")
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}
	
	sm.config = config
	return nil
}

// GetSetting gets a custom setting value
func (sm *SettingsManager) GetSetting(key string) (interface{}, bool) {
	if sm.config.Custom == nil {
		return nil, false
	}
	
	value, exists := sm.config.Custom[key]
	return value, exists
}

// SetSetting sets a custom setting value
func (sm *SettingsManager) SetSetting(key string, value interface{}) error {
	if sm.config.Custom == nil {
		sm.config.Custom = make(map[string]interface{})
	}
	
	sm.config.Custom[key] = value
	
	// Save the updated configuration
	return sm.SaveConfig(sm.config)
}

// GetStringSetting gets a string setting with default value
func (sm *SettingsManager) GetStringSetting(key, defaultValue string) string {
	if value, exists := sm.GetSetting(key); exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// GetIntSetting gets an integer setting with default value
func (sm *SettingsManager) GetIntSetting(key string, defaultValue int) int {
	if value, exists := sm.GetSetting(key); exists {
		if intValue, ok := value.(float64); ok { // JSON numbers are float64
			return int(intValue)
		}
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}

// GetBoolSetting gets a boolean setting with default value
func (sm *SettingsManager) GetBoolSetting(key string, defaultValue bool) bool {
	if value, exists := sm.GetSetting(key); exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// ResetToDefaults resets all settings to their default values
func (sm *SettingsManager) ResetToDefaults() error {
	defaultConfig := &models.AppConfig{
		RecentProjects: sm.config.RecentProjects, // Keep recent projects
		WindowWidth:    1200,
		WindowHeight:   800,
		Theme:          "default",
		AutoSave:       true,
		PreviewPort:    1313,
		Custom:         make(map[string]interface{}),
	}
	
	// Set default custom settings
	defaultConfig.Custom["editor_font_size"] = 12
	defaultConfig.Custom["preview_auto_refresh"] = true
	defaultConfig.Custom["show_line_numbers"] = true
	defaultConfig.Custom["hugo_path"] = ""
	defaultConfig.Custom["default_theme"] = ""
	
	return sm.SaveConfig(defaultConfig)
}

// ApplyWindowSettings applies window-related settings to the application
func (sm *SettingsManager) ApplyWindowSettings(window interface{}) error {
	// This would be implemented to apply window settings
	// For now, we'll just validate that the settings are reasonable
	if sm.config.WindowWidth < 800 || sm.config.WindowWidth > 3840 {
		return fmt.Errorf("invalid window width: %d", sm.config.WindowWidth)
	}
	if sm.config.WindowHeight < 600 || sm.config.WindowHeight > 2160 {
		return fmt.Errorf("invalid window height: %d", sm.config.WindowHeight)
	}
	
	// In a real implementation, this would resize the window
	// window.Resize(fyne.NewSize(float32(sm.config.WindowWidth), float32(sm.config.WindowHeight)))
	
	return nil
}

// GetEditorSettings returns editor-specific settings
func (sm *SettingsManager) GetEditorSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	
	settings["font_size"] = sm.GetIntSetting("editor_font_size", 12)
	settings["show_line_numbers"] = sm.GetBoolSetting("show_line_numbers", true)
	settings["auto_save"] = sm.config.AutoSave
	
	return settings
}

// GetPreviewSettings returns preview-specific settings
func (sm *SettingsManager) GetPreviewSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	
	settings["port"] = sm.config.PreviewPort
	settings["auto_refresh"] = sm.GetBoolSetting("preview_auto_refresh", true)
	
	return settings
}

// GetHugoSettings returns Hugo-specific settings
func (sm *SettingsManager) GetHugoSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	
	settings["hugo_path"] = sm.GetStringSetting("hugo_path", "")
	settings["default_theme"] = sm.GetStringSetting("default_theme", "")
	
	return settings
}

// UpdateRecentProjects updates the recent projects list
func (sm *SettingsManager) UpdateRecentProjects(projects []string) error {
	sm.config.RecentProjects = projects
	return sm.SaveConfig(sm.config)
}

// AddRecentProject adds a project to the recent projects list
func (sm *SettingsManager) AddRecentProject(projectPath string) error {
	// Remove if already exists
	for i, path := range sm.config.RecentProjects {
		if path == projectPath {
			sm.config.RecentProjects = append(sm.config.RecentProjects[:i], sm.config.RecentProjects[i+1:]...)
			break
		}
	}
	
	// Add to front
	sm.config.RecentProjects = append([]string{projectPath}, sm.config.RecentProjects...)
	
	// Limit to 10 recent projects
	if len(sm.config.RecentProjects) > 10 {
		sm.config.RecentProjects = sm.config.RecentProjects[:10]
	}
	
	return sm.SaveConfig(sm.config)
}

// GetConfigPath returns the configuration file path
func (sm *SettingsManager) GetConfigPath() string {
	return sm.configPath
}