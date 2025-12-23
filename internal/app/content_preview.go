package app

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/interfaces"
)

// ContentPreview handles the preview functionality for content
type ContentPreview struct {
	// UI components
	container    *container.Scroll
	htmlContent  *widget.RichText
	webView      fyne.CanvasObject // Placeholder for web view
	
	// State
	currentContent   string
	currentFrontMatter interfaces.FrontMatter
	hugoService      interfaces.HugoService
	projectPath      string
	previewMode      PreviewMode
	
	// Callbacks
	onError func(error)
}

// PreviewMode defines the preview rendering mode
type PreviewMode int

const (
	PreviewModeHTML PreviewMode = iota
	PreviewModeLive
)

// NewContentPreview creates a new content preview component
func NewContentPreview(hugoService interfaces.HugoService, projectPath string) *ContentPreview {
	cp := &ContentPreview{
		hugoService: hugoService,
		projectPath: projectPath,
		previewMode: PreviewModeHTML,
	}
	
	cp.setupUI()
	return cp
}

// setupUI initializes the preview UI
func (cp *ContentPreview) setupUI() {
	// Create HTML content widget
	cp.htmlContent = widget.NewRichText()
	cp.htmlContent.Wrapping = fyne.TextWrapWord
	
	// Create container
	cp.container = container.NewScroll(cp.htmlContent)
	cp.container.SetMinSize(fyne.NewSize(400, 300))
	
	// Set initial content
	cp.showWelcomeMessage()
}

// showWelcomeMessage displays a welcome message when no content is loaded
func (cp *ContentPreview) showWelcomeMessage() {
	welcomeText := `# Preview

Start editing content to see a live preview here.

The preview will update automatically as you type.`
	
	cp.htmlContent.ParseMarkdown(welcomeText)
}

// UpdatePreview updates the preview with new content and front matter
func (cp *ContentPreview) UpdatePreview(content string, frontMatter interfaces.FrontMatter) error {
	cp.currentContent = content
	cp.currentFrontMatter = frontMatter
	
	switch cp.previewMode {
	case PreviewModeHTML:
		return cp.updateHTMLPreview()
	case PreviewModeLive:
		return cp.updateLivePreview()
	default:
		return fmt.Errorf("unknown preview mode: %d", cp.previewMode)
	}
}

// updateHTMLPreview updates the HTML preview mode
func (cp *ContentPreview) updateHTMLPreview() error {
	if cp.currentContent == "" {
		cp.showWelcomeMessage()
		return nil
	}
	
	// Create a complete markdown document with front matter context
	previewContent := cp.buildPreviewContent()
	
	// Parse and display markdown
	cp.htmlContent.ParseMarkdown(previewContent)
	
	return nil
}

// updateLivePreview updates the live preview mode (Hugo server integration)
func (cp *ContentPreview) updateLivePreview() error {
	// This would integrate with Hugo's live server
	// For now, fall back to HTML preview
	return cp.updateHTMLPreview()
}

// SetProjectPath updates the project path for the preview component
func (cp *ContentPreview) SetProjectPath(projectPath string) {
	cp.projectPath = projectPath
}

// buildPreviewContent builds the complete content for preview
func (cp *ContentPreview) buildPreviewContent() string {
	var builder strings.Builder
	
	// Add title if available
	if cp.currentFrontMatter.Title != "" {
		builder.WriteString("# ")
		builder.WriteString(cp.currentFrontMatter.Title)
		builder.WriteString("\n\n")
	}
	
	// Add metadata section
	if cp.shouldShowMetadata() {
		builder.WriteString(cp.buildMetadataSection())
		builder.WriteString("\n---\n\n")
	}
	
	// Add main content
	builder.WriteString(cp.currentContent)
	
	return builder.String()
}

// shouldShowMetadata determines if metadata should be shown in preview
func (cp *ContentPreview) shouldShowMetadata() bool {
	fm := cp.currentFrontMatter
	return fm.Date != (time.Time{}) || len(fm.Tags) > 0 || len(fm.Categories) > 0 || fm.Author != ""
}

