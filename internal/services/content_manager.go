package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"hugo-visual-client/internal/interfaces"
	"hugo-visual-client/internal/repository"
	"gopkg.in/yaml.v3"
)

// ContentManagerService implements the ContentManager interface
type ContentManagerService struct {
	fileRepo    *repository.FileRepository
	projectPath string
}

// NewContentManager creates a new content manager service
func NewContentManager(projectPath string) *ContentManagerService {
	return &ContentManagerService{
		fileRepo:    repository.NewFileRepository(projectPath),
		projectPath: projectPath,
	}
}

// ListContent returns all content items in the project
func (cm *ContentManagerService) ListContent(projectPath string) ([]interfaces.ContentItem, error) {
	contentDir := filepath.Join(projectPath, "content")
	
	// Check if content directory exists
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return []interfaces.ContentItem{}, nil
	}
	
	var contentItems []interfaces.ContentItem
	
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		
		// Get relative path from project root
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return err
		}
		
		// Normalize path separators to forward slashes for consistency
		relPath = filepath.ToSlash(relPath)
		
		// Read and parse the content file
		contentItem, err := cm.parseContentFile(path, relPath, info.ModTime())
		if err != nil {
			// Log error but continue processing other files
			fmt.Printf("Warning: failed to parse content file %s: %v\n", relPath, err)
			return nil
		}
		
		contentItems = append(contentItems, *contentItem)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk content directory: %w", err)
	}
	
	return contentItems, nil
}

// CreateContent creates a new content file with front matter and content
func (cm *ContentManagerService) CreateContent(path string, frontMatter interfaces.FrontMatter, content string) error {
	// Serialize front matter to YAML
	frontMatterYAML, err := cm.SerializeFrontMatter(frontMatter)
	if err != nil {
		return fmt.Errorf("failed to serialize front matter: %w", err)
	}
	
	// Combine front matter and content
	fullContent := fmt.Sprintf("---\n%s---\n%s", frontMatterYAML, content)
	
	// Write to file
	err = cm.fileRepo.CreateFile(path, []byte(fullContent))
	if err != nil {
		return fmt.Errorf("failed to create content file: %w", err)
	}
	
	return nil
}

// UpdateContent updates an existing content file
func (cm *ContentManagerService) UpdateContent(path string, frontMatter interfaces.FrontMatter, content string) error {
	// Serialize front matter to YAML
	frontMatterYAML, err := cm.SerializeFrontMatter(frontMatter)
	if err != nil {
		return fmt.Errorf("failed to serialize front matter: %w", err)
	}
	
	// Combine front matter and content
	fullContent := fmt.Sprintf("---\n%s---\n%s", frontMatterYAML, content)
	
	// Write to file
	err = cm.fileRepo.WriteFile(path, []byte(fullContent))
	if err != nil {
		return fmt.Errorf("failed to update content file: %w", err)
	}
	
	return nil
}

// DeleteContent deletes a content file
func (cm *ContentManagerService) DeleteContent(path string) error {
	err := cm.fileRepo.DeleteFile(path)
	if err != nil {
		return fmt.Errorf("failed to delete content file: %w", err)
	}
	
	return nil
}

// GetContent retrieves a specific content item
func (cm *ContentManagerService) GetContent(path string) (*interfaces.ContentItem, error) {
	fullPath := filepath.Join(cm.projectPath, path)
	
	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	
	// Parse the content file
	contentItem, err := cm.parseContentFile(fullPath, path, info.ModTime())
	if err != nil {
		return nil, fmt.Errorf("failed to parse content file: %w", err)
	}
	
	return contentItem, nil
}

// SearchContent searches for content items matching the query
func (cm *ContentManagerService) SearchContent(query string) ([]interfaces.ContentItem, error) {
	// Get all content items
	allContent, err := cm.ListContent(cm.projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list content: %w", err)
	}
	
	if query == "" {
		return allContent, nil
	}
	
	// Convert query to lowercase for case-insensitive search
	queryLower := strings.ToLower(query)
	
	var matchingItems []interfaces.ContentItem
	
	for _, item := range allContent {
		// Search in title, tags, categories, and content
		if cm.matchesQuery(item, queryLower) {
			matchingItems = append(matchingItems, item)
		}
	}
	
	return matchingItems, nil
}

