package app

import (
	"fmt"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/models"
	"hugo-visual-client/internal/repository"
)

// MenuEditor provides a visual interface for editing Hugo site menus
type MenuEditor struct {
	widget.BaseWidget
	
	// Configuration data
	config     *models.SiteConfig
	configRepo *repository.ConfigRepository
	configPath string
	
	// UI components
	container     *container.Split
	menusList     *widget.List
	menuDetails   *fyne.Container
	itemsList     *widget.List
	
	// Data
	menuNames     []string
	selectedMenu  string
	menuItems     []models.MenuItem
	selectedItem  *models.MenuItem
	
	// Callbacks
	onSave   func(*models.SiteConfig) error
	onReset  func()
	
	// State
	isModified bool
}

// MenuItemEditor provides an interface for editing individual menu items
type MenuItemEditor struct {
	widget.BaseWidget
	
	item      *models.MenuItem
	container *fyne.Container
	
	// Form fields
	nameEntry       *widget.Entry
	urlEntry        *widget.Entry
	weightEntry     *widget.Entry
	identifierEntry *widget.Entry
	parentSelect    *widget.Select
	
	onChanged func()
}

// NewMenuEditor creates a new menu editor
func NewMenuEditor(configRepo *repository.ConfigRepository, configPath string) *MenuEditor {
	editor := &MenuEditor{
		configRepo: configRepo,
		configPath: configPath,
		menuNames:  []string{},
		menuItems:  []models.MenuItem{},
		isModified: false,
	}
	
	editor.ExtendBaseWidget(editor)
	return editor
}

// LoadConfig loads configuration from file
func (me *MenuEditor) LoadConfig() error {
	config, err := me.configRepo.LoadSiteConfig(me.configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	me.config = config
	me.updateMenusList()
	me.isModified = false
	
	return nil
}

// SaveConfig saves the current configuration
func (me *MenuEditor) SaveConfig() error {
	if me.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Validate configuration
	if err := me.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Save to file
	err := me.configRepo.SaveSiteConfig(me.configPath, me.config)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	me.isModified = false
	
	// Call save callback if set
	if me.onSave != nil {
		return me.onSave(me.config)
	}
	
	return nil
}

// CreateRenderer creates the visual representation of the menu editor
func (me *MenuEditor) CreateRenderer() fyne.WidgetRenderer {
	me.setupUI()
	return widget.NewSimpleRenderer(me.container)
}

// setupUI creates the user interface components
func (me *MenuEditor) setupUI() {
	// Create menus list
	me.menusList = widget.NewList(
		func() int {
			return len(me.menuNames)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.MenuIcon()),
				widget.NewLabel("Menu Name"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(me.menuNames) {
				return
			}
			
			menuName := me.menuNames[id]
			if hbox, ok := obj.(*fyne.Container); ok {
				nameLabel := hbox.Objects[1].(*widget.Label)
				nameLabel.SetText(menuName)
			}
		},
	)
	
	me.menusList.OnSelected = func(id widget.ListItemID) {
		if id < len(me.menuNames) {
			me.selectedMenu = me.menuNames[id]
			me.loadMenuItems()
			me.updateMenuDetails()
		}
	}
	
	// Create menu items list
	me.itemsList = widget.NewList(
		func() int {
			return len(me.menuItems)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.NavigateNextIcon()),
				widget.NewLabel("Item Name"),
				widget.NewLabel("URL"),
				widget.NewLabel("Weight"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(me.menuItems) {
				return
			}
			
			item := me.menuItems[id]
			if hbox, ok := obj.(*fyne.Container); ok {
				nameLabel := hbox.Objects[1].(*widget.Label)
				urlLabel := hbox.Objects[2].(*widget.Label)
				weightLabel := hbox.Objects[3].(*widget.Label)
				
				nameLabel.SetText(item.Name)
				urlLabel.SetText(item.URL)
				weightLabel.SetText(fmt.Sprintf("%d", item.Weight))
			}
		},
	)
	
	me.itemsList.OnSelected = func(id widget.ListItemID) {
		if id < len(me.menuItems) {
			me.selectedItem = &me.menuItems[id]
			me.updateMenuDetails()
		}
	}
	
	// Create menu details panel
	me.menuDetails = container.NewVBox(
		widget.NewLabel("Select a menu to view details"),
	)
	
	// Create action buttons for menus
	addMenuButton := widget.NewButtonWithIcon("Add Menu", theme.ContentAddIcon(), func() {
		me.handleAddMenu()
	})
	
	deleteMenuButton := widget.NewButtonWithIcon("Delete Menu", theme.DeleteIcon(), func() {
		me.handleDeleteMenu()
	})
	
	menuButtonContainer := container.NewHBox(
		addMenuButton,
		deleteMenuButton,
	)
	
	// Create action buttons for menu items
	addItemButton := widget.NewButtonWithIcon("Add Item", theme.ContentAddIcon(), func() {
		me.handleAddMenuItem()
	})
	
	editItemButton := widget.NewButtonWithIcon("Edit Item", theme.DocumentCreateIcon(), func() {
		me.handleEditMenuItem()
	})
	
	deleteItemButton := widget.NewButtonWithIcon("Delete Item", theme.DeleteIcon(), func() {
		me.handleDeleteMenuItem()
	})
	
	moveUpButton := widget.NewButtonWithIcon("Move Up", theme.MoveUpIcon(), func() {
		me.handleMoveItemUp()
	})
	
	moveDownButton := widget.NewButtonWithIcon("Move Down", theme.MoveDownIcon(), func() {
		me.handleMoveItemDown()
	})
	
	itemButtonContainer := container.NewHBox(
		addItemButton,
		editItemButton,
		deleteItemButton,
		moveUpButton,
		moveDownButton,
	)
	
	// Create save/reset buttons
	saveButton := widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), func() {
		me.handleSave()
	})
	saveButton.Importance = widget.HighImportance
	
	resetButton := widget.NewButtonWithIcon("Reset Changes", theme.ViewRefreshIcon(), func() {
		me.handleReset()
	})
	
	mainButtonContainer := container.NewHBox(
		saveButton,
		resetButton,
	)
	
	// Create left panel with menus and items
	leftPanel := container.NewVBox(
		widget.NewCard("Menus", "", container.NewVBox(
			me.menusList,
			menuButtonContainer,
		)),
		widget.NewCard("Menu Items", "", container.NewVBox(
			me.itemsList,
			itemButtonContainer,
		)),
	)
	
	// Create right panel with details
	rightPanel := container.NewVBox(
		widget.NewCard("Menu Details", "", me.menuDetails),
		widget.NewSeparator(),
		mainButtonContainer,
	)
	
	// Create main split container
	me.container = container.NewHSplit(leftPanel, rightPanel)
	me.container.SetOffset(0.6) // 60% for left panel
}

