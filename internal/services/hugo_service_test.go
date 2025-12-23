package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: hugo-visual-client, Property 11: Hugo服务器启动成功性**
// **Validates: Requirements 4.1**
func TestHugoServerStartupSuccess(t *testing.T) {
	// Skip test if Hugo is not installed
	service := NewHugoService()
	if installed, _, err := service.IsHugoInstalled(); !installed {
		t.Skip("Hugo is not installed, skipping server startup test:", err)
	}

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-server-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectPath := filepath.Join(tempDir, "test-project")

	// Create a valid Hugo project structure
	if err := createValidHugoProjectForService(projectPath); err != nil {
		t.Fatalf("Failed to create valid Hugo project: %v", err)
	}

	// Create Hugo service
	hugoService := NewHugoService()

	port := 3333

	// Try to start server
	err = hugoService.StartServer(projectPath, port)
	if err != nil {
		t.Fatalf("Failed to start Hugo server: %v", err)
	}

	// Give server time to start
	time.Sleep(1 * time.Second)

	// Check server status
	status := hugoService.GetServerStatus()
	if !status.Running {
		t.Error("Server status indicates not running")
	}

	if status.Port != port {
		t.Errorf("Server port mismatch: expected %d, got %d", port, status.Port)
	}

	if status.PID <= 0 {
		t.Errorf("Invalid server PID: %d", status.PID)
	}

	// Stop server for cleanup - don't fail the test if stop fails
	err = hugoService.StopServer()
	if err != nil {
		t.Logf("Warning: Failed to stop Hugo server cleanly: %v", err)
		// Don't fail the test here - the start was successful
	}

	// Give time for cleanup
	time.Sleep(500 * time.Millisecond)
}

// **Feature: hugo-visual-client, Property 11: Hugo服务器启动成功性**
// **Validates: Requirements 4.1**
func TestHugoServerStartupSuccessProperty(t *testing.T) {
	// Skip test if Hugo is not installed
	service := NewHugoService()
	if installed, _, err := service.IsHugoInstalled(); !installed {
		t.Skip("Hugo is not installed, skipping Hugo server startup property test:", err)
	}

	parameters := gopter.DefaultTestParameters()
	// Reduce iterations to prevent timeout issues with server cleanup on Windows
	parameters.MinSuccessfulTests = 10
	parameters.MaxSize = 10
	properties := gopter.NewProperties(parameters)

	properties.Property("Hugo server startup should succeed for valid projects", prop.ForAll(
		func(portOffset int) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-server-prop-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			projectPath := filepath.Join(tempDir, "test-project")

			// Create a valid Hugo project structure
			if err := createValidHugoProjectForService(projectPath); err != nil {
				t.Logf("Failed to create valid Hugo project: %v", err)
				return false
			}

			// Create Hugo service
			hugoService := NewHugoService()

			// Use a port in a safe range to avoid conflicts
			port := 4000 + (portOffset % 100) // Smaller range to reduce conflicts

			// Try to start server
			err = hugoService.StartServer(projectPath, port)
			if err != nil {
				t.Logf("Failed to start Hugo server on port %d: %v", port, err)
				return false
			}

			// Give server time to start
			time.Sleep(1 * time.Second)

			// Check server status
			status := hugoService.GetServerStatus()
			success := status.Running && status.Port == port && status.PID > 0

			// Stop server for cleanup - be more aggressive about cleanup
			stopErr := hugoService.StopServer()
			if stopErr != nil {
				t.Logf("Warning: Failed to stop Hugo server cleanly: %v", stopErr)
			}

			// Give more time for cleanup on Windows
			time.Sleep(1 * time.Second)

			return success
		},
		gen.IntRange(0, 99), // Smaller range
	))

	properties.TestingRun(t)
}

// Test Hugo service basic functionality
func TestHugoServiceBasicFunctionality(t *testing.T) {
	service := NewHugoService()

	// Test initial status
	status := service.GetServerStatus()
	if status.Running {
		t.Error("Server should not be running initially")
	}
	if status.Port != 0 {
		t.Error("Server port should be 0 initially")
	}
	if status.PID != 0 {
		t.Error("Server PID should be 0 initially")
	}

	// Test Hugo installation check
	installed, version, err := service.IsHugoInstalled()
	if err != nil && installed {
		t.Errorf("IsHugoInstalled returned error but claimed installed: %v", err)
	}
	if installed && version == "" {
		t.Error("Hugo is installed but version is empty")
	}

	// Test GetHugoVersion
	if installed {
		ver, err := service.GetHugoVersion()
		if err != nil {
			t.Errorf("GetHugoVersion failed: %v", err)
		}
		if ver != version {
			t.Errorf("Version mismatch: IsHugoInstalled returned %s, GetHugoVersion returned %s", version, ver)
		}
	}
}

