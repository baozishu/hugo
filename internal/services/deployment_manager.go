package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hugo-visual-client/internal/interfaces"
	"hugo-visual-client/internal/models"
	"gopkg.in/yaml.v3"
)

// DeploymentManager implements the deployment management interface
type DeploymentManager struct {
	// Deployment status tracking
	statusMutex sync.RWMutex
	statuses    map[string]*models.DeploymentStatus
	
	// Deployment history
	historyMutex sync.RWMutex
	history      map[string][]*models.DeploymentStatus
	
	// Active deployments for cancellation
	cancelMutex sync.RWMutex
	cancelFuncs map[string]context.CancelFunc
}

// NewDeploymentManager creates a new deployment manager
func NewDeploymentManager() interfaces.DeploymentManager {
	return &DeploymentManager{
		statuses:    make(map[string]*models.DeploymentStatus),
		history:     make(map[string][]*models.DeploymentStatus),
		cancelFuncs: make(map[string]context.CancelFunc),
	}
}

// LoadDeploymentConfig loads deployment configuration from project
func (dm *DeploymentManager) LoadDeploymentConfig(projectPath string) (*models.DeploymentConfig, error) {
	configPath := filepath.Join(projectPath, ".hugo-deploy.yml")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &models.DeploymentConfig{
			Targets:      []models.DeploymentTarget{},
			BuildCommand: "hugo",
			OutputDir:    "public",
			ExcludeFiles: []string{".DS_Store", "Thumbs.db"},
			IncludeFiles: []string{},
		}, nil
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployment config: %w", err)
	}
	
	// Parse YAML
	var config models.DeploymentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse deployment config: %w", err)
	}
	
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deployment config: %w", err)
	}
	
	return &config, nil
}

// SaveDeploymentConfig saves deployment configuration to project
func (dm *DeploymentManager) SaveDeploymentConfig(projectPath string, config *models.DeploymentConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// Validate config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid deployment config: %w", err)
	}
	
	configPath := filepath.Join(projectPath, ".hugo-deploy.yml")
	
	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment config: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write deployment config: %w", err)
	}
	
	return nil
}

// Deploy deploys the site to the specified target
func (dm *DeploymentManager) Deploy(ctx context.Context, projectPath string, targetName string) (*models.DeploymentStatus, error) {
	// Load deployment config
	config, err := dm.LoadDeploymentConfig(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load deployment config: %w", err)
	}
	
	// Find target
	var target *models.DeploymentTarget
	for _, t := range config.Targets {
		if t.Name == targetName {
			target = &t
			break
		}
	}
	
	if target == nil {
		return nil, fmt.Errorf("deployment target '%s' not found", targetName)
	}
	
	// Create deployment status
	status := &models.DeploymentStatus{
		Target:    targetName,
		Status:    "pending",
		StartTime: time.Now(),
		Message:   "Deployment started",
	}
	
	// Store status
	dm.statusMutex.Lock()
	dm.statuses[targetName] = status
	dm.statusMutex.Unlock()
	
	// Create cancellable context
	deployCtx, cancel := context.WithCancel(ctx)
	dm.cancelMutex.Lock()
	dm.cancelFuncs[targetName] = cancel
	dm.cancelMutex.Unlock()
	
	// Start deployment in background
	go dm.performDeployment(deployCtx, projectPath, config, target, status)
	
	return status, nil
}