// ParseFrontMatter parses front matter from a content file
func (cm *ContentManagerService) ParseFrontMatter(content string) (interfaces.FrontMatter, string, error) {
	// Check if content starts with front matter delimiter
	if !strings.HasPrefix(content, "---\n") {
		return interfaces.FrontMatter{}, content, nil
	}
	
	// Find the end of front matter
	lines := strings.Split(content, "\n")
	var frontMatterLines []string
	var contentLines []string
	var inFrontMatter = true
	var frontMatterEnded = false
	
	for i, line := range lines {
		if i == 0 && line == "---" {
			continue // Skip the opening delimiter
		}
		
		if inFrontMatter && line == "---" {
			inFrontMatter = false
			frontMatterEnded = true
			continue
		}
		
		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
		} else {
			contentLines = append(contentLines, line)
		}
	}
	
	if !frontMatterEnded {
		return interfaces.FrontMatter{}, content, fmt.Errorf("front matter not properly closed")
	}
	
	// Parse YAML front matter
	var frontMatter interfaces.FrontMatter
	if len(frontMatterLines) > 0 {
		frontMatterYAML := strings.Join(frontMatterLines, "\n")
		err := yaml.Unmarshal([]byte(frontMatterYAML), &frontMatter)
		if err != nil {
			return interfaces.FrontMatter{}, content, fmt.Errorf("failed to parse front matter YAML: %w", err)
		}
	}
	
	// Join content lines
	contentBody := strings.Join(contentLines, "\n")
	
	return frontMatter, contentBody, nil
}

// SerializeFrontMatter serializes front matter to YAML
func (cm *ContentManagerService) SerializeFrontMatter(frontMatter interfaces.FrontMatter) (string, error) {
	data, err := yaml.Marshal(frontMatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal front matter to YAML: %w", err)
	}
	
	return string(data), nil
}

// parseContentFile parses a content file and returns a ContentItem
func (cm *ContentManagerService) parseContentFile(fullPath, relativePath string, modTime time.Time) (*interfaces.ContentItem, error) {
	// Read file content
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	content := string(data)
	
	// Parse front matter and content
	frontMatter, contentBody, err := cm.ParseFrontMatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}
	
	// Count words in content
	wordCount := cm.countWords(contentBody)
	
	// Convert front matter to map for JSON serialization
	frontMatterMap := make(map[string]interface{})
	frontMatterData, _ := yaml.Marshal(frontMatter)
	yaml.Unmarshal(frontMatterData, &frontMatterMap)
	
	contentItem := &interfaces.ContentItem{
		Path:         relativePath,
		Title:        frontMatter.Title,
		Date:         frontMatter.Date,
		Draft:        frontMatter.Draft,
		Tags:         frontMatter.Tags,
		Categories:   frontMatter.Categories,
		FrontMatter:  frontMatterMap,
		Content:      contentBody,
		WordCount:    wordCount,
		ModifiedTime: modTime,
	}
	
	return contentItem, nil
}

// matchesQuery checks if a content item matches the search query
func (cm *ContentManagerService) matchesQuery(item interfaces.ContentItem, queryLower string) bool {
	// Search in title
	if strings.Contains(strings.ToLower(item.Title), queryLower) {
		return true
	}
	
	// Search in tags
	for _, tag := range item.Tags {
		if strings.Contains(strings.ToLower(tag), queryLower) {
			return true
		}
	}
	
	// Search in categories
	for _, category := range item.Categories {
		if strings.Contains(strings.ToLower(category), queryLower) {
			return true
		}
	}
	
	// Search in content
	if strings.Contains(strings.ToLower(item.Content), queryLower) {
		return true
	}
	
	return false
}

