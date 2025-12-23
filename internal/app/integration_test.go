package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/test"

	"hugo-visual-client/internal/interfaces"
	"hugo-visual-client/internal/services"
)

// TestEndToEndWorkflow tests the complete end-to-end workflow
// from project creation to content editing and Hugo server management
func TestEndToEndWorkflow(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test application
	testApp := app.New()
	testWindow := testApp.NewWindow("Test Hugo Client")
	defer testWindow.Close()

	// Initialize app controller
	controller, err := NewAppController(testWindow)
	if err != nil {
		t.Fatalf("Failed to create app controller: %v", err)
	}

	// Initialize UI
	controller.InitializeUI()

	// Test 1: Project Creation Workflow
	t.Run("ProjectCreationWorkflow", func(t *testing.T) {
		projectPath := filepath.Join(tempDir, "test-project")
		config := interfaces.ProjectConfig{
			Name:        "Test Project",
			Title:       "Test Hugo Site",
			BaseURL:     "https://test.example.com",
			Theme:       "",
			Language:    "en",
			Description: "A test Hugo site for integration testing",
		}

		// Create project through controller
		project, err := controller.projectManager.CreateProject(projectPath, config)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		// Set as current project
		controller.SetCurrentProject(project)

		// Verify project is loaded correctly
		currentProject := controller.GetCurrentProject()
		if currentProject == nil {
			t.Fatal("Current project should not be nil after setting")
		}

		if currentProject.Name != config.Name {
			t.Errorf("Expected project name %s, got %s", config.Name, currentProject.Name)
		}

		// Verify project structure was created
		requiredDirs := []string{"content", "themes", "static", "layouts", "data", "assets"}
		for _, dir := range requiredDirs {
			dirPath := filepath.Join(projectPath, dir)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				t.Errorf("Required directory %s was not created", dir)
			}
		}

		// Verify config file exists
		configPath := filepath.Join(projectPath, "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}
	})

	// Test 2: Content Management Workflow
	t.Run("ContentManagementWorkflow", func(t *testing.T) {
		if controller.GetCurrentProject() == nil {
			t.Skip("No project loaded, skipping content management test")
		}

		// Create new content through controller
		contentPath := "content/posts/test-post.md"
		
		// Simulate creating new content
		controller.createNewContent("Test Post", contentPath)

		// Verify content was created
		fullPath := filepath.Join(controller.GetCurrentProject().Path, contentPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Error("Content file was not created")
		}

		// Verify content editor was opened
		if _, exists := controller.contentEditors[contentPath]; !exists {
			t.Error("Content editor was not created for new content")
		}

		// Test content saving
		if editor, exists := controller.contentEditors[contentPath]; exists {
			// Modify content
			editor.contentBinding.Set("# Test Post\n\nThis is a test post content.")
			
			// Save content
			err := editor.SaveContent()
			if err != nil {
				t.Errorf("Failed to save content: %v", err)
			}

			// Verify file was updated
			content, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("Failed to read saved content: %v", err)
			}

			if !strings.Contains(string(content), "This is a test post content.") {
				t.Error("Saved content does not contain expected text")
			}
		}
	})

	// Test 3: Hugo Service Integration Workflow
	t.Run("HugoServiceIntegrationWorkflow", func(t *testing.T) {
		if controller.GetCurrentProject() == nil {
			t.Skip("No project loaded, skipping Hugo service test")
		}

		// Check if Hugo is installed (skip if not available)
		installed, _, err := controller.hugoService.IsHugoInstalled()
		if err != nil || !installed {
			t.Skip("Hugo not installed, skipping Hugo service integration test")
		}

		projectPath := controller.GetCurrentProject().Path

		// Test server startup
		err = controller.hugoService.StartServer(projectPath, 1314) // Use different port to avoid conflicts
		if err != nil {
			t.Errorf("Failed to start Hugo server: %v", err)
		} else {
			// Verify server is running
			status := controller.hugoService.GetServerStatus()
			if !status.Running {
				t.Error("Hugo server should be running after startup")
			}

			// Test server stop
			err = controller.hugoService.StopServer()
			if err != nil {
				t.Errorf("Failed to stop Hugo server: %v", err)
			}

			// Verify server is stopped
			status = controller.hugoService.GetServerStatus()
			if status.Running {
				t.Error("Hugo server should be stopped after stop command")
			}
		}

		// Test site building
		outputPath := filepath.Join(projectPath, "public")
		err = controller.hugoService.BuildSite(projectPath, outputPath)
		if err != nil {
			t.Errorf("Failed to build Hugo site: %v", err)
		} else {
			// Verify build output exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("Build output directory was not created")
			}
		}
	})

	// Test 4: Component Integration
	t.Run("ComponentIntegration", func(t *testing.T) {
		// Test project explorer integration
		if controller.projectExplorer != nil && controller.GetCurrentProject() != nil {
			err := controller.projectExplorer.LoadProject(controller.GetCurrentProject())
			if err != nil {
				t.Errorf("Failed to load project in explorer: %v", err)
			}

			// Test file selection callback
			testPath := "content/posts/test-post.md"
			controller.onFileSelected(testPath)

			// Verify status was updated
			// Note: In a real test, we'd check the actual status bar content
		}

		// Test synchronization between components
		controller.SynchronizeProjectState()

		// Test refresh functionality
		controller.RefreshAllComponents()
	})

	// Test 5: Application Lifecycle
	t.Run("ApplicationLifecycle", func(t *testing.T) {
		// Test shutdown process
		err := controller.Shutdown()
		if err != nil {
			t.Errorf("Failed to shutdown application: %v", err)
		}

		// Verify cleanup was performed
		if controller.hugoService != nil {
			status := controller.hugoService.GetServerStatus()
			if status.Running {
				t.Error("Hugo server should be stopped after application shutdown")
			}
		}
	})
}

