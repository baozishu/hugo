package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/interfaces"
)

// ProjectExplorer represents the project file browser tree
type ProjectExplorer struct {
	tree        *widget.Tree
	projectPath string
	selectedPath string // Track selected path
	onFileSelect func(string)
	onFileAction func(string, string) // action, path
	
	// File system data
	fileNodes map[string][]string // parent -> children mapping
	fileTypes map[string]string   // path -> type mapping
}

// FileNode represents a file or directory in the project
type FileNode struct {
	Path     string
	Name     string
	IsDir    bool
	Children []*FileNode
}

// NewProjectExplorer creates a new project explorer widget
func NewProjectExplorer() *ProjectExplorer {
	pe := &ProjectExplorer{
		fileNodes: make(map[string][]string),
		fileTypes: make(map[string]string),
	}
	
	pe.setupTree()
	return pe
}

// setupTree initializes the tree widget
func (pe *ProjectExplorer) setupTree() {
	pe.tree = widget.NewTree(
		// Child UIDs function
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return pe.getChildren(uid)
		},
		// Is Branch function
		func(uid widget.TreeNodeID) bool {
			return pe.isBranch(uid)
		},
		// Create Node function
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			label := widget.NewLabel("Node")
			return container.NewHBox(icon, label)
		},
		// Update Node function
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			pe.updateNode(uid, branch, node)
		},
	)
	
	// Set up selection handler
	pe.tree.OnSelected = func(uid widget.TreeNodeID) {
		pe.selectedPath = uid
		if pe.onFileSelect != nil {
			pe.onFileSelect(uid)
		}
	}
}

// LoadProject loads a project into the explorer
func (pe *ProjectExplorer) LoadProject(project *interfaces.Project) error {
	if project == nil {
		return fmt.Errorf("project is nil")
	}
	
	pe.projectPath = project.Path
	
	// Clear existing data
	pe.fileNodes = make(map[string][]string)
	pe.fileTypes = make(map[string]string)
	
	// Scan project directory
	err := pe.scanDirectory(project.Path, "")
	if err != nil {
		return fmt.Errorf("failed to scan project directory: %w", err)
	}
	
	// Refresh the tree
	pe.tree.Refresh()
	
	return nil
}

// scanDirectory recursively scans a directory and builds the file tree
func (pe *ProjectExplorer) scanDirectory(basePath, relativePath string) error {
	fullPath := filepath.Join(basePath, relativePath)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return err
	}
	
	var children []string
	
	for _, entry := range entries {
		// Skip hidden files and common ignore patterns
		if pe.shouldIgnoreFile(entry.Name()) {
			continue
		}
		
		childPath := filepath.Join(relativePath, entry.Name())
		children = append(children, childPath)
		
		if entry.IsDir() {
			pe.fileTypes[childPath] = "directory"
			// Recursively scan subdirectories
			err := pe.scanDirectory(basePath, childPath)
			if err != nil {
				// Log error but continue with other files
				continue
			}
		} else {
			pe.fileTypes[childPath] = pe.getFileType(entry.Name())
		}
	}
	
	// Sort children: directories first, then files
	sort.Slice(children, func(i, j int) bool {
		iIsDir := pe.fileTypes[children[i]] == "directory"
		jIsDir := pe.fileTypes[children[j]] == "directory"
		
		if iIsDir && !jIsDir {
			return true
		}
		if !iIsDir && jIsDir {
			return false
		}
		
		return strings.ToLower(filepath.Base(children[i])) < strings.ToLower(filepath.Base(children[j]))
	})
	
	pe.fileNodes[relativePath] = children
	
	return nil
}

// shouldIgnoreFile determines if a file should be ignored in the tree
func (pe *ProjectExplorer) shouldIgnoreFile(name string) bool {
	ignorePatterns := []string{
		".",
		".git",
		".gitignore",
		"node_modules",
		".DS_Store",
		"Thumbs.db",
		"public", // Hugo output directory
		"resources", // Hugo resources directory
	}
	
	for _, pattern := range ignorePatterns {
		if strings.HasPrefix(name, pattern) {
			return true
		}
	}
	
	return false
}

// getFileType determines the file type based on extension
func (pe *ProjectExplorer) getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".md", ".markdown":
		return "markdown"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".json":
		return "json"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".js":
		return "javascript"
	case ".go":
		return "go"
	case ".png", ".jpg", ".jpeg", ".gif", ".svg":
		return "image"
	default:
		return "file"
	}
}

