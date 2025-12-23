package app

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/interfaces"
	"hugo-visual-client/internal/models"
)

// DeploymentConfigEditor provides a visual interface for editing deployment configuration
type DeploymentConfigEditor struct {
	widget.BaseWidget
	
	// Dependencies
	deploymentManager interfaces.DeploymentManager
	projectPath       string
	
	// Configuration data
	config *models.DeploymentConfig
	
	// UI components
	container       *container.Scroll
	targetsList     *widget.List
	targetDetails   *container.Split
	addButton       *widget.Button
	removeButton    *widget.Button
	testButton      *widget.Button
	saveButton      *widget.Button
	resetButton     *widget.Button
	
	// Target editor components
	targetEditor    *DeploymentTargetEditor
	
	// State
	selectedTarget  int
	isModified      bool
	
	// Callbacks
	onSave   func(*models.DeploymentConfig) error
	onReset  func()
}

// DeploymentTargetEditor provides an interface for editing individual deployment targets
type DeploymentTargetEditor struct {
	widget.BaseWidget
	
	// Target data
	target *models.DeploymentTarget
	
	// Data bindings
	nameBinding     binding.String
	typeBinding     binding.String
	urlBinding      binding.String
	usernameBinding binding.String
	passwordBinding binding.String
	tokenBinding    binding.String
	regionBinding   binding.String
	bucketBinding   binding.String
	pathBinding     binding.String
	portBinding     binding.String
	
	// UI components
	container    *fyne.Container
	form         *widget.Form
	typeSelect   *widget.Select
	presetSelect *widget.Select
	
	// Callbacks
	onChanged func()
}

// NewDeploymentConfigEditor creates a new deployment configuration editor
func NewDeploymentConfigEditor(deploymentManager interfaces.DeploymentManager, projectPath string) *DeploymentConfigEditor {
	editor := &DeploymentConfigEditor{
		deploymentManager: deploymentManager,
		projectPath:       projectPath,
		selectedTarget:    -1,
		isModified:        false,
	}
	
	editor.ExtendBaseWidget(editor)
	return editor
}

// LoadConfig loads deployment configuration from project
func (dce *DeploymentConfigEditor) LoadConfig() error {
	config, err := dce.deploymentManager.LoadDeploymentConfig(dce.projectPath)
	if err != nil {
		// Create default config if none exists
		config = &models.DeploymentConfig{
			Targets:      []models.DeploymentTarget{},
			BuildCommand: "hugo",
			OutputDir:    "public",
		}
	}
	
	dce.config = config
	dce.isModified = false
	dce.refreshTargetsList()
	
	return nil
}

// SaveConfig saves the current deployment configuration
func (dce *DeploymentConfigEditor) SaveConfig() error {
	if dce.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Update current target from editor
	if dce.selectedTarget >= 0 && dce.targetEditor != nil {
		updatedTarget := dce.targetEditor.GetTarget()
		if updatedTarget != nil {
			dce.config.Targets[dce.selectedTarget] = *updatedTarget
		}
	}
	
	// Validate configuration
	if err := dce.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Save to project
	err := dce.deploymentManager.SaveDeploymentConfig(dce.projectPath, dce.config)
	if err != nil {
		return fmt.Errorf("failed to save deployment configuration: %w", err)
	}
	
	dce.isModified = false
	
	// Call save callback if set
	if dce.onSave != nil {
		return dce.onSave(dce.config)
	}
	
	return nil
}

// CreateRenderer creates the visual representation of the deployment config editor
func (dce *DeploymentConfigEditor) CreateRenderer() fyne.WidgetRenderer {
	dce.setupUI()
	return widget.NewSimpleRenderer(dce.container)
}

