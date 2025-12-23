package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/interfaces"
	"hugo-visual-client/internal/services"
)

// MainWindow represents the main application window with all UI components
type MainWindow struct {
	window       fyne.Window
	toolbar      *widget.Toolbar
	sidebar      *container.Split
	contentArea  *container.AppTabs
	statusBar    *widget.Label
	mainSplit    *container.Split
	
	// Menu items for connecting actions
	newProjectMenuItem  *fyne.MenuItem
	openProjectMenuItem *fyne.MenuItem
	newContentMenuItem  *fyne.MenuItem
}

// AppController manages the main application state and UI
type AppController struct {
	window         fyne.Window
	projectManager interfaces.ProjectManager
	contentManager interfaces.ContentManager
	hugoService    interfaces.HugoService
	deploymentManager interfaces.DeploymentManager
	settingsManager *services.SettingsManager
	currentProject *interfaces.Project
	
	// UI components
	mainWindow      *MainWindow
	projectExplorer *ProjectExplorer
	contentEditors  map[string]*ContentEditor // path -> editor mapping
	
	// Application lifecycle
	isInitialized bool
	configPath    string
	
	// Status and notification system
	statusCallbacks []func(string)
	errorCallbacks  []func(error)
}

// NewMainWindow creates a new main window with all UI components
func NewMainWindow(window fyne.Window) *MainWindow {
	mw := &MainWindow{
		window: window,
	}
	
	mw.setupUI()
	return mw
}

// SetSidebarContent updates the sidebar content
func (mw *MainWindow) SetSidebarContent(content fyne.CanvasObject) {
	if mw.mainSplit != nil {
		mw.mainSplit.Leading = content
		mw.mainSplit.Refresh()
	}
}

// setupUI initializes all UI components and layout
func (mw *MainWindow) setupUI() {
	// Create toolbar with main actions
	mw.toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			// Will be connected to controller actions
		}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			// Will be connected to controller actions
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DocumentIcon(), func() {
			// Will be connected to controller actions
		}),
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
			// Will be connected to controller actions
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			// Will be connected to controller actions
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			// Will be connected to controller actions
		}),
	)
	
	// Create status bar
	mw.statusBar = widget.NewLabel("Ready")
	mw.statusBar.Alignment = fyne.TextAlignLeading
	
	// Create content area with tabs
	mw.contentArea = container.NewAppTabs()
	mw.contentArea.SetTabLocation(container.TabLocationTop)
	
	// Add welcome tab initially
	welcomeContent := container.NewVBox(
		widget.NewCard("Welcome to Hugo Visual Client", "",
			container.NewVBox(
				widget.NewLabel("Get started by creating a new Hugo project or opening an existing one."),
				container.NewHBox(
					widget.NewButton("New Project", func() {
						// Connect to controller action
					}),
					widget.NewButton("Open Project", func() {
						// Connect to controller action
					}),
				),
			),
		),
	)
	mw.contentArea.Append(container.NewTabItem("Welcome", welcomeContent))
	
	// Create sidebar placeholder (will be enhanced in subtask 8.3)
	sidebarContent := container.NewVBox(
		widget.NewCard("Project Explorer", "", 
			widget.NewLabel("No project loaded")),
	)
	
	// Create main split container
	mw.mainSplit = container.NewHSplit(sidebarContent, mw.contentArea)
	mw.mainSplit.SetOffset(0.25) // 25% for sidebar
	
	// Create the main layout using border container
	mainContent := container.NewBorder(
		mw.toolbar,   // top
		mw.statusBar, // bottom
		nil,          // left
		nil,          // right
		mw.mainSplit, // center
	)
	
	// Set up window properties
	mw.window.SetContent(mainContent)
	mw.setupMenus()
	mw.setupShortcuts()
}

