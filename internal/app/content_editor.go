package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/interfaces"
)

// ContentEditor represents the content editing interface
type ContentEditor struct {
	// UI components
	container       *container.Split
	frontMatterForm *widget.Form
	markdownEditor  *widget.Entry
	previewComponent *ContentPreview
	
	// Data bindings
	titleBinding       binding.String
	dateBinding        binding.String
	draftBinding       binding.Bool
	tagsBinding        binding.String
	categoriesBinding  binding.String
	descriptionBinding binding.String
	authorBinding      binding.String
	imageBinding       binding.String
	contentBinding     binding.String
	
	// State
	currentPath    string
	contentManager interfaces.ContentManager
	hugoService    interfaces.HugoService
	projectPath    string
	isModified     bool
	onSave         func(string) // callback when content is saved
	onClose        func(string) // callback when editor is closed
	
	// Preview update timer
	previewUpdateTimer *time.Timer
}

// NewContentEditor creates a new content editor
func NewContentEditor(contentManager interfaces.ContentManager, hugoService interfaces.HugoService, projectPath string) *ContentEditor {
	ce := &ContentEditor{
		contentManager: contentManager,
		hugoService:    hugoService,
		projectPath:    projectPath,
	}
	
	ce.setupBindings()
	ce.setupUI()
	
	return ce
}

// setupBindings initializes data bindings
func (ce *ContentEditor) setupBindings() {
	ce.titleBinding = binding.NewString()
	ce.dateBinding = binding.NewString()
	ce.draftBinding = binding.NewBool()
	ce.tagsBinding = binding.NewString()
	ce.categoriesBinding = binding.NewString()
	ce.descriptionBinding = binding.NewString()
	ce.authorBinding = binding.NewString()
	ce.imageBinding = binding.NewString()
	ce.contentBinding = binding.NewString()
	
	// Set default values
	ce.titleBinding.Set("")
	ce.dateBinding.Set(time.Now().Format("2006-01-02T15:04:05Z07:00"))
	ce.draftBinding.Set(true)
	ce.tagsBinding.Set("")
	ce.categoriesBinding.Set("")
	ce.descriptionBinding.Set("")
	ce.authorBinding.Set("")
	ce.imageBinding.Set("")
	ce.contentBinding.Set("")
}

// setupUI initializes the user interface
func (ce *ContentEditor) setupUI() {
	// Create front matter form
	ce.createFrontMatterForm()
	
	// Create markdown editor
	ce.createMarkdownEditor()
	
	// Create preview component
	ce.createPreviewComponent()
	
	// Create the main split container
	leftPanel := container.NewVBox(
		widget.NewCard("Front Matter", "", ce.frontMatterForm),
		widget.NewCard("Content", "", container.NewBorder(
			nil, nil, nil, nil, ce.markdownEditor,
		)),
	)
	
	rightPanel := container.NewVBox(
		widget.NewCard("Preview", "", ce.previewComponent.GetWidget()),
	)
	
	ce.container = container.NewHSplit(leftPanel, rightPanel)
	ce.container.SetOffset(0.6) // 60% for editor, 40% for preview
	
	// Set up real-time preview updates
	ce.setupPreviewUpdates()
}

// createFrontMatterForm creates the front matter editing form
func (ce *ContentEditor) createFrontMatterForm() {
	// Title entry
	titleEntry := widget.NewEntryWithData(ce.titleBinding)
	titleEntry.SetPlaceHolder("Enter post title...")
	titleEntry.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Date entry
	dateEntry := widget.NewEntryWithData(ce.dateBinding)
	dateEntry.SetPlaceHolder("YYYY-MM-DDTHH:MM:SSZ")
	dateEntry.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Draft checkbox
	draftCheck := widget.NewCheckWithData("Draft", ce.draftBinding)
	draftCheck.OnChanged = func(bool) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Tags entry
	tagsEntry := widget.NewEntryWithData(ce.tagsBinding)
	tagsEntry.SetPlaceHolder("tag1, tag2, tag3")
	tagsEntry.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Categories entry
	categoriesEntry := widget.NewEntryWithData(ce.categoriesBinding)
	categoriesEntry.SetPlaceHolder("category1, category2")
	categoriesEntry.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Description entry
	descriptionEntry := widget.NewEntryWithData(ce.descriptionBinding)
	descriptionEntry.SetPlaceHolder("Brief description of the post...")
	descriptionEntry.MultiLine = true
	descriptionEntry.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Author entry
	authorEntry := widget.NewEntryWithData(ce.authorBinding)
	authorEntry.SetPlaceHolder("Author name")
	authorEntry.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Image entry
	imageEntry := widget.NewEntryWithData(ce.imageBinding)
	imageEntry.SetPlaceHolder("Path to featured image")
	imageEntry.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Create form items
	ce.frontMatterForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Title", Widget: titleEntry, HintText: "The title of your post"},
			{Text: "Date", Widget: dateEntry, HintText: "Publication date (ISO 8601 format)"},
			{Text: "Draft", Widget: draftCheck, HintText: "Check if this is a draft"},
			{Text: "Tags", Widget: tagsEntry, HintText: "Comma-separated tags"},
			{Text: "Categories", Widget: categoriesEntry, HintText: "Comma-separated categories"},
			{Text: "Description", Widget: descriptionEntry, HintText: "Meta description for SEO"},
			{Text: "Author", Widget: authorEntry, HintText: "Post author"},
			{Text: "Image", Widget: imageEntry, HintText: "Featured image path"},
		},
		SubmitText: "Save",
		OnSubmit:   ce.saveContent,
		OnCancel:   ce.cancelEdit,
	}
}