// setupUI creates the user interface components
func (dce *DeploymentConfigEditor) setupUI() {
	// Create targets list
	dce.targetsList = widget.NewList(
		func() int {
			if dce.config == nil {
				return 0
			}
			return len(dce.config.Targets)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.ComputerIcon()),
				widget.NewLabel("Target Name"),
				widget.NewLabel("(Type)"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if dce.config == nil || id >= len(dce.config.Targets) {
				return
			}
			
			target := dce.config.Targets[id]
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
			typeLabel.SetText(fmt.Sprintf("(%s)", strings.ToUpper(target.Type)))
			
			// Highlight default target
			if target.IsDefault {
				nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				nameLabel.TextStyle = fyne.TextStyle{}
			}
		},
	)
	
	dce.targetsList.OnSelected = func(id widget.ListItemID) {
		dce.selectTarget(id)
	}
	
	// Create target editor
	dce.targetEditor = NewDeploymentTargetEditor()
	dce.targetEditor.SetOnChanged(func() {
		dce.markModified()
	})
	
	// Create action buttons
	dce.addButton = widget.NewButtonWithIcon("Add Target", theme.ContentAddIcon(), func() {
		dce.showAddTargetDialog()
	})
	
	dce.removeButton = widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), func() {
		dce.removeSelectedTarget()
	})
	dce.removeButton.Disable()
	
	dce.testButton = widget.NewButtonWithIcon("Test Connection", theme.ConfirmIcon(), func() {
		dce.testSelectedTarget()
	})
	dce.testButton.Disable()
	
	dce.saveButton = widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), func() {
		dce.handleSave()
	})
	dce.saveButton.Importance = widget.HighImportance
	
	dce.resetButton = widget.NewButtonWithIcon("Reset Changes", theme.ViewRefreshIcon(), func() {
		dce.handleReset()
	})
	
	// Create button containers
	targetButtons := container.NewHBox(
		dce.addButton,
		dce.removeButton,
		dce.testButton,
	)
	
	configButtons := container.NewHBox(
		dce.saveButton,
		dce.resetButton,
	)
	
	// Create targets panel
	targetsPanel := container.NewVBox(
		widget.NewLabel("Deployment Targets"),
		dce.targetsList,
		targetButtons,
	)
	
	// Create details panel
	detailsPanel := container.NewVBox(
		widget.NewLabel("Target Configuration"),
		dce.targetEditor,
	)
	
	// Create split container
	dce.targetDetails = container.NewHSplit(targetsPanel, detailsPanel)
	dce.targetDetails.SetOffset(0.3) // 30% for targets list
	
	// Create main content
	content := container.NewVBox(
		dce.targetDetails,
		widget.NewSeparator(),
		configButtons,
	)
	
	dce.container = container.NewScroll(content)
	dce.container.SetMinSize(fyne.NewSize(800, 600))
}

// refreshTargetsList refreshes the targets list display
func (dce *DeploymentConfigEditor) refreshTargetsList() {
	if dce.targetsList != nil {
		dce.targetsList.Refresh()
	}
	
	// Update button states
	// hasTargets := dce.config != nil && len(dce.config.Targets) > 0
	hasSelection := dce.selectedTarget >= 0
	
	if dce.removeButton != nil {
		if hasSelection {
			dce.removeButton.Enable()
		} else {
			dce.removeButton.Disable()
		}
	}
	
	if dce.testButton != nil {
		if hasSelection {
			dce.testButton.Enable()
		} else {
			dce.testButton.Disable()
		}
	}
}

// selectTarget selects a target for editing
func (dce *DeploymentConfigEditor) selectTarget(index int) {
	if dce.config == nil || index < 0 || index >= len(dce.config.Targets) {
		dce.selectedTarget = -1
		dce.targetEditor.SetTarget(nil)
		dce.refreshTargetsList()
		return
	}
	
	dce.selectedTarget = index
	target := dce.config.Targets[index]
	dce.targetEditor.SetTarget(&target)
	dce.refreshTargetsList()
}