// setupMenus creates the main menu bar
func (mw *MainWindow) setupMenus() {
	// Create menu items and store references
	mw.newProjectMenuItem = fyne.NewMenuItem("New Project...", func() {
		// Will be connected to controller actions
	})
	mw.openProjectMenuItem = fyne.NewMenuItem("Open Project...", func() {
		// Will be connected to controller actions
	})
	mw.newContentMenuItem = fyne.NewMenuItem("New Content...", func() {
		// Will be connected to controller actions
	})
	
	// File menu
	fileMenu := fyne.NewMenu("File",
		mw.newProjectMenuItem,
		mw.openProjectMenuItem,
		fyne.NewMenuItemSeparator(),
		mw.newContentMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() {
			mw.window.Close()
		}),
	)
	
	// Edit menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Undo", func() {
			// Will be implemented later
		}),
		fyne.NewMenuItem("Redo", func() {
			// Will be implemented later
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Cut", func() {
			// Will be implemented later
		}),
		fyne.NewMenuItem("Copy", func() {
			// Will be implemented later
		}),
		fyne.NewMenuItem("Paste", func() {
			// Will be implemented later
		}),
	)
	
	// View menu
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Refresh", func() {
			// Will be connected to controller actions
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Toggle Sidebar", func() {
			// Will be implemented later
		}),
	)
	
	// Hugo menu
	hugoMenu := fyne.NewMenu("Hugo",
		fyne.NewMenuItem("Start Server", func() {
			// Will be connected to controller actions
		}),
		fyne.NewMenuItem("Stop Server", func() {
			// Will be connected to controller actions
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Build Site", func() {
			// Will be connected to controller actions
		}),
	)
	
	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			// Will be implemented later
		}),
	)
	
	mainMenu := fyne.NewMainMenu(fileMenu, editMenu, viewMenu, hugoMenu, helpMenu)
	mw.window.SetMainMenu(mainMenu)
}

// setupShortcuts configures keyboard shortcuts
func (mw *MainWindow) setupShortcuts() {
	// Ctrl+N - New Project
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		Modifier: fyne.KeyModifierControl,
		KeyName:  fyne.KeyN,
	}, func(shortcut fyne.Shortcut) {
		// Will be connected to controller actions
	})
	
	// Ctrl+O - Open Project
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		Modifier: fyne.KeyModifierControl,
		KeyName:  fyne.KeyO,
	}, func(shortcut fyne.Shortcut) {
		// Will be connected to controller actions
	})
	
	// Ctrl+Shift+N - New Content
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift,
		KeyName:  fyne.KeyN,
	}, func(shortcut fyne.Shortcut) {
		// Will be connected to controller actions
	})
	
	// F5 - Refresh
	mw.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		Modifier: 0,
		KeyName:  fyne.KeyF5,
	}, func(shortcut fyne.Shortcut) {
		// Will be connected to controller actions
	})
}

// UpdateStatusBar updates the status bar text
func (mw *MainWindow) UpdateStatusBar(text string) {
	mw.statusBar.SetText(text)
}

// AddContentTab adds a new tab to the content area
func (mw *MainWindow) AddContentTab(title string, content fyne.CanvasObject) {
	mw.contentArea.Append(container.NewTabItem(title, content))
	mw.contentArea.SelectTab(mw.contentArea.Items[len(mw.contentArea.Items)-1])
}

// RemoveContentTab removes a tab from the content area
func (mw *MainWindow) RemoveContentTab(title string) {
	for i, item := range mw.contentArea.Items {
		if item.Text == title {
			mw.contentArea.RemoveIndex(i)
			break
		}
	}
}

// GetContentArea returns the main content area
func (mw *MainWindow) GetContentArea() *container.AppTabs {
	return mw.contentArea
}

// GetSidebar returns the sidebar container
func (mw *MainWindow) GetSidebar() *container.Split {
	return mw.sidebar
}

// NewAppController creates a new application controller
func NewAppController(window fyne.Window) (*AppController, error) {
	controller := &AppController{
		window: window,
		contentEditors: make(map[string]*ContentEditor),
		statusCallbacks: make([]func(string), 0),
		errorCallbacks: make([]func(error), 0),
	}
	
	// Initialize configuration path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	controller.configPath = filepath.Join(homeDir, ".hugo-client")
	
	// Initialize all managers
	if err := controller.initializeManagers(); err != nil {
		return nil, fmt.Errorf("failed to initialize managers: %w", err)
	}
	
	// Initialize the project explorer
	controller.projectExplorer = NewProjectExplorer()
	controller.projectExplorer.SetOnFileSelect(controller.onFileSelected)
	controller.projectExplorer.SetOnFileAction(controller.onFileAction)
	
	// Initialize the main window
	controller.mainWindow = NewMainWindow(window)
	
	// Update the sidebar with project explorer
	controller.updateSidebar()
	
	// Set up status and error callbacks
	controller.setupCallbacks()
	
	return controller, nil
}

