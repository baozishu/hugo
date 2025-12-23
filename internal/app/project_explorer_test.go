package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectExplorer_GetFileType(t *testing.T) {
	pe := &ProjectExplorer{}
	
	tests := []struct {
		filename string
		expected string
	}{
		{"test.md", "markdown"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.toml", "toml"},
		{"data.json", "json"},
		{"index.html", "html"},
		{"style.css", "css"},
		{"script.js", "javascript"},
		{"main.go", "go"},
		{"image.png", "image"},
		{"unknown.xyz", "file"},
	}
	
	for _, test := range tests {
		result := pe.getFileType(test.filename)
		if result != test.expected {
			t.Errorf("getFileType(%s) = %s, expected %s", test.filename, result, test.expected)
		}
	}
}

func TestProjectExplorer_ShouldIgnoreFile(t *testing.T) {
	pe := &ProjectExplorer{}
	
	tests := []struct {
		filename string
		expected bool
	}{
		{".git", true},
		{".gitignore", true},
		{"node_modules", true},
		{".DS_Store", true},
		{"public", true},
		{"resources", true},
		{"content", false},
		{"config.yaml", false},
		{"index.md", false},
	}
	
	for _, test := range tests {
		result := pe.shouldIgnoreFile(test.filename)
		if result != test.expected {
			t.Errorf("shouldIgnoreFile(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

func TestProjectExplorer_ScanDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create some test files and directories
	contentDir := filepath.Join(tempDir, "content")
	os.MkdirAll(contentDir, 0755)
	
	testFile := filepath.Join(contentDir, "test.md")
	err = os.WriteFile(testFile, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	configFile := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configFile, []byte("title: Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Create project explorer and scan directory
	pe := &ProjectExplorer{
		fileNodes: make(map[string][]string),
		fileTypes: make(map[string]string),
	}
	
	err = pe.scanDirectory(tempDir, "")
	if err != nil {
		t.Errorf("Failed to scan directory: %v", err)
	}
	
	// Check that files were found
	rootChildren, exists := pe.fileNodes[""]
	if !exists {
		t.Error("Root directory not found in file nodes")
	}
	
	if len(rootChildren) == 0 {
		t.Error("No children found in root directory")
	}
	
	// Check that content directory was found
	foundContent := false
	foundConfig := false
	for _, child := range rootChildren {
		if child == "content" {
			foundContent = true
		}
		if child == "config.yaml" {
			foundConfig = true
		}
	}
	
	if !foundContent {
		t.Error("Content directory not found")
	}
	
	if !foundConfig {
		t.Error("Config file not found")
	}
	
	// Check file types
	if pe.fileTypes["content"] != "directory" {
		t.Errorf("Expected content to be directory, got %s", pe.fileTypes["content"])
	}
	
	if pe.fileTypes["config.yaml"] != "yaml" {
		t.Errorf("Expected config.yaml to be yaml, got %s", pe.fileTypes["config.yaml"])
	}
}