// createMarkdownEditor creates the markdown content editor
func (ce *ContentEditor) createMarkdownEditor() {
	ce.markdownEditor = widget.NewMultiLineEntry()
	ce.markdownEditor.Wrapping = fyne.TextWrapWord
	ce.markdownEditor.SetPlaceHolder("Write your content in Markdown...")
	
	// Bind to content binding
	ce.markdownEditor.Bind(ce.contentBinding)
	ce.markdownEditor.OnChanged = func(string) { 
		ce.markModified()
		ce.schedulePreviewUpdate()
	}
	
	// Set a reasonable size
	ce.markdownEditor.Resize(fyne.NewSize(600, 400))
}

// createPreviewComponent creates the preview component
func (ce *ContentEditor) createPreviewComponent() {
	ce.previewComponent = NewContentPreview(ce.hugoService, ce.projectPath)
	ce.previewComponent.SetOnError(func(err error) {
		fmt.Printf("Preview error: %v\n", err)
	})
}

// LoadContent loads content from a file path
func (ce *ContentEditor) LoadContent(path string) error {
	if ce.contentManager == nil {
		return fmt.Errorf("content manager not initialized")
	}
	
	// Get content item
	contentItem, err := ce.contentManager.GetContent(path)
	if err != nil {
		return fmt.Errorf("failed to load content: %w", err)
	}
	
	ce.currentPath = path
	
	// Parse front matter and populate form
	err = ce.populateFromContentItem(contentItem)
	if err != nil {
		return fmt.Errorf("failed to populate editor: %w", err)
	}
	
	ce.isModified = false
	return nil
}

// CreateNewContent creates a new content file
func (ce *ContentEditor) CreateNewContent(path string) error {
	ce.currentPath = path
	
	// Set default values
	ce.titleBinding.Set("New Post")
	ce.dateBinding.Set(time.Now().Format("2006-01-02T15:04:05Z07:00"))
	ce.draftBinding.Set(true)
	ce.tagsBinding.Set("")
	ce.categoriesBinding.Set("")
	ce.descriptionBinding.Set("")
	ce.authorBinding.Set("")
	ce.imageBinding.Set("")
	ce.contentBinding.Set("# New Post\n\nWrite your content here...")
	
	ce.isModified = true
	return nil
}

// populateFromContentItem populates the editor from a content item
func (ce *ContentEditor) populateFromContentItem(item *interfaces.ContentItem) error {
	// Parse front matter from the content item
	frontMatter, content, err := ce.contentManager.ParseFrontMatter(item.Content)
	if err != nil {
		// If parsing fails, use the raw content
		frontMatter = interfaces.FrontMatter{
			Title: item.Title,
			Date:  item.Date,
			Draft: item.Draft,
			Tags:  item.Tags,
			Categories: item.Categories,
		}
		content = item.Content
	}
	
	// Populate bindings
	ce.titleBinding.Set(frontMatter.Title)
	ce.dateBinding.Set(frontMatter.Date.Format("2006-01-02T15:04:05Z07:00"))
	ce.draftBinding.Set(frontMatter.Draft)
	ce.tagsBinding.Set(strings.Join(frontMatter.Tags, ", "))
	ce.categoriesBinding.Set(strings.Join(frontMatter.Categories, ", "))
	ce.descriptionBinding.Set(frontMatter.Description)
	ce.authorBinding.Set(frontMatter.Author)
	ce.imageBinding.Set(frontMatter.Image)
	ce.contentBinding.Set(content)
	
	return nil
}