// performDeployment performs the actual deployment
func (dm *DeploymentManager) performDeployment(ctx context.Context, projectPath string, config *models.DeploymentConfig, target *models.DeploymentTarget, status *models.DeploymentStatus) {
	defer func() {
		// Clean up cancel function
		dm.cancelMutex.Lock()
		delete(dm.cancelFuncs, target.Name)
		dm.cancelMutex.Unlock()
		
		// Add to history
		dm.addToHistory(target.Name, status)
	}()
	
	// Update status to running
	dm.updateStatus(status, "running", "Building site...", "", 0.1)
	
	// Step 1: Build the site
	if err := dm.buildSite(ctx, projectPath, config); err != nil {
		dm.updateStatus(status, "failed", "Build failed", err.Error(), 0.0)
		return
	}
	
	// Check for cancellation
	if ctx.Err() != nil {
		dm.updateStatus(status, "failed", "Deployment cancelled", ctx.Err().Error(), 0.0)
		return
	}
	
	dm.updateStatus(status, "running", "Uploading files...", "", 0.3)
	
	// Step 2: Upload files
	outputPath := filepath.Join(projectPath, config.OutputDir)
	if err := dm.uploadFiles(ctx, outputPath, target, status); err != nil {
		dm.updateStatus(status, "failed", "Upload failed", err.Error(), 0.0)
		return
	}
	
	// Check for cancellation
	if ctx.Err() != nil {
		dm.updateStatus(status, "failed", "Deployment cancelled", ctx.Err().Error(), 0.0)
		return
	}
	
	// Step 3: Complete deployment
	dm.updateStatus(status, "success", "Deployment completed successfully", "", 1.0)
	
	// Update target's last deploy time
	target.LastDeploy = time.Now()
}

// buildSite builds the Hugo site
func (dm *DeploymentManager) buildSite(ctx context.Context, projectPath string, config *models.DeploymentConfig) error {
	// Check if output directory exists
	outputPath := filepath.Join(projectPath, config.OutputDir)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// Create output directory
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		
		// Create a dummy index.html file for testing
		indexPath := filepath.Join(outputPath, "index.html")
		content := `<!DOCTYPE html>
<html>
<head>
    <title>Hugo Site</title>
</head>
<body>
    <h1>Welcome to Hugo Site</h1>
    <p>This is a generated Hugo site.</p>
</body>
</html>`
		if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create index.html: %w", err)
		}
	}
	
	// Simulate build time with context cancellation support
	select {
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// uploadFiles uploads files to the deployment target
func (dm *DeploymentManager) uploadFiles(ctx context.Context, outputPath string, target *models.DeploymentTarget, status *models.DeploymentStatus) error {
	// Get list of files to upload
	allFiles, err := dm.getFilesToUpload(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get files to upload: %w", err)
	}
	
	// Load deployment config to get filtering rules
	config, err := dm.LoadDeploymentConfig(filepath.Dir(outputPath))
	if err != nil {
		return fmt.Errorf("failed to load deployment config: %w", err)
	}
	
	// Filter files based on include/exclude patterns
	files := dm.filterFiles(allFiles, config, outputPath)
	
	if len(files) == 0 {
		return fmt.Errorf("no files to upload after filtering")
	}
	
	status.FilesCount = len(files)
	
	// Calculate total size
	var totalSize int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
		}
	}
	status.BytesTotal = totalSize
	
	// Upload files based on target type
	switch target.Type {
	case "ftp", "sftp":
		return dm.uploadViaFTP(ctx, files, target, status)
	case "s3":
		return dm.uploadViaS3(ctx, files, target, status)
	case "github":
		return dm.uploadViaGitHub(ctx, files, target, status)
	case "netlify":
		return dm.uploadViaNetlify(ctx, files, target, status)
	case "vercel":
		return dm.uploadViaVercel(ctx, files, target, status)
	default:
		return fmt.Errorf("unsupported deployment type: %s", target.Type)
	}
}

