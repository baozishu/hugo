package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"hugo-visual-client/internal/models"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestDeploymentManager_LoadSaveConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-deploy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	dm := NewDeploymentManager()
	
	// Test loading non-existent config (should return default)
	config, err := dm.LoadDeploymentConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	
	if config == nil {
		t.Fatal("Config should not be nil")
	}
	
	if config.BuildCommand != "hugo" {
		t.Errorf("Expected default build command 'hugo', got '%s'", config.BuildCommand)
	}
	
	// Add a test target
	testTarget := models.DeploymentTarget{
		Name:     "test-target",
		Type:     "ftp",
		URL:      "ftp.example.com",
		Username: "testuser",
		Password: "testpass",
		Path:     "/public_html",
		Port:     21,
	}
	
	err = config.AddTarget(testTarget)
	if err != nil {
		t.Fatalf("Failed to add target: %v", err)
	}
	
	// Save config
	err = dm.SaveDeploymentConfig(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Load config again
	loadedConfig, err := dm.LoadDeploymentConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	
	// Verify config was saved and loaded correctly
	if len(loadedConfig.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(loadedConfig.Targets))
	}
	
	if loadedConfig.Targets[0].Name != "test-target" {
		t.Errorf("Expected target name 'test-target', got '%s'", loadedConfig.Targets[0].Name)
	}
}

func TestDeploymentManager_TestConnection(t *testing.T) {
	dm := NewDeploymentManager()
	
	// Test valid FTP target
	ftpTarget := &models.DeploymentTarget{
		Name:     "test-ftp",
		Type:     "ftp",
		URL:      "ftp.example.com",
		Username: "testuser",
		Password: "testpass",
		Path:     "/",
		Port:     21,
	}
	
	err := dm.TestConnection(ftpTarget)
	if err != nil {
		t.Errorf("FTP connection test should succeed: %v", err)
	}
	
	// Test invalid target (missing required fields)
	invalidTarget := &models.DeploymentTarget{
		Name: "invalid",
		Type: "ftp",
		// Missing URL and Username
	}
	
	err = dm.TestConnection(invalidTarget)
	if err == nil {
		t.Error("Invalid target connection test should fail")
	}
}

