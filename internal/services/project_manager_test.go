package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"hugo-visual-client/internal/interfaces"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: hugo-visual-client, Property 2: 项目验证准确性**
// **Validates: Requirements 1.3**
func TestProjectValidationAccuracy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("project validation should accurately identify valid Hugo projects", prop.ForAll(
		func(isValid bool, projectName string) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-validation-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			projectPath := filepath.Join(tempDir, projectName)

			// Create project manager
			pm, err := NewProjectManagerService("")
			if err != nil {
				t.Logf("Failed to create project manager: %v", err)
				return false
			}

			if isValid {
				// Create a valid Hugo project structure
				if err := createValidHugoProject(projectPath); err != nil {
					t.Logf("Failed to create valid Hugo project: %v", err)
					return false
				}

				// Validation should succeed
				err := pm.ValidateProject(projectPath)
				if err != nil {
					t.Logf("Valid project failed validation: %v", err)
					return false
				}
			} else {
				// Create an invalid project structure (missing required components)
				if err := createInvalidHugoProject(projectPath); err != nil {
					t.Logf("Failed to create invalid Hugo project: %v", err)
					return false
				}

				// Validation should fail
				err := pm.ValidateProject(projectPath)
				if err == nil {
					t.Logf("Invalid project passed validation")
					return false
				}
			}

			return true
		},
		gen.Bool(),
		gen.AlphaString().SuchThat(func(s string) bool { 
			return len(s) > 0 && len(s) < 50 
		}),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// createValidHugoProject creates a valid Hugo project structure for testing
func createValidHugoProject(projectPath string) error {
	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return err
	}

	// Create required directories
	dirs := []string{"content", "themes", "static", "layouts", "data", "assets"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return err
		}
	}

	// Create config file
	configContent := `baseURL: "http://localhost:1313"
title: "Test Site"
languageCode: "en"
theme: ""
description: "A test Hugo site"
`
	configPath := filepath.Join(projectPath, "config.yaml")
	return os.WriteFile(configPath, []byte(configContent), 0644)
}

// createInvalidHugoProject creates an invalid Hugo project structure for testing
func createInvalidHugoProject(projectPath string) error {
	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return err
	}

	// Create some directories but not all required ones (missing content directory)
	dirs := []string{"themes", "static"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return err
		}
	}

	// Don't create config file or create an invalid one
	// This makes it an invalid Hugo project
	return nil
}

// Test the actual ProjectManagerService implementation
func TestProjectManagerServiceIntegration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-pm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project manager
	configPath := filepath.Join(tempDir, "config.json")
	pm, err := NewProjectManagerService(configPath)
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}

	// Test creating a new project
	projectPath := filepath.Join(tempDir, "test-project")
	config := interfaces.ProjectConfig{
		Name:        "test-project",
		Title:       "Test Project",
		BaseURL:     "http://localhost:1313",
		Theme:       "",
		Language:    "en",
		Description: "A test project",
	}

	project, err := pm.CreateProject(projectPath, config)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify project was created correctly
	if project.Name != config.Name {
		t.Errorf("Expected project name %s, got %s", config.Name, project.Name)
	}

	// Test opening the created project
	openedProject, err := pm.OpenProject(projectPath)
	if err != nil {
		t.Fatalf("Failed to open project: %v", err)
	}

	if openedProject.Name != project.Name {
		t.Errorf("Expected opened project name %s, got %s", project.Name, openedProject.Name)
	}

	// Test validation of the created project
	err = pm.ValidateProject(projectPath)
	if err != nil {
		t.Errorf("Created project failed validation: %v", err)
	}

	// Test recent projects functionality
	recentProjects := pm.GetRecentProjects()
	if len(recentProjects) == 0 {
		t.Error("Expected at least one recent project")
	}

	// Test saving project
	err = pm.SaveProject(project)
	if err != nil {
		t.Errorf("Failed to save project: %v", err)
	}
}