// updateMenusList updates the menus list from configuration
func (me *MenuEditor) updateMenusList() {
	me.menuNames = []string{}
	
	if me.config != nil && me.config.Menu != nil {
		for menuName := range me.config.Menu {
			me.menuNames = append(me.menuNames, menuName)
		}
		sort.Strings(me.menuNames)
	}
	
	if me.menusList != nil {
		me.menusList.Refresh()
	}
}

// loadMenuItems loads menu items for the selected menu
func (me *MenuEditor) loadMenuItems() {
	me.menuItems = []models.MenuItem{}
	
	if me.config != nil && me.config.Menu != nil && me.selectedMenu != "" {
		if items, exists := me.config.Menu[me.selectedMenu]; exists {
			me.menuItems = make([]models.MenuItem, len(items))
			copy(me.menuItems, items)
			
			// Sort by weight
			sort.Slice(me.menuItems, func(i, j int) bool {
				return me.menuItems[i].Weight < me.menuItems[j].Weight
			})
		}
	}
	
	if me.itemsList != nil {
		me.itemsList.Refresh()
	}
}

// updateMenuDetails updates the menu details panel
func (me *MenuEditor) updateMenuDetails() {
	if me.selectedMenu == "" {
		me.menuDetails = container.NewVBox(
			widget.NewLabel("Select a menu to view details"),
		)
		me.menuDetails.Refresh()
		return
	}
	
	details := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Menu:"),
			widget.NewLabel(me.selectedMenu),
		),
		container.NewHBox(
			widget.NewLabel("Items:"),
			widget.NewLabel(fmt.Sprintf("%d", len(me.menuItems))),
		),
		widget.NewSeparator(),
	)
	
	if me.selectedItem != nil {
		itemDetails := container.NewVBox(
			widget.NewLabel("Selected Item:"),
			container.NewHBox(
				widget.NewLabel("Name:"),
				widget.NewLabel(me.selectedItem.Name),
			),
			container.NewHBox(
				widget.NewLabel("URL:"),
				widget.NewLabel(me.selectedItem.URL),
			),
			container.NewHBox(
				widget.NewLabel("Weight:"),
				widget.NewLabel(fmt.Sprintf("%d", me.selectedItem.Weight)),
			),
			container.NewHBox(
				widget.NewLabel("Identifier:"),
				widget.NewLabel(me.selectedItem.Identifier),
			),
		)
		
		if me.selectedItem.Parent != "" {
			itemDetails.Add(container.NewHBox(
				widget.NewLabel("Parent:"),
				widget.NewLabel(me.selectedItem.Parent),
			))
		}
		
		details.Add(itemDetails)
	}
	
	me.menuDetails = details
	me.menuDetails.Refresh()
}

