package models

import (
	"fmt"
	"strings"
	"time"
)

// DeploymentTarget represents a deployment target configuration
type DeploymentTarget struct {
	Name        string                 `json:"name" yaml:"name"`
	Type        string                 `json:"type" yaml:"type"` // "ftp", "sftp", "s3", "github", "netlify", "vercel"
	URL         string                 `json:"url" yaml:"url"`
	Username    string                 `json:"username" yaml:"username"`
	Password    string                 `json:"password,omitempty" yaml:"password,omitempty"`
	Token       string                 `json:"token,omitempty" yaml:"token,omitempty"`
	Region      string                 `json:"region,omitempty" yaml:"region,omitempty"`
	Bucket      string                 `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Path        string                 `json:"path" yaml:"path"`
	Port        int                    `json:"port,omitempty" yaml:"port,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	LastDeploy  time.Time              `json:"last_deploy" yaml:"last_deploy"`
	IsDefault   bool                   `json:"is_default" yaml:"is_default"`
}

// Validate checks if the deployment target configuration is valid
func (dt *DeploymentTarget) Validate() error {
	if strings.TrimSpace(dt.Name) == "" {
		return fmt.Errorf("deployment target name cannot be empty")
	}
	
	if strings.TrimSpace(dt.Type) == "" {
		return fmt.Errorf("deployment target type cannot be empty")
	}
	
	// Validate based on deployment type
	switch dt.Type {
	case "ftp", "sftp":
		if strings.TrimSpace(dt.URL) == "" {
			return fmt.Errorf("URL is required for %s deployment", dt.Type)
		}
		if strings.TrimSpace(dt.Username) == "" {
			return fmt.Errorf("username is required for %s deployment", dt.Type)
		}
		if dt.Port <= 0 {
			if dt.Type == "ftp" {
				dt.Port = 21
			} else {
				dt.Port = 22
			}
		}
		
	case "s3":
		if strings.TrimSpace(dt.Bucket) == "" {
			return fmt.Errorf("bucket is required for S3 deployment")
		}
		if strings.TrimSpace(dt.Region) == "" {
			dt.Region = "us-east-1" // Default region
		}
		
	case "github":
		if strings.TrimSpace(dt.Token) == "" {
			return fmt.Errorf("token is required for GitHub deployment")
		}
		if strings.TrimSpace(dt.URL) == "" {
			return fmt.Errorf("repository URL is required for GitHub deployment")
		}
		
	case "netlify", "vercel":
		if strings.TrimSpace(dt.Token) == "" {
			return fmt.Errorf("token is required for %s deployment", dt.Type)
		}
		
	default:
		return fmt.Errorf("unsupported deployment type: %s", dt.Type)
	}
	
	return nil
}

// GetDisplayName returns a user-friendly display name for the deployment target
func (dt *DeploymentTarget) GetDisplayName() string {
	if dt.Name != "" {
		return dt.Name
	}
	return fmt.Sprintf("%s (%s)", dt.Type, dt.URL)
}

// DeploymentConfig represents the overall deployment configuration for a project
type DeploymentConfig struct {
	Targets       []DeploymentTarget `json:"targets" yaml:"targets"`
	DefaultTarget string             `json:"default_target" yaml:"default_target"`
	BuildCommand  string             `json:"build_command" yaml:"build_command"`
	OutputDir     string             `json:"output_dir" yaml:"output_dir"`
	ExcludeFiles  []string           `json:"exclude_files" yaml:"exclude_files"`
	IncludeFiles  []string           `json:"include_files" yaml:"include_files"`
}

