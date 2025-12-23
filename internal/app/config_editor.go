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

	"hugo-visual-client/internal/models"
	"hugo-visual-client/internal/repository"
)

// ConfigEditor provides a visual interface for editing Hugo site configuration
type ConfigEditor struct {
	widget.BaseWidget
	
	// Configuration data
	config     *models.SiteConfig
	configRepo *repository.ConfigRepository
	configPath string
	
	// Data bindings for form fields
	baseURLBinding     binding.String
	titleBinding       binding.String
	descriptionBinding binding.String
	languageBinding    binding.String
	themeBinding       binding.String
	
	// UI components
	container     *container.Scroll
	form          *widget.Form
	paramsEditor  *DynamicParamsEditor
	saveButton    *widget.Button
	resetButton   *widget.Button
	
	// Callbacks
	onSave   func(*models.SiteConfig) error
	onReset  func()
	onClose  func()
	
	// State
	isModified bool
}

// DynamicParamsEditor provides an interface for editing dynamic parameters
type DynamicParamsEditor struct {
	widget.BaseWidget
	
	params    map[string]interface{}
	container *fyne.Container
	entries   map[string]*widget.Entry
	
	onChanged func()
}

// NewConfigEditor creates a new configuration editor
func NewConfigEditor(configRepo *repository.ConfigRepository, configPath string) *ConfigEditor {
	editor := &ConfigEditor{
		configRepo: configRepo,
		configPath: configPath,
		isModified: false,
	}
	
	// Initialize data bindings
	editor.baseURLBinding = binding.NewString()
	editor.titleBinding = binding.NewString()
	editor.descriptionBinding = binding.NewString()
	editor.languageBinding = binding.NewString()
	editor.themeBinding = binding.NewString()
	
	// Set up change listeners
	editor.setupBindingListeners()
	
	editor.ExtendBaseWidget(editor)
	return editor
}

// LoadConfig loads configuration from file
func (ce *ConfigEditor) LoadConfig() error {
	config, err := ce.configRepo.LoadSiteConfig(ce.configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	ce.config = config
	ce.updateBindingsFromConfig()
	ce.isModified = false
	
	return nil
}

// SaveConfig saves the current configuration
func (ce *ConfigEditor) SaveConfig() error {
	if ce.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Update config from bindings
	ce.updateConfigFromBindings()
	
	// Validate configuration
	if err := ce.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Save to file
	err := ce.configRepo.SaveSiteConfig(ce.configPath, ce.config)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	ce.isModified = false
	
	// Call save callback if set
	if ce.onSave != nil {
		return ce.onSave(ce.config)
	}
	
	return nil
}

// ResetConfig resets the configuration to the last saved state
func (ce *ConfigEditor) ResetConfig() error {
	if ce.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Reload from file
	err := ce.LoadConfig()
	if err != nil {
		return err
	}
	
	// Call reset callback if set
	if ce.onReset != nil {
		ce.onReset()
	}
	
	return nil
}

// CreateRenderer creates the visual representation of the config editor
func (ce *ConfigEditor) CreateRenderer() fyne.WidgetRenderer {
	ce.setupUI()
	return widget.NewSimpleRenderer(ce.container)
}

// setupUI creates the user interface components
func (ce *ConfigEditor) setupUI() {
	// Create form for basic configuration
	ce.form = &widget.Form{}
	
	// Basic site information
	baseURLEntry := widget.NewEntryWithData(ce.baseURLBinding)
	baseURLEntry.SetPlaceHolder("https://example.com")
	ce.form.Append("Base URL", baseURLEntry)
	
	titleEntry := widget.NewEntryWithData(ce.titleBinding)
	titleEntry.SetPlaceHolder("My Hugo Site")
	ce.form.Append("Site Title", titleEntry)
	
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.Bind(ce.descriptionBinding)
	descriptionEntry.SetPlaceHolder("A brief description of your site")
	descriptionEntry.Resize(fyne.NewSize(400, 80))
	ce.form.Append("Description", descriptionEntry)
	
	languageEntry := widget.NewEntryWithData(ce.languageBinding)
	languageEntry.SetPlaceHolder("en")
	ce.form.Append("Language Code", languageEntry)
	
	themeEntry := widget.NewEntryWithData(ce.themeBinding)
	themeEntry.SetPlaceHolder("theme-name")
	ce.form.Append("Theme", themeEntry)
	
	// Create dynamic parameters editor
	ce.paramsEditor = NewDynamicParamsEditor()
	ce.paramsEditor.SetOnChanged(ce.onParameterChanged)
	
	// Create action buttons
	ce.saveButton = widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), func() {
		ce.handleSave()
	})
	ce.saveButton.Importance = widget.HighImportance
	
	ce.resetButton = widget.NewButtonWithIcon("Reset Changes", theme.ViewRefreshIcon(), func() {
		ce.handleReset()
	})
	
	validateButton := widget.NewButtonWithIcon("Validate", theme.ConfirmIcon(), func() {
		ce.handleValidate()
	})
	
	buttonContainer := container.NewHBox(
		ce.saveButton,
		ce.resetButton,
		validateButton,
	)
	
	// Create main content
	content := container.NewVBox(
		widget.NewCard("Basic Site Information", "", ce.form),
		widget.NewCard("Custom Parameters", "", ce.paramsEditor),
		widget.NewSeparator(),
		buttonContainer,
	)
	
	ce.container = container.NewScroll(content)
	ce.container.SetMinSize(fyne.NewSize(600, 500))
}

