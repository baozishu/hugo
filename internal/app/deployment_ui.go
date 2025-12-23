package app

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/interfaces"
	"hugo-visual-client/internal/models"
)

// DeploymentUI provides a user interface for deployment operations
type DeploymentUI struct {
	widget.BaseWidget
	
	// Dependencies
	deploymentManager interfaces.DeploymentManager
	projectPath       string
	
	// Configuration
	config *models.DeploymentConfig
	
	// UI components
	container       *container.Scroll
	targetsList     *widget.List
	deployButton    *widget.Button
	cancelButton    *widget.Button
	statusLabel     *widget.Label
	progressBar     *widget.ProgressBar
	logText         *widget.Entry
	
	// State
	selectedTarget  int
	isDeploying     bool
	deployContext   context.Context
	deployCancel    context.CancelFunc
	
	// Callbacks
	onDeployStart   func(string)
	onDeployComplete func(string, bool)
}

// NewDeploymentUI creates a new deployment UI
func NewDeploymentUI(deploymentManager interfaces.DeploymentManager, projectPath string) *DeploymentUI {
	ui := &DeploymentUI{
		deploymentManager: deploymentManager,
		projectPath:       projectPath,
		selectedTarget:    -1,
		isDeploying:       false,
	}
	
	ui.ExtendBaseWidget(ui)
	return ui
}

// LoadConfig loads the deployment configuration
func (dui *DeploymentUI) LoadConfig() error {
	config, err := dui.deploymentManager.LoadDeploymentConfig(dui.projectPath)
	if err != nil {
		return fmt.Errorf("failed to load deployment config: %w", err)
	}
	
	dui.config = config
	dui.refreshTargetsList()
	
	return nil
}

// CreateRenderer creates the visual representation of the deployment UI
func (dui *DeploymentUI) CreateRenderer() fyne.WidgetRenderer {
	dui.setupUI()
	return widget.NewSimpleRenderer(dui.container)
}

// setupUI creates the user interface components
func (dui *DeploymentUI) setupUI() {
	// Create targets list
	dui.targetsList = widget.NewList(
		func() int {
			if dui.config == nil {
				return 0
			}
			return len(dui.config.Targets)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.ComputerIcon()),
				widget.NewLabel("Target Name"),
				widget.NewLabel("(Type)"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if dui.config == nil || id >= len(dui.config.Targets) {
				return
			}
			
			target := dui.config.Targets[id]
			container := obj.(*fyne.Container)
			
			// Update icon based on type
			icon := container.Objects[0].(*widget.Icon)
			switch target.Type {
			case "ftp", "sftp":
				icon.SetResource(theme.ComputerIcon())
			case "s3":
				icon.SetResource(theme.StorageIcon())
			case "github":
				icon.SetResource(theme.InfoIcon())
			case "netlify", "vercel":
				icon.SetResource(theme.MailSendIcon())
			default:
				icon.SetResource(theme.ComputerIcon())
			}
			
			// Update labels
			nameLabel := container.Objects[1].(*widget.Label)
			typeLabel := container.Objects[2].(*widget.Label)
			
			nameLabel.SetText(target.Name)
			typeLabel.SetText(fmt.Sprintf("(%s)", target.Type))
			
			// Highlight default target
			if target.IsDefault {
				nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				nameLabel.TextStyle = fyne.TextStyle{}
			}
		},
	)
	
	dui.targetsList.OnSelected = func(id widget.ListItemID) {
		dui.selectTarget(id)
	}
	
	// Create deployment controls
	dui.deployButton = widget.NewButtonWithIcon("Deploy", theme.MailSendIcon(), func() {
		dui.startDeployment()
	})
	dui.deployButton.Importance = widget.HighImportance
	dui.deployButton.Disable()
	
	dui.cancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		dui.cancelDeployment()
	})
	dui.cancelButton.Disable()
	
	// Create status display
	dui.statusLabel = widget.NewLabel("Select a target to deploy")
	dui.statusLabel.Alignment = fyne.TextAlignCenter
	
	dui.progressBar = widget.NewProgressBar()
	dui.progressBar.Hide()
	
	// Create log display
	dui.logText = widget.NewMultiLineEntry()
	dui.logText.SetText("Deployment log will appear here...")
	dui.logText.Disable()
	
	// Create control buttons
	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		dui.refreshStatus()
	})
	
	configButton := widget.NewButtonWithIcon("Configure", theme.SettingsIcon(), func() {
		dui.showConfigDialog()
	})
	
	historyButton := widget.NewButtonWithIcon("History", theme.HistoryIcon(), func() {
		dui.showHistoryDialog()
	})
	
	controlButtons := container.NewHBox(
		refreshButton,
		configButton,
		historyButton,
	)
	
	deploymentButtons := container.NewHBox(
		dui.deployButton,
		dui.cancelButton,
	)
	
	// Create targets panel
	targetsPanel := container.NewVBox(
		widget.NewLabel("Deployment Targets"),
		dui.targetsList,
		controlButtons,
	)
	
	// Create status panel
	statusPanel := container.NewVBox(
		widget.NewLabel("Deployment Status"),
		dui.statusLabel,
		dui.progressBar,
		deploymentButtons,
	)
	
	// Create log panel
	logPanel := container.NewVBox(
		widget.NewLabel("Deployment Log"),
		container.NewScroll(dui.logText),
	)
	
	// Create main layout
	leftPanel := container.NewVBox(targetsPanel, statusPanel)
	mainSplit := container.NewHSplit(leftPanel, logPanel)
	mainSplit.SetOffset(0.4) // 40% for left panel
	
	dui.container = container.NewScroll(mainSplit)
	dui.container.SetMinSize(fyne.NewSize(800, 600))
}

