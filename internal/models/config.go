package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SiteConfig represents Hugo site configuration
type SiteConfig struct {
	BaseURL         string                 `yaml:"baseURL" json:"baseURL"`
	Title           string                 `yaml:"title" json:"title"`
	Description     string                 `yaml:"description" json:"description"`
	LanguageCode    string                 `yaml:"languageCode" json:"languageCode"`
	Theme           string                 `yaml:"theme" json:"theme"`
	Params          map[string]interface{} `yaml:"params" json:"params"`
	Menu            map[string][]MenuItem  `yaml:"menu" json:"menu"`
	Taxonomies      map[string]string      `yaml:"taxonomies" json:"taxonomies"`
	OutputFormats   map[string]interface{} `yaml:"outputFormats" json:"outputFormats"`
}

// Validate checks if the site configuration is valid
func (sc *SiteConfig) Validate() error {
	if strings.TrimSpace(sc.Title) == "" {
		return fmt.Errorf("site title cannot be empty")
	}
	if strings.TrimSpace(sc.BaseURL) == "" {
		return fmt.Errorf("baseURL cannot be empty")
	}
	if strings.TrimSpace(sc.LanguageCode) == "" {
		sc.LanguageCode = "en" // Default to English
	}
	return nil
}

// MenuItem represents a menu item in Hugo configuration
type MenuItem struct {
	Name       string `yaml:"name" json:"name"`
	URL        string `yaml:"url" json:"url"`
	Weight     int    `yaml:"weight" json:"weight"`
	Identifier string `yaml:"identifier" json:"identifier"`
	Parent     string `yaml:"parent,omitempty" json:"parent,omitempty"`
}

// Validate checks if the menu item is valid
func (mi *MenuItem) Validate() error {
	if strings.TrimSpace(mi.Name) == "" {
		return fmt.Errorf("menu item name cannot be empty")
	}
	if strings.TrimSpace(mi.URL) == "" {
		return fmt.Errorf("menu item URL cannot be empty")
	}
	return nil
}

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

// Validate checks if the project configuration is valid
func (p *Project) Validate() error {
	if strings.TrimSpace(p.Path) == "" {
		return fmt.Errorf("project path cannot be empty")
	}
	
	// Check if path exists
	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", p.Path)
	}
	
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	
	// Validate required directories exist
	requiredDirs := []string{p.ContentDir, p.ThemeDir}
	for _, dir := range requiredDirs {
		if dir != "" {
			fullPath := filepath.Join(p.Path, dir)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				return fmt.Errorf("required directory does not exist: %s", fullPath)
			}
		}
	}
	
	// Check config file exists
	if p.ConfigFile != "" {
		configPath := filepath.Join(p.Path, p.ConfigFile)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", configPath)
		}
	}
	
	return nil
}

// ContentItem represents a content file in Hugo
type ContentItem struct {
	Path         string                 `json:"path"`
	Title        string                 `json:"title"`
	Date         time.Time              `json:"date"`
	Draft        bool                   `json:"draft"`
	Tags         []string               `json:"tags"`
	Categories   []string               `json:"categories"`
	FrontMatter  map[string]interface{} `json:"front_matter"`
	Content      string                 `json:"content"`
	WordCount    int                    `json:"word_count"`
	ModifiedTime time.Time              `json:"modified_time"`
}

// Validate checks if the content item is valid
func (ci *ContentItem) Validate() error {
	if strings.TrimSpace(ci.Path) == "" {
		return fmt.Errorf("content item path cannot be empty")
	}
	
	if strings.TrimSpace(ci.Title) == "" {
		return fmt.Errorf("content item title cannot be empty")
	}
	
	// Check if path has .md extension
	if !strings.HasSuffix(strings.ToLower(ci.Path), ".md") {
		return fmt.Errorf("content item must be a markdown file (.md)")
	}
	
	return nil
}

// UpdateWordCount calculates and updates the word count based on content
func (ci *ContentItem) UpdateWordCount() {
	words := strings.Fields(ci.Content)
	ci.WordCount = len(words)
}

// FrontMatter represents the YAML front matter of a content file
type FrontMatter struct {
	Title       string                 `yaml:"title" json:"title"`
	Date        time.Time              `yaml:"date" json:"date"`
	Draft       bool                   `yaml:"draft" json:"draft"`
	Tags        []string               `yaml:"tags" json:"tags"`
	Categories  []string               `yaml:"categories" json:"categories"`
	Description string                 `yaml:"description" json:"description"`
	Author      string                 `yaml:"author" json:"author"`
	Image       string                 `yaml:"image" json:"image"`
	Custom      map[string]interface{} `yaml:",inline" json:"custom"`
}

// Validate checks if the front matter is valid
func (fm *FrontMatter) Validate() error {
	if strings.TrimSpace(fm.Title) == "" {
		return fmt.Errorf("front matter title cannot be empty")
	}
	
	if fm.Date.IsZero() {
		fm.Date = time.Now() // Default to current time
	}
	
	return nil
}

// ToMap converts FrontMatter to a map for generic processing
func (fm *FrontMatter) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	
	result["title"] = fm.Title
	result["date"] = fm.Date
	result["draft"] = fm.Draft
	
	if len(fm.Tags) > 0 {
		result["tags"] = fm.Tags
	}
	if len(fm.Categories) > 0 {
		result["categories"] = fm.Categories
	}
	if fm.Description != "" {
		result["description"] = fm.Description
	}
	if fm.Author != "" {
		result["author"] = fm.Author
	}
	if fm.Image != "" {
		result["image"] = fm.Image
	}
	
	// Add custom fields
	for k, v := range fm.Custom {
		result[k] = v
	}
	
	return result
}

// AppConfig represents application-specific configuration
type AppConfig struct {
	RecentProjects []string               `json:"recent_projects"`
	WindowWidth    int                    `json:"window_width"`
	WindowHeight   int                    `json:"window_height"`
	Theme          string                 `json:"theme"`
	AutoSave       bool                   `json:"auto_save"`
	PreviewPort    int                    `json:"preview_port"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

// Validate checks if the app configuration is valid
func (ac *AppConfig) Validate() error {
	if ac.WindowWidth <= 0 {
		ac.WindowWidth = 1200 // Default width
	}
	if ac.WindowHeight <= 0 {
		ac.WindowHeight = 800 // Default height
	}
	if ac.PreviewPort <= 0 || ac.PreviewPort > 65535 {
		ac.PreviewPort = 1313 // Hugo default port
	}
	if ac.Theme == "" {
		ac.Theme = "default"
	}
	if ac.Custom == nil {
		ac.Custom = make(map[string]interface{})
	}
	return nil
}

// ToMap converts SiteConfig to a generic map
func (sc *SiteConfig) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	
	result["baseURL"] = sc.BaseURL
	result["title"] = sc.Title
	result["description"] = sc.Description
	result["languageCode"] = sc.LanguageCode
	result["theme"] = sc.Theme
	
	if sc.Params != nil {
		result["params"] = sc.Params
	}
	if sc.Menu != nil {
		result["menu"] = sc.Menu
	}
	if sc.Taxonomies != nil {
		result["taxonomies"] = sc.Taxonomies
	}
	if sc.OutputFormats != nil {
		result["outputFormats"] = sc.OutputFormats
	}
	
	return result
}