// Event handlers
func (me *MenuEditor) handleAddMenu() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Menu name (e.g., main, footer)")
	
	dialog.ShowForm("Add Menu", "Add", "Cancel", []*widget.FormItem{
		{Text: "Name:", Widget: nameEntry},
	}, func(confirmed bool) {
		if confirmed && nameEntry.Text != "" {
			menuName := nameEntry.Text
			
			// Initialize menu if config doesn't have menus
			if me.config.Menu == nil {
				me.config.Menu = make(map[string][]models.MenuItem)
			}
			
			// Don't overwrite existing menus
			if _, exists := me.config.Menu[menuName]; !exists {
				me.config.Menu[menuName] = []models.MenuItem{}
				me.updateMenusList()
				me.markModified()
			}
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (me *MenuEditor) handleDeleteMenu() {
	if me.selectedMenu == "" {
		dialog.ShowInformation("No Menu Selected", "Please select a menu to delete.", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	dialog.ShowConfirm("Delete Menu", 
		fmt.Sprintf("Are you sure you want to delete the menu '%s' and all its items?", me.selectedMenu),
		func(confirmed bool) {
			if confirmed {
				delete(me.config.Menu, me.selectedMenu)
				me.selectedMenu = ""
				me.selectedItem = nil
				me.updateMenusList()
				me.loadMenuItems()
				me.updateMenuDetails()
				me.markModified()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (me *MenuEditor) handleAddMenuItem() {
	if me.selectedMenu == "" {
		dialog.ShowInformation("No Menu Selected", "Please select a menu first.", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	me.showMenuItemDialog(nil)
}

func (me *MenuEditor) handleEditMenuItem() {
	if me.selectedItem == nil {
		dialog.ShowInformation("No Item Selected", "Please select a menu item to edit.", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	me.showMenuItemDialog(me.selectedItem)
}

func (me *MenuEditor) handleDeleteMenuItem() {
	if me.selectedItem == nil {
		dialog.ShowInformation("No Item Selected", "Please select a menu item to delete.", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	dialog.ShowConfirm("Delete Menu Item", 
		fmt.Sprintf("Are you sure you want to delete the menu item '%s'?", me.selectedItem.Name),
		func(confirmed bool) {
			if confirmed {
				me.removeMenuItem(me.selectedItem)
				me.selectedItem = nil
				me.saveMenuItems()
				me.loadMenuItems()
				me.updateMenuDetails()
				me.markModified()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (me *MenuEditor) handleMoveItemUp() {
	if me.selectedItem == nil {
		return
	}
	
	// Find item index
	for i, item := range me.menuItems {
		if item.Identifier == me.selectedItem.Identifier && item.Name == me.selectedItem.Name {
			if i > 0 {
				// Swap weights
				me.menuItems[i].Weight, me.menuItems[i-1].Weight = me.menuItems[i-1].Weight, me.menuItems[i].Weight
				me.saveMenuItems()
				me.loadMenuItems()
				me.markModified()
			}
			break
		}
	}
}

func (me *MenuEditor) handleMoveItemDown() {
	if me.selectedItem == nil {
		return
	}
	
	// Find item index
	for i, item := range me.menuItems {
		if item.Identifier == me.selectedItem.Identifier && item.Name == me.selectedItem.Name {
			if i < len(me.menuItems)-1 {
				// Swap weights
				me.menuItems[i].Weight, me.menuItems[i+1].Weight = me.menuItems[i+1].Weight, me.menuItems[i].Weight
				me.saveMenuItems()
				me.loadMenuItems()
				me.markModified()
			}
			break
		}
	}
}

func (me *MenuEditor) handleSave() {
	err := me.SaveConfig()
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	dialog.ShowInformation("Success", "Menu configuration saved successfully!", fyne.CurrentApp().Driver().AllWindows()[0])
}

func (me *MenuEditor) handleReset() {
	dialog.ShowConfirm("Reset Configuration", 
		"Are you sure you want to reset all changes? This will discard any unsaved modifications.",
		func(confirmed bool) {
			if confirmed {
				err := me.LoadConfig()
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				} else {
					me.selectedMenu = ""
					me.selectedItem = nil
					me.updateMenuDetails()
				}
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// showMenuItemDialog shows a dialog for adding/editing menu items
func (me *MenuEditor) showMenuItemDialog(item *models.MenuItem) {
	nameEntry := widget.NewEntry()
	urlEntry := widget.NewEntry()
	weightEntry := widget.NewEntry()
	identifierEntry := widget.NewEntry()
	parentSelect := widget.NewSelect(me.getParentOptions(), nil)
	
	// Set placeholders
	nameEntry.SetPlaceHolder("Menu item name")
	urlEntry.SetPlaceHolder("/path/to/page or https://example.com")
	weightEntry.SetPlaceHolder("10")
	identifierEntry.SetPlaceHolder("unique-identifier")
	
	// If editing, populate fields
	if item != nil {
		nameEntry.SetText(item.Name)
		urlEntry.SetText(item.URL)
		weightEntry.SetText(fmt.Sprintf("%d", item.Weight))
		identifierEntry.SetText(item.Identifier)
		if item.Parent != "" {
			parentSelect.SetSelected(item.Parent)
		}
	}
	
	form := []*widget.FormItem{
		{Text: "Name:", Widget: nameEntry},
		{Text: "URL:", Widget: urlEntry},
		{Text: "Weight:", Widget: weightEntry},
		{Text: "Identifier:", Widget: identifierEntry},
		{Text: "Parent:", Widget: parentSelect},
	}
	
	title := "Add Menu Item"
	if item != nil {
		title = "Edit Menu Item"
	}
	
	dialog.ShowForm(title, "Save", "Cancel", form, func(confirmed bool) {
		if confirmed && nameEntry.Text != "" && urlEntry.Text != "" {
			weight := 0
			if weightEntry.Text != "" {
				if w, err := strconv.Atoi(weightEntry.Text); err == nil {
					weight = w
				}
			}
			
			newItem := models.MenuItem{
				Name:       nameEntry.Text,
				URL:        urlEntry.Text,
				Weight:     weight,
				Identifier: identifierEntry.Text,
				Parent:     parentSelect.Selected,
			}
			
			// Validate the menu item
			if err := newItem.Validate(); err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
			
			if item != nil {
				// Update existing item
				me.updateMenuItem(item, &newItem)
			} else {
				// Add new item
				me.addMenuItem(&newItem)
			}
			
			me.saveMenuItems()
			me.loadMenuItems()
			me.updateMenuDetails()
			me.markModified()
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// getParentOptions returns available parent menu items
func (me *MenuEditor) getParentOptions() []string {
	options := []string{""}
	
	for _, item := range me.menuItems {
		if item.Identifier != "" {
			options = append(options, item.Identifier)
		}
	}
	
	return options
}

// addMenuItem adds a new menu item
func (me *MenuEditor) addMenuItem(item *models.MenuItem) {
	me.menuItems = append(me.menuItems, *item)
}

// updateMenuItem updates an existing menu item
func (me *MenuEditor) updateMenuItem(oldItem, newItem *models.MenuItem) {
	for i, item := range me.menuItems {
		if item.Identifier == oldItem.Identifier && item.Name == oldItem.Name {
			me.menuItems[i] = *newItem
			break
		}
	}
}

// removeMenuItem removes a menu item
func (me *MenuEditor) removeMenuItem(item *models.MenuItem) {
	for i, menuItem := range me.menuItems {
		if menuItem.Identifier == item.Identifier && menuItem.Name == item.Name {
			me.menuItems = append(me.menuItems[:i], me.menuItems[i+1:]...)
			break
		}
	}
}

// saveMenuItems saves the current menu items back to the configuration
func (me *MenuEditor) saveMenuItems() {
	if me.config.Menu == nil {
		me.config.Menu = make(map[string][]models.MenuItem)
	}
	
	me.config.Menu[me.selectedMenu] = make([]models.MenuItem, len(me.menuItems))
	copy(me.config.Menu[me.selectedMenu], me.menuItems)
}

// Utility methods
func (me *MenuEditor) markModified() {
	me.isModified = true
}

func (me *MenuEditor) IsModified() bool {
	return me.isModified
}

// Callback setters
func (me *MenuEditor) SetOnSave(callback func(*models.SiteConfig) error) {
	me.onSave = callback
}

func (me *MenuEditor) SetOnReset(callback func()) {
	me.onReset = callback
}

// GetWidget returns the widget for embedding in other containers
func (me *MenuEditor) GetWidget() fyne.CanvasObject {
	return me
}

// GetConfig returns the current configuration
func (me *MenuEditor) GetConfig() *models.SiteConfig {
	return me.config
}