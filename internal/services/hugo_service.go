package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"hugo-visual-client/internal/interfaces"
)

// HugoServiceImpl implements the HugoService interface
type HugoServiceImpl struct {
	serverCmd    *exec.Cmd
	serverCancel context.CancelFunc
	serverStatus interfaces.ServerStatus
	watcher      *fsnotify.Watcher
	watchCancel  context.CancelFunc
	mutex        sync.RWMutex
}

// NewHugoService creates a new Hugo service instance
func NewHugoService() *HugoServiceImpl {
	return &HugoServiceImpl{
		serverStatus: interfaces.ServerStatus{
			Running: false,
			Port:    0,
			URL:     "",
			PID:     0,
		},
	}
}

// IsHugoInstalled checks if Hugo is installed and accessible
func (h *HugoServiceImpl) IsHugoInstalled() (bool, string, error) {
	cmd := exec.Command("hugo", "version")
	output, err := cmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("hugo not found or not executable: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return true, version, nil
}

// GetHugoVersion returns the installed Hugo version
func (h *HugoServiceImpl) GetHugoVersion() (string, error) {
	installed, version, err := h.IsHugoInstalled()
	if !installed {
		return "", err
	}
	return version, nil
}

// StartServer starts the Hugo development server
func (h *HugoServiceImpl) StartServer(projectPath string, port int) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check if server is already running
	if h.serverStatus.Running {
		return fmt.Errorf("server is already running on port %d", h.serverStatus.Port)
	}

	// Validate project path
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", projectPath)
	}

	// Check if Hugo is installed
	if installed, _, err := h.IsHugoInstalled(); !installed {
		return fmt.Errorf("hugo is not installed: %w", err)
	}

	// Create context for server process
	ctx, cancel := context.WithCancel(context.Background())
	h.serverCancel = cancel

	// Prepare Hugo server command
	args := []string{"server", "--bind", "0.0.0.0", "--port", strconv.Itoa(port), "--buildDrafts", "--buildFuture"}
	cmd := exec.CommandContext(ctx, "hugo", args...)
	cmd.Dir = projectPath

	// Set up pipes for output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start hugo server: %w", err)
	}

	h.serverCmd = cmd

	// Monitor server process
	go func() {
		defer cancel()
		err := cmd.Wait()
		
		// Use a separate goroutine to update status to avoid deadlock
		go func() {
			h.mutex.Lock()
			h.serverStatus.Running = false
			h.serverStatus.Port = 0
			h.serverStatus.URL = ""
			h.serverStatus.PID = 0
			h.serverCmd = nil
			h.mutex.Unlock()
		}()

		if err != nil && ctx.Err() == nil {
			// Server exited unexpectedly
			fmt.Printf("Hugo server exited unexpectedly: %v\n", err)
		}
	}()

	// Monitor server output in goroutines
	go h.monitorServerOutput(stdout, "stdout")
	go h.monitorServerOutput(stderr, "stderr")

	// Wait a moment to see if server starts successfully
	time.Sleep(2 * time.Second)

	// Check if the process is still running
	if cmd.Process == nil || cmd.ProcessState != nil {
		// Process has already exited
		cancel()
		return fmt.Errorf("hugo server failed to start - process exited")
	}

	// Update server status only if server is still running
	h.serverStatus = interfaces.ServerStatus{
		Running: true,
		Port:    port,
		URL:     fmt.Sprintf("http://localhost:%d", port),
		PID:     cmd.Process.Pid,
	}

	return nil
}

// StopServer stops the Hugo development server
func (h *HugoServiceImpl) StopServer() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.serverStatus.Running || h.serverCmd == nil {
		return fmt.Errorf("server is not running")
	}

	// Cancel the context to signal shutdown
	if h.serverCancel != nil {
		h.serverCancel()
	}

	// On Windows, try a gentler approach first
	if h.serverCmd.Process != nil {
		// Try to terminate gracefully by closing stdin
		if h.serverCmd.Process.Pid > 0 {
			// Just kill the process directly on Windows for testing
			err := h.serverCmd.Process.Kill()
			if err != nil {
				fmt.Printf("Warning: failed to kill server process: %v\n", err)
			}
		}
	}

	// Reset server status immediately to prevent deadlock
	h.serverStatus = interfaces.ServerStatus{
		Running: false,
		Port:    0,
		URL:     "",
		PID:     0,
	}
	h.serverCmd = nil
	h.serverCancel = nil

	return nil
}

// GetServerStatus returns the current status of the development server
func (h *HugoServiceImpl) GetServerStatus() interfaces.ServerStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.serverStatus
}