// Validate checks if the deployment configuration is valid
func (dc *DeploymentConfig) Validate() error {
	if len(dc.Targets) == 0 {
		return fmt.Errorf("at least one deployment target must be configured")
	}
	
	// Validate each target
	for i, target := range dc.Targets {
		if err := target.Validate(); err != nil {
			return fmt.Errorf("target %d validation failed: %w", i, err)
		}
	}
	
	// Check if default target exists
	if dc.DefaultTarget != "" {
		found := false
		for _, target := range dc.Targets {
			if target.Name == dc.DefaultTarget {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default target '%s' not found in targets list", dc.DefaultTarget)
		}
	}
	
	// Set default values
	if dc.BuildCommand == "" {
		dc.BuildCommand = "hugo"
	}
	if dc.OutputDir == "" {
		dc.OutputDir = "public"
	}
	
	return nil
}

// GetDefaultTarget returns the default deployment target
func (dc *DeploymentConfig) GetDefaultTarget() *DeploymentTarget {
	if dc.DefaultTarget == "" && len(dc.Targets) > 0 {
		return &dc.Targets[0]
	}
	
	for _, target := range dc.Targets {
		if target.Name == dc.DefaultTarget {
			return &target
		}
	}
	
	return nil
}

// AddTarget adds a new deployment target
func (dc *DeploymentConfig) AddTarget(target DeploymentTarget) error {
	// Check for duplicate names
	for _, existing := range dc.Targets {
		if existing.Name == target.Name {
			return fmt.Errorf("deployment target with name '%s' already exists", target.Name)
		}
	}
	
	// Validate the target
	if err := target.Validate(); err != nil {
		return fmt.Errorf("invalid deployment target: %w", err)
	}
	
	// If this is the first target or marked as default, make it the default
	if len(dc.Targets) == 0 || target.IsDefault {
		dc.DefaultTarget = target.Name
		target.IsDefault = true
		
		// Unmark other targets as default
		for i := range dc.Targets {
			dc.Targets[i].IsDefault = false
		}
	}
	
	dc.Targets = append(dc.Targets, target)
	return nil
}

// RemoveTarget removes a deployment target by name
func (dc *DeploymentConfig) RemoveTarget(name string) error {
	for i, target := range dc.Targets {
		if target.Name == name {
			// Remove the target
			dc.Targets = append(dc.Targets[:i], dc.Targets[i+1:]...)
			
			// If this was the default target, clear the default
			if dc.DefaultTarget == name {
				dc.DefaultTarget = ""
				if len(dc.Targets) > 0 {
					dc.DefaultTarget = dc.Targets[0].Name
					dc.Targets[0].IsDefault = true
				}
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("deployment target '%s' not found", name)
}

// UpdateTarget updates an existing deployment target
func (dc *DeploymentConfig) UpdateTarget(name string, updatedTarget DeploymentTarget) error {
	for i, target := range dc.Targets {
		if target.Name == name {
			// Validate the updated target
			if err := updatedTarget.Validate(); err != nil {
				return fmt.Errorf("invalid deployment target: %w", err)
			}
			
			// Update the target
			dc.Targets[i] = updatedTarget
			
			// Handle default target changes
			if updatedTarget.IsDefault {
				dc.DefaultTarget = updatedTarget.Name
				// Unmark other targets as default
				for j := range dc.Targets {
					if j != i {
						dc.Targets[j].IsDefault = false
					}
				}
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("deployment target '%s' not found", name)
}

// DeploymentStatus represents the status of a deployment operation
type DeploymentStatus struct {
	Target      string    `json:"target"`
	Status      string    `json:"status"` // "pending", "running", "success", "failed"
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Message     string    `json:"message"`
	Error       string    `json:"error,omitempty"`
	FilesCount  int       `json:"files_count"`
	BytesTotal  int64     `json:"bytes_total"`
	Progress    float64   `json:"progress"` // 0.0 to 1.0
}

// IsRunning returns true if the deployment is currently running
func (ds *DeploymentStatus) IsRunning() bool {
	return ds.Status == "running" || ds.Status == "pending"
}

// IsComplete returns true if the deployment has completed (success or failure)
func (ds *DeploymentStatus) IsComplete() bool {
	return ds.Status == "success" || ds.Status == "failed"
}

// Duration returns the duration of the deployment
func (ds *DeploymentStatus) Duration() time.Duration {
	if ds.EndTime.IsZero() {
		if ds.StartTime.IsZero() {
			return 0
		}
		return time.Since(ds.StartTime)
	}
	return ds.EndTime.Sub(ds.StartTime)
}

// DeploymentStatistics provides detailed statistics about a deployment
type DeploymentStatistics struct {
	Target            string        `json:"target"`
	Status            string        `json:"status"`
	FilesCount        int           `json:"files_count"`
	BytesTotal        int64         `json:"bytes_total"`
	Progress          float64       `json:"progress"`
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`
	Duration          time.Duration `json:"duration"`
	TransferRate      float64       `json:"transfer_rate"`      // bytes per second
	EstimatedTimeLeft time.Duration `json:"estimated_time_left"`
}

// GetPresetConfigurations returns common deployment target presets
func GetPresetConfigurations() map[string]DeploymentTarget {
	return map[string]DeploymentTarget{
		"netlify": {
			Name: "Netlify",
			Type: "netlify",
			URL:  "https://api.netlify.com",
			Path: "/",
			Config: map[string]interface{}{
				"site_id": "",
			},
		},
		"vercel": {
			Name: "Vercel",
			Type: "vercel",
			URL:  "https://api.vercel.com",
			Path: "/",
			Config: map[string]interface{}{
				"project_id": "",
			},
		},
		"github_pages": {
			Name: "GitHub Pages",
			Type: "github",
			URL:  "",
			Path: "/",
			Config: map[string]interface{}{
				"branch": "gh-pages",
			},
		},
		"aws_s3": {
			Name:   "AWS S3",
			Type:   "s3",
			Region: "us-east-1",
			Path:   "/",
			Config: map[string]interface{}{
				"cloudfront_distribution": "",
			},
		},
		"ftp": {
			Name: "FTP Server",
			Type: "ftp",
			Port: 21,
			Path: "/public_html",
		},
		"sftp": {
			Name: "SFTP Server",
			Type: "sftp",
			Port: 22,
			Path: "/var/www/html",
		},
	}
}