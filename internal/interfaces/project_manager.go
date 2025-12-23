package interfaces

import (
	"time"
)

// Project represents a Hugo project
type Project struct {
	Path        string                 `json:"path"`
	Name        string                 `json:"name"`
	ConfigFile  string                 `json:"config_file"`
	ContentDir  string                 `json:"content_dir"`
	ThemeDir    string                 `json:"theme_dir"`
	PublicDir   string                 `json:"public_dir"`
	Config      map[string]interface{} `json:"config"`
	LastOpened  time.Time              `json:"last_opened"`
}

// ProjectConfig represents configuration for creating a new project
type ProjectConfig struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	BaseURL     string `json:"base_url"`
	Theme       string `json:"theme"`
	Language    string `json:"language"`
	Description string `json:"description"`
}

// ProjectManager interface defines project management operations
type ProjectManager interface {
	// CreateProject creates a new Hugo project with the given configuration
	CreateProject(path string, config ProjectConfig) (*Project, error)
	
	// OpenProject opens an existing Hugo project from the given path
	OpenProject(path string) (*Project, error)
	
	// ValidateProject validates if the given path contains a valid Hugo project
	ValidateProject(path string) error
	
	// GetRecentProjects returns a list of recently opened project paths
	GetRecentProjects() []string
	
	// AddRecentProject adds a project path to the recent projects list
	AddRecentProject(path string) error
	
	// SaveProject saves project configuration and metadata
	SaveProject(project *Project) error
}