// saveContent saves the current content
func (ce *ContentEditor) saveContent() {
	if ce.contentManager == nil {
		return
	}
	
	// Build front matter from form data
	frontMatter, err := ce.buildFrontMatter()
	if err != nil {
		// Show error dialog
		ce.showError("Invalid Front Matter", err.Error())
		return
	}
	
	// Get content
	content, _ := ce.contentBinding.Get()
	
	// Save content
	var saveErr error
	if ce.currentPath == "" {
		saveErr = fmt.Errorf("no file path specified")
	} else {
		// Check if file exists to determine create vs update
		_, getErr := ce.contentManager.GetContent(ce.currentPath)
		if getErr != nil {
			// File doesn't exist, create new
			saveErr = ce.contentManager.CreateContent(ce.currentPath, frontMatter, content)
		} else {
			// File exists, update
			saveErr = ce.contentManager.UpdateContent(ce.currentPath, frontMatter, content)
		}
	}
	
	if saveErr != nil {
		ce.showError("Save Failed", saveErr.Error())
		return
	}
	
	ce.isModified = false
	
	// Call save callback
	if ce.onSave != nil {
		ce.onSave(ce.currentPath)
	}
}

// cancelEdit cancels the current edit
func (ce *ContentEditor) cancelEdit() {
	if ce.isModified {
		// TODO: Show confirmation dialog
		// For now, just proceed with cancel
	}
	
	// Call close callback
	if ce.onClose != nil {
		ce.onClose(ce.currentPath)
	}
}

// buildFrontMatter builds a FrontMatter struct from form data
func (ce *ContentEditor) buildFrontMatter() (interfaces.FrontMatter, error) {
	title, _ := ce.titleBinding.Get()
	dateStr, _ := ce.dateBinding.Get()
	draft, _ := ce.draftBinding.Get()
	tagsStr, _ := ce.tagsBinding.Get()
	categoriesStr, _ := ce.categoriesBinding.Get()
	description, _ := ce.descriptionBinding.Get()
	author, _ := ce.authorBinding.Get()
	image, _ := ce.imageBinding.Get()
	
	// Parse date
	var date time.Time
	var err error
	if dateStr != "" {
		date, err = time.Parse("2006-01-02T15:04:05Z07:00", dateStr)
		if err != nil {
			// Try alternative formats
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return interfaces.FrontMatter{}, fmt.Errorf("invalid date format: %s", dateStr)
			}
		}
	} else {
		date = time.Now()
	}
	
	// Parse tags and categories
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	
	var categories []string
	if categoriesStr != "" {
		for _, category := range strings.Split(categoriesStr, ",") {
			category = strings.TrimSpace(category)
			if category != "" {
				categories = append(categories, category)
			}
		}
	}
	
	frontMatter := interfaces.FrontMatter{
		Title:       title,
		Date:        date,
		Draft:       draft,
		Tags:        tags,
		Categories:  categories,
		Description: description,
		Author:      author,
		Image:       image,
		Custom:      make(map[string]interface{}),
	}
	
	return frontMatter, nil
}

// markModified marks the content as modified
func (ce *ContentEditor) markModified() {
	ce.isModified = true
}

// showError shows an error dialog
func (ce *ContentEditor) showError(title, message string) {
	// Create error dialog
	errorLabel := widget.NewLabel(message)
	errorLabel.Wrapping = fyne.TextWrapWord
	
	// For now, just print to console
	fmt.Printf("Error - %s: %s\n", title, message)
}

// GetWidget returns the main widget for the content editor
func (ce *ContentEditor) GetWidget() fyne.CanvasObject {
	return ce.container
}

// SetOnSave sets the callback for when content is saved
func (ce *ContentEditor) SetOnSave(callback func(string)) {
	ce.onSave = callback
}

// SetOnClose sets the callback for when the editor is closed
func (ce *ContentEditor) SetOnClose(callback func(string)) {
	ce.onClose = callback
}

// IsModified returns whether the content has been modified
func (ce *ContentEditor) IsModified() bool {
	return ce.isModified
}

// GetCurrentPath returns the current file path being edited
func (ce *ContentEditor) GetCurrentPath() string {
	return ce.currentPath
}

// GetTitle returns the current title for display purposes
func (ce *ContentEditor) GetTitle() string {
	title, _ := ce.titleBinding.Get()
	if title == "" {
		if ce.currentPath != "" {
			return filepath.Base(ce.currentPath)
		}
		return "New Content"
	}
	return title
}

