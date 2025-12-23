package repository

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"hugo-visual-client/internal/models"
	"gopkg.in/yaml.v3"
)

// ConfigRepository handles configuration file operations
type ConfigRepository struct {
	fileRepo *FileRepository
}

// NewConfigRepository creates a new configuration repository
func NewConfigRepository(basePath string) *ConfigRepository {
	return &ConfigRepository{
		fileRepo: NewFileRepository(basePath),
	}
}

// LoadSiteConfig loads Hugo site configuration from config file
func (cr *ConfigRepository) LoadSiteConfig(configPath string) (*models.SiteConfig, error) {
	data, err := cr.fileRepo.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config models.SiteConfig
	
	// Determine file format based on extension
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		err = json.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// SaveSiteConfig saves Hugo site configuration to config file
func (cr *ConfigRepository) SaveSiteConfig(configPath string, config *models.SiteConfig) error {
	// Validate configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	var data []byte
	var err error
	
	// Determine file format based on extension
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
	
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	err = cr.fileRepo.WriteFile(configPath, data)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// LoadAppConfig loads application configuration
func (cr *ConfigRepository) LoadAppConfig(configPath string) (*models.AppConfig, error) {
	data, err := cr.fileRepo.ReadFile(configPath)
	if err != nil {
		// If config file doesn't exist, return default config
		if cr.fileRepo.FileExists(configPath) {
			return nil, fmt.Errorf("failed to read app config file: %w", err)
		}
		
		// Return default configuration
		defaultConfig := &models.AppConfig{
			RecentProjects: []string{},
			WindowWidth:    1200,
			WindowHeight:   800,
			Theme:          "default",
			AutoSave:       true,
			PreviewPort:    1313,
			Custom:         make(map[string]interface{}),
		}
		return defaultConfig, nil
	}
	
	var config models.AppConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse app config file: %w", err)
	}
	
	// Validate and set defaults
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid app configuration: %w", err)
	}
	
	return &config, nil
}

// SaveAppConfig saves application configuration
func (cr *ConfigRepository) SaveAppConfig(configPath string, config *models.AppConfig) error {
	// Validate configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid app configuration: %w", err)
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal app config: %w", err)
	}
	
	err = cr.fileRepo.WriteFile(configPath, data)
	if err != nil {
		return fmt.Errorf("failed to write app config file: %w", err)
	}
	
	return nil
}

// LoadFrontMatter loads front matter from a markdown file
func (cr *ConfigRepository) LoadFrontMatter(filePath string) (*models.FrontMatter, string, error) {
	data, err := cr.fileRepo.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}
	
	content := string(data)
	
	// Check if file has front matter
	if !strings.HasPrefix(content, "---") {
		return nil, content, nil
	}
	
	// Find the end of front matter
	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) != 2 {
		return nil, content, nil
	}
	
	frontMatterYAML := strings.TrimSpace(parts[0])
	markdownContent := strings.TrimSpace(parts[1])
	
	var frontMatter models.FrontMatter
	err = yaml.Unmarshal([]byte(frontMatterYAML), &frontMatter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse front matter: %w", err)
	}
	
	// Validate front matter
	if err := frontMatter.Validate(); err != nil {
		return nil, "", fmt.Errorf("invalid front matter: %w", err)
	}
	
	return &frontMatter, markdownContent, nil
}

// SaveContentWithFrontMatter saves content with front matter to a markdown file
func (cr *ConfigRepository) SaveContentWithFrontMatter(filePath string, frontMatter *models.FrontMatter, content string) error {
	// Validate front matter before saving
	if err := frontMatter.Validate(); err != nil {
		return fmt.Errorf("invalid front matter: %w", err)
	}
	
	// Marshal front matter to YAML
	frontMatterData, err := yaml.Marshal(frontMatter)
	if err != nil {
		return fmt.Errorf("failed to marshal front matter: %w", err)
	}
	
	// Combine front matter and content
	fullContent := fmt.Sprintf("---\n%s---\n\n%s", string(frontMatterData), content)
	
	err = cr.fileRepo.WriteFile(filePath, []byte(fullContent))
	if err != nil {
		return fmt.Errorf("failed to write content file: %w", err)
	}
	
	return nil
}

// ConfigExists checks if a configuration file exists
func (cr *ConfigRepository) ConfigExists(configPath string) bool {
	return cr.fileRepo.FileExists(configPath)
}

// FindConfigFile finds the Hugo configuration file in the project directory
func (cr *ConfigRepository) FindConfigFile() (string, error) {
	// Common Hugo config file names
	configFiles := []string{
		"config.yaml",
		"config.yml",
		"config.json",
		"config.toml",
		"hugo.yaml",
		"hugo.yml",
		"hugo.json",
		"hugo.toml",
	}
	
	for _, configFile := range configFiles {
		if cr.fileRepo.FileExists(configFile) {
			return configFile, nil
		}
	}
	
	return "", fmt.Errorf("no Hugo configuration file found")
}