// initializeManagers initializes all service managers
func (ac *AppController) initializeManagers() error {
	var err error
	
	// Initialize project manager
	ac.projectManager, err = services.NewProjectManagerService(filepath.Join(ac.configPath, "config.json"))
	if err != nil {
		return fmt.Errorf("failed to initialize project manager: %w", err)
	}
	
	// Initialize Hugo service
	ac.hugoService = services.NewHugoService()
	
	// Initialize deployment manager
	ac.deploymentManager = services.NewDeploymentManager()
	
	// Content manager will be initialized when a project is loaded
	// since it needs a project path
	
	log.Println("All managers initialized successfully")
	return nil
}

// setupCallbacks sets up status and error handling callbacks
func (ac *AppController) setupCallbacks() {
	// Add default status callback to update status bar
	ac.AddStatusCallback(func(message string) {
		if ac.mainWindow != nil {
			ac.mainWindow.UpdateStatusBar(message)
		}
	})
	
	// Add default error callback to show errors in status bar
	ac.AddErrorCallback(func(err error) {
		if ac.mainWindow != nil {
			ac.mainWindow.UpdateStatusBar("Error: " + err.Error())
		}
		log.Printf("Application error: %v", err)
	})
}

// connectShortcutActions connects keyboard shortcuts to controller methods
func (ac *AppController) connectShortcutActions() {
	// Shortcuts are already connected in setupShortcuts method
	// This method can be used for dynamic shortcut updates
}

// InitializeUI sets up the main user interface
func (ac *AppController) InitializeUI() {
	if ac.isInitialized {
		return
	}
	
	// Connect toolbar actions to controller methods
	ac.connectToolbarActions()
	ac.connectMenuActions()
	ac.connectShortcutActions()
	ac.connectWelcomeActions()
	
	// Load recent projects and update UI
	ac.loadRecentProjects()
	
	// Check Hugo installation status
	ac.checkHugoInstallation()
	
	// Update status
	ac.notifyStatus("Hugo Visual Client initialized")
	ac.isInitialized = true
}

// loadRecentProjects loads and displays recent projects
func (ac *AppController) loadRecentProjects() {
	if ac.projectManager == nil {
		return
	}
	
	recentProjects := ac.projectManager.GetRecentProjects()
	if len(recentProjects) > 0 {
		ac.notifyStatus(fmt.Sprintf("Found %d recent projects", len(recentProjects)))
		// TODO: Update welcome screen with recent projects
	}
}

// checkHugoInstallation checks if Hugo is installed and shows status
func (ac *AppController) checkHugoInstallation() {
	if ac.hugoService == nil {
		return
	}
	
	installed, version, err := ac.hugoService.IsHugoInstalled()
	if err != nil {
		ac.notifyError(fmt.Errorf("failed to check Hugo installation: %w", err))
		return
	}
	
	if !installed {
		ac.notifyStatus("Warning: Hugo not found - please install Hugo to use all features")
	} else {
		ac.notifyStatus(fmt.Sprintf("Hugo %s detected", version))
	}
}

// connectToolbarActions connects toolbar buttons to controller methods
func (ac *AppController) connectToolbarActions() {
	// Get toolbar actions and connect them
	toolbar := ac.mainWindow.toolbar
	if len(toolbar.Items) >= 7 {
		// New Project
		if action, ok := toolbar.Items[0].(*widget.ToolbarAction); ok {
			action.OnActivated = ac.onNewProject
		}
		// Open Project
		if action, ok := toolbar.Items[1].(*widget.ToolbarAction); ok {
			action.OnActivated = ac.onOpenProject
		}
		// New Content
		if action, ok := toolbar.Items[3].(*widget.ToolbarAction); ok {
			action.OnActivated = ac.onNewContent
		}
		// Start/Stop Server
		if action, ok := toolbar.Items[4].(*widget.ToolbarAction); ok {
			action.OnActivated = ac.onToggleServer
		}
		// Refresh Preview
		if action, ok := toolbar.Items[5].(*widget.ToolbarAction); ok {
			action.OnActivated = ac.onRefreshPreview
		}
		// Settings
		if action, ok := toolbar.Items[7].(*widget.ToolbarAction); ok {
			action.OnActivated = ac.onShowSettings
		}
	}
}

