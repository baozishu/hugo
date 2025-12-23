package interfaces

import (
	"context"
	"hugo-visual-client/internal/models"
)

// DeploymentManager interface defines deployment operations
type DeploymentManager interface {
	// LoadDeploymentConfig loads deployment configuration from project
	LoadDeploymentConfig(projectPath string) (*models.DeploymentConfig, error)
	
	// SaveDeploymentConfig saves deployment configuration to project
	SaveDeploymentConfig(projectPath string, config *models.DeploymentConfig) error
	
	// Deploy deploys the site to the specified target
	Deploy(ctx context.Context, projectPath string, targetName string) (*models.DeploymentStatus, error)
	
	// GetDeploymentStatus returns the current deployment status
	GetDeploymentStatus(targetName string) (*models.DeploymentStatus, error)
	
	// CancelDeployment cancels an ongoing deployment
	CancelDeployment(targetName string) error
	
	// TestConnection tests the connection to a deployment target
	TestConnection(target *models.DeploymentTarget) error
	
	// GetDeploymentHistory returns deployment history for a target
	GetDeploymentHistory(targetName string, limit int) ([]*models.DeploymentStatus, error)
	
	// ValidateTarget validates a deployment target configuration
	ValidateTarget(target *models.DeploymentTarget) error
	
	// IsDeploymentActive checks if a deployment is currently active
	IsDeploymentActive(targetName string) bool
	
	// GetDeploymentStatistics returns detailed deployment statistics
	GetDeploymentStatistics(targetName string) (*models.DeploymentStatistics, error)
	
	// GetAllDeploymentStatistics returns statistics for all active deployments
	GetAllDeploymentStatistics() map[string]*models.DeploymentStatistics
}

// DeploymentProgressCallback is called during deployment to report progress
type DeploymentProgressCallback func(status *models.DeploymentStatus)