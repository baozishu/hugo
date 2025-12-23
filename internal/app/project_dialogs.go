package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/interfaces"
)

// ProjectSelectionDialog shows the project selection interface
type ProjectSelectionDialog struct {
	window     fyne.Window
	controller *AppController
	dialog     *dialog.CustomDialog
}

// NewProjectWizard handles new project creation
type NewProjectWizard struct {
	window     fyne.Window
	controller *AppController
	dialog     *dialog.CustomDialog
	
	// Form fields
	nameEntry     *widget.Entry
	pathEntry     *widget.Entry
	themeSelect   *widget.Select
	authorEntry   *widget.Entry
	descEntry     *widget.Entry
}

// RecentProjectsList manages the recent projects display
type RecentProjectsList struct {
	list       *widget.List
	projects   []string
	onSelect   func(string)
}

// ShowProjectSelectionDialog displays the main project selection interface
func (ac *AppController) ShowProjectSelectionDialog() {
	psd := &ProjectSelectionDialog{
		window:     ac.window,
		controller: ac,
	}
	
	psd.show()
}

// show creates and displays the project selection dialog
func (psd *ProjectSelectionDialog) show() {
	// Create recent projects list
	recentList := psd.createRecentProjectsList()
	
	// Create action buttons
	newProjectBtn := widget.NewButton("New Project", func() {
		psd.dialog.Hide()
		psd.controller.ShowNewProjectWizard()
	})
	newProjectBtn.Importance = widget.HighImportance
	
	openProjectBtn := widget.NewButton("Open Project", func() {
		psd.dialog.Hide()
		psd.controller.ShowOpenProjectDialog()
	})
	
	// Create main content
	content := container.NewVBox(
		widget.NewCard("Recent Projects", "", recentList),
		container.NewHBox(
			newProjectBtn,
			openProjectBtn,
		),
	)
	
	// Create dialog
	psd.dialog = dialog.NewCustom("Select Project", "Cancel", content, psd.window)
	psd.dialog.Resize(fyne.NewSize(500, 400))
	psd.dialog.Show()
}

// createRecentProjectsList creates the recent projects list widget
func (psd *ProjectSelectionDialog) createRecentProjectsList() fyne.CanvasObject {
	// Get recent projects (placeholder - will be implemented with actual storage)
	recentProjects := []string{
		"My Blog",
		"Company Website", 
		"Documentation Site",
	}
	
	if len(recentProjects) == 0 {
		return widget.NewLabel("No recent projects")
	}
	
	list := widget.NewList(
		func() int {
			return len(recentProjects)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.FolderIcon()),
				widget.NewLabel("Project Name"),
				widget.NewLabel("Path"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(recentProjects) {
				// Use type assertion for fyne.Container
				if hbox, ok := obj.(*fyne.Container); ok && len(hbox.Objects) >= 3 {
					if nameLabel, ok := hbox.Objects[1].(*widget.Label); ok {
						nameLabel.SetText(recentProjects[id])
					}
					if pathLabel, ok := hbox.Objects[2].(*widget.Label); ok {
						pathLabel.SetText("/path/to/" + strings.ToLower(strings.ReplaceAll(recentProjects[id], " ", "-")))
					}
				}
			}
		},
	)
	
	list.OnSelected = func(id widget.ListItemID) {
		if id < len(recentProjects) {
			psd.dialog.Hide()
			// Open the selected project (placeholder)
			psd.controller.mainWindow.UpdateStatusBar("Opening project: " + recentProjects[id])
		}
	}
	
	return list
}