// showAddTargetDialog shows a dialog to add a new deployment target
func (dce *DeploymentConfigEditor) showAddTargetDialog() {
	presets := models.GetPresetConfigurations()
	presetNames := make([]string, 0, len(presets))
	presetNames = append(presetNames, "Custom")
	
	for name := range presets {
		presetNames = append(presetNames, name)
	}
	
	presetSelect := widget.NewSelect(presetNames, nil)
	presetSelect.SetSelected("Custom")
	
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter target name...")
	
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Preset:", Widget: presetSelect},
			{Text: "Name:", Widget: nameEntry},
		},
	}
	
	dialog.ShowForm("Add Deployment Target", "Add", "Cancel", form.Items, func(confirmed bool) {
		if confirmed && nameEntry.Text != "" {
			var newTarget models.DeploymentTarget
			
			if presetSelect.Selected != "Custom" {
				if preset, exists := presets[presetSelect.Selected]; exists {
					newTarget = preset
				}
			}
			
			newTarget.Name = strings.TrimSpace(nameEntry.Text)
			
			// Add the target
			err := dce.config.AddTarget(newTarget)
			if err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
			
			dce.markModified()
			dce.refreshTargetsList()
			
			// Select the new target
			dce.selectTarget(len(dce.config.Targets) - 1)
			dce.targetsList.Select(len(dce.config.Targets) - 1)
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// removeSelectedTarget removes the currently selected target
func (dce *DeploymentConfigEditor) removeSelectedTarget() {
	if dce.selectedTarget < 0 || dce.config == nil {
		return
	}
	
	target := dce.config.Targets[dce.selectedTarget]
	
	dialog.ShowConfirm("Remove Target", 
		fmt.Sprintf("Are you sure you want to remove the deployment target '%s'?", target.Name),
		func(confirmed bool) {
			if confirmed {
				err := dce.config.RemoveTarget(target.Name)
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
					return
				}
				
				dce.markModified()
				dce.selectedTarget = -1
				dce.targetEditor.SetTarget(nil)
				dce.refreshTargetsList()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// testSelectedTarget tests the connection to the selected target
func (dce *DeploymentConfigEditor) testSelectedTarget() {
	if dce.selectedTarget < 0 || dce.config == nil {
		return
	}
	
	target := dce.config.Targets[dce.selectedTarget]
	
	// Update target from editor
	if dce.targetEditor != nil {
		updatedTarget := dce.targetEditor.GetTarget()
		if updatedTarget != nil {
			target = *updatedTarget
		}
	}
	
	// Show progress dialog
	progress := dialog.NewProgressInfinite("Testing Connection", 
		fmt.Sprintf("Testing connection to %s...", target.Name), 
		fyne.CurrentApp().Driver().AllWindows()[0])
	progress.Show()
	
	// Test connection in background
	go func() {
		err := dce.deploymentManager.TestConnection(&target)
		progress.Hide()
		
		if err != nil {
			dialog.ShowError(fmt.Errorf("connection test failed: %w", err), 
				fyne.CurrentApp().Driver().AllWindows()[0])
		} else {
			dialog.ShowInformation("Connection Test", 
				fmt.Sprintf("Successfully connected to %s!", target.Name), 
				fyne.CurrentApp().Driver().AllWindows()[0])
		}
	}()
}

// Event handlers
func (dce *DeploymentConfigEditor) handleSave() {
	err := dce.SaveConfig()
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	dialog.ShowInformation("Success", "Deployment configuration saved successfully!", 
		fyne.CurrentApp().Driver().AllWindows()[0])
}

func (dce *DeploymentConfigEditor) handleReset() {
	dialog.ShowConfirm("Reset Configuration", 
		"Are you sure you want to reset all changes? This will discard any unsaved modifications.",
		func(confirmed bool) {
			if confirmed {
				err := dce.LoadConfig()
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				}
				
				dce.selectedTarget = -1
				dce.targetEditor.SetTarget(nil)
				dce.refreshTargetsList()
				
				if dce.onReset != nil {
					dce.onReset()
				}
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// Utility methods
func (dce *DeploymentConfigEditor) markModified() {
	dce.isModified = true
}

func (dce *DeploymentConfigEditor) IsModified() bool {
	return dce.isModified
}

// Callback setters
func (dce *DeploymentConfigEditor) SetOnSave(callback func(*models.DeploymentConfig) error) {
	dce.onSave = callback
}

func (dce *DeploymentConfigEditor) SetOnReset(callback func()) {
	dce.onReset = callback
}

// GetWidget returns the widget for embedding in other containers
func (dce *DeploymentConfigEditor) GetWidget() fyne.CanvasObject {
	return dce
}

// GetConfig returns the current deployment configuration
func (dce *DeploymentConfigEditor) GetConfig() *models.DeploymentConfig {
	return dce.config
}

// NewDeploymentTargetEditor creates a new deployment target editor
func NewDeploymentTargetEditor() *DeploymentTargetEditor {
	editor := &DeploymentTargetEditor{}
	
	// Initialize data bindings
	editor.nameBinding = binding.NewString()
	editor.typeBinding = binding.NewString()
	editor.urlBinding = binding.NewString()
	editor.usernameBinding = binding.NewString()
	editor.passwordBinding = binding.NewString()
	editor.tokenBinding = binding.NewString()
	editor.regionBinding = binding.NewString()
	editor.bucketBinding = binding.NewString()
	editor.pathBinding = binding.NewString()
	editor.portBinding = binding.NewString()
	
	// Set up change listeners
	editor.setupBindingListeners()
	
	editor.ExtendBaseWidget(editor)
	return editor
}

// CreateRenderer creates the visual representation of the target editor
func (dte *DeploymentTargetEditor) CreateRenderer() fyne.WidgetRenderer {
	dte.setupUI()
	return widget.NewSimpleRenderer(dte.container)
}

// setupUI creates the user interface for the target editor
func (dte *DeploymentTargetEditor) setupUI() {
	// Create form for target configuration
	dte.form = &widget.Form{}
	
	// Basic information
	nameEntry := widget.NewEntryWithData(dte.nameBinding)
	nameEntry.SetPlaceHolder("Target name")
	dte.form.Append("Name", nameEntry)
	
	// Type selection
	dte.typeSelect = widget.NewSelect([]string{"ftp", "sftp", "s3", "github", "netlify", "vercel"}, func(selected string) {
		dte.typeBinding.Set(selected)
		dte.updateFormForType(selected)
	})
	dte.form.Append("Type", dte.typeSelect)
	
	// Preset selection
	presets := models.GetPresetConfigurations()
	presetNames := make([]string, 0, len(presets))
	presetNames = append(presetNames, "Custom")
	for name := range presets {
		presetNames = append(presetNames, name)
	}
	
	dte.presetSelect = widget.NewSelect(presetNames, func(selected string) {
		if selected != "Custom" {
			if preset, exists := presets[selected]; exists {
				dte.loadPreset(preset)
			}
		}
	})
	dte.form.Append("Preset", dte.presetSelect)
	
	// Connection details
	urlEntry := widget.NewEntryWithData(dte.urlBinding)
	urlEntry.SetPlaceHolder("Server URL or endpoint")
	dte.form.Append("URL", urlEntry)
	
	usernameEntry := widget.NewEntryWithData(dte.usernameBinding)
	usernameEntry.SetPlaceHolder("Username")
	dte.form.Append("Username", usernameEntry)
	
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.Bind(dte.passwordBinding)
	passwordEntry.SetPlaceHolder("Password")
	dte.form.Append("Password", passwordEntry)
	
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.Bind(dte.tokenBinding)
	tokenEntry.SetPlaceHolder("API Token")
	dte.form.Append("Token", tokenEntry)
	
	regionEntry := widget.NewEntryWithData(dte.regionBinding)
	regionEntry.SetPlaceHolder("Region (e.g., us-east-1)")
	dte.form.Append("Region", regionEntry)
	
	bucketEntry := widget.NewEntryWithData(dte.bucketBinding)
	bucketEntry.SetPlaceHolder("Bucket name")
	dte.form.Append("Bucket", bucketEntry)
	
	pathEntry := widget.NewEntryWithData(dte.pathBinding)
	pathEntry.SetPlaceHolder("Remote path")
	dte.form.Append("Path", pathEntry)
	
	portEntry := widget.NewEntryWithData(dte.portBinding)
	portEntry.SetPlaceHolder("Port number")
	dte.form.Append("Port", portEntry)
	
	dte.container = container.NewVBox(dte.form)
}

// setupBindingListeners sets up listeners for data binding changes
func (dte *DeploymentTargetEditor) setupBindingListeners() {
	bindings := []binding.String{
		dte.nameBinding, dte.typeBinding, dte.urlBinding, dte.usernameBinding,
		dte.passwordBinding, dte.tokenBinding, dte.regionBinding, 
		dte.bucketBinding, dte.pathBinding, dte.portBinding,
	}
	
	for _, b := range bindings {
		b.AddListener(binding.NewDataListener(func() {
			if dte.onChanged != nil {
				dte.onChanged()
			}
		}))
	}
}

// SetTarget sets the target to edit
func (dte *DeploymentTargetEditor) SetTarget(target *models.DeploymentTarget) {
	dte.target = target
	
	if target == nil {
		// Clear all fields
		dte.nameBinding.Set("")
		dte.typeBinding.Set("")
		dte.urlBinding.Set("")
		dte.usernameBinding.Set("")
		dte.passwordBinding.Set("")
		dte.tokenBinding.Set("")
		dte.regionBinding.Set("")
		dte.bucketBinding.Set("")
		dte.pathBinding.Set("")
		dte.portBinding.Set("")
		
		if dte.typeSelect != nil {
			dte.typeSelect.ClearSelected()
		}
		if dte.presetSelect != nil {
			dte.presetSelect.SetSelected("Custom")
		}
		
		return
	}
	
	// Update bindings from target
	dte.nameBinding.Set(target.Name)
	dte.typeBinding.Set(target.Type)
	dte.urlBinding.Set(target.URL)
	dte.usernameBinding.Set(target.Username)
	dte.passwordBinding.Set(target.Password)
	dte.tokenBinding.Set(target.Token)
	dte.regionBinding.Set(target.Region)
	dte.bucketBinding.Set(target.Bucket)
	dte.pathBinding.Set(target.Path)
	
	if target.Port > 0 {
		dte.portBinding.Set(strconv.Itoa(target.Port))
	} else {
		dte.portBinding.Set("")
	}
	
	// Update UI components
	if dte.typeSelect != nil {
		dte.typeSelect.SetSelected(target.Type)
		dte.updateFormForType(target.Type)
	}
	if dte.presetSelect != nil {
		dte.presetSelect.SetSelected("Custom")
	}
}

// GetTarget returns the current target configuration
func (dte *DeploymentTargetEditor) GetTarget() *models.DeploymentTarget {
	if dte.target == nil {
		dte.target = &models.DeploymentTarget{}
	}
	
	// Update target from bindings
	name, _ := dte.nameBinding.Get()
	targetType, _ := dte.typeBinding.Get()
	url, _ := dte.urlBinding.Get()
	username, _ := dte.usernameBinding.Get()
	password, _ := dte.passwordBinding.Get()
	token, _ := dte.tokenBinding.Get()
	region, _ := dte.regionBinding.Get()
	bucket, _ := dte.bucketBinding.Get()
	path, _ := dte.pathBinding.Get()
	portStr, _ := dte.portBinding.Get()
	
	dte.target.Name = name
	dte.target.Type = targetType
	dte.target.URL = url
	dte.target.Username = username
	dte.target.Password = password
	dte.target.Token = token
	dte.target.Region = region
	dte.target.Bucket = bucket
	dte.target.Path = path
	
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			dte.target.Port = port
		}
	}
	
	return dte.target
}

// updateFormForType updates form visibility based on deployment type
func (dte *DeploymentTargetEditor) updateFormForType(deploymentType string) {
	if dte.form == nil {
		return
	}
	
	// Show/hide form items based on type
	for _, item := range dte.form.Items {
		switch item.Text {
		case "Username", "Password", "Port":
			item.Widget.Show()
			if deploymentType == "s3" || deploymentType == "netlify" || deploymentType == "vercel" {
				item.Widget.Hide()
			}
		case "Token":
			item.Widget.Hide()
			if deploymentType == "github" || deploymentType == "netlify" || deploymentType == "vercel" {
				item.Widget.Show()
			}
		case "Region", "Bucket":
			item.Widget.Hide()
			if deploymentType == "s3" {
				item.Widget.Show()
			}
		}
	}
	
	dte.form.Refresh()
}

// loadPreset loads a preset configuration
func (dte *DeploymentTargetEditor) loadPreset(preset models.DeploymentTarget) {
	dte.typeBinding.Set(preset.Type)
	dte.urlBinding.Set(preset.URL)
	dte.regionBinding.Set(preset.Region)
	dte.pathBinding.Set(preset.Path)
	
	if preset.Port > 0 {
		dte.portBinding.Set(strconv.Itoa(preset.Port))
	}
	
	if dte.typeSelect != nil {
		dte.typeSelect.SetSelected(preset.Type)
		dte.updateFormForType(preset.Type)
	}
}

// SetOnChanged sets the callback for when the target changes
func (dte *DeploymentTargetEditor) SetOnChanged(callback func()) {
	dte.onChanged = callback
}