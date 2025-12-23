package interfaces

// ServerStatus represents the status of the Hugo development server
type ServerStatus struct {
	Running bool   `json:"running"`
	Port    int    `json:"port"`
	URL     string `json:"url"`
	PID     int    `json:"pid"`
}

// BuildResult represents the result of a Hugo build operation
type BuildResult struct {
	Success      bool     `json:"success"`
	OutputPath   string   `json:"output_path"`
	FilesWritten int      `json:"files_written"`
	Duration     string   `json:"duration"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
}

// HugoService interface defines Hugo-related operations
type HugoService interface {
	// StartServer starts the Hugo development server
	StartServer(projectPath string, port int) error
	
	// StopServer stops the Hugo development server
	StopServer() error
	
	// GetServerStatus returns the current status of the development server
	GetServerStatus() ServerStatus
	
	// BuildSite builds the Hugo site to the specified output path
	BuildSite(projectPath string, outputPath string) (*BuildResult, error)
	
	// IsHugoInstalled checks if Hugo is installed and accessible
	IsHugoInstalled() (bool, string, error)
	
	// GetHugoVersion returns the installed Hugo version
	GetHugoVersion() (string, error)
	
	// WatchFiles starts watching for file changes in the project
	WatchFiles(projectPath string, callback func(string)) error
	
	// StopWatching stops file watching
	StopWatching() error
}