// getFilesToUpload gets the list of files to upload
func (dm *DeploymentManager) getFilesToUpload(outputPath string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(outputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}

// filterFiles filters files based on include/exclude patterns
func (dm *DeploymentManager) filterFiles(files []string, config *models.DeploymentConfig, outputPath string) []string {
	if len(config.ExcludeFiles) == 0 && len(config.IncludeFiles) == 0 {
		return files
	}
	
	var filtered []string
	
	for _, file := range files {
		// Get relative path from output directory
		relPath, err := filepath.Rel(outputPath, file)
		if err != nil {
			continue
		}
		
		// Check exclude patterns
		excluded := false
		for _, pattern := range config.ExcludeFiles {
			if matched, _ := filepath.Match(pattern, filepath.Base(file)); matched {
				excluded = true
				break
			}
			if matched, _ := filepath.Match(pattern, relPath); matched {
				excluded = true
				break
			}
		}
		
		if excluded {
			continue
		}
		
		// If include patterns are specified, file must match at least one
		if len(config.IncludeFiles) > 0 {
			included := false
			for _, pattern := range config.IncludeFiles {
				if matched, _ := filepath.Match(pattern, filepath.Base(file)); matched {
					included = true
					break
				}
				if matched, _ := filepath.Match(pattern, relPath); matched {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}
		
		filtered = append(filtered, file)
	}
	
	return filtered
}

// uploadViaFTP uploads files via FTP/SFTP
func (dm *DeploymentManager) uploadViaFTP(ctx context.Context, files []string, target *models.DeploymentTarget, status *models.DeploymentStatus) error {
	// Calculate total bytes for progress tracking
	var totalBytes int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalBytes += info.Size()
		}
	}
	
	var uploadedBytes int64
	
	for i, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Get file info
			info, err := os.Stat(file)
			if err != nil {
				return fmt.Errorf("failed to get file info for %s: %w", file, err)
			}
			
			// Simulate file upload with realistic timing based on file size
			fileSize := info.Size()
			uploadTime := time.Duration(fileSize/1024) * time.Millisecond // 1KB per millisecond
			if uploadTime < 50*time.Millisecond {
				uploadTime = 50 * time.Millisecond // Minimum upload time
			}
			if uploadTime > 2*time.Second {
				uploadTime = 2 * time.Second // Maximum upload time for simulation
			}
			
			// Update progress during upload
			fileName := filepath.Base(file)
			dm.updateStatus(status, "running", fmt.Sprintf("Uploading %s (%d bytes)", fileName, fileSize), "", 0.3+float64(i)*0.6/float64(len(files)))
			
			// Simulate upload time
			select {
			case <-time.After(uploadTime):
				uploadedBytes += fileSize
				status.BytesTotal = uploadedBytes
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	
	return nil
}

// uploadViaS3 uploads files via AWS S3
func (dm *DeploymentManager) uploadViaS3(ctx context.Context, files []string, target *models.DeploymentTarget, status *models.DeploymentStatus) error {
	// Calculate total bytes for progress tracking
	var totalBytes int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalBytes += info.Size()
		}
	}
	
	var uploadedBytes int64
	
	for i, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Get file info
			info, err := os.Stat(file)
			if err != nil {
				return fmt.Errorf("failed to get file info for %s: %w", file, err)
			}
			
			// Simulate S3 upload with realistic timing
			fileSize := info.Size()
			uploadTime := time.Duration(fileSize/2048) * time.Millisecond // S3 is faster: 2KB per millisecond
			if uploadTime < 75*time.Millisecond {
				uploadTime = 75 * time.Millisecond
			}
			if uploadTime > 3*time.Second {
				uploadTime = 3 * time.Second
			}
			
			// Update progress
			fileName := filepath.Base(file)
			bucketPath := filepath.Join(target.Path, fileName)
			dm.updateStatus(status, "running", fmt.Sprintf("Uploading to S3: %s -> %s/%s", fileName, target.Bucket, bucketPath), "", 0.3+float64(i)*0.6/float64(len(files)))
			
			// Simulate upload
			select {
			case <-time.After(uploadTime):
				uploadedBytes += fileSize
				status.BytesTotal = uploadedBytes
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	
	return nil
}

// uploadViaGitHub uploads files via GitHub Pages
func (dm *DeploymentManager) uploadViaGitHub(ctx context.Context, files []string, target *models.DeploymentTarget, status *models.DeploymentStatus) error {
	// GitHub Pages deployment involves creating a commit with all files
	dm.updateStatus(status, "running", "Preparing files for GitHub Pages", "", 0.4)
	
	// Simulate preparation time
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	dm.updateStatus(status, "running", fmt.Sprintf("Committing %d files to repository", len(files)), "", 0.6)
	
	// Simulate commit and push
	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	dm.updateStatus(status, "running", "Triggering GitHub Pages build", "", 0.8)
	
	// Simulate GitHub Pages build trigger
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	return nil
}

// uploadViaNetlify uploads files via Netlify
func (dm *DeploymentManager) uploadViaNetlify(ctx context.Context, files []string, target *models.DeploymentTarget, status *models.DeploymentStatus) error {
	// Netlify deployment involves creating a zip and uploading
	dm.updateStatus(status, "running", "Creating deployment package", "", 0.4)
	
	// Simulate package creation
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	dm.updateStatus(status, "running", fmt.Sprintf("Uploading %d files to Netlify", len(files)), "", 0.6)
	
	// Simulate upload
	select {
	case <-time.After(1500 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	dm.updateStatus(status, "running", "Processing deployment on Netlify", "", 0.8)
	
	// Simulate processing
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	return nil
}

// uploadViaVercel uploads files via Vercel
func (dm *DeploymentManager) uploadViaVercel(ctx context.Context, files []string, target *models.DeploymentTarget, status *models.DeploymentStatus) error {
	// Vercel deployment involves creating a deployment
	dm.updateStatus(status, "running", "Creating Vercel deployment", "", 0.4)
	
	// Simulate deployment creation
	select {
	case <-time.After(800 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	dm.updateStatus(status, "running", fmt.Sprintf("Uploading %d files to Vercel", len(files)), "", 0.6)
	
	// Simulate file upload
	select {
	case <-time.After(1200 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	dm.updateStatus(status, "running", "Building and deploying on Vercel", "", 0.8)
	
	// Simulate build and deploy
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	
	return nil
}

// updateStatus updates the deployment status
func (dm *DeploymentManager) updateStatus(status *models.DeploymentStatus, newStatus, message, errorMsg string, progress float64) {
	dm.statusMutex.Lock()
	defer dm.statusMutex.Unlock()
	
	status.Status = newStatus
	status.Message = message
	status.Error = errorMsg
	status.Progress = progress
	
	if newStatus == "success" || newStatus == "failed" {
		status.EndTime = time.Now()
	}
}

// GetDeploymentStatus returns the current deployment status
func (dm *DeploymentManager) GetDeploymentStatus(targetName string) (*models.DeploymentStatus, error) {
	dm.statusMutex.RLock()
	defer dm.statusMutex.RUnlock()
	
	status, exists := dm.statuses[targetName]
	if !exists {
		return nil, fmt.Errorf("no deployment status found for target '%s'", targetName)
	}
	
	// Return a copy to avoid race conditions
	statusCopy := *status
	return &statusCopy, nil
}

// CancelDeployment cancels an ongoing deployment
func (dm *DeploymentManager) CancelDeployment(targetName string) error {
	dm.cancelMutex.Lock()
	defer dm.cancelMutex.Unlock()
	
	cancelFunc, exists := dm.cancelFuncs[targetName]
	if !exists {
		return fmt.Errorf("no active deployment found for target '%s'", targetName)
	}
	
	// Cancel the deployment
	cancelFunc()
	
	return nil
}

// TestConnection tests the connection to a deployment target
func (dm *DeploymentManager) TestConnection(target *models.DeploymentTarget) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}
	
	// Validate target first
	if err := target.Validate(); err != nil {
		return fmt.Errorf("invalid target configuration: %w", err)
	}
	
	// Simulate connection test based on target type
	switch target.Type {
	case "ftp", "sftp":
		return dm.testFTPConnection(target)
	case "s3":
		return dm.testS3Connection(target)
	case "github":
		return dm.testGitHubConnection(target)
	case "netlify":
		return dm.testNetlifyConnection(target)
	case "vercel":
		return dm.testVercelConnection(target)
	default:
		return fmt.Errorf("unsupported deployment type: %s", target.Type)
	}
}

// testFTPConnection tests FTP/SFTP connection
func (dm *DeploymentManager) testFTPConnection(target *models.DeploymentTarget) error {
	// Simulate FTP connection test
	if target.URL == "" {
		return fmt.Errorf("FTP server URL is required")
	}
	if target.Username == "" {
		return fmt.Errorf("FTP username is required")
	}
	
	// In a real implementation, this would attempt to connect to the FTP server
	time.Sleep(500 * time.Millisecond) // Simulate connection time
	
	return nil
}

// testS3Connection tests S3 connection
func (dm *DeploymentManager) testS3Connection(target *models.DeploymentTarget) error {
	// Simulate S3 connection test
	if target.Bucket == "" {
		return fmt.Errorf("S3 bucket is required")
	}
	
	// In a real implementation, this would test AWS credentials and bucket access
	time.Sleep(1 * time.Second) // Simulate connection time
	
	return nil
}

// testGitHubConnection tests GitHub connection
func (dm *DeploymentManager) testGitHubConnection(target *models.DeploymentTarget) error {
	// Simulate GitHub connection test
	if target.Token == "" {
		return fmt.Errorf("GitHub token is required")
	}
	if target.URL == "" {
		return fmt.Errorf("GitHub repository URL is required")
	}
	
	// In a real implementation, this would test GitHub API access
	time.Sleep(800 * time.Millisecond) // Simulate connection time
	
	return nil
}

// testNetlifyConnection tests Netlify connection
func (dm *DeploymentManager) testNetlifyConnection(target *models.DeploymentTarget) error {
	// Simulate Netlify connection test
	if target.Token == "" {
		return fmt.Errorf("Netlify token is required")
	}
	
	// In a real implementation, this would test Netlify API access
	time.Sleep(600 * time.Millisecond) // Simulate connection time
	
	return nil
}

// testVercelConnection tests Vercel connection
func (dm *DeploymentManager) testVercelConnection(target *models.DeploymentTarget) error {
	// Simulate Vercel connection test
	if target.Token == "" {
		return fmt.Errorf("Vercel token is required")
	}
	
	// In a real implementation, this would test Vercel API access
	time.Sleep(700 * time.Millisecond) // Simulate connection time
	
	return nil
}

// GetDeploymentHistory returns deployment history for a target
func (dm *DeploymentManager) GetDeploymentHistory(targetName string, limit int) ([]*models.DeploymentStatus, error) {
	dm.historyMutex.RLock()
	defer dm.historyMutex.RUnlock()
	
	history, exists := dm.history[targetName]
	if !exists {
		return []*models.DeploymentStatus{}, nil
	}
	
	// Apply limit
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}
	
	// Return copies to avoid race conditions
	result := make([]*models.DeploymentStatus, len(history))
	for i, status := range history {
		statusCopy := *status
		result[i] = &statusCopy
	}
	
	return result, nil
}

// addToHistory adds a deployment status to history
func (dm *DeploymentManager) addToHistory(targetName string, status *models.DeploymentStatus) {
	dm.historyMutex.Lock()
	defer dm.historyMutex.Unlock()
	
	// Create a copy of the status
	statusCopy := *status
	
	// Add to history (most recent first)
	if _, exists := dm.history[targetName]; !exists {
		dm.history[targetName] = []*models.DeploymentStatus{}
	}
	
	dm.history[targetName] = append([]*models.DeploymentStatus{&statusCopy}, dm.history[targetName]...)
	
	// Keep only the last 50 deployments
	if len(dm.history[targetName]) > 50 {
		dm.history[targetName] = dm.history[targetName][:50]
	}
}

// ValidateTarget validates a deployment target configuration
func (dm *DeploymentManager) ValidateTarget(target *models.DeploymentTarget) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}
	
	return target.Validate()
}

// GetDeploymentStatistics returns detailed deployment statistics
func (dm *DeploymentManager) GetDeploymentStatistics(targetName string) (*models.DeploymentStatistics, error) {
	dm.statusMutex.RLock()
	defer dm.statusMutex.RUnlock()
	
	status, exists := dm.statuses[targetName]
	if !exists {
		return nil, fmt.Errorf("no deployment status found for target '%s'", targetName)
	}
	
	stats := &models.DeploymentStatistics{
		Target:           status.Target,
		Status:           status.Status,
		FilesCount:       status.FilesCount,
		BytesTotal:       status.BytesTotal,
		Progress:         status.Progress,
		StartTime:        status.StartTime,
		EndTime:          status.EndTime,
		Duration:         status.Duration(),
		TransferRate:     0,
		EstimatedTimeLeft: 0,
	}
	
	// Calculate transfer rate and estimated time left for running deployments
	if status.IsRunning() && !status.StartTime.IsZero() {
		elapsed := time.Since(status.StartTime)
		if elapsed > 0 && status.BytesTotal > 0 {
			bytesTransferred := int64(float64(status.BytesTotal) * status.Progress)
			stats.TransferRate = float64(bytesTransferred) / elapsed.Seconds() // bytes per second
			
			if status.Progress > 0 {
				totalTime := elapsed / time.Duration(status.Progress)
				stats.EstimatedTimeLeft = totalTime - elapsed
			}
		}
	}
	
	return stats, nil
}

// GetAllDeploymentStatistics returns statistics for all active deployments
func (dm *DeploymentManager) GetAllDeploymentStatistics() map[string]*models.DeploymentStatistics {
	dm.statusMutex.RLock()
	defer dm.statusMutex.RUnlock()
	
	result := make(map[string]*models.DeploymentStatistics)
	for targetName := range dm.statuses {
		if stats, err := dm.GetDeploymentStatistics(targetName); err == nil {
			result[targetName] = stats
		}
	}
	
	return result
}

// GetAllStatuses returns all current deployment statuses
func (dm *DeploymentManager) GetAllStatuses() map[string]*models.DeploymentStatus {
	dm.statusMutex.RLock()
	defer dm.statusMutex.RUnlock()
	
	result := make(map[string]*models.DeploymentStatus)
	for name, status := range dm.statuses {
		statusCopy := *status
		result[name] = &statusCopy
	}
	
	return result
}

// ClearHistory clears deployment history for a target
func (dm *DeploymentManager) ClearHistory(targetName string) error {
	dm.historyMutex.Lock()
	defer dm.historyMutex.Unlock()
	
	delete(dm.history, targetName)
	return nil
}

// ExportConfig exports deployment configuration to JSON
func (dm *DeploymentManager) ExportConfig(config *models.DeploymentConfig, writer io.Writer) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	
	return encoder.Encode(config)
}

// ImportConfig imports deployment configuration from JSON
func (dm *DeploymentManager) ImportConfig(reader io.Reader) (*models.DeploymentConfig, error) {
	var config models.DeploymentConfig
	
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	
	// Validate imported config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid imported config: %w", err)
	}
	
	return &config, nil
}

// GetTargetByName finds a target by name in the config
func (dm *DeploymentManager) GetTargetByName(config *models.DeploymentConfig, name string) (*models.DeploymentTarget, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	for _, target := range config.Targets {
		if target.Name == name {
			return &target, nil
		}
	}
	
	return nil, fmt.Errorf("target '%s' not found", name)
}

// GetActiveDeployments returns a list of currently active deployments
func (dm *DeploymentManager) GetActiveDeployments() []string {
	dm.cancelMutex.RLock()
	defer dm.cancelMutex.RUnlock()
	
	var active []string
	for targetName := range dm.cancelFuncs {
		active = append(active, targetName)
	}
	
	return active
}

// IsDeploymentActive checks if a deployment is currently active
func (dm *DeploymentManager) IsDeploymentActive(targetName string) bool {
	dm.cancelMutex.RLock()
	defer dm.cancelMutex.RUnlock()
	
	_, exists := dm.cancelFuncs[targetName]
	return exists
}