// countWords counts the number of words in the content
func (cm *ContentManagerService) countWords(content string) int {
	// Remove markdown syntax for more accurate word count
	content = cm.stripMarkdown(content)
	
	// Split by whitespace and count non-empty strings
	words := strings.Fields(content)
	return len(words)
}

// stripMarkdown removes basic markdown syntax for word counting
func (cm *ContentManagerService) stripMarkdown(content string) string {
	// Remove headers
	re := regexp.MustCompile(`#{1,6}\s+`)
	content = re.ReplaceAllString(content, "")
	
	// Remove bold and italic
	re = regexp.MustCompile(`\*{1,2}([^*]+)\*{1,2}`)
	content = re.ReplaceAllString(content, "$1")
	
	// Remove links
	re = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	content = re.ReplaceAllString(content, "$1")
	
	// Remove code blocks
	re = regexp.MustCompile("```[\\s\\S]*?```")
	content = re.ReplaceAllString(content, "")
	
	// Remove inline code
	re = regexp.MustCompile("`([^`]+)`")
	content = re.ReplaceAllString(content, "$1")
	
	return content
}
// Resource management methods

// CopyResource copies a media file to the static resources directory
func (cm *ContentManagerService) CopyResource(sourcePath, destPath string) error {
	// Validate the destination path before joining
	if cm.isUnsafePath(destPath) {
		return fmt.Errorf("invalid destination path: %s", destPath)
	}
	
	// Read the source file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", sourcePath, err)
	}
	
	// Ensure the destination path is within the static directory
	staticDir := filepath.Join(cm.projectPath, "static")
	fullDestPath := filepath.Join(staticDir, destPath)
	
	// Validate the destination path is safe
	if !cm.isValidResourcePath(fullDestPath, staticDir) {
		return fmt.Errorf("invalid destination path: %s", destPath)
	}
	
	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(fullDestPath)
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Write the file to the destination
	err = os.WriteFile(fullDestPath, sourceData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write resource file: %w", err)
	}
	
	return nil
}

// GetResourcePath returns the correct path for a resource file
func (cm *ContentManagerService) GetResourcePath(resourceName string) string {
	// Hugo static files are served from the root of the site
	// So a file in static/images/photo.jpg is accessible as /images/photo.jpg
	return "/" + strings.TrimPrefix(resourceName, "/")
}

// ListResources lists all resource files in the static directory
func (cm *ContentManagerService) ListResources() ([]interfaces.ResourceInfo, error) {
	staticDir := filepath.Join(cm.projectPath, "static")
	
	// Check if static directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return []interfaces.ResourceInfo{}, nil
	}
	
	var resources []interfaces.ResourceInfo
	
	err := filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Get relative path from static directory
		relPath, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}
		
		// Determine if it's an image file
		isImage := cm.isImageFile(path)
		
		resource := interfaces.ResourceInfo{
			Path:         relPath,
			Name:         info.Name(),
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
			IsImage:      isImage,
		}
		
		resources = append(resources, resource)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk static directory: %w", err)
	}
	
	return resources, nil
}

// DeleteResource deletes a resource file
func (cm *ContentManagerService) DeleteResource(resourcePath string) error {
	// Validate the path before joining
	if cm.isUnsafePath(resourcePath) {
		return fmt.Errorf("invalid resource path: %s", resourcePath)
	}
	
	staticDir := filepath.Join(cm.projectPath, "static")
	fullPath := filepath.Join(staticDir, resourcePath)
	
	// Validate the path is safe
	if !cm.isValidResourcePath(fullPath, staticDir) {
		return fmt.Errorf("invalid resource path: %s", resourcePath)
	}
	
	err := os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete resource file: %w", err)
	}
	
	return nil
}

// Helper methods

// isUnsafePath checks for various unsafe path patterns
func (cm *ContentManagerService) isUnsafePath(path string) bool {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return true
	}
	
	// Check for absolute paths (both Unix and Windows style)
	if filepath.IsAbs(path) {
		return true
	}
	
	// Check for paths that start with / (Unix absolute)
	if strings.HasPrefix(path, "/") {
		return true
	}
	
	// Check for Windows drive letters
	if len(path) >= 2 && path[1] == ':' {
		return true
	}
	
	// Check for UNC paths
	if strings.HasPrefix(path, "\\\\") {
		return true
	}
	
	return false
}