// Test edge cases for project validation
func TestProjectValidationEdgeCases(t *testing.T) {
	pm, err := NewProjectManagerService("")
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}

	// Test empty path
	err = pm.ValidateProject("")
	if err == nil {
		t.Error("Expected validation to fail for empty path")
	}

	// Test non-existent path
	err = pm.ValidateProject("/non/existent/path")
	if err == nil {
		t.Error("Expected validation to fail for non-existent path")
	}

	// Test file instead of directory
	tempFile, err := os.CreateTemp("", "test-file-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = pm.ValidateProject(tempFile.Name())
	if err == nil {
		t.Error("Expected validation to fail for file path")
	}
}

// Test recent projects management functionality
func TestRecentProjectsManagement(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-recent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project manager with custom config path
	configPath := filepath.Join(tempDir, "config.json")
	pm, err := NewProjectManagerService(configPath)
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}

	// Initially should have no recent projects
	recentProjects := pm.GetRecentProjects()
	if len(recentProjects) != 0 {
		t.Errorf("Expected 0 recent projects, got %d", len(recentProjects))
	}

	// Create some test project directories
	project1Path := filepath.Join(tempDir, "project1")
	project2Path := filepath.Join(tempDir, "project2")
	project3Path := filepath.Join(tempDir, "project3")

	// Create valid Hugo projects
	for _, path := range []string{project1Path, project2Path, project3Path} {
		if err := createValidHugoProject(path); err != nil {
			t.Fatalf("Failed to create test project at %s: %v", path, err)
		}
	}

	// Add projects to recent list
	err = pm.AddRecentProject(project1Path)
	if err != nil {
		t.Fatalf("Failed to add recent project: %v", err)
	}

	err = pm.AddRecentProject(project2Path)
	if err != nil {
		t.Fatalf("Failed to add recent project: %v", err)
	}

	err = pm.AddRecentProject(project3Path)
	if err != nil {
		t.Fatalf("Failed to add recent project: %v", err)
	}

	// Check recent projects count
	if pm.GetRecentProjectsCount() != 3 {
		t.Errorf("Expected 3 recent projects, got %d", pm.GetRecentProjectsCount())
	}

	// Get recent projects and verify order (most recent first)
	recentProjects = pm.GetRecentProjects()
	if len(recentProjects) != 3 {
		t.Errorf("Expected 3 recent projects, got %d", len(recentProjects))
	}

	// Most recent should be project3
	expectedPath3, _ := filepath.Abs(project3Path)
	if recentProjects[0] != expectedPath3 {
		t.Errorf("Expected first recent project to be %s, got %s", expectedPath3, recentProjects[0])
	}

	// Add project1 again - it should move to front
	err = pm.AddRecentProject(project1Path)
	if err != nil {
		t.Fatalf("Failed to re-add recent project: %v", err)
	}

	recentProjects = pm.GetRecentProjects()
	expectedPath1, _ := filepath.Abs(project1Path)
	if recentProjects[0] != expectedPath1 {
		t.Errorf("Expected first recent project to be %s after re-adding, got %s", expectedPath1, recentProjects[0])
	}

	// Should still have 3 projects total
	if len(recentProjects) != 3 {
		t.Errorf("Expected 3 recent projects after re-adding, got %d", len(recentProjects))
	}

	// Test removing a project
	err = pm.RemoveRecentProject(project2Path)
	if err != nil {
		t.Fatalf("Failed to remove recent project: %v", err)
	}

	recentProjects = pm.GetRecentProjects()
	if len(recentProjects) != 2 {
		t.Errorf("Expected 2 recent projects after removal, got %d", len(recentProjects))
	}

	// Verify project2 is not in the list
	expectedPath2, _ := filepath.Abs(project2Path)
	for _, path := range recentProjects {
		if path == expectedPath2 {
			t.Error("Project2 should have been removed from recent projects")
		}
	}

	// Test clearing all recent projects
	err = pm.ClearRecentProjects()
	if err != nil {
		t.Fatalf("Failed to clear recent projects: %v", err)
	}

	recentProjects = pm.GetRecentProjects()
	if len(recentProjects) != 0 {
		t.Errorf("Expected 0 recent projects after clearing, got %d", len(recentProjects))
	}

	if pm.GetRecentProjectsCount() != 0 {
		t.Errorf("Expected 0 recent projects count after clearing, got %d", pm.GetRecentProjectsCount())
	}
}