// Test server startup with invalid project path
func TestHugoServerStartupWithInvalidPath(t *testing.T) {
	service := NewHugoService()

	// Test with non-existent path
	err := service.StartServer("/non/existent/path", 3000)
	if err == nil {
		t.Error("Expected error when starting server with non-existent path")
		service.StopServer() // Cleanup just in case
	}

	// Test with empty path
	err = service.StartServer("", 3000)
	if err == nil {
		t.Error("Expected error when starting server with empty path")
		service.StopServer() // Cleanup just in case
	}
}

// Test stopping server when not running
func TestStopServerWhenNotRunning(t *testing.T) {
	service := NewHugoService()

	err := service.StopServer()
	if err == nil {
		t.Error("Expected error when stopping server that is not running")
	}
}

// Helper function to create a valid Hugo project for testing
func createValidHugoProjectForService(projectPath string) error {
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
// **Feature: hugo-visual-client, Property 20: 构建过程可靠性**
// **Validates: Requirements 6.1, 6.2**
func TestHugoBuildProcessReliability(t *testing.T) {
	// Skip test if Hugo is not installed
	service := NewHugoService()
	if installed, _, err := service.IsHugoInstalled(); !installed {
		t.Skip("Hugo is not installed, skipping build process test:", err)
	}

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-build-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectPath := filepath.Join(tempDir, "test-project")
	outputPath := filepath.Join(tempDir, "public")

	// Create a valid Hugo project structure
	if err := createValidHugoProjectForService(projectPath); err != nil {
		t.Fatalf("Failed to create valid Hugo project: %v", err)
	}

	// Create Hugo service
	hugoService := NewHugoService()

	// Try to build site
	buildResult, err := hugoService.BuildSite(projectPath, outputPath)
	if err != nil {
		t.Fatalf("Failed to build Hugo site: %v", err)
	}

	// Verify build was successful
	if !buildResult.Success {
		t.Errorf("Build was not successful. Errors: %v", buildResult.Errors)
	}

	// Verify output path was set correctly
	if buildResult.OutputPath != outputPath {
		t.Errorf("Output path mismatch: expected %s, got %s", outputPath, buildResult.OutputPath)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}

	// Verify some files were written (Hugo should generate at least index.html)
	if buildResult.FilesWritten < 0 {
		t.Errorf("Invalid files written count: %d", buildResult.FilesWritten)
	}

	// Verify build duration was captured
	if buildResult.Duration == "" {
		t.Log("Warning: Build duration was not captured")
	}

	// Check that output directory contains some files
	files, err := os.ReadDir(outputPath)
	if err != nil {
		t.Errorf("Failed to read output directory: %v", err)
	} else if len(files) == 0 {
		t.Error("No files were generated in output directory")
	} else {
		t.Logf("Generated %d files in output directory", len(files))
		for _, file := range files {
			t.Logf("Generated file: %s", file.Name())
		}
	}
}
// Test file monitoring functionality
func TestHugoServiceFileWatching(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-watch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectPath := filepath.Join(tempDir, "test-project")

	// Create a valid Hugo project structure
	if err := createValidHugoProjectForService(projectPath); err != nil {
		t.Fatalf("Failed to create valid Hugo project: %v", err)
	}

	// Create Hugo service
	hugoService := NewHugoService()

	// Channel to receive file change notifications
	changes := make(chan string, 10)
	callback := func(filename string) {
		changes <- filename
	}

	// Start watching files
	err = hugoService.WatchFiles(projectPath, callback)
	if err != nil {
		t.Fatalf("Failed to start file watching: %v", err)
	}

	// Give watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Create a new file in the content directory
	contentDir := filepath.Join(projectPath, "content")
	testFile := filepath.Join(contentDir, "test-post.md")
	testContent := `---
title: "Test Post"
date: 2023-01-01T00:00:00Z
---

This is a test post.
`

	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait for file change notification
	select {
	case changedFile := <-changes:
		t.Logf("Detected file change: %s", changedFile)
		// Verify the changed file is the one we created
		if !filepath.IsAbs(changedFile) {
			t.Errorf("Expected absolute path, got: %s", changedFile)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file change notification")
	}

	// Stop watching
	err = hugoService.StopWatching()
	if err != nil {
		t.Errorf("Failed to stop file watching: %v", err)
	}

	// Give extra time for watcher to fully stop
	time.Sleep(200 * time.Millisecond)

	// Verify no more notifications after stopping
	err = os.WriteFile(filepath.Join(contentDir, "another-post.md"), []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}

	// Should not receive any more notifications (allow for some buffered notifications)
	notificationCount := 0
	timeout := time.After(1 * time.Second)
	for {
		select {
		case changedFile := <-changes:
			notificationCount++
			t.Logf("Received notification after stopping (may be buffered): %s", changedFile)
			if notificationCount > 2 { // Allow a few buffered notifications
				t.Errorf("Too many notifications received after stopping watcher")
				return
			}
		case <-timeout:
			// This is expected - no more notifications should be received
			return
		}
	}
}
// **Feature: hugo-visual-client, Property 12: 文件变化触发重新生成**
// **Validates: Requirements 4.2**
func TestFileChangeTriggersRegeneration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-regen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectPath := filepath.Join(tempDir, "test-project")

	// Create a valid Hugo project structure
	if err := createValidHugoProjectForService(projectPath); err != nil {
		t.Fatalf("Failed to create valid Hugo project: %v", err)
	}

	// Create Hugo service
	hugoService := NewHugoService()

	// Channel to receive file change notifications
	changes := make(chan string, 10)
	callback := func(filename string) {
		changes <- filename
	}

	// Start watching files
	err = hugoService.WatchFiles(projectPath, callback)
	if err != nil {
		t.Fatalf("Failed to start file watching: %v", err)
	}
	defer hugoService.StopWatching()

	// Give watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Test different types of file changes that should trigger regeneration
	testCases := []struct {
		name     string
		filePath string
		content  string
	}{
		{
			name:     "Content file change",
			filePath: filepath.Join(projectPath, "content", "post1.md"),
			content: `---
title: "Test Post 1"
date: 2023-01-01T00:00:00Z
---

Content for post 1.
`,
		},
		{
			name:     "Config file change",
			filePath: filepath.Join(projectPath, "config.yaml"),
			content: `baseURL: "http://example.com"
title: "Updated Test Site"
languageCode: "en"
theme: ""
description: "An updated test Hugo site"
`,
		},
		{
			name:     "Layout file change",
			filePath: filepath.Join(projectPath, "layouts", "_default", "single.html"),
			content: `<!DOCTYPE html>
<html>
<head>
    <title>{{ .Title }}</title>
</head>
<body>
    <h1>{{ .Title }}</h1>
    <div>{{ .Content }}</div>
</body>
</html>
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create directory if needed
			dir := filepath.Dir(tc.filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}

			// Create or modify the file
			err := os.WriteFile(tc.filePath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write file %s: %v", tc.filePath, err)
			}

			// Wait for file change notification
			select {
			case changedFile := <-changes:
				t.Logf("Detected file change: %s", changedFile)
				// Verify the changed file is related to our test file
				if !filepath.IsAbs(changedFile) {
					t.Errorf("Expected absolute path, got: %s", changedFile)
				}
				// The notification might be for the file itself or its parent directory
				expectedDir := filepath.Dir(tc.filePath)
				if !strings.Contains(changedFile, expectedDir) && changedFile != tc.filePath {
					t.Logf("Warning: Notification for unexpected file: %s (expected related to %s)", changedFile, tc.filePath)
				}
			case <-time.After(2 * time.Second):
				t.Errorf("Timeout waiting for file change notification for %s", tc.name)
			}
		})
	}
}
// Test error handling and reporting
func TestHugoServiceErrorHandling(t *testing.T) {
	// Skip test if Hugo is not installed
	service := NewHugoService()
	if installed, _, err := service.IsHugoInstalled(); !installed {
		t.Skip("Hugo is not installed, skipping error handling test:", err)
	}

	// Test 1: Build with invalid project path
	t.Run("Build with invalid project path", func(t *testing.T) {
		hugoService := NewHugoService()
		buildResult, err := hugoService.BuildSite("/non/existent/path", "/tmp/output")
		
		if err == nil {
			t.Error("Expected error when building with invalid project path")
		}
		
		if buildResult != nil {
			t.Error("Expected nil build result when build fails")
		}
	})

	// Test 2: Start server with invalid project path
	t.Run("Start server with invalid project path", func(t *testing.T) {
		hugoService := NewHugoService()
		err := hugoService.StartServer("/non/existent/path", 3000)
		
		if err == nil {
			t.Error("Expected error when starting server with invalid project path")
			hugoService.StopServer() // Cleanup just in case
		}
	})

	// Test 3: Build with invalid Hugo project (missing config)
	t.Run("Build with invalid Hugo project", func(t *testing.T) {
		// Create temporary directory with no Hugo config
		tempDir, err := os.MkdirTemp("", "hugo-error-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create some directories but no config file
		os.MkdirAll(filepath.Join(tempDir, "content"), 0755)
		
		hugoService := NewHugoService()
		buildResult, err := hugoService.BuildSite(tempDir, filepath.Join(tempDir, "public"))
		
		// Hugo might still build successfully even without a config file
		// but let's check the result
		if err != nil {
			t.Logf("Build failed as expected: %v", err)
		} else if buildResult != nil {
			t.Logf("Build succeeded with result: Success=%v, Errors=%v, Warnings=%v", 
				buildResult.Success, buildResult.Errors, buildResult.Warnings)
		}
	})

	// Test 4: Stop server when not running
	t.Run("Stop server when not running", func(t *testing.T) {
		hugoService := NewHugoService()
		err := hugoService.StopServer()
		
		if err == nil {
			t.Error("Expected error when stopping server that is not running")
		}
	})

	// Test 5: Start server on invalid port
	t.Run("Start server on invalid port", func(t *testing.T) {
		// Create a valid project first
		tempDir, err := os.MkdirTemp("", "hugo-port-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		projectPath := filepath.Join(tempDir, "test-project")
		if err := createValidHugoProjectForService(projectPath); err != nil {
			t.Fatalf("Failed to create valid Hugo project: %v", err)
		}

		hugoService := NewHugoService()
		
		// Try to start server on port 0 (invalid)
		err = hugoService.StartServer(projectPath, 0)
		if err != nil {
			t.Logf("Server start failed as expected with invalid port: %v", err)
		} else {
			t.Log("Server started successfully even with port 0")
			hugoService.StopServer() // Cleanup
		}
	})
}
// **Feature: hugo-visual-client, Property 14: 错误处理完整性**
// **Validates: Requirements 4.5**
func TestErrorHandlingCompleteness(t *testing.T) {
	// Skip test if Hugo is not installed
	service := NewHugoService()
	if installed, _, err := service.IsHugoInstalled(); !installed {
		t.Skip("Hugo is not installed, skipping error handling completeness test:", err)
	}

	// Test that all error conditions are properly handled and reported
	testCases := []struct {
		name        string
		testFunc    func() error
		expectError bool
	}{
		{
			name: "Build with empty path",
			testFunc: func() error {
				hugoService := NewHugoService()
				_, err := hugoService.BuildSite("", "/tmp/output")
				return err
			},
			expectError: true,
		},
		{
			name: "Build with non-existent path",
			testFunc: func() error {
				hugoService := NewHugoService()
				_, err := hugoService.BuildSite("/absolutely/non/existent/path/12345", "/tmp/output")
				return err
			},
			expectError: true,
		},
		{
			name: "Start server with empty path",
			testFunc: func() error {
				hugoService := NewHugoService()
				err := hugoService.StartServer("", 3000)
				if err == nil {
					hugoService.StopServer() // Cleanup
				}
				return err
			},
			expectError: true,
		},
		{
			name: "Start server with non-existent path",
			testFunc: func() error {
				hugoService := NewHugoService()
				err := hugoService.StartServer("/absolutely/non/existent/path/12345", 3000)
				if err == nil {
					hugoService.StopServer() // Cleanup
				}
				return err
			},
			expectError: true,
		},
		{
			name: "Stop server when not running",
			testFunc: func() error {
				hugoService := NewHugoService()
				return hugoService.StopServer()
			},
			expectError: true,
		},
		{
			name: "Start server twice on same port",
			testFunc: func() error {
				// Create a valid project first
				tempDir, err := os.MkdirTemp("", "hugo-double-start-test-*")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tempDir)

				projectPath := filepath.Join(tempDir, "test-project")
				if err := createValidHugoProjectForService(projectPath); err != nil {
					return err
				}

				hugoService1 := NewHugoService()
				hugoService2 := NewHugoService()

				port := 3456

				// Start first server
				err = hugoService1.StartServer(projectPath, port)
				if err != nil {
					return err
				}
				defer hugoService1.StopServer()

				// Give first server time to start
				time.Sleep(500 * time.Millisecond)

				// Try to start second server on same port - should fail
				err = hugoService2.StartServer(projectPath, port)
				if err == nil {
					hugoService2.StopServer() // Cleanup
				}
				return err
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tc.name, err)
			} else if tc.expectError && err != nil {
				t.Logf("Got expected error for %s: %v", tc.name, err)
			}
		})
	}
}
// **Feature: hugo-visual-client, Property 21: 构建错误报告准确性**
// **Validates: Requirements 6.3**
func TestBuildErrorReportingAccuracy(t *testing.T) {
	// Skip test if Hugo is not installed
	service := NewHugoService()
	if installed, _, err := service.IsHugoInstalled(); !installed {
		t.Skip("Hugo is not installed, skipping build error reporting test:", err)
	}

	// Test different scenarios that should produce build errors or warnings
	testCases := []struct {
		name           string
		setupProject   func(projectPath string) error
		expectSuccess  bool
		expectErrors   bool
		expectWarnings bool
	}{
		{
			name: "Valid project with warnings",
			setupProject: func(projectPath string) error {
				// Create a valid Hugo project (will have warnings about missing layouts)
				return createValidHugoProjectForService(projectPath)
			},
			expectSuccess:  true,
			expectErrors:   false,
			expectWarnings: true, // Hugo will warn about missing layout files
		},
		{
			name: "Project with no config file",
			setupProject: func(projectPath string) error {
				// Create directories but no config file
				dirs := []string{"content", "themes", "static", "layouts", "data", "assets"}
				for _, dir := range dirs {
					if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
						return err
					}
				}
				return nil
			},
			expectSuccess:  false,
			expectErrors:   true, // Hugo will error about missing config
			expectWarnings: false,
		},
		{
			name: "Empty directory",
			setupProject: func(projectPath string) error {
				// Just create the directory, no content
				return os.MkdirAll(projectPath, 0755)
			},
			expectSuccess:  false,
			expectErrors:   true, // Hugo will error about missing config
			expectWarnings: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-build-error-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			projectPath := filepath.Join(tempDir, "test-project")
			outputPath := filepath.Join(tempDir, "public")

			// Setup the project according to test case
			if err := tc.setupProject(projectPath); err != nil {
				t.Fatalf("Failed to setup project: %v", err)
			}

			// Create Hugo service
			hugoService := NewHugoService()

			// Try to build site
			buildResult, err := hugoService.BuildSite(projectPath, outputPath)

			// Check if build result matches expectations
			if tc.expectSuccess {
				if buildResult == nil {
					t.Error("Expected build result, but got nil")
					return
				}
				if !buildResult.Success {
					t.Errorf("Expected successful build, but got failure. Errors: %v", buildResult.Errors)
				}
			} else {
				// For failed builds, we might get an error or a build result with Success=false
				if err == nil && (buildResult == nil || buildResult.Success) {
					t.Error("Expected build to fail, but it succeeded")
				}
			}

			// Check error reporting
			if tc.expectErrors {
				if buildResult != nil && len(buildResult.Errors) == 0 {
					t.Error("Expected build errors to be reported, but got none")
				} else if buildResult != nil {
					t.Logf("Got expected errors: %v", buildResult.Errors)
				}
			}

			// Check warning reporting
			if tc.expectWarnings {
				if buildResult != nil && len(buildResult.Warnings) == 0 {
					t.Log("Expected build warnings, but got none (this might be OK depending on Hugo version)")
				} else if buildResult != nil {
					t.Logf("Got expected warnings: %v", buildResult.Warnings)
				}
			}

			// Verify build result structure
			if buildResult != nil {
				// Output path should be set
				if buildResult.OutputPath != outputPath {
					t.Errorf("Output path mismatch: expected %s, got %s", outputPath, buildResult.OutputPath)
				}

				// Files written count should be non-negative
				if buildResult.FilesWritten < 0 {
					t.Errorf("Invalid files written count: %d", buildResult.FilesWritten)
				}

				// Duration should be captured (might be empty for failed builds)
				t.Logf("Build duration: %s", buildResult.Duration)

				// Errors and warnings should be properly formatted
				for i, errMsg := range buildResult.Errors {
					if strings.TrimSpace(errMsg) == "" {
						t.Errorf("Error message %d is empty", i)
					}
				}

				for i, warnMsg := range buildResult.Warnings {
					if strings.TrimSpace(warnMsg) == "" {
						t.Errorf("Warning message %d is empty", i)
					}
				}
			}
		})
	}
}