// TestComponentIntegration tests the integration between different components
func TestComponentIntegration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-component-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Project Manager and Content Manager Integration
	t.Run("ProjectManagerContentManagerIntegration", func(t *testing.T) {
		// Create project manager
		configPath := filepath.Join(tempDir, "config.json")
		pm, err := services.NewProjectManagerService(configPath)
		if err != nil {
			t.Fatalf("Failed to create project manager: %v", err)
		}

		// Create a test project
		projectPath := filepath.Join(tempDir, "integration-project")
		config := interfaces.ProjectConfig{
			Name:        "Integration Test Project",
			Title:       "Integration Test",
			BaseURL:     "https://integration.test",
			Theme:       "",
			Language:    "en",
			Description: "Project for integration testing",
		}

		project, err := pm.CreateProject(projectPath, config)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		// Create content manager for the project
		cm := services.NewContentManager(project.Path)

		// Test content creation through content manager
		frontMatter := interfaces.FrontMatter{
			Title:       "Integration Test Post",
			Date:        time.Now(),
			Draft:       false,
			Tags:        []string{"integration", "test"},
			Categories:  []string{"testing"},
			Description: "A post for integration testing",
		}

		contentPath := "content/posts/integration-test.md"
		content := "# Integration Test Post\n\nThis post tests the integration between components."

		err = cm.CreateContent(contentPath, frontMatter, content)
		if err != nil {
			t.Errorf("Failed to create content: %v", err)
		}

		// Verify content was created in project
		fullPath := filepath.Join(project.Path, contentPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Error("Content file was not created in project directory")
		}

		// Test content listing
		contentItems, err := cm.ListContent(project.Path)
		if err != nil {
			t.Errorf("Failed to list content: %v", err)
		}

		found := false
		for _, item := range contentItems {
			if item.Path == contentPath {
				found = true
				if item.Title != frontMatter.Title {
					t.Errorf("Expected title %s, got %s", frontMatter.Title, item.Title)
				}
				break
			}
		}

		if !found {
			t.Error("Created content was not found in content listing")
		}
	})

	// Test 2: Hugo Service and Content Manager Integration
	t.Run("HugoServiceContentManagerIntegration", func(t *testing.T) {
		// Check if Hugo is installed
		hugoService := services.NewHugoService()
		installed, _, err := hugoService.IsHugoInstalled()
		if err != nil || !installed {
			t.Skip("Hugo not installed, skipping Hugo service integration test")
		}

		// Create a valid Hugo project
		projectPath := filepath.Join(tempDir, "hugo-integration-project")
		err = createValidHugoProject(projectPath)
		if err != nil {
			t.Fatalf("Failed to create Hugo project: %v", err)
		}

		// Create content manager
		cm := services.NewContentManager(projectPath)

		// Create some content
		frontMatter := interfaces.FrontMatter{
			Title:       "Hugo Integration Test",
			Date:        time.Now(),
			Draft:       false,
			Tags:        []string{"hugo", "integration"},
			Categories:  []string{"testing"},
			Description: "Testing Hugo and content manager integration",
		}

		contentPath := "content/posts/hugo-integration.md"
		content := "# Hugo Integration Test\n\nTesting the integration between Hugo service and content manager."

		err = cm.CreateContent(contentPath, frontMatter, content)
		if err != nil {
			t.Errorf("Failed to create content: %v", err)
		}

		// Test building the site with the new content
		outputPath := filepath.Join(projectPath, "public")
		err = hugoService.BuildSite(projectPath, outputPath)
		if err != nil {
			t.Errorf("Failed to build site with new content: %v", err)
		}

		// Verify build output contains our content
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Build output directory was not created")
		}

		// Test file watching integration
		err = hugoService.WatchFiles(projectPath, func(path string) {
			t.Logf("File changed: %s", path)
		})
		if err != nil {
			t.Errorf("Failed to start file watching: %v", err)
		}

		// Modify content and verify watching detects it
		modifiedContent := content + "\n\nThis is additional content for testing file watching."
		err = cm.UpdateContent(contentPath, frontMatter, modifiedContent)
		if err != nil {
			t.Errorf("Failed to update content: %v", err)
		}

		// Give file watcher time to detect changes
		time.Sleep(100 * time.Millisecond)

		// Stop file watching
		err = hugoService.StopWatching()
		if err != nil {
			t.Errorf("Failed to stop file watching: %v", err)
		}
	})

	// Test 3: All Components Integration
	t.Run("AllComponentsIntegration", func(t *testing.T) {
		// Create test application
		testApp := app.New()
		testWindow := testApp.NewWindow("Integration Test")
		defer testWindow.Close()

		// Create app controller (integrates all components)
		controller, err := NewAppController(testWindow)
		if err != nil {
			t.Fatalf("Failed to create app controller: %v", err)
		}

		// Create a project through the controller
		projectPath := filepath.Join(tempDir, "full-integration-project")
		config := interfaces.ProjectConfig{
			Name:        "Full Integration Project",
			Title:       "Full Integration Test",
			BaseURL:     "https://full-integration.test",
			Theme:       "",
			Language:    "en",
			Description: "Project for full integration testing",
		}

		project, err := controller.projectManager.CreateProject(projectPath, config)
		if err != nil {
			t.Fatalf("Failed to create project through controller: %v", err)
		}

		// Set as current project
		controller.SetCurrentProject(project)

		// Test that all components are synchronized
		if controller.GetCurrentProject() != project {
			t.Error("Current project not set correctly")
		}

		if controller.contentManager == nil {
			t.Error("Content manager should be initialized when project is set")
		}

		// Test creating content through the integrated system
		controller.createNewContent("Full Integration Test", "content/posts/full-integration.md")

		// Verify content editor was created
		if len(controller.contentEditors) == 0 {
			t.Error("No content editors were created")
		}

		// Test project explorer integration
		if controller.projectExplorer != nil {
			err = controller.projectExplorer.LoadProject(project)
			if err != nil {
				t.Errorf("Failed to load project in explorer: %v", err)
			}
		}

		// Test component synchronization
		controller.SynchronizeProjectState()

		// Test refresh all components
		controller.RefreshAllComponents()

		// Test shutdown
		err = controller.Shutdown()
		if err != nil {
			t.Errorf("Failed to shutdown integrated system: %v", err)
		}
	})
}