// connectMenuActions connects menu items to controller methods
func (ac *AppController) connectMenuActions() {
	// Connect menu items to controller actions
	if ac.mainWindow.newProjectMenuItem != nil {
		ac.mainWindow.newProjectMenuItem.Action = ac.onNewProject
	}
	if ac.mainWindow.openProjectMenuItem != nil {
		ac.mainWindow.openProjectMenuItem.Action = ac.onOpenProject
	}
	if ac.mainWindow.newContentMenuItem != nil {
		ac.mainWindow.newContentMenuItem.Action = ac.onNewContent
	}
}

// connectWelcomeActions connects welcome screen buttons to controller methods
func (ac *AppController) connectWelcomeActions() {
	// For now, we'll skip the complex navigation and connect actions later
	// This will be improved when we have better access to the UI components
}

// Event handlers
func (ac *AppController) onNewProject() {
	if ac.projectManager == nil {
		ac.notifyError(fmt.Errorf("project manager not initialized"))
		return
	}
	ac.ShowNewProjectWizard()
}

func (ac *AppController) onOpenProject() {
	if ac.projectManager == nil {
		ac.notifyError(fmt.Errorf("project manager not initialized"))
		return
	}
	ac.ShowOpenProjectDialog()
}

func (ac *AppController) onNewContent() {
	if ac.currentProject == nil {
		ac.notifyStatus("No project loaded - cannot create content")
		return
	}
	
	if ac.contentManager == nil {
		ac.notifyError(fmt.Errorf("content manager not initialized"))
		return
	}
	
	// Show new content dialog
	ac.showNewContentDialog()
}

func (ac *AppController) onToggleServer() {
	if ac.hugoService == nil {
		ac.notifyError(fmt.Errorf("Hugo service not initialized"))
		return
	}
	
	if ac.currentProject == nil {
		ac.notifyStatus("No project loaded - cannot start server")
		return
	}
	
	status := ac.hugoService.GetServerStatus()
	if status.Running {
		// Stop server
		if err := ac.hugoService.StopServer(); err != nil {
			ac.notifyError(fmt.Errorf("failed to stop Hugo server: %w", err))
		} else {
			ac.notifyStatus("Hugo server stopped")
		}
	} else {
		// Start server
		if err := ac.hugoService.StartServer(ac.currentProject.Path, 1313); err != nil {
			ac.notifyError(fmt.Errorf("failed to start Hugo server: %w", err))
		} else {
			ac.notifyStatus("Hugo server started on http://localhost:1313")
		}
	}
}

func (ac *AppController) onRefreshPreview() {
	if ac.hugoService == nil {
		ac.notifyError(fmt.Errorf("Hugo service not initialized"))
		return
	}
	
	if ac.currentProject == nil {
		ac.notifyStatus("No project loaded - cannot refresh preview")
		return
	}
	
	// Trigger a rebuild by stopping and starting file watching
	ac.hugoService.StopWatching()
	if err := ac.hugoService.WatchFiles(ac.currentProject.Path, func(path string) {
		ac.notifyStatus("File changed: " + path)
	}); err != nil {
		ac.notifyError(fmt.Errorf("failed to start file watching: %w", err))
	} else {
		ac.notifyStatus("Preview refreshed")
	}
}

func (ac *AppController) onShowSettings() {
	ac.notifyStatus("Opening application settings...")
	// TODO: Implement settings dialog
}

// GetCurrentProject returns the currently loaded project
func (ac *AppController) GetCurrentProject() *interfaces.Project {
	return ac.currentProject
}

// GetMainWindow returns the main window instance
func (ac *AppController) GetMainWindow() *MainWindow {
	return ac.mainWindow
}

// SetCurrentProject sets the current project
func (ac *AppController) SetCurrentProject(project *interfaces.Project) {
	ac.currentProject = project
	
	if project != nil {
		// Add to recent projects
		if ac.projectManager != nil {
			if err := ac.projectManager.AddRecentProject(project.Path); err != nil {
				ac.notifyError(fmt.Errorf("failed to add to recent projects: %w", err))
			}
		}
		
		// Initialize content manager for this project
		ac.contentManager = services.NewContentManager(project.Path)
		
		// Synchronize all components with new project
		ac.SynchronizeProjectState()
		
		ac.notifyStatus("Project loaded: " + project.Name)
	} else {
		ac.window.SetTitle("Hugo Visual Client")
		ac.contentManager = nil
		ac.updateSidebar()
		ac.notifyStatus("No project loaded")
	}
}