// refreshTargetsList refreshes the targets list display
func (dui *DeploymentUI) refreshTargetsList() {
	if dui.targetsList != nil {
		dui.targetsList.Refresh()
	}
	
	// Update button states
	// hasTargets := dui.config != nil && len(dui.config.Targets) > 0
	hasSelection := dui.selectedTarget >= 0
	
	if dui.deployButton != nil {
		if hasSelection && !dui.isDeploying {
			dui.deployButton.Enable()
		} else {
			dui.deployButton.Disable()
		}
	}
	
	if dui.cancelButton != nil {
		if dui.isDeploying {
			dui.cancelButton.Enable()
		} else {
			dui.cancelButton.Disable()
		}
	}
}

// selectTarget selects a target for deployment
func (dui *DeploymentUI) selectTarget(index int) {
	if dui.config == nil || index < 0 || index >= len(dui.config.Targets) {
		dui.selectedTarget = -1
		dui.statusLabel.SetText("Select a target to deploy")
		dui.refreshTargetsList()
		return
	}
	
	dui.selectedTarget = index
	target := dui.config.Targets[index]
	dui.statusLabel.SetText(fmt.Sprintf("Ready to deploy to %s (%s)", target.Name, target.Type))
	dui.refreshTargetsList()
}

// startDeployment starts the deployment process
func (dui *DeploymentUI) startDeployment() {
	if dui.selectedTarget < 0 || dui.config == nil {
		return
	}
	
	target := dui.config.Targets[dui.selectedTarget]
	
	// Confirm deployment
	dialog.ShowConfirm("Deploy Site", 
		fmt.Sprintf("Are you sure you want to deploy to %s (%s)?", target.Name, target.Type),
		func(confirmed bool) {
			if confirmed {
				dui.performDeployment(target.Name)
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// performDeployment performs the actual deployment
func (dui *DeploymentUI) performDeployment(targetName string) {
	// Set deploying state
	dui.isDeploying = true
	dui.refreshTargetsList()
	
	// Show progress bar
	dui.progressBar.Show()
	dui.progressBar.SetValue(0)
	
	// Clear log
	dui.logText.SetText("")
	dui.addLogMessage("Starting deployment to " + targetName + "...")
	
	// Create deployment context
	dui.deployContext, dui.deployCancel = context.WithCancel(context.Background())
	
	// Call deployment start callback
	if dui.onDeployStart != nil {
		dui.onDeployStart(targetName)
	}
	
	// Start deployment
	status, err := dui.deploymentManager.Deploy(dui.deployContext, dui.projectPath, targetName)
	if err != nil {
		dui.addLogMessage(fmt.Sprintf("Failed to start deployment: %v", err))
		dui.finishDeployment(targetName, false)
		return
	}
	
	// Monitor deployment progress
	go dui.monitorDeployment(targetName, status)
}

// monitorDeployment monitors the deployment progress
func (dui *DeploymentUI) monitorDeployment(targetName string, initialStatus *models.DeploymentStatus) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-dui.deployContext.Done():
			dui.addLogMessage("Deployment cancelled")
			dui.finishDeployment(targetName, false)
			return
			
		case <-ticker.C:
			status, err := dui.deploymentManager.GetDeploymentStatus(targetName)
			if err != nil {
				dui.addLogMessage(fmt.Sprintf("Error getting deployment status: %v", err))
				continue
			}
			
			// Update UI
			dui.updateDeploymentUI(status)
			
			// Check if deployment is complete
			if status.IsComplete() {
				success := status.Status == "success"
				if success {
					dui.addLogMessage("Deployment completed successfully!")
				} else {
					dui.addLogMessage(fmt.Sprintf("Deployment failed: %s", status.Error))
				}
				dui.finishDeployment(targetName, success)
				return
			}
		}
	}
}

// updateDeploymentUI updates the UI based on deployment status
func (dui *DeploymentUI) updateDeploymentUI(status *models.DeploymentStatus) {
	// Update status label
	dui.statusLabel.SetText(status.Message)
	
	// Update progress bar
	dui.progressBar.SetValue(status.Progress)
	
	// Add log message if status changed
	if status.Message != "" {
		dui.addLogMessage(status.Message)
	}
}

// finishDeployment finishes the deployment process
func (dui *DeploymentUI) finishDeployment(targetName string, success bool) {
	// Reset deploying state
	dui.isDeploying = false
	dui.deployCancel = nil
	dui.deployContext = nil
	
	// Hide progress bar
	dui.progressBar.Hide()
	
	// Update status
	if success {
		dui.statusLabel.SetText(fmt.Sprintf("Successfully deployed to %s", targetName))
	} else {
		dui.statusLabel.SetText(fmt.Sprintf("Deployment to %s failed", targetName))
	}
	
	// Refresh UI
	dui.refreshTargetsList()
	
	// Call completion callback
	if dui.onDeployComplete != nil {
		dui.onDeployComplete(targetName, success)
	}
	
	// Show completion dialog
	if success {
		dialog.ShowInformation("Deployment Complete", 
			fmt.Sprintf("Successfully deployed to %s!", targetName), 
			fyne.CurrentApp().Driver().AllWindows()[0])
	} else {
		dialog.ShowError(fmt.Errorf("deployment to %s failed", targetName), 
			fyne.CurrentApp().Driver().AllWindows()[0])
	}
}

// cancelDeployment cancels the current deployment
func (dui *DeploymentUI) cancelDeployment() {
	if !dui.isDeploying || dui.deployCancel == nil {
		return
	}
	
	dialog.ShowConfirm("Cancel Deployment", 
		"Are you sure you want to cancel the current deployment?",
		func(confirmed bool) {
			if confirmed && dui.deployCancel != nil {
				dui.deployCancel()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// addLogMessage adds a message to the deployment log
func (dui *DeploymentUI) addLogMessage(message string) {
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	currentText := dui.logText.Text
	dui.logText.SetText(currentText + logEntry)
	
	// Scroll to bottom
	dui.logText.CursorRow = len(dui.logText.Text)
}

// refreshStatus refreshes the deployment status
func (dui *DeploymentUI) refreshStatus() {
	if dui.selectedTarget < 0 || dui.config == nil {
		return
	}
	
	target := dui.config.Targets[dui.selectedTarget]
	
	// Check if there's an active deployment
	if dui.deploymentManager.IsDeploymentActive(target.Name) {
		status, err := dui.deploymentManager.GetDeploymentStatus(target.Name)
		if err == nil {
			dui.updateDeploymentUI(status)
		}
	}
	
	dui.addLogMessage("Status refreshed")
}

// showConfigDialog shows the deployment configuration dialog
func (dui *DeploymentUI) showConfigDialog() {
	// Create a new deployment config editor
	configEditor := NewDeploymentConfigEditor(dui.deploymentManager, dui.projectPath)
	
	// Load current config
	if err := configEditor.LoadConfig(); err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	// Show in a dialog
	dialog.ShowCustom("Deployment Configuration", "Close", configEditor.GetWidget(), 
		fyne.CurrentApp().Driver().AllWindows()[0])
}

// showHistoryDialog shows the deployment history dialog
func (dui *DeploymentUI) showHistoryDialog() {
	if dui.selectedTarget < 0 || dui.config == nil {
		dialog.ShowInformation("No Target Selected", 
			"Please select a deployment target to view its history.", 
			fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	target := dui.config.Targets[dui.selectedTarget]
	
	// Get deployment history
	history, err := dui.deploymentManager.GetDeploymentHistory(target.Name, 10)
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	// Create history list
	historyList := widget.NewList(
		func() int { return len(history) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Status"),
				widget.NewLabel("Time"),
				widget.NewLabel("Message"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(history) {
				return
			}
			
			status := history[id]
			container := obj.(*fyne.Container)
			
			statusLabel := container.Objects[0].(*widget.Label)
			timeLabel := container.Objects[1].(*widget.Label)
			messageLabel := container.Objects[2].(*widget.Label)
			
			statusLabel.SetText(status.Status)
			timeLabel.SetText(status.StartTime.Format("2006-01-02 15:04"))
			messageLabel.SetText(status.Message)
			
			// Color code status
			switch status.Status {
			case "success":
				statusLabel.Importance = widget.SuccessImportance
			case "failed":
				statusLabel.Importance = widget.DangerImportance
			default:
				statusLabel.Importance = widget.MediumImportance
			}
		},
	)
	
	historyContent := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Deployment History for %s", target.Name)),
		container.NewScroll(historyList),
	)
	
	dialog.ShowCustom("Deployment History", "Close", historyContent, 
		fyne.CurrentApp().Driver().AllWindows()[0])
}

// Callback setters
func (dui *DeploymentUI) SetOnDeployStart(callback func(string)) {
	dui.onDeployStart = callback
}

func (dui *DeploymentUI) SetOnDeployComplete(callback func(string, bool)) {
	dui.onDeployComplete = callback
}

// GetWidget returns the widget for embedding in other containers
func (dui *DeploymentUI) GetWidget() fyne.CanvasObject {
	return dui
}

// GetSelectedTarget returns the currently selected target
func (dui *DeploymentUI) GetSelectedTarget() *models.DeploymentTarget {
	if dui.selectedTarget < 0 || dui.config == nil || dui.selectedTarget >= len(dui.config.Targets) {
		return nil
	}
	
	return &dui.config.Targets[dui.selectedTarget]
}

// IsDeploying returns whether a deployment is currently in progress
func (dui *DeploymentUI) IsDeploying() bool {
	return dui.isDeploying
}