// buildMetadataSection builds the metadata section for preview
func (cp *ContentPreview) buildMetadataSection() string {
	var builder strings.Builder
	
	builder.WriteString("**Post Information**\n\n")
	
	// Date
	if cp.currentFrontMatter.Date != (time.Time{}) {
		builder.WriteString("**Date:** ")
		builder.WriteString(cp.currentFrontMatter.Date.Format("January 2, 2006"))
		builder.WriteString("\n\n")
	}
	
	// Author
	if cp.currentFrontMatter.Author != "" {
		builder.WriteString("**Author:** ")
		builder.WriteString(cp.currentFrontMatter.Author)
		builder.WriteString("\n\n")
	}
	
	// Draft status
	if cp.currentFrontMatter.Draft {
		builder.WriteString("**Status:** Draft\n\n")
	}
	
	// Tags
	if len(cp.currentFrontMatter.Tags) > 0 {
		builder.WriteString("**Tags:** ")
		builder.WriteString(strings.Join(cp.currentFrontMatter.Tags, ", "))
		builder.WriteString("\n\n")
	}
	
	// Categories
	if len(cp.currentFrontMatter.Categories) > 0 {
		builder.WriteString("**Categories:** ")
		builder.WriteString(strings.Join(cp.currentFrontMatter.Categories, ", "))
		builder.WriteString("\n\n")
	}
	
	// Description
	if cp.currentFrontMatter.Description != "" {
		builder.WriteString("**Description:** ")
		builder.WriteString(cp.currentFrontMatter.Description)
		builder.WriteString("\n\n")
	}
	
	return builder.String()
}

// SetPreviewMode sets the preview mode
func (cp *ContentPreview) SetPreviewMode(mode PreviewMode) {
	cp.previewMode = mode
	
	// Update preview with current content
	if cp.currentContent != "" {
		cp.UpdatePreview(cp.currentContent, cp.currentFrontMatter)
	}
}

// GetWidget returns the preview widget
func (cp *ContentPreview) GetWidget() fyne.CanvasObject {
	return cp.container
}

// SetOnError sets the error callback
func (cp *ContentPreview) SetOnError(callback func(error)) {
	cp.onError = callback
}

// Clear clears the preview content
func (cp *ContentPreview) Clear() {
	cp.currentContent = ""
	cp.currentFrontMatter = interfaces.FrontMatter{}
	cp.showWelcomeMessage()
}

// ScrollToTop scrolls the preview to the top
func (cp *ContentPreview) ScrollToTop() {
	cp.container.ScrollToTop()
}

// ScrollToBottom scrolls the preview to the bottom
func (cp *ContentPreview) ScrollToBottom() {
	cp.container.ScrollToBottom()
}

// GetScrollPosition returns the current scroll position (0.0 to 1.0)
func (cp *ContentPreview) GetScrollPosition() float32 {
	// This is a simplified implementation
	// In a real implementation, we'd need to access the scroll container's position
	return 0.0
}

// SetScrollPosition sets the scroll position (0.0 to 1.0)
func (cp *ContentPreview) SetScrollPosition(position float32) {
	// This is a simplified implementation
	// In a real implementation, we'd need to set the scroll container's position
	if position <= 0.0 {
		cp.ScrollToTop()
	} else if position >= 1.0 {
		cp.ScrollToBottom()
	}
}

// PreviewTemplate represents a template for preview rendering
type PreviewTemplate struct {
	Title       string
	Content     template.HTML
	FrontMatter interfaces.FrontMatter
	Date        string
	Tags        []string
	Categories  []string
}

// buildPreviewTemplate builds a template for advanced preview rendering
func (cp *ContentPreview) buildPreviewTemplate() PreviewTemplate {
	dateStr := ""
	if cp.currentFrontMatter.Date != (time.Time{}) {
		dateStr = cp.currentFrontMatter.Date.Format("January 2, 2006")
	}
	
	return PreviewTemplate{
		Title:       cp.currentFrontMatter.Title,
		Content:     template.HTML(cp.currentContent),
		FrontMatter: cp.currentFrontMatter,
		Date:        dateStr,
		Tags:        cp.currentFrontMatter.Tags,
		Categories:  cp.currentFrontMatter.Categories,
	}
}

// RefreshPreview forces a refresh of the preview
func (cp *ContentPreview) RefreshPreview() error {
	return cp.UpdatePreview(cp.currentContent, cp.currentFrontMatter)
}

// IsEmpty returns true if the preview has no content
func (cp *ContentPreview) IsEmpty() bool {
	return cp.currentContent == ""
}

// GetCurrentContent returns the current content being previewed
func (cp *ContentPreview) GetCurrentContent() string {
	return cp.currentContent
}

// GetCurrentFrontMatter returns the current front matter being previewed
func (cp *ContentPreview) GetCurrentFrontMatter() interfaces.FrontMatter {
	return cp.currentFrontMatter
}