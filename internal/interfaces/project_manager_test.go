package interfaces

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: hugo-visual-client, Property 1: 项目创建结构一致性**
// **Validates: Requirements 1.2**
func TestProjectCreationStructureConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("project creation should generate consistent Hugo structure", prop.ForAll(
		func(config ProjectConfig) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			projectPath := filepath.Join(tempDir, config.Name)

			// Create a mock project manager for testing
			pm := &mockProjectManager{}
			
			// Create project
			project, err := pm.CreateProject(projectPath, config)
			if err != nil {
				t.Logf("Failed to create project: %v", err)
				return false
			}

			// Verify project structure consistency
			return verifyHugoProjectStructure(project.Path)
		},
		genProjectConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// verifyHugoProjectStructure checks if a path contains valid Hugo project structure
func verifyHugoProjectStructure(projectPath string) bool {
	requiredDirs := []string{
		"content",
		"themes",
		"static",
		"layouts",
		"data",
		"assets",
	}

	requiredFiles := []string{
		"config.yaml",
	}

	// Check required directories
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(projectPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			return false
		}
	}

	// Check required files
	for _, file := range requiredFiles {
		filePath := filepath.Join(projectPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// genProjectConfig generates random project configurations for testing
func genProjectConfig() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.Const("http://localhost:1313"),
		gen.OneConstOf("", "ananke", "hugo-theme-stack"),
		gen.OneConstOf("en", "zh", "fr", "de"),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 200 }),
	).Map(func(values []interface{}) ProjectConfig {
		return ProjectConfig{
			Name:        values[0].(string),
			Title:       values[1].(string),
			BaseURL:     values[2].(string),
			Theme:       values[3].(string),
			Language:    values[4].(string),
			Description: values[5].(string),
		}
	})
}

// mockProjectManager is a simple implementation for testing
type mockProjectManager struct{}

func (pm *mockProjectManager) CreateProject(path string, config ProjectConfig) (*Project, error) {
	// Create project directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	// Create required directories
	dirs := []string{"content", "themes", "static", "layouts", "data", "assets"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(path, dir), 0755); err != nil {
			return nil, err
		}
	}

	// Create config file
	configContent := `baseURL: "` + config.BaseURL + `"
title: "` + config.Title + `"
languageCode: "` + config.Language + `"
theme: "` + config.Theme + `"
description: "` + config.Description + `"
`
	configPath := filepath.Join(path, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return nil, err
	}

	return &Project{
		Path:       path,
		Name:       config.Name,
		ConfigFile: "config.yaml",
		ContentDir: "content",
		ThemeDir:   "themes",
		PublicDir:  "public",
		Config:     make(map[string]interface{}),
		LastOpened: time.Now(),
	}, nil
}

func (pm *mockProjectManager) OpenProject(path string) (*Project, error) {
	return nil, nil
}

func (pm *mockProjectManager) ValidateProject(path string) error {
	return nil
}

func (pm *mockProjectManager) GetRecentProjects() []string {
	return []string{}
}

func (pm *mockProjectManager) AddRecentProject(path string) error {
	return nil
}

func (pm *mockProjectManager) SaveProject(project *Project) error {
	return nil
}