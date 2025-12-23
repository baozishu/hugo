package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hugo-visual-client/internal/interfaces"
	"hugo-visual-client/internal/models"
	"hugo-visual-client/internal/repository"
)

// ProjectManagerService implements the ProjectManager interface
type ProjectManagerService struct {
	configRepo *repository.ConfigRepository
	appConfig  *models.AppConfig
	configPath string
}

// NewProjectManagerService creates a new project manager service
func NewProjectManagerService(configPath string) (*ProjectManagerService, error) {
	// Use user's home directory for app config if no path specified
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".hugo-client", "config.json")
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configRepo := repository.NewConfigRepository(configDir)
	
	// Load or create app config
	appConfig, err := configRepo.LoadAppConfig(filepath.Base(configPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	return &ProjectManagerService{
		configRepo: configRepo,
		appConfig:  appConfig,
		configPath: configPath,
	}, nil
}

// CreateProject creates a new Hugo project with the given configuration
func (pm *ProjectManagerService) CreateProject(path string, config interfaces.ProjectConfig) (*interfaces.Project, error) {
	// Validate input
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("project path cannot be empty")
	}
	if strings.TrimSpace(config.Name) == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	if strings.TrimSpace(config.Title) == "" {
		config.Title = config.Name // Default title to name
	}
	if strings.TrimSpace(config.BaseURL) == "" {
		config.BaseURL = "http://localhost:1313" // Default base URL
	}
	if strings.TrimSpace(config.Language) == "" {
		config.Language = "en" // Default language
	}

	// Check if directory already exists
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("directory already exists: %s", path)
	}

	// Create project directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create Hugo standard directory structure
	dirs := []string{
		"content",
		"themes",
		"static",
		"layouts",
		"data",
		"assets",
		"archetypes",
		"public",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(path, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create site configuration
	siteConfig := &models.SiteConfig{
		BaseURL:      config.BaseURL,
		Title:        config.Title,
		Description:  config.Description,
		LanguageCode: config.Language,
		Theme:        config.Theme,
		Params:       make(map[string]interface{}),
		Menu:         make(map[string][]models.MenuItem),
		Taxonomies: map[string]string{
			"tags":       "tags",
			"categories": "categories",
		},
		OutputFormats: make(map[string]interface{}),
	}

	// Save configuration file
	configRepo := repository.NewConfigRepository(path)
	configFile := "config.yaml"
	if err := configRepo.SaveSiteConfig(configFile, siteConfig); err != nil {
		return nil, fmt.Errorf("failed to save site config: %w", err)
	}

	// Create default archetype
	defaultArchetype := `---
title: "{{ replace .Name "-" " " | title }}"
date: {{ .Date }}
draft: true
---

`
	archetypePath := filepath.Join(path, "archetypes", "default.md")
	if err := os.WriteFile(archetypePath, []byte(defaultArchetype), 0644); err != nil {
		return nil, fmt.Errorf("failed to create default archetype: %w", err)
	}

	// Create project object
	project := &interfaces.Project{
		Path:       path,
		Name:       config.Name,
		ConfigFile: configFile,
		ContentDir: "content",
		ThemeDir:   "themes",
		PublicDir:  "public",
		Config:     siteConfig.ToMap(),
		LastOpened: time.Now(),
	}

	// Add to recent projects
	if err := pm.AddRecentProject(path); err != nil {
		// Log error but don't fail project creation
		fmt.Printf("Warning: failed to add project to recent list: %v\n", err)
	}

	return project, nil
}

// OpenProject opens an existing Hugo project from the given path
func (pm *ProjectManagerService) OpenProject(path string) (*interfaces.Project, error) {
	// Validate project first
	if err := pm.ValidateProject(path); err != nil {
		return nil, fmt.Errorf("invalid Hugo project: %w", err)
	}

	// Find configuration file
	configRepo := repository.NewConfigRepository(path)
	configFile, err := configRepo.FindConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to find config file: %w", err)
	}

	// Load site configuration
	siteConfig, err := configRepo.LoadSiteConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load site config: %w", err)
	}

	// Extract project name from path
	projectName := filepath.Base(path)

	// Create project object
	project := &interfaces.Project{
		Path:       path,
		Name:       projectName,
		ConfigFile: configFile,
		ContentDir: "content",
		ThemeDir:   "themes",
		PublicDir:  "public",
		Config:     siteConfig.ToMap(),
		LastOpened: time.Now(),
	}

	// Add to recent projects
	if err := pm.AddRecentProject(path); err != nil {
		// Log error but don't fail project opening
		fmt.Printf("Warning: failed to add project to recent list: %v\n", err)
	}

	return project, nil
}

