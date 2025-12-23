package interfaces

import (
	"time"
)

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

// FrontMatter represents the YAML front matter of a content file
type FrontMatter struct {
	Title       string                 `yaml:"title"`
	Date        time.Time              `yaml:"date"`
	Draft       bool                   `yaml:"draft"`
	Tags        []string               `yaml:"tags"`
	Categories  []string               `yaml:"categories"`
	Description string                 `yaml:"description"`
	Author      string                 `yaml:"author"`
	Image       string                 `yaml:"image"`
	Custom      map[string]interface{} `yaml:",inline"`
}

// ContentManager interface defines content management operations
type ContentManager interface {
	// ListContent returns all content items in the project
	ListContent(projectPath string) ([]ContentItem, error)
	
	// CreateContent creates a new content file with front matter and content
	CreateContent(path string, frontMatter FrontMatter, content string) error
	
	// UpdateContent updates an existing content file
	UpdateContent(path string, frontMatter FrontMatter, content string) error
	
	// DeleteContent deletes a content file
	DeleteContent(path string) error
	
	// GetContent retrieves a specific content item
	GetContent(path string) (*ContentItem, error)
	
	// SearchContent searches for content items matching the query
	SearchContent(query string) ([]ContentItem, error)
	
	// ParseFrontMatter parses front matter from a content file
	ParseFrontMatter(content string) (FrontMatter, string, error)
	
	// SerializeFrontMatter serializes front matter to YAML
	SerializeFrontMatter(frontMatter FrontMatter) (string, error)
	
	// Resource management methods
	
	// CopyResource copies a media file to the static resources directory
	CopyResource(sourcePath, destPath string) error
	
	// GetResourcePath returns the correct path for a resource file
	GetResourcePath(resourceName string) string
	
	// ListResources lists all resource files in the static directory
	ListResources() ([]ResourceInfo, error)
	
	// DeleteResource deletes a resource file
	DeleteResource(resourcePath string) error
	
	// Batch operation methods
	
	// BatchDeleteContent deletes multiple content files atomically
	BatchDeleteContent(paths []string) error
	
	// BatchMoveContent moves multiple content files to new locations atomically
	BatchMoveContent(moves []ContentMove) error
	
	// BatchUpdateTags updates tags for multiple content files atomically
	BatchUpdateTags(updates []TagUpdate) error
}

// ContentMove represents a file move operation
type ContentMove struct {
	SourcePath string `json:"source_path"`
	DestPath   string `json:"dest_path"`
}

// TagUpdate represents a tag update operation
type TagUpdate struct {
	Path string   `json:"path"`
	Tags []string `json:"tags"`
}

// ResourceInfo represents information about a resource file
type ResourceInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	IsImage      bool      `json:"is_image"`
}