// getChildren returns the child nodes for a given parent
func (pe *ProjectExplorer) getChildren(uid widget.TreeNodeID) []widget.TreeNodeID {
	children, exists := pe.fileNodes[uid]
	if !exists {
		return []widget.TreeNodeID{}
	}
	
	result := make([]widget.TreeNodeID, len(children))
	for i, child := range children {
		result[i] = child
	}
	
	return result
}

// isBranch determines if a node is a branch (directory)
func (pe *ProjectExplorer) isBranch(uid widget.TreeNodeID) bool {
	fileType, exists := pe.fileTypes[uid]
	if !exists {
		return false
	}
	
	return fileType == "directory"
}

// updateNode updates the visual representation of a tree node
func (pe *ProjectExplorer) updateNode(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
	if container, ok := node.(*fyne.Container); ok && len(container.Objects) >= 2 {
		icon := container.Objects[0].(*widget.Icon)
		label := container.Objects[1].(*widget.Label)
		
		// Set the label text
		if uid == "" {
			label.SetText(filepath.Base(pe.projectPath))
		} else {
			label.SetText(filepath.Base(uid))
		}
		
		// Set the appropriate icon
		if branch {
			icon.SetResource(theme.FolderIcon())
		} else {
			icon.SetResource(pe.getFileIcon(uid))
		}
	}
}

// getFileIcon returns the appropriate icon for a file type
func (pe *ProjectExplorer) getFileIcon(path string) fyne.Resource {
	fileType, exists := pe.fileTypes[path]
	if !exists {
		return theme.DocumentIcon()
	}
	
	switch fileType {
	case "markdown":
		return theme.DocumentIcon()
	case "yaml", "toml", "json":
		return theme.SettingsIcon()
	case "html":
		return theme.ComputerIcon()
	case "css":
		return theme.ColorPaletteIcon()
	case "javascript":
		return theme.ComputerIcon()
	case "image":
		return theme.VisibilityIcon()
	default:
		return theme.DocumentIcon()
	}
}

// GetWidget returns the tree widget
func (pe *ProjectExplorer) GetWidget() fyne.CanvasObject {
	if pe.projectPath == "" {
		// Show empty state
		return container.NewVBox(
			widget.NewCard("Project Explorer", "", 
				widget.NewLabel("No project loaded\n\nOpen a project to see its files here.")),
		)
	}
	
	return container.NewVBox(
		widget.NewCard("Project Explorer", "", pe.tree),
	)
}

// SetOnFileSelect sets the callback for file selection
func (pe *ProjectExplorer) SetOnFileSelect(callback func(string)) {
	pe.onFileSelect = callback
}

// SetOnFileAction sets the callback for file actions
func (pe *ProjectExplorer) SetOnFileAction(callback func(string, string)) {
	pe.onFileAction = callback
}

// GetSelectedPath returns the currently selected file path
func (pe *ProjectExplorer) GetSelectedPath() string {
	if pe.selectedPath == "" {
		return ""
	}
	
	return filepath.Join(pe.projectPath, pe.selectedPath)
}

// RefreshProject refreshes the project tree
func (pe *ProjectExplorer) RefreshProject() {
	if pe.projectPath != "" {
		// Re-scan the directory
		pe.fileNodes = make(map[string][]string)
		pe.fileTypes = make(map[string]string)
		
		pe.scanDirectory(pe.projectPath, "")
		pe.tree.Refresh()
	}
}

// CreateContextMenu creates a context menu for file operations
func (pe *ProjectExplorer) CreateContextMenu(path string) *fyne.Menu {
	isDir := pe.fileTypes[path] == "directory"
	
	var menuItems []*fyne.MenuItem
	
	if !isDir {
		// File operations
		menuItems = append(menuItems,
			fyne.NewMenuItem("Open", func() {
				if pe.onFileAction != nil {
					pe.onFileAction("open", path)
				}
			}),
			fyne.NewMenuItem("Edit", func() {
				if pe.onFileAction != nil {
					pe.onFileAction("edit", path)
				}
			}),
		)
	}
	
	// Common operations
	menuItems = append(menuItems,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Rename", func() {
			if pe.onFileAction != nil {
				pe.onFileAction("rename", path)
			}
		}),
		fyne.NewMenuItem("Delete", func() {
			if pe.onFileAction != nil {
				pe.onFileAction("delete", path)
			}
		}),
	)
	
	if isDir {
		// Directory operations
		menuItems = append(menuItems,
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("New File", func() {
				if pe.onFileAction != nil {
					pe.onFileAction("new_file", path)
				}
			}),
			fyne.NewMenuItem("New Folder", func() {
				if pe.onFileAction != nil {
					pe.onFileAction("new_folder", path)
				}
			}),
		)
	}
	
	return fyne.NewMenu("", menuItems...)
}