// updateSidebar updates the sidebar with the project explorer
func (ac *AppController) updateSidebar() {
	explorerWidget := ac.projectExplorer.GetWidget()
	ac.mainWindow.SetSidebarContent(explorerWidget)
}

// onFileSelected handles file selection in the project explorer
func (ac *AppController) onFileSelected(path string) {
	fullPath := ac.projectExplorer.GetSelectedPath()
	ac.notifyStatus("Selected: " + fullPath)
	
	// If it's a markdown file, we could auto-open it for editing
	if strings.HasSuffix(strings.ToLower(path), ".md") {
		// Auto-open markdown files for editing
		ac.openContentEditor(path)
	}
}

// onFileAction handles file actions from the project explorer
func (ac *AppController) onFileAction(action, path string) {
	fullPath := filepath.Join(ac.projectExplorer.projectPath, path)
	
	switch action {
	case "open", "edit":
		ac.notifyStatus("Opening file: " + fullPath)
		if strings.HasSuffix(strings.ToLower(path), ".md") {
			ac.openContentEditor(path)
		} else {
			ac.notifyStatus("File type not supported for editing: " + path)
		}
		
	case "rename":
		ac.notifyStatus("Rename action for: " + fullPath)
		// TODO: Implement rename dialog using content manager
		
	case "delete":
		ac.notifyStatus("Delete action for: " + fullPath)
		if ac.contentManager != nil {
			// TODO: Show confirmation dialog then use content manager to delete
			ac.notifyStatus("Delete confirmation dialog would appear here")
		}
		
	case "new_file":
		ac.notifyStatus("New file in: " + fullPath)
		ac.showNewContentDialog()
		
	case "new_folder":
		ac.notifyStatus("New folder in: " + fullPath)
		// TODO: Implement new folder dialog
		
	default:
		ac.notifyStatus("Unknown action: " + action)
	}
}

// Content Editor Methods

// openContentEditor opens a content file in the editor
func (ac *AppController) openContentEditor(relativePath string) {
	if ac.contentManager == nil {
		ac.notifyError(fmt.Errorf("content manager not initialized"))
		return
	}
	
	// Check if editor is already open for this file
	if editor, exists := ac.contentEditors[relativePath]; exists {
		// Switch to existing tab
		ac.switchToEditorTab(relativePath, editor)
		return
	}
	
	// Create new content editor
	var projectPath string
	if ac.currentProject != nil {
		projectPath = ac.currentProject.Path
	}
	editor := NewContentEditor(ac.contentManager, ac.hugoService, projectPath)
	editor.SetOnSave(ac.onContentSaved)
	editor.SetOnClose(ac.onContentClosed)
	
	// Load the content
	err := editor.LoadContent(relativePath)
	if err != nil {
		ac.notifyError(fmt.Errorf("failed to load content: %w", err))
		return
	}
	
	// Store the editor
	ac.contentEditors[relativePath] = editor
	
	// Add tab to main window
	tabTitle := editor.GetTitle()
	if editor.IsModified() {
		tabTitle = "* " + tabTitle
	}
	
	ac.mainWindow.AddContentTab(tabTitle, editor.GetWidget())
	ac.notifyStatus("Opened: " + relativePath)
}

// showNewContentDialog shows a dialog to create new content
func (ac *AppController) showNewContentDialog() {
	if ac.currentProject == nil {
		ac.mainWindow.UpdateStatusBar("No project loaded")
		return
	}
	
	// Create a simple dialog for new content
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter post title...")
	
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("posts/my-new-post.md")
	
	// Auto-generate path from title
	titleEntry.OnChanged = func(title string) {
		if title != "" {
			// Simple slug generation
			slug := strings.ToLower(title)
			slug = strings.ReplaceAll(slug, " ", "-")
			slug = strings.ReplaceAll(slug, "'", "")
			pathEntry.SetText("content/posts/" + slug + ".md")
		}
	}
	
	// For now, we'll create the content directly since we don't have a proper dialog system
	// In a real implementation, this would show a modal dialog
	ac.createNewContent("New Post", "content/posts/new-post.md")
}