func TestDeploymentManager_Deploy(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-deploy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create public directory with test file
	publicDir := filepath.Join(tempDir, "public")
	err = os.MkdirAll(publicDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create public dir: %v", err)
	}
	
	testFile := filepath.Join(publicDir, "index.html")
	err = os.WriteFile(testFile, []byte("<html><body>Test</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	dm := NewDeploymentManager()
	
	// Create deployment config
	config := &models.DeploymentConfig{
		BuildCommand: "hugo",
		OutputDir:    "public",
		Targets: []models.DeploymentTarget{
			{
				Name:     "test-target",
				Type:     "ftp",
				URL:      "ftp.example.com",
				Username: "testuser",
				Password: "testpass",
				Path:     "/",
				Port:     21,
			},
		},
		DefaultTarget: "test-target",
	}
	
	// Save config
	err = dm.SaveDeploymentConfig(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Start deployment
	ctx := context.Background()
	status, err := dm.Deploy(ctx, tempDir, "test-target")
	if err != nil {
		t.Fatalf("Failed to start deployment: %v", err)
	}
	
	if status == nil {
		t.Fatal("Deployment status should not be nil")
	}
	
	if status.Target != "test-target" {
		t.Errorf("Expected target 'test-target', got '%s'", status.Target)
	}
	
	// Wait for deployment to complete (with timeout)
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			t.Fatal("Deployment timed out")
		case <-ticker.C:
			currentStatus, err := dm.GetDeploymentStatus("test-target")
			if err != nil {
				t.Fatalf("Failed to get deployment status: %v", err)
			}
			
			if currentStatus.IsComplete() {
				if currentStatus.Status != "success" {
					t.Errorf("Expected successful deployment, got status: %s, error: %s", 
						currentStatus.Status, currentStatus.Error)
				}
				return
			}
		}
	}
}

func TestDeploymentManager_CancelDeployment(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-deploy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	dm := NewDeploymentManager()
	
	// Create deployment config
	config := &models.DeploymentConfig{
		BuildCommand: "hugo",
		OutputDir:    "public",
		Targets: []models.DeploymentTarget{
			{
				Name:     "test-target",
				Type:     "ftp",
				URL:      "ftp.example.com",
				Username: "testuser",
				Password: "testpass",
				Path:     "/",
				Port:     21,
			},
		},
		DefaultTarget: "test-target",
	}
	
	// Save config
	err = dm.SaveDeploymentConfig(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Start deployment
	ctx := context.Background()
	_, err = dm.Deploy(ctx, tempDir, "test-target")
	if err != nil {
		t.Fatalf("Failed to start deployment: %v", err)
	}
	
	// Wait a bit for deployment to start
	time.Sleep(100 * time.Millisecond)
	
	// Cancel deployment
	err = dm.CancelDeployment("test-target")
	if err != nil {
		t.Fatalf("Failed to cancel deployment: %v", err)
	}
	
	// Wait for cancellation to take effect
	time.Sleep(500 * time.Millisecond)
	
	// Check that deployment was cancelled
	currentStatus, err := dm.GetDeploymentStatus("test-target")
	if err != nil {
		t.Fatalf("Failed to get deployment status: %v", err)
	}
	
	if currentStatus.Status != "failed" {
		t.Errorf("Expected cancelled deployment to have 'failed' status, got '%s'", currentStatus.Status)
	}
}

func TestDeploymentManager_History(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-deploy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	dm := NewDeploymentManager()
	
	// Create deployment config
	config := &models.DeploymentConfig{
		BuildCommand: "hugo",
		OutputDir:    "public",
		Targets: []models.DeploymentTarget{
			{
				Name:     "test-target",
				Type:     "ftp",
				URL:      "ftp.example.com",
				Username: "testuser",
				Password: "testpass",
				Path:     "/",
				Port:     21,
			},
		},
		DefaultTarget: "test-target",
	}
	
	// Save config
	err = dm.SaveDeploymentConfig(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Initially, history should be empty
	history, err := dm.GetDeploymentHistory("test-target", 10)
	if err != nil {
		t.Fatalf("Failed to get deployment history: %v", err)
	}
	
	if len(history) != 0 {
		t.Errorf("Expected 0 history entries initially, got %d", len(history))
	}
	
	// Perform a deployment to generate history
	ctx := context.Background()
	_, err = dm.Deploy(ctx, tempDir, "test-target")
	if err != nil {
		t.Fatalf("Failed to start deployment: %v", err)
	}
	
	// Wait for deployment to complete
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			t.Fatal("Deployment timed out")
		case <-ticker.C:
			currentStatus, err := dm.GetDeploymentStatus("test-target")
			if err != nil {
				t.Fatalf("Failed to get deployment status: %v", err)
			}
			
			if currentStatus.IsComplete() {
				// Now check history
				history, err := dm.GetDeploymentHistory("test-target", 10)
				if err != nil {
					t.Fatalf("Failed to get deployment history: %v", err)
				}
				
				if len(history) != 1 {
					t.Errorf("Expected 1 history entry after deployment, got %d", len(history))
				}
				
				if len(history) > 0 && history[0].Target != "test-target" {
					t.Errorf("Expected history target 'test-target', got '%s'", history[0].Target)
				}
				
				return
			}
		}
	}
}

func TestDeploymentManager_Statistics(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-deploy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	dm := NewDeploymentManager()
	
	// Create deployment config
	config := &models.DeploymentConfig{
		BuildCommand: "hugo",
		OutputDir:    "public",
		Targets: []models.DeploymentTarget{
			{
				Name:     "test-target",
				Type:     "ftp",
				URL:      "ftp.example.com",
				Username: "testuser",
				Password: "testpass",
				Path:     "/",
				Port:     21,
			},
		},
		DefaultTarget: "test-target",
	}
	
	// Save config
	err = dm.SaveDeploymentConfig(tempDir, config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Start deployment
	ctx := context.Background()
	_, err = dm.Deploy(ctx, tempDir, "test-target")
	if err != nil {
		t.Fatalf("Failed to start deployment: %v", err)
	}
	
	// Wait a bit for deployment to start
	time.Sleep(200 * time.Millisecond)
	
	// Get deployment statistics
	stats, err := dm.GetDeploymentStatistics("test-target")
	if err != nil {
		t.Fatalf("Failed to get deployment statistics: %v", err)
	}
	
	if stats == nil {
		t.Fatal("Statistics should not be nil")
	}
	
	if stats.Target != "test-target" {
		t.Errorf("Expected target 'test-target', got '%s'", stats.Target)
	}
	
	if stats.Status != "running" && stats.Status != "success" {
		t.Errorf("Expected status 'running' or 'success', got '%s'", stats.Status)
	}
	
	// Test getting all statistics
	allStats := dm.GetAllDeploymentStatistics()
	if len(allStats) != 1 {
		t.Errorf("Expected 1 deployment statistics, got %d", len(allStats))
	}
	
	if _, exists := allStats["test-target"]; !exists {
		t.Error("Expected statistics for 'test-target' to exist")
	}
}

// **Feature: hugo-visual-client, Property 23: 部署过程完整性**
// **Validates: Requirements 6.5**
func TestDeploymentProcessIntegrity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("deployment should transfer all files from build output directory to deployment target", prop.ForAll(
		func(targetType string, fileCount int, fileSizes []int64) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-deploy-integrity-test")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create public directory with test files
			publicDir := filepath.Join(tempDir, "public")
			err = os.MkdirAll(publicDir, 0755)
			if err != nil {
				t.Logf("Failed to create public dir: %v", err)
				return false
			}

			// Create test files with specified sizes
			expectedFiles := make(map[string]int64)
			for i := 0; i < fileCount; i++ {
				fileName := fmt.Sprintf("file%d.html", i)
				filePath := filepath.Join(publicDir, fileName)
				
				// Create file content based on specified size
				size := fileSizes[i%len(fileSizes)]
				if size <= 0 {
					size = 100 // Minimum size
				}
				content := make([]byte, size)
				for j := range content {
					content[j] = byte('A' + (j % 26)) // Fill with repeating alphabet
				}
				
				err = os.WriteFile(filePath, content, 0644)
				if err != nil {
					t.Logf("Failed to create test file %s: %v", fileName, err)
					return false
				}
				
				expectedFiles[fileName] = size
			}

			// Create deployment manager
			dm := NewDeploymentManager()

			// Create deployment config with test target
			config := &models.DeploymentConfig{
				BuildCommand: "hugo",
				OutputDir:    "public",
				Targets: []models.DeploymentTarget{
					{
						Name:     "test-target",
						Type:     targetType,
						URL:      "test.example.com",
						Username: "testuser",
						Password: "testpass",
						Token:    "test-token",
						Bucket:   "test-bucket",
						Region:   "us-east-1",
						Path:     "/",
						Port:     21,
					},
				},
				DefaultTarget: "test-target",
			}

			// Save config
			err = dm.SaveDeploymentConfig(tempDir, config)
			if err != nil {
				t.Logf("Failed to save config: %v", err)
				return false
			}

			// Start deployment
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			_, err = dm.Deploy(ctx, tempDir, "test-target")
			if err != nil {
				t.Logf("Failed to start deployment: %v", err)
				return false
			}

			// Wait for deployment to complete
			timeout := time.After(25 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					t.Logf("Deployment timed out")
					return false
				case <-ticker.C:
					currentStatus, err := dm.GetDeploymentStatus("test-target")
					if err != nil {
						t.Logf("Failed to get deployment status: %v", err)
						return false
					}

					if currentStatus.IsComplete() {
						// Verify deployment completed successfully
						if currentStatus.Status != "success" {
							t.Logf("Deployment failed: %s - %s", currentStatus.Status, currentStatus.Error)
							return false
						}

						// Verify all files were processed
						if currentStatus.FilesCount != len(expectedFiles) {
							t.Logf("Expected %d files to be processed, but got %d", len(expectedFiles), currentStatus.FilesCount)
							return false
						}

						// Verify total bytes match expected
						var expectedTotalBytes int64
						for _, size := range expectedFiles {
							expectedTotalBytes += size
						}
						
						// Allow some tolerance for metadata overhead in deployment
						if currentStatus.BytesTotal < expectedTotalBytes {
							t.Logf("Expected at least %d bytes to be transferred, but got %d", expectedTotalBytes, currentStatus.BytesTotal)
							return false
						}

						// Verify deployment progress reached completion
						if currentStatus.Progress < 1.0 {
							t.Logf("Expected deployment progress to be 1.0, but got %f", currentStatus.Progress)
							return false
						}

						// Verify deployment has end time set
						if currentStatus.EndTime.IsZero() {
							t.Logf("Expected deployment to have end time set")
							return false
						}

						// Verify deployment duration is reasonable
						duration := currentStatus.Duration()
						if duration <= 0 {
							t.Logf("Expected positive deployment duration, got %v", duration)
							return false
						}

						return true
					}
				}
			}
		},
		genValidDeploymentType(),
		gen.IntRange(1, 10), // File count between 1 and 10
		gen.SliceOfN(5, gen.Int64Range(100, 10000)), // File sizes between 100 bytes and 10KB
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genValidDeploymentType generates valid deployment types for testing
func genValidDeploymentType() gopter.Gen {
	return gen.OneConstOf("ftp", "sftp", "s3", "github", "netlify", "vercel")
}