// setupBindingListeners sets up listeners for data binding changes
func (ce *ConfigEditor) setupBindingListeners() {
	ce.baseURLBinding.AddListener(binding.NewDataListener(func() {
		ce.markModified()
	}))
	
	ce.titleBinding.AddListener(binding.NewDataListener(func() {
		ce.markModified()
	}))
	
	ce.descriptionBinding.AddListener(binding.NewDataListener(func() {
		ce.markModified()
	}))
	
	ce.languageBinding.AddListener(binding.NewDataListener(func() {
		ce.markModified()
	}))
	
	ce.themeBinding.AddListener(binding.NewDataListener(func() {
		ce.markModified()
	}))
}

// updateBindingsFromConfig updates the UI bindings from the loaded configuration
func (ce *ConfigEditor) updateBindingsFromConfig() {
	if ce.config == nil {
		return
	}
	
	ce.baseURLBinding.Set(ce.config.BaseURL)
	ce.titleBinding.Set(ce.config.Title)
	ce.descriptionBinding.Set(ce.config.Description)
	ce.languageBinding.Set(ce.config.LanguageCode)
	ce.themeBinding.Set(ce.config.Theme)
	
	// Update parameters editor
	if ce.paramsEditor != nil {
		ce.paramsEditor.SetParams(ce.config.Params)
	}
}

// updateConfigFromBindings updates the configuration from the UI bindings
func (ce *ConfigEditor) updateConfigFromBindings() {
	if ce.config == nil {
		ce.config = &models.SiteConfig{}
	}
	
	baseURL, _ := ce.baseURLBinding.Get()
	title, _ := ce.titleBinding.Get()
	description, _ := ce.descriptionBinding.Get()
	language, _ := ce.languageBinding.Get()
	theme, _ := ce.themeBinding.Get()
	
	ce.config.BaseURL = baseURL
	ce.config.Title = title
	ce.config.Description = description
	ce.config.LanguageCode = language
	ce.config.Theme = theme
	
	// Update parameters from editor
	if ce.paramsEditor != nil {
		ce.config.Params = ce.paramsEditor.GetParams()
	}
}

// Event handlers
func (ce *ConfigEditor) handleSave() {
	err := ce.SaveConfig()
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	dialog.ShowInformation("Success", "Configuration saved successfully!", fyne.CurrentApp().Driver().AllWindows()[0])
}