// createNewContent creates a new content file and opens it in the editor
func (ac *AppController) createNewContent(title, relativePath string) {
	if ac.contentManager == nil {
		ac.notifyError(fmt.Errorf("content manager not initialized"))
		return
	}
	
	// Check if editor is already open for this file
	if editor, exists := ac.contentEditors[relativePath]; exists {
		ac.switchToEditorTab(relativePath, editor)
		return
	}
	
	// Create new content editor
	var projectPath string
	if ac.currentProject != nil {
		projectPath = ac.currentProject.Path
	}
	editor := NewContentEditor(ac.contentManager, ac.hugoService, projectPath)
	editor.SetOnSave(ac.onContentSaved)
	editor.SetOnClose(ac.onContentClosed)
	
	// Create new content
	err := editor.CreateNewContent(relativePath)
	if err != nil {
		ac.notifyError(fmt.Errorf("failed to create new content: %w", err))
		return
	}
	
	// Set the title
	editor.titleBinding.Set(title)
	
	// Store the editor
	ac.contentEditors[relativePath] = editor
	
	// Add tab to main window
	tabTitle := editor.GetTitle()
	if editor.IsModified() {
		tabTitle = "* " + tabTitle
	}
	
	ac.mainWindow.AddContentTab(tabTitle, editor.GetWidget())
	ac.notifyStatus("Created new content: " + relativePath)
	
	// Focus the editor
	editor.Focus()
}

// switchToEditorTab switches to an existing editor tab
func (ac *AppController) switchToEditorTab(relativePath string, editor *ContentEditor) {
	// Find and select the tab
	contentArea := ac.mainWindow.GetContentArea()
	tabTitle := editor.GetTitle()
	if editor.IsModified() {
		tabTitle = "* " + tabTitle
	}
	
	// Look for the tab and select it
	for _, item := range contentArea.Items {
		if strings.Contains(item.Text, editor.GetTitle()) {
			contentArea.SelectTab(item)
			break
		}
	}
	
	ac.notifyStatus("Switched to: " + relativePath)
}

// onContentSaved handles when content is saved
func (ac *AppController) onContentSaved(relativePath string) {
	ac.notifyStatus("Saved: " + relativePath)
	
	// Update tab title to remove modified indicator
	if editor, exists := ac.contentEditors[relativePath]; exists {
		ac.updateEditorTabTitle(relativePath, editor)
	}
	
	// Refresh project explorer to show any new files
	ac.projectExplorer.RefreshProject()
	
	// Trigger Hugo rebuild if server is running
	if ac.hugoService != nil {
		if status := ac.hugoService.GetServerStatus(); status.Running {
			ac.notifyStatus("Content saved - Hugo will rebuild automatically")
		}
	}
}

// onContentClosed handles when content editor is closed
func (ac *AppController) onContentClosed(relativePath string) {
	// Remove from editors map
	delete(ac.contentEditors, relativePath)
	
	// Remove tab from main window
	if editor, exists := ac.contentEditors[relativePath]; exists {
		tabTitle := editor.GetTitle()
		ac.mainWindow.RemoveContentTab(tabTitle)
	}
	
	ac.notifyStatus("Closed: " + relativePath)
}

// updateEditorTabTitle updates the tab title for an editor
func (ac *AppController) updateEditorTabTitle(relativePath string, editor *ContentEditor) {
	contentArea := ac.mainWindow.GetContentArea()
	tabTitle := editor.GetTitle()
	if editor.IsModified() {
		tabTitle = "* " + tabTitle
	}
	
	// Find and update the tab title
	for _, item := range contentArea.Items {
		if strings.Contains(item.Text, editor.GetTitle()) {
			item.Text = tabTitle
			contentArea.Refresh()
			break
		}
	}
}

// SetContentManager sets the content manager for the controller
func (ac *AppController) SetContentManager(cm interfaces.ContentManager) {
	ac.contentManager = cm
}

// Status and Notification Management

// AddStatusCallback adds a callback for status updates
func (ac *AppController) AddStatusCallback(callback func(string)) {
	ac.statusCallbacks = append(ac.statusCallbacks, callback)
}

// AddErrorCallback adds a callback for error notifications
func (ac *AppController) AddErrorCallback(callback func(error)) {
	ac.errorCallbacks = append(ac.errorCallbacks, callback)
}

// notifyStatus sends a status message to all registered callbacks
func (ac *AppController) notifyStatus(message string) {
	for _, callback := range ac.statusCallbacks {
		callback(message)
	}
}