// isValidResourcePath checks if the resource path is within the static directory
func (cm *ContentManagerService) isValidResourcePath(fullPath, staticDir string) bool {
	// Check for absolute paths that try to escape the static directory
	if filepath.IsAbs(fullPath) && !strings.HasPrefix(fullPath, staticDir) {
		return false
	}
	
	absStaticDir, err := filepath.Abs(staticDir)
	if err != nil {
		return false
	}
	
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false
	}
	
	// Check if the resolved path is within the static directory
	return strings.HasPrefix(absFullPath, absStaticDir)
}

// isImageFile determines if a file is an image based on its extension
func (cm *ContentManagerService) isImageFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".ico"}
	
	for _, imgExt := range imageExtensions {
		if ext == imgExt {
			return true
		}
	}
	
	return false
}

// Helper types for batch operations
type moveInfo struct {
	sourcePath string
	destPath   string
	content    []byte
}

type updateInfo struct {
	path            string
	fullPath        string
	originalContent []byte
	frontMatter     interfaces.FrontMatter
	content         string
}

// Batch operation methods

// BatchDeleteContent deletes multiple content files atomically
func (cm *ContentManagerService) BatchDeleteContent(paths []string) error {
	// Validate all paths first
	var fullPaths []string
	for _, path := range paths {
		if cm.isUnsafePath(path) {
			return fmt.Errorf("invalid path: %s", path)
		}
		
		fullPath := filepath.Join(cm.projectPath, path)
		
		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		
		fullPaths = append(fullPaths, fullPath)
	}
	
	// Create backup information for rollback
	type backupInfo struct {
		path    string
		content []byte
	}
	var backups []backupInfo
	
	// Read all files for backup before deletion
	for i, fullPath := range fullPaths {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read file for backup %s: %w", paths[i], err)
		}
		backups = append(backups, backupInfo{
			path:    fullPath,
			content: content,
		})
	}
	
	// Perform deletions
	var deletedFiles []string
	for i, fullPath := range fullPaths {
		err := os.Remove(fullPath)
		if err != nil {
			// Rollback: restore previously deleted files
			for j, backup := range backups[:len(deletedFiles)] {
				if restoreErr := os.WriteFile(backup.path, backup.content, 0644); restoreErr != nil {
					fmt.Printf("Warning: failed to restore file during rollback %s: %v\n", deletedFiles[j], restoreErr)
				}
			}
			return fmt.Errorf("failed to delete file %s: %w", paths[i], err)
		}
		deletedFiles = append(deletedFiles, fullPath)
	}
	
	return nil
}

// BatchMoveContent moves multiple content files to new locations atomically
func (cm *ContentManagerService) BatchMoveContent(moves []interfaces.ContentMove) error {
	// Validate all moves first
	var moveInfos []moveInfo
	
	for _, move := range moves {
		// Validate paths
		if cm.isUnsafePath(move.SourcePath) || cm.isUnsafePath(move.DestPath) {
			return fmt.Errorf("invalid path in move operation: %s -> %s", move.SourcePath, move.DestPath)
		}
		
		sourcePath := filepath.Join(cm.projectPath, move.SourcePath)
		destPath := filepath.Join(cm.projectPath, move.DestPath)
		
		// Check if source exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %s", move.SourcePath)
		}
		
		// Check if destination already exists
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("destination file already exists: %s", move.DestPath)
		}
		
		// Read source content for backup
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read source file %s: %w", move.SourcePath, err)
		}
		
		moveInfos = append(moveInfos, moveInfo{
			sourcePath: sourcePath,
			destPath:   destPath,
			content:    content,
		})
	}
	
	// Perform moves
	var completedMoves []moveInfo
	for i, moveInfo := range moveInfos {
		// Create destination directory if needed
		destDir := filepath.Dir(moveInfo.destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			// Rollback completed moves
			cm.rollbackMoves(completedMoves)
			return fmt.Errorf("failed to create destination directory for %s: %w", moves[i].DestPath, err)
		}
		
		// Write to destination
		if err := os.WriteFile(moveInfo.destPath, moveInfo.content, 0644); err != nil {
			// Rollback completed moves
			cm.rollbackMoves(completedMoves)
			return fmt.Errorf("failed to write destination file %s: %w", moves[i].DestPath, err)
		}
		
		// Remove source
		if err := os.Remove(moveInfo.sourcePath); err != nil {
			// Remove the destination file we just created and rollback
			os.Remove(moveInfo.destPath)
			cm.rollbackMoves(completedMoves)
			return fmt.Errorf("failed to remove source file %s: %w", moves[i].SourcePath, err)
		}
		
		completedMoves = append(completedMoves, moveInfo)
	}
	
	return nil
}