func (ce *ConfigEditor) handleReset() {
	dialog.ShowConfirm("Reset Configuration", 
		"Are you sure you want to reset all changes? This will discard any unsaved modifications.",
		func(confirmed bool) {
			if confirmed {
				err := ce.ResetConfig()
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				}
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (ce *ConfigEditor) handleValidate() {
	ce.updateConfigFromBindings()
	
	if ce.config == nil {
		dialog.ShowError(fmt.Errorf("no configuration to validate"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	err := ce.config.Validate()
	if err != nil {
		dialog.ShowError(fmt.Errorf("validation failed: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	
	dialog.ShowInformation("Validation", "Configuration is valid!", fyne.CurrentApp().Driver().AllWindows()[0])
}

func (ce *ConfigEditor) onParameterChanged() {
	ce.markModified()
}

// Utility methods
func (ce *ConfigEditor) markModified() {
	ce.isModified = true
}

func (ce *ConfigEditor) IsModified() bool {
	return ce.isModified
}

// Callback setters
func (ce *ConfigEditor) SetOnSave(callback func(*models.SiteConfig) error) {
	ce.onSave = callback
}

func (ce *ConfigEditor) SetOnReset(callback func()) {
	ce.onReset = callback
}

func (ce *ConfigEditor) SetOnClose(callback func()) {
	ce.onClose = callback
}

// GetWidget returns the widget for embedding in other containers
func (ce *ConfigEditor) GetWidget() fyne.CanvasObject {
	return ce
}

// GetConfig returns the current configuration
func (ce *ConfigEditor) GetConfig() *models.SiteConfig {
	return ce.config
}

// NewDynamicParamsEditor creates a new dynamic parameters editor
func NewDynamicParamsEditor() *DynamicParamsEditor {
	editor := &DynamicParamsEditor{
		params:  make(map[string]interface{}),
		entries: make(map[string]*widget.Entry),
	}
	
	editor.ExtendBaseWidget(editor)
	return editor
}

// CreateRenderer creates the visual representation of the dynamic params editor
func (dpe *DynamicParamsEditor) CreateRenderer() fyne.WidgetRenderer {
	dpe.setupUI()
	return widget.NewSimpleRenderer(dpe.container)
}

// setupUI creates the user interface for the dynamic parameters editor
func (dpe *DynamicParamsEditor) setupUI() {
	dpe.container = container.NewVBox()
	
	// Add button to add new parameters
	addButton := widget.NewButtonWithIcon("Add Parameter", theme.ContentAddIcon(), func() {
		dpe.showAddParameterDialog()
	})
	
	dpe.container.Add(addButton)
	dpe.refreshParametersList()
}

// SetParams sets the parameters to edit
func (dpe *DynamicParamsEditor) SetParams(params map[string]interface{}) {
	if params == nil {
		dpe.params = make(map[string]interface{})
	} else {
		dpe.params = make(map[string]interface{})
		for k, v := range params {
			dpe.params[k] = v
		}
	}
	
	dpe.refreshParametersList()
}

// GetParams returns the current parameters
func (dpe *DynamicParamsEditor) GetParams() map[string]interface{} {
	// Update params from entries
	for key, entry := range dpe.entries {
		text := entry.Text
		
		// Try to parse as different types
		if text == "true" || text == "false" {
			dpe.params[key] = text == "true"
		} else if val, err := strconv.Atoi(text); err == nil {
			dpe.params[key] = val
		} else if val, err := strconv.ParseFloat(text, 64); err == nil {
			dpe.params[key] = val
		} else {
			dpe.params[key] = text
		}
	}
	
	return dpe.params
}

// refreshParametersList refreshes the parameters list UI
func (dpe *DynamicParamsEditor) refreshParametersList() {
	// Clear existing entries (except add button)
	if len(dpe.container.Objects) > 1 {
		dpe.container.Objects = dpe.container.Objects[:1]
	}
	
	// Clear entries map
	dpe.entries = make(map[string]*widget.Entry)
	
	// Add parameter entries
	for key, value := range dpe.params {
		dpe.addParameterEntry(key, value)
	}
	
	dpe.container.Refresh()
}

// addParameterEntry adds a parameter entry to the UI
func (dpe *DynamicParamsEditor) addParameterEntry(key string, value interface{}) {
	entry := widget.NewEntry()
	entry.SetText(fmt.Sprintf("%v", value))
	entry.OnChanged = func(text string) {
		if dpe.onChanged != nil {
			dpe.onChanged()
		}
	}
	
	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dpe.removeParameter(key)
	})
	deleteButton.Resize(fyne.NewSize(32, 32))
	
	paramContainer := container.NewBorder(
		nil, nil, 
		widget.NewLabel(key+":"), 
		deleteButton,
		entry,
	)
	
	dpe.container.Add(paramContainer)
	dpe.entries[key] = entry
}

// showAddParameterDialog shows a dialog to add a new parameter
func (dpe *DynamicParamsEditor) showAddParameterDialog() {
	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("Parameter name")
	
	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Parameter value")
	
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name:", Widget: keyEntry},
			{Text: "Value:", Widget: valueEntry},
		},
	}
	
	dialog.ShowForm("Add Parameter", "Add", "Cancel", form.Items, func(confirmed bool) {
		if confirmed && keyEntry.Text != "" {
			key := strings.TrimSpace(keyEntry.Text)
			value := strings.TrimSpace(valueEntry.Text)
			
			// Don't overwrite existing parameters
			if _, exists := dpe.params[key]; !exists {
				dpe.params[key] = value
				dpe.addParameterEntry(key, value)
				
				if dpe.onChanged != nil {
					dpe.onChanged()
				}
			}
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

// removeParameter removes a parameter from the editor
func (dpe *DynamicParamsEditor) removeParameter(key string) {
	delete(dpe.params, key)
	delete(dpe.entries, key)
	dpe.refreshParametersList()
	
	if dpe.onChanged != nil {
		dpe.onChanged()
	}
}

// SetOnChanged sets the callback for when parameters change
func (dpe *DynamicParamsEditor) SetOnChanged(callback func()) {
	dpe.onChanged = callback
}