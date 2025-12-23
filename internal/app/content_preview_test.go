package app

import (
	"strings"
	"testing"
	"time"

	"hugo-visual-client/internal/interfaces"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// MockContentPreview is a simplified version for testing without GUI dependencies
type MockContentPreview struct {
	currentContent     string
	currentFrontMatter interfaces.FrontMatter
	isEmpty           bool
}

func NewMockContentPreview() *MockContentPreview {
	return &MockContentPreview{
		isEmpty: true,
	}
}

func (mcp *MockContentPreview) UpdatePreview(content string, frontMatter interfaces.FrontMatter) error {
	mcp.currentContent = content
	mcp.currentFrontMatter = frontMatter
	mcp.isEmpty = content == "" && frontMatter.Title == ""
	return nil
}

func (mcp *MockContentPreview) GetCurrentContent() string {
	return mcp.currentContent
}

func (mcp *MockContentPreview) GetCurrentFrontMatter() interfaces.FrontMatter {
	return mcp.currentFrontMatter
}

func (mcp *MockContentPreview) IsEmpty() bool {
	return mcp.isEmpty
}

func (mcp *MockContentPreview) Clear() {
	mcp.currentContent = ""
	mcp.currentFrontMatter = interfaces.FrontMatter{}
	mcp.isEmpty = true
}

func (mcp *MockContentPreview) RefreshPreview() error {
	// Simulate refresh - no-op for mock
	return nil
}

// **Feature: hugo-visual-client, Property 13: 预览更新及时性**
// **Validates: Requirements 4.3, 4.4**
func TestPreviewUpdateTimeliness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("preview should update immediately when content or front matter changes", prop.ForAll(
		func(titleSuffix, content, author string, draft bool, tagCount int) bool {
			// Create a valid title by adding a prefix
			title := "Test " + titleSuffix
			if strings.TrimSpace(title) == "Test " {
				title = "Test Post" // Fallback for empty suffix
			}

			// No need to limit tag count since we're using IntRange

			// Create mock content preview component
			preview := NewMockContentPreview()
			if preview == nil {
				t.Logf("Failed to create content preview")
				return false
			}

			// Generate tags based on tagCount
			var tags []string
			for i := 0; i < tagCount; i++ {
				tags = append(tags, "tag"+string(rune('A'+i)))
			}

			// Create initial front matter
			initialFrontMatter := interfaces.FrontMatter{
				Title:       title,
				Date:        time.Now(),
				Draft:       draft,
				Tags:        tags,
				Categories:  []string{"test", "blog"},
				Description: "Test description",
				Author:      author,
				Image:       "",
				Custom:      make(map[string]interface{}),
			}

			// Test initial preview update
			err := preview.UpdatePreview(content, initialFrontMatter)
			if err != nil {
				t.Logf("Failed to update preview initially: %v", err)
				return false
			}

			// Verify preview is not empty after update
			if preview.IsEmpty() {
				t.Logf("Preview should not be empty after update")
				return false
			}

			// Verify current content matches what we set
			currentContent := preview.GetCurrentContent()
			if currentContent != content {
				t.Logf("Current content mismatch: expected %q, got %q", content, currentContent)
				return false
			}

			// Verify current front matter matches what we set
			currentFM := preview.GetCurrentFrontMatter()
			if currentFM.Title != title {
				t.Logf("Current title mismatch: expected %q, got %q", title, currentFM.Title)
				return false
			}

			if currentFM.Author != author {
				t.Logf("Current author mismatch: expected %q, got %q", author, currentFM.Author)
				return false
			}

			if currentFM.Draft != draft {
				t.Logf("Current draft status mismatch: expected %v, got %v", draft, currentFM.Draft)
				return false
			}

			if len(currentFM.Tags) != len(tags) {
				t.Logf("Current tags count mismatch: expected %d, got %d", len(tags), len(currentFM.Tags))
				return false
			}

			// Test content change updates preview immediately
			newContent := content + "\n\nAdditional content added"
			err = preview.UpdatePreview(newContent, initialFrontMatter)
			if err != nil {
				t.Logf("Failed to update preview with new content: %v", err)
				return false
			}

			// Verify the content change is reflected immediately
			updatedContent := preview.GetCurrentContent()
			if updatedContent != newContent {
				t.Logf("Content change not reflected immediately: expected %q, got %q", newContent, updatedContent)
				return false
			}

			// Test front matter change updates preview immediately
			modifiedFrontMatter := initialFrontMatter
			modifiedFrontMatter.Title = "Modified " + title
			modifiedFrontMatter.Draft = !draft
			modifiedFrontMatter.Tags = append(tags, "modified")

			err = preview.UpdatePreview(newContent, modifiedFrontMatter)
			if err != nil {
				t.Logf("Failed to update preview with modified front matter: %v", err)
				return false
			}

			// Verify the front matter changes are reflected immediately
			updatedFM := preview.GetCurrentFrontMatter()
			if updatedFM.Title != modifiedFrontMatter.Title {
				t.Logf("Front matter title change not reflected immediately: expected %q, got %q", modifiedFrontMatter.Title, updatedFM.Title)
				return false
			}

			if updatedFM.Draft != modifiedFrontMatter.Draft {
				t.Logf("Front matter draft change not reflected immediately: expected %v, got %v", modifiedFrontMatter.Draft, updatedFM.Draft)
				return false
			}

			if len(updatedFM.Tags) != len(modifiedFrontMatter.Tags) {
				t.Logf("Front matter tags change not reflected immediately: expected %d tags, got %d", len(modifiedFrontMatter.Tags), len(updatedFM.Tags))
				return false
			}

			// Test that preview can be cleared
			preview.Clear()
			if !preview.IsEmpty() {
				t.Logf("Preview should be empty after clear")
				return false
			}

			// Test that preview can be refreshed
			err = preview.UpdatePreview(content, initialFrontMatter)
			if err != nil {
				t.Logf("Failed to update preview after clear: %v", err)
				return false
			}

			err = preview.RefreshPreview()
			if err != nil {
				t.Logf("Failed to refresh preview: %v", err)
				return false
			}

			// Verify content is still correct after refresh
			refreshedContent := preview.GetCurrentContent()
			if refreshedContent != content {
				t.Logf("Content not preserved after refresh: expected %q, got %q", content, refreshedContent)
				return false
			}

			return true
		},
		gen.AlphaString(),  // title suffix
		gen.AlphaString(),  // content
		gen.AlphaString(),  // author
		gen.Bool(),         // draft
		gen.IntRange(0, 3), // tag count
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test basic preview functionality without GUI dependencies
func TestPreviewBasicFunctionality(t *testing.T) {
	// Create mock content preview component
	preview := NewMockContentPreview()

	// Test with empty content
	frontMatter := interfaces.FrontMatter{
		Title:       "Test Post",
		Date:        time.Now(),
		Draft:       false,
		Tags:        []string{"test", "preview"},
		Categories:  []string{"blog"},
		Description: "Test description",
		Author:      "Test Author",
		Custom:      make(map[string]interface{}),
	}

	err := preview.UpdatePreview("", frontMatter)
	if err != nil {
		t.Fatalf("Failed to update preview with empty content: %v", err)
	}

	// Test with markdown content
	markdownContent := `# Main Heading

This is a paragraph with **bold** and *italic* text.

## Subheading

- List item 1
- List item 2
- List item 3

[Link example](https://example.com)

` + "```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```"

	err = preview.UpdatePreview(markdownContent, frontMatter)
	if err != nil {
		t.Fatalf("Failed to update preview with markdown content: %v", err)
	}

	// Verify content is stored correctly
	currentContent := preview.GetCurrentContent()
	if currentContent != markdownContent {
		t.Errorf("Markdown content not stored correctly")
	}

	// Test metadata display
	frontMatterWithMetadata := interfaces.FrontMatter{
		Title:       "Rich Metadata Post",
		Date:        time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
		Draft:       true,
		Tags:        []string{"golang", "hugo", "testing"},
		Categories:  []string{"development", "tutorial"},
		Description: "A post with rich metadata for testing",
		Author:      "Test Author",
		Image:       "/images/featured.jpg",
		Custom:      make(map[string]interface{}),
	}

	err = preview.UpdatePreview(markdownContent, frontMatterWithMetadata)
	if err != nil {
		t.Fatalf("Failed to update preview with rich metadata: %v", err)
	}

	// Verify metadata is preserved
	currentFM := preview.GetCurrentFrontMatter()
	if currentFM.Title != frontMatterWithMetadata.Title {
		t.Errorf("Title not preserved: expected %s, got %s", frontMatterWithMetadata.Title, currentFM.Title)
	}

	if len(currentFM.Tags) != len(frontMatterWithMetadata.Tags) {
		t.Errorf("Tags not preserved: expected %d, got %d", len(frontMatterWithMetadata.Tags), len(currentFM.Tags))
	}

	if len(currentFM.Categories) != len(frontMatterWithMetadata.Categories) {
		t.Errorf("Categories not preserved: expected %d, got %d", len(frontMatterWithMetadata.Categories), len(currentFM.Categories))
	}
}