// BatchUpdateTags updates tags for multiple content files atomically
func (cm *ContentManagerService) BatchUpdateTags(updates []interfaces.TagUpdate) error {
	// Validate all updates and prepare data
	var updateInfos []updateInfo
	
	for _, update := range updates {
		// Validate path
		if cm.isUnsafePath(update.Path) {
			return fmt.Errorf("invalid path: %s", update.Path)
		}
		
		fullPath := filepath.Join(cm.projectPath, update.Path)
		
		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", update.Path)
		}
		
		// Read and parse the file
		originalContent, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", update.Path, err)
		}
		
		frontMatter, content, err := cm.ParseFrontMatter(string(originalContent))
		if err != nil {
			return fmt.Errorf("failed to parse front matter for %s: %w", update.Path, err)
		}
		
		// Update tags
		frontMatter.Tags = update.Tags
		
		updateInfos = append(updateInfos, updateInfo{
			path:            update.Path,
			fullPath:        fullPath,
			originalContent: originalContent,
			frontMatter:     frontMatter,
			content:         content,
		})
	}
	
	// Perform updates
	var updatedFiles []updateInfo
	for _, updateInfo := range updateInfos {
		// Serialize front matter
		frontMatterYAML, err := cm.SerializeFrontMatter(updateInfo.frontMatter)
		if err != nil {
			// Rollback completed updates
			cm.rollbackTagUpdates(updatedFiles)
			return fmt.Errorf("failed to serialize front matter for %s: %w", updateInfo.path, err)
		}
		
		// Combine front matter and content
		fullContent := fmt.Sprintf("---\n%s---\n%s", frontMatterYAML, updateInfo.content)
		
		// Write updated content
		if err := os.WriteFile(updateInfo.fullPath, []byte(fullContent), 0644); err != nil {
			// Rollback completed updates
			cm.rollbackTagUpdates(updatedFiles)
			return fmt.Errorf("failed to write updated file %s: %w", updateInfo.path, err)
		}
		
		updatedFiles = append(updatedFiles, updateInfo)
	}
	
	return nil
}

// Helper methods for rollback operations

// rollbackMoves restores files to their original locations
func (cm *ContentManagerService) rollbackMoves(completedMoves []moveInfo) {
	for _, moveInfo := range completedMoves {
		// Remove destination file
		os.Remove(moveInfo.destPath)
		// Restore source file
		if err := os.WriteFile(moveInfo.sourcePath, moveInfo.content, 0644); err != nil {
			fmt.Printf("Warning: failed to restore file during rollback %s: %v\n", moveInfo.sourcePath, err)
		}
	}
}

// rollbackTagUpdates restores files to their original content
func (cm *ContentManagerService) rollbackTagUpdates(updatedFiles []updateInfo) {
	for _, updateInfo := range updatedFiles {
		if err := os.WriteFile(updateInfo.fullPath, updateInfo.originalContent, 0644); err != nil {
			fmt.Printf("Warning: failed to restore file during rollback %s: %v\n", updateInfo.path, err)
		}
	}
}