// show creates and displays the new project wizard
func (npw *NewProjectWizard) show() {
	// Create form fields
	npw.nameEntry = widget.NewEntry()
	npw.nameEntry.SetPlaceHolder("My Hugo Blog")
	
	npw.pathEntry = widget.NewEntry()
	npw.pathEntry.SetPlaceHolder("C:\\Users\\Username\\Documents\\my-hugo-blog")
	
	npw.themeSelect = widget.NewSelect([]string{
		"ananke (default)",
		"paper",
		"terminal",
		"hermit",
		"hyde",
	}, nil)
	npw.themeSelect.SetSelected("ananke (default)")
	
	npw.authorEntry = widget.NewEntry()
	npw.authorEntry.SetPlaceHolder("Your Name")
	
	npw.descEntry = widget.NewMultiLineEntry()
	npw.descEntry.SetPlaceHolder("A brief description of your site")
	npw.descEntry.Resize(fyne.NewSize(400, 80))
	
	// Create browse button for path
	browseBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				return
			}
			if uri != nil {
				path := uri.Path()
				if npw.nameEntry.Text != "" {
					path = filepath.Join(path, strings.ToLower(strings.ReplaceAll(npw.nameEntry.Text, " ", "-")))
				}
				npw.pathEntry.SetText(path)
			}
		}, npw.window)
	})
	
	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Project Name:", Widget: npw.nameEntry},
			{Text: "Project Path:", Widget: container.NewBorder(nil, nil, nil, browseBtn, npw.pathEntry)},
			{Text: "Theme:", Widget: npw.themeSelect},
			{Text: "Author:", Widget: npw.authorEntry},
			{Text: "Description:", Widget: npw.descEntry},
		},
		OnSubmit: npw.createProject,
		OnCancel: func() {
			npw.dialog.Hide()
		},
	}
	
	// Create dialog
	content := container.NewVBox(
		widget.NewCard("New Hugo Project", "Create a new Hugo static site project", form),
	)
	
	npw.dialog = dialog.NewCustom("New Project Wizard", "", content, npw.window)
	npw.dialog.Resize(fyne.NewSize(600, 500))
	npw.dialog.Show()
}

// createProject handles the project creation process
func (npw *NewProjectWizard) createProject() {
	// Validate inputs
	if npw.nameEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("project name is required"), npw.window)
		return
	}
	
	if npw.pathEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("project path is required"), npw.window)
		return
	}
	
	// Check if path already exists
	if _, err := os.Stat(npw.pathEntry.Text); err == nil {
		dialog.ShowError(fmt.Errorf("path already exists: %s", npw.pathEntry.Text), npw.window)
		return
	}
	
	// Create project configuration
	config := interfaces.ProjectConfig{
		Name:        npw.nameEntry.Text,
		Title:       npw.nameEntry.Text,
		Theme:       strings.Split(npw.themeSelect.Selected, " ")[0], // Extract theme name
		Description: npw.descEntry.Text,
	}
	
	// Hide dialog
	npw.dialog.Hide()
	
	// Show progress dialog
	progress := dialog.NewProgressInfinite("Creating Project", "Creating new Hugo project...", npw.window)
	progress.Show()
	
	// Create project (placeholder - will use actual ProjectManager)
	go func() {
		// Simulate project creation
		// In real implementation, this would call:
		// project, err := npw.controller.projectManager.CreateProject(config.Path, config)
		
		// For now, just simulate success
		progress.Hide()
		
		// Show success message
		dialog.ShowInformation("Success", 
			fmt.Sprintf("Project '%s' created successfully at:\n%s", config.Name, npw.pathEntry.Text), 
			npw.window)
		
		// Update status
		npw.controller.mainWindow.UpdateStatusBar("Project created: " + config.Name)
	}()
}

// ShowOpenProjectDialog displays the open project dialog
func (ac *AppController) ShowOpenProjectDialog() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, ac.window)
			return
		}
		if uri != nil {
			path := uri.Path()
			
			// Validate Hugo project (placeholder)
			if ac.validateHugoProject(path) {
				// Load project (placeholder)
				ac.mainWindow.UpdateStatusBar("Opening project at: " + path)
				
				// Create project object (placeholder)
				project := &interfaces.Project{
					Name: filepath.Base(path),
					Path: path,
				}
				
				ac.SetCurrentProject(project)
			} else {
				dialog.ShowError(fmt.Errorf("selected folder is not a valid Hugo project"), ac.window)
			}
		}
	}, ac.window)
}

// validateHugoProject checks if a directory contains a valid Hugo project
func (ac *AppController) validateHugoProject(path string) bool {
	// Check for config file (placeholder implementation)
	configFiles := []string{"config.yaml", "config.yml", "config.toml", "hugo.yaml", "hugo.yml", "hugo.toml"}
	
	for _, configFile := range configFiles {
		if _, err := os.Stat(filepath.Join(path, configFile)); err == nil {
			return true
		}
	}
	
	return false
}