// BuildSite builds the Hugo site to the specified output path
func (h *HugoServiceImpl) BuildSite(projectPath string, outputPath string) (*interfaces.BuildResult, error) {
	// Validate project path
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", projectPath)
	}

	// Check if Hugo is installed
	if installed, _, err := h.IsHugoInstalled(); !installed {
		return nil, fmt.Errorf("hugo is not installed: %w", err)
	}

	// Prepare build command
	args := []string{"--destination", outputPath}
	cmd := exec.Command("hugo", args...)
	cmd.Dir = projectPath

	// Capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	buildResult := &interfaces.BuildResult{
		Success:      err == nil,
		OutputPath:   outputPath,
		FilesWritten: 0,
		Duration:     "",
		Errors:       []string{},
		Warnings:     []string{},
	}

	// Parse Hugo output for statistics
	if err == nil {
		buildResult.FilesWritten = h.parseFilesWritten(outputStr)
		buildResult.Duration = h.parseBuildDuration(outputStr)
	} else {
		// Parse errors from output
		buildResult.Errors = h.parseErrors(outputStr)
	}

	// Parse warnings (present even in successful builds)
	buildResult.Warnings = h.parseWarnings(outputStr)

	return buildResult, nil
}

// WatchFiles starts watching for file changes in the project
func (h *HugoServiceImpl) WatchFiles(projectPath string, callback func(string)) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Stop existing watcher if running
	if h.watcher != nil {
		h.StopWatching()
	}

	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	h.watcher = watcher

	// Create context for watching
	ctx, cancel := context.WithCancel(context.Background())
	h.watchCancel = cancel

	// Add directories to watch
	watchDirs := []string{
		filepath.Join(projectPath, "content"),
		filepath.Join(projectPath, "layouts"),
		filepath.Join(projectPath, "static"),
		filepath.Join(projectPath, "data"),
		filepath.Join(projectPath, "assets"),
		projectPath, // Watch root for config changes
	}

	for _, dir := range watchDirs {
		if _, err := os.Stat(dir); err == nil {
			if err := watcher.Add(dir); err != nil {
				cancel()
				watcher.Close()
				return fmt.Errorf("failed to watch directory %s: %w", dir, err)
			}

			// Also watch subdirectories
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip errors
				}
				if info.IsDir() {
					watcher.Add(path)
				}
				return nil
			})
		}
	}

	// Start watching in goroutine
	go func() {
		defer watcher.Close()
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				
				// Filter relevant file changes
				if h.isRelevantFileChange(event) {
					callback(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("File watcher error: %v\n", err)
			}
		}
	}()

	return nil
}

// StopWatching stops file watching
func (h *HugoServiceImpl) StopWatching() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.watchCancel != nil {
		h.watchCancel()
		h.watchCancel = nil
	}

	if h.watcher != nil {
		err := h.watcher.Close()
		h.watcher = nil
		// Give some time for the watcher to fully close
		time.Sleep(100 * time.Millisecond)
		return err
	}

	return nil
}

// monitorServerOutput monitors Hugo server output
func (h *HugoServiceImpl) monitorServerOutput(pipe interface{}, source string) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))
	for scanner.Scan() {
		line := scanner.Text()
		// You can process server output here if needed
		// For now, just print to help with debugging
		fmt.Printf("[Hugo %s] %s\n", source, line)
	}
}

// parseFilesWritten extracts the number of files written from Hugo output
func (h *HugoServiceImpl) parseFilesWritten(output string) int {
	// Look for patterns like "Total in 123 ms" or similar Hugo output
	re := regexp.MustCompile(`(\d+)\s+(?:static files?|pages?|files?)\s+(?:copied|written|created)`)
	matches := re.FindAllStringSubmatch(output, -1)
	
	total := 0
	for _, match := range matches {
		if len(match) > 1 {
			if count, err := strconv.Atoi(match[1]); err == nil {
				total += count
			}
		}
	}
	
	return total
}

// parseBuildDuration extracts build duration from Hugo output
func (h *HugoServiceImpl) parseBuildDuration(output string) string {
	// Look for patterns like "Total in 123 ms"
	re := regexp.MustCompile(`Total in (\d+(?:\.\d+)?\s*(?:ms|s))`)
	matches := re.FindStringSubmatch(output)
	
	if len(matches) > 1 {
		return matches[1]
	}
	
	return ""
}

// parseErrors extracts error messages from Hugo output
func (h *HugoServiceImpl) parseErrors(output string) []string {
	var errors []string
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "error") && line != "" {
			errors = append(errors, line)
		}
	}
	
	return errors
}

// parseWarnings extracts warning messages from Hugo output
func (h *HugoServiceImpl) parseWarnings(output string) []string {
	var warnings []string
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "warn") && line != "" {
			warnings = append(warnings, line)
		}
	}
	
	return warnings
}

// isRelevantFileChange determines if a file change should trigger regeneration
func (h *HugoServiceImpl) isRelevantFileChange(event fsnotify.Event) bool {
	// Ignore temporary files and hidden files
	filename := filepath.Base(event.Name)
	if strings.HasPrefix(filename, ".") || strings.HasPrefix(filename, "~") {
		return false
	}
	
	// Ignore certain file extensions
	ext := strings.ToLower(filepath.Ext(event.Name))
	ignoredExts := []string{".tmp", ".swp", ".log", ".lock"}
	for _, ignoredExt := range ignoredExts {
		if ext == ignoredExt {
			return false
		}
	}
	
	// Only watch for write and create events
	return event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create
}