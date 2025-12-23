package repository

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileRepository handles file system operations
type FileRepository struct {
	basePath string
}

// NewFileRepository creates a new file repository
func NewFileRepository(basePath string) *FileRepository {
	return &FileRepository{
		basePath: basePath,
	}
}

// ReadFile reads the content of a file
func (fr *FileRepository) ReadFile(relativePath string) ([]byte, error) {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	// Validate path is within base directory
	if !fr.isValidPath(fullPath) {
		return nil, fmt.Errorf("invalid path: %s", relativePath)
	}
	
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", relativePath, err)
	}
	
	return data, nil
}

// WriteFile writes content to a file
func (fr *FileRepository) WriteFile(relativePath string, data []byte) error {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	// Validate path is within base directory
	if !fr.isValidPath(fullPath) {
		return fmt.Errorf("invalid path: %s", relativePath)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	err := os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", relativePath, err)
	}
	
	return nil
}

// CreateFile creates a new file with the given content
func (fr *FileRepository) CreateFile(relativePath string, data []byte) error {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	// Validate path is within base directory
	if !fr.isValidPath(fullPath) {
		return fmt.Errorf("invalid path: %s", relativePath)
	}
	
	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("file already exists: %s", relativePath)
	}
	
	return fr.WriteFile(relativePath, data)
}

// DeleteFile deletes a file
func (fr *FileRepository) DeleteFile(relativePath string) error {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	// Validate path is within base directory
	if !fr.isValidPath(fullPath) {
		return fmt.Errorf("invalid path: %s", relativePath)
	}
	
	err := os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", relativePath, err)
	}
	
	return nil
}

// FileExists checks if a file exists
func (fr *FileRepository) FileExists(relativePath string) bool {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	if !fr.isValidPath(fullPath) {
		return false
	}
	
	_, err := os.Stat(fullPath)
	return err == nil
}

// ListFiles lists all files in a directory
func (fr *FileRepository) ListFiles(relativePath string) ([]FileInfo, error) {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	// Validate path is within base directory
	if !fr.isValidPath(fullPath) {
		return nil, fmt.Errorf("invalid path: %s", relativePath)
	}
	
	var files []FileInfo
	
	err := filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Get relative path from base
		relPath, err := filepath.Rel(fr.basePath, path)
		if err != nil {
			return err
		}
		
		info, err := d.Info()
		if err != nil {
			return err
		}
		
		files = append(files, FileInfo{
			Path:         relPath,
			Name:         d.Name(),
			IsDirectory:  d.IsDir(),
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
		})
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %w", relativePath, err)
	}
	
	return files, nil
}

// CreateDirectory creates a directory
func (fr *FileRepository) CreateDirectory(relativePath string) error {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	// Validate path is within base directory
	if !fr.isValidPath(fullPath) {
		return fmt.Errorf("invalid path: %s", relativePath)
	}
	
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", relativePath, err)
	}
	
	return nil
}

// DeleteDirectory deletes a directory and all its contents
func (fr *FileRepository) DeleteDirectory(relativePath string) error {
	fullPath := filepath.Join(fr.basePath, relativePath)
	
	// Validate path is within base directory
	if !fr.isValidPath(fullPath) {
		return fmt.Errorf("invalid path: %s", relativePath)
	}
	
	err := os.RemoveAll(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete directory %s: %w", relativePath, err)
	}
	
	return nil
}

// MoveFile moves a file from source to destination
func (fr *FileRepository) MoveFile(sourcePath, destPath string) error {
	fullSourcePath := filepath.Join(fr.basePath, sourcePath)
	fullDestPath := filepath.Join(fr.basePath, destPath)
	
	// Validate both paths are within base directory
	if !fr.isValidPath(fullSourcePath) || !fr.isValidPath(fullDestPath) {
		return fmt.Errorf("invalid path: source=%s, dest=%s", sourcePath, destPath)
	}
	
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(fullDestPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	err := os.Rename(fullSourcePath, fullDestPath)
	if err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", sourcePath, destPath, err)
	}
	
	return nil
}

// isValidPath checks if the path is within the base directory (prevents path traversal)
func (fr *FileRepository) isValidPath(fullPath string) bool {
	absBasePath, err := filepath.Abs(fr.basePath)
	if err != nil {
		return false
	}
	
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false
	}
	
	return strings.HasPrefix(absFullPath, absBasePath)
}

// FileInfo represents information about a file
type FileInfo struct {
	Path         string
	Name         string
	IsDirectory  bool
	Size         int64
	ModifiedTime time.Time
}