// ValidateProject validates if the given path contains a valid Hugo project
func (pm *ProjectManagerService) ValidateProject(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", path)
	}

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat project path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory: %s", path)
	}

	// Check for Hugo configuration file
	configRepo := repository.NewConfigRepository(path)
	_, err = configRepo.FindConfigFile()
	if err != nil {
		return fmt.Errorf("no Hugo configuration file found in %s", path)
	}

	// Check for required directories
	requiredDirs := []string{"content"}
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(path, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			return fmt.Errorf("required directory missing: %s", dir)
		}
	}

	return nil
}

// GetRecentProjects returns a list of recently opened project paths
func (pm *ProjectManagerService) GetRecentProjects() []string {
	// Filter out non-existent projects
	var validProjects []string
	for _, projectPath := range pm.appConfig.RecentProjects {
		if pm.ValidateProject(projectPath) == nil {
			validProjects = append(validProjects, projectPath)
		}
	}

	// Update app config if we filtered out invalid projects
	if len(validProjects) != len(pm.appConfig.RecentProjects) {
		pm.appConfig.RecentProjects = validProjects
		pm.saveAppConfig()
	}

	return validProjects
}

// AddRecentProject adds a project path to the recent projects list
func (pm *ProjectManagerService) AddRecentProject(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Remove if already exists (to move to front)
	for i, existingPath := range pm.appConfig.RecentProjects {
		if existingPath == absPath {
			pm.appConfig.RecentProjects = append(
				pm.appConfig.RecentProjects[:i],
				pm.appConfig.RecentProjects[i+1:]...,
			)
			break
		}
	}

	// Add to front
	pm.appConfig.RecentProjects = append([]string{absPath}, pm.appConfig.RecentProjects...)

	// Limit to 10 recent projects
	if len(pm.appConfig.RecentProjects) > 10 {
		pm.appConfig.RecentProjects = pm.appConfig.RecentProjects[:10]
	}

	// Save app config
	return pm.saveAppConfig()
}

// RemoveRecentProject removes a project path from the recent projects list
func (pm *ProjectManagerService) RemoveRecentProject(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	// Convert to absolute path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find and remove the project
	for i, existingPath := range pm.appConfig.RecentProjects {
		if existingPath == absPath {
			pm.appConfig.RecentProjects = append(
				pm.appConfig.RecentProjects[:i],
				pm.appConfig.RecentProjects[i+1:]...,
			)
			break
		}
	}

	// Save app config
	return pm.saveAppConfig()
}

// ClearRecentProjects clears all recent projects
func (pm *ProjectManagerService) ClearRecentProjects() error {
	pm.appConfig.RecentProjects = []string{}
	return pm.saveAppConfig()
}

// GetRecentProjectsCount returns the number of recent projects
func (pm *ProjectManagerService) GetRecentProjectsCount() int {
	return len(pm.GetRecentProjects())
}

// SaveProject saves project configuration and metadata
func (pm *ProjectManagerService) SaveProject(project *interfaces.Project) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	// Validate project
	projectModel := &models.Project{
		Path:       project.Path,
		Name:       project.Name,
		ConfigFile: project.ConfigFile,
		ContentDir: project.ContentDir,
		ThemeDir:   project.ThemeDir,
		PublicDir:  project.PublicDir,
		Config:     project.Config,
		LastOpened: project.LastOpened,
	}

	if err := projectModel.Validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	// Convert config map to SiteConfig
	siteConfig, err := pm.mapToSiteConfig(project.Config)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Save site configuration
	configRepo := repository.NewConfigRepository(project.Path)
	if err := configRepo.SaveSiteConfig(project.ConfigFile, siteConfig); err != nil {
		return fmt.Errorf("failed to save site config: %w", err)
	}

	// Update last opened time
	project.LastOpened = time.Now()

	// Add to recent projects
	return pm.AddRecentProject(project.Path)
}

// saveAppConfig saves the application configuration
func (pm *ProjectManagerService) saveAppConfig() error {
	return pm.configRepo.SaveAppConfig(filepath.Base(pm.configPath), pm.appConfig)
}

// mapToSiteConfig converts a generic map to SiteConfig
func (pm *ProjectManagerService) mapToSiteConfig(configMap map[string]interface{}) (*models.SiteConfig, error) {
	// Convert map to JSON and back to SiteConfig for type safety
	jsonData, err := json.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config map: %w", err)
	}

	var siteConfig models.SiteConfig
	if err := json.Unmarshal(jsonData, &siteConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to SiteConfig: %w", err)
	}

	return &siteConfig, nil
}