// notifyError sends an error to all registered callbacks
func (ac *AppController) notifyError(err error) {
	for _, callback := range ac.errorCallbacks {
		callback(err)
	}
}

// Application Lifecycle Management

// Shutdown gracefully shuts down the application
func (ac *AppController) Shutdown() error {
	ac.notifyStatus("Shutting down application...")
	
	// Stop Hugo server if running
	if ac.hugoService != nil {
		if status := ac.hugoService.GetServerStatus(); status.Running {
			if err := ac.hugoService.StopServer(); err != nil {
				log.Printf("Error stopping Hugo server: %v", err)
			}
		}
		
		// Stop file watching
		if err := ac.hugoService.StopWatching(); err != nil {
			log.Printf("Error stopping file watcher: %v", err)
		}
	}
	
	// Save current project if loaded
	if ac.currentProject != nil && ac.projectManager != nil {
		if err := ac.projectManager.SaveProject(ac.currentProject); err != nil {
			log.Printf("Error saving current project: %v", err)
		}
	}
	
	// Close all content editors
	for path, editor := range ac.contentEditors {
		if editor.IsModified() {
			// TODO: Show save dialog for modified files
			log.Printf("Warning: Unsaved changes in %s", path)
		}
	}
	
	ac.notifyStatus("Application shutdown complete")
	return nil
}

// IsInitialized returns whether the application has been initialized
func (ac *AppController) IsInitialized() bool {
	return ac.isInitialized
}

// GetConfigPath returns the application configuration path
func (ac *AppController) GetConfigPath() string {
	return ac.configPath
}

// Component Communication and Coordination

// RefreshAllComponents refreshes all UI components
func (ac *AppController) RefreshAllComponents() {
	// Refresh project explorer
	if ac.projectExplorer != nil {
		ac.projectExplorer.RefreshProject()
	}
	
	// Refresh all open content editors
	for _, editor := range ac.contentEditors {
		editor.RefreshContent()
	}
	
	// Update main window
	if ac.mainWindow != nil {
		ac.mainWindow.GetContentArea().Refresh()
	}
	
	ac.notifyStatus("All components refreshed")
}

// SynchronizeProjectState synchronizes project state across all components
func (ac *AppController) SynchronizeProjectState() {
	if ac.currentProject == nil {
		return
	}
	
	// Update window title
	ac.window.SetTitle("Hugo Visual Client - " + ac.currentProject.Name)
	
	// Update project explorer
	if ac.projectExplorer != nil {
		err := ac.projectExplorer.LoadProject(ac.currentProject)
		if err != nil {
			ac.notifyError(fmt.Errorf("failed to load project in explorer: %w", err))
		}
	}
	
	// Initialize content manager for the project
	if ac.contentManager == nil {
		ac.contentManager = services.NewContentManager(ac.currentProject.Path)
	}
	
	// Update all content editors with new project context
	for _, editor := range ac.contentEditors {
		editor.SetProjectPath(ac.currentProject.Path)
	}
	
	// Update sidebar
	ac.updateSidebar()
	
	ac.notifyStatus("Project state synchronized")
}

// Project Dialog Methods

// ShowNewProjectWizard shows the new project creation wizard
func (ac *AppController) ShowNewProjectWizard() {
	if ac.projectManager == nil {
		ac.notifyError(fmt.Errorf("project manager not initialized"))
		return
	}
	
	// For now, create a simple project dialog
	// In a full implementation, this would show a proper wizard dialog
	ac.notifyStatus("New Project Wizard - Creating sample project...")
	
	// Create a sample project configuration
	config := interfaces.ProjectConfig{
		Name:        "My Hugo Site",
		Title:       "My Hugo Site",
		BaseURL:     "https://example.com",
		Theme:       "",
		Language:    "en",
		Description: "A new Hugo site",
	}
	
	// For demonstration, create in a temp directory
	// In real implementation, user would choose the path
	homeDir, _ := os.UserHomeDir()
	projectPath := filepath.Join(homeDir, "hugo-sites", "my-hugo-site")
	
	project, err := ac.projectManager.CreateProject(projectPath, config)
	if err != nil {
		ac.notifyError(fmt.Errorf("failed to create project: %w", err))
		return
	}
	
	ac.SetCurrentProject(project)
	ac.notifyStatus("New project created successfully")
}