// TestUserInterfaceInteraction tests UI component interactions
func TestUserInterfaceInteraction(t *testing.T) {
	// Create test application
	testApp := app.New()
	testWindow := testApp.NewWindow("UI Test")
	defer testWindow.Close()

	// Test 1: Main Window Creation and Layout
	t.Run("MainWindowCreationAndLayout", func(t *testing.T) {
		mainWindow := NewMainWindow(testWindow)
		if mainWindow == nil {
			t.Fatal("Failed to create main window")
		}

		// Verify main window components
		if mainWindow.toolbar == nil {
			t.Error("Toolbar was not created")
		}

		if mainWindow.contentArea == nil {
			t.Error("Content area was not created")
		}

		if mainWindow.statusBar == nil {
			t.Error("Status bar was not created")
		}

		if mainWindow.mainSplit == nil {
			t.Error("Main split container was not created")
		}

		// Test status bar update
		testMessage := "Test status message"
		mainWindow.UpdateStatusBar(testMessage)
		if mainWindow.statusBar.Text != testMessage {
			t.Errorf("Expected status bar text %s, got %s", testMessage, mainWindow.statusBar.Text)
		}
	})

	// Test 2: Content Tab Management
	t.Run("ContentTabManagement", func(t *testing.T) {
		mainWindow := NewMainWindow(testWindow)

		// Test adding content tab
		testContent := test.NewTappableLabel("Test Content", func() {})
		mainWindow.AddContentTab("Test Tab", testContent)

		// Verify tab was added
		if len(mainWindow.contentArea.Items) < 2 { // Welcome tab + our tab
			t.Error("Content tab was not added")
		}

		// Find our tab
		found := false
		for _, item := range mainWindow.contentArea.Items {
			if item.Text == "Test Tab" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Added tab was not found in content area")
		}

		// Test removing content tab
		mainWindow.RemoveContentTab("Test Tab")

		// Verify tab was removed
		for _, item := range mainWindow.contentArea.Items {
			if item.Text == "Test Tab" {
				t.Error("Tab was not removed from content area")
			}
		}
	})

	// Test 3: App Controller UI Integration
	t.Run("AppControllerUIIntegration", func(t *testing.T) {
		controller, err := NewAppController(testWindow)
		if err != nil {
			t.Fatalf("Failed to create app controller: %v", err)
		}

		// Initialize UI
		controller.InitializeUI()

		// Verify controller has main window
		if controller.mainWindow == nil {
			t.Error("App controller should have main window")
		}

		// Verify project explorer is created
		if controller.projectExplorer == nil {
			t.Error("App controller should have project explorer")
		}

		// Test status notification system
		testMessage := "Test notification"
		controller.notifyStatus(testMessage)

		// Verify status was updated in UI
		if controller.mainWindow.statusBar.Text != testMessage {
			t.Errorf("Expected status %s, got %s", testMessage, controller.mainWindow.statusBar.Text)
		}

		// Test error notification system
		testError := fmt.Errorf("test error")
		controller.notifyError(testError)

		// Verify error was displayed in status bar
		expectedErrorText := "Error: " + testError.Error()
		if controller.mainWindow.statusBar.Text != expectedErrorText {
			t.Errorf("Expected error status %s, got %s", expectedErrorText, controller.mainWindow.statusBar.Text)
		}
	})

	// Test 4: Project Explorer UI Integration
	t.Run("ProjectExplorerUIIntegration", func(t *testing.T) {
		// Create temporary directory for testing
		tempDir, err := os.MkdirTemp("", "hugo-ui-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a test project
		projectPath := filepath.Join(tempDir, "ui-test-project")
		err = createValidHugoProject(projectPath)
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}

		// Create project explorer
		explorer := NewProjectExplorer()
		if explorer == nil {
			t.Fatal("Failed to create project explorer")
		}

		// Create a mock project
		project := &interfaces.Project{
			Path:       projectPath,
			Name:       "UI Test Project",
			ConfigFile: "config.yaml",
			ContentDir: "content",
			ThemeDir:   "themes",
			PublicDir:  "public",
		}

		// Load project in explorer
		err = explorer.LoadProject(project)
		if err != nil {
			t.Errorf("Failed to load project in explorer: %v", err)
		}

		// Verify explorer widget is created
		widget := explorer.GetWidget()
		if widget == nil {
			t.Error("Project explorer widget should not be nil")
		}

		// Test callback setup
		callbackCalled := false
		explorer.SetOnFileSelect(func(path string) {
			callbackCalled = true
		})

		// Simulate file selection (in a real test, we'd interact with the UI)
		explorer.onFileSelect("content/posts/test.md")
		if !callbackCalled {
			t.Error("File select callback was not called")
		}
	})

	// Test 5: Content Editor UI Integration
	t.Run("ContentEditorUIIntegration", func(t *testing.T) {
		// Create temporary directory for testing
		tempDir, err := os.MkdirTemp("", "hugo-editor-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create content manager
		cm := services.NewContentManager(tempDir)

		// Create Hugo service
		hs := services.NewHugoService()

		// Create content editor
		editor := NewContentEditor(cm, hs, tempDir)
		if editor == nil {
			t.Fatal("Failed to create content editor")
		}

		// Verify editor widget is created
		widget := editor.GetWidget()
		if widget == nil {
			t.Error("Content editor widget should not be nil")
		}

		// Test creating new content
		err = editor.CreateNewContent("content/posts/ui-test.md")
		if err != nil {
			t.Errorf("Failed to create new content: %v", err)
		}

		// Verify content was created
		if !editor.IsModified() {
			t.Error("Editor should be marked as modified after creating new content")
		}

		// Test setting title
		testTitle := "UI Test Post"
		editor.titleBinding.Set(testTitle)

		title, _ := editor.titleBinding.Get()
		if title != testTitle {
			t.Errorf("Expected title %s, got %s", testTitle, title)
		}

		// Test callback setup
		saveCalled := false
		editor.SetOnSave(func(path string) {
			saveCalled = true
		})

		// Save content
		err = editor.SaveContent()
		if err != nil {
			t.Errorf("Failed to save content: %v", err)
		}

		if !saveCalled {
			t.Error("Save callback was not called")
		}
	})
}

// Helper function to create a valid Hugo project structure for testing
func createValidHugoProject(projectPath string) error {
	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return err
	}

	// Create required directories
	dirs := []string{"content", "themes", "static", "layouts", "data", "assets", "content/posts"}
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

// TestIntegrationTestSuite runs all integration tests as a suite
func TestIntegrationTestSuite(t *testing.T) {
	// This test ensures all integration tests can run together
	// and don't interfere with each other

	t.Run("EndToEndWorkflow", TestEndToEndWorkflow)
	t.Run("ComponentIntegration", TestComponentIntegration)
	t.Run("UserInterfaceInteraction", TestUserInterfaceInteraction)
}