// Test recent projects limit (max 10 projects)
func TestRecentProjectsLimit(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-limit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project manager
	configPath := filepath.Join(tempDir, "config.json")
	pm, err := NewProjectManagerService(configPath)
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}

	// Create 12 test projects (more than the limit of 10)
	var projectPaths []string
	for i := 0; i < 12; i++ {
		projectPath := filepath.Join(tempDir, fmt.Sprintf("project%d", i))
		if err := createValidHugoProject(projectPath); err != nil {
			t.Fatalf("Failed to create test project %d: %v", i, err)
		}
		projectPaths = append(projectPaths, projectPath)
	}

	// Add all projects to recent list
	for _, path := range projectPaths {
		err = pm.AddRecentProject(path)
		if err != nil {
			t.Fatalf("Failed to add recent project %s: %v", path, err)
		}
	}

	// Should only keep the last 10 projects
	recentProjects := pm.GetRecentProjects()
	if len(recentProjects) != 10 {
		t.Errorf("Expected 10 recent projects (limit), got %d", len(recentProjects))
	}

	// The most recent should be the last project added
	expectedLastPath, _ := filepath.Abs(projectPaths[11])
	if recentProjects[0] != expectedLastPath {
		t.Errorf("Expected first recent project to be %s, got %s", expectedLastPath, recentProjects[0])
	}

	// The first two projects should not be in the list (they were pushed out)
	expectedFirstPath, _ := filepath.Abs(projectPaths[0])
	expectedSecondPath, _ := filepath.Abs(projectPaths[1])
	
	for _, path := range recentProjects {
		if path == expectedFirstPath || path == expectedSecondPath {
			t.Error("First two projects should have been removed due to limit")
		}
	}
}

// Test recent projects with invalid/deleted projects
func TestRecentProjectsWithInvalidProjects(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-invalid-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create project manager
	configPath := filepath.Join(tempDir, "config.json")
	pm, err := NewProjectManagerService(configPath)
	if err != nil {
		t.Fatalf("Failed to create project manager: %v", err)
	}

	// Create valid project
	validProjectPath := filepath.Join(tempDir, "valid-project")
	if err := createValidHugoProject(validProjectPath); err != nil {
		t.Fatalf("Failed to create valid project: %v", err)
	}

	// Create invalid project (will be deleted)
	invalidProjectPath := filepath.Join(tempDir, "invalid-project")
	if err := createValidHugoProject(invalidProjectPath); err != nil {
		t.Fatalf("Failed to create project to be deleted: %v", err)
	}

	// Add both projects to recent list
	err = pm.AddRecentProject(validProjectPath)
	if err != nil {
		t.Fatalf("Failed to add valid project: %v", err)
	}

	err = pm.AddRecentProject(invalidProjectPath)
	if err != nil {
		t.Fatalf("Failed to add project to be deleted: %v", err)
	}

	// Should have 2 projects
	if pm.GetRecentProjectsCount() != 2 {
		t.Errorf("Expected 2 recent projects, got %d", pm.GetRecentProjectsCount())
	}

	// Delete the invalid project
	os.RemoveAll(invalidProjectPath)

	// GetRecentProjects should filter out the deleted project
	recentProjects := pm.GetRecentProjects()
	if len(recentProjects) != 1 {
		t.Errorf("Expected 1 valid recent project after filtering, got %d", len(recentProjects))
	}

	// The remaining project should be the valid one
	expectedValidPath, _ := filepath.Abs(validProjectPath)
	if recentProjects[0] != expectedValidPath {
		t.Errorf("Expected remaining project to be %s, got %s", expectedValidPath, recentProjects[0])
	}
}