// Focus sets focus to the appropriate widget
func (ce *ContentEditor) Focus() {
	// Focus on the title field if it's empty, otherwise focus on content
	title, _ := ce.titleBinding.Get()
	if title == "" {
		// Focus on title entry - this would need access to the actual widget
		// For now, we'll skip this implementation detail
	} else {
		ce.markdownEditor.FocusGained()
	}
}

// Validate validates the current content
func (ce *ContentEditor) Validate() error {
	title, _ := ce.titleBinding.Get()
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	
	dateStr, _ := ce.dateBinding.Get()
	if dateStr != "" {
		_, err := time.Parse("2006-01-02T15:04:05Z07:00", dateStr)
		if err != nil {
			_, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %s", dateStr)
			}
		}
	}
	
	return nil
}

// Preview Update Methods

// setupPreviewUpdates sets up real-time preview updates
func (ce *ContentEditor) setupPreviewUpdates() {
	// Initial preview update
	ce.updatePreview()
}

// schedulePreviewUpdate schedules a preview update with debouncing
func (ce *ContentEditor) schedulePreviewUpdate() {
	// Cancel existing timer
	if ce.previewUpdateTimer != nil {
		ce.previewUpdateTimer.Stop()
	}
	
	// Schedule new update with 500ms delay to debounce rapid changes
	ce.previewUpdateTimer = time.AfterFunc(500*time.Millisecond, func() {
		ce.updatePreview()
	})
}

// updatePreview updates the preview with current content and front matter
func (ce *ContentEditor) updatePreview() {
	if ce.previewComponent == nil {
		return
	}
	
	// Build front matter from current form data
	frontMatter, err := ce.buildFrontMatter()
	if err != nil {
		// If front matter is invalid, show error in preview
		ce.previewComponent.Clear()
		return
	}
	
	// Get current content
	content, _ := ce.contentBinding.Get()
	
	// Update preview
	err = ce.previewComponent.UpdatePreview(content, frontMatter)
	if err != nil {
		fmt.Printf("Failed to update preview: %v\n", err)
	}
}

// syncScrollPosition synchronizes scroll position between editor and preview
func (ce *ContentEditor) syncScrollPosition() {
	// This is a placeholder for scroll synchronization
	// In a full implementation, we would:
	// 1. Get the current cursor position in the markdown editor
	// 2. Calculate the corresponding position in the preview
	// 3. Scroll the preview to match
	
	// For now, we'll implement a simple version
	if ce.previewComponent != nil {
		// Get approximate scroll position based on cursor position
		// This is a simplified implementation
		cursorPos := ce.markdownEditor.CursorRow
		totalLines := len(strings.Split(ce.markdownEditor.Text, "\n"))
		
		if totalLines > 0 {
			scrollRatio := float32(cursorPos) / float32(totalLines)
			ce.previewComponent.SetScrollPosition(scrollRatio)
		}
	}
}

// GetPreviewComponent returns the preview component
func (ce *ContentEditor) GetPreviewComponent() *ContentPreview {
	return ce.previewComponent
}

// SetPreviewMode sets the preview mode
func (ce *ContentEditor) SetPreviewMode(mode PreviewMode) {
	if ce.previewComponent != nil {
		ce.previewComponent.SetPreviewMode(mode)
	}
}

// RefreshPreview forces a refresh of the preview
func (ce *ContentEditor) RefreshPreview() {
	ce.updatePreview()
}

// SetProjectPath updates the project path for the editor
func (ce *ContentEditor) SetProjectPath(projectPath string) {
	ce.projectPath = projectPath
	
	// Update preview component with new project path
	if ce.previewComponent != nil {
		ce.previewComponent.SetProjectPath(projectPath)
	}
	
	// Refresh preview with new project context
	ce.updatePreview()
}

// RefreshContent refreshes the content from the file system
func (ce *ContentEditor) RefreshContent() {
	if ce.currentPath == "" || ce.contentManager == nil {
		return
	}
	
	// Check if content has been modified externally
	contentItem, err := ce.contentManager.GetContent(ce.currentPath)
	if err != nil {
		// File might have been deleted
		fmt.Printf("Warning: Could not refresh content for %s: %v\n", ce.currentPath, err)
		return
	}
	
	// If the editor has unsaved changes, don't overwrite them
	if ce.isModified {
		fmt.Printf("Content has local modifications, skipping refresh for %s\n", ce.currentPath)
		return
	}
	
	// Reload content from file
	err = ce.populateFromContentItem(contentItem)
	if err != nil {
		fmt.Printf("Failed to refresh content: %v\n", err)
		return
	}
	
	// Update preview
	ce.updatePreview()
}