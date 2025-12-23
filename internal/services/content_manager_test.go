package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hugo-visual-client/internal/interfaces"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestContentManager_CreateAndGetContent(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create content directory
	contentDir := filepath.Join(tempDir, "content")
	err = os.MkdirAll(contentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create content dir: %v", err)
	}

	// Create content manager
	cm := NewContentManager(tempDir)

	// Test data
	frontMatter := interfaces.FrontMatter{
		Title:       "Test Post",
		Date:        time.Now(),
		Draft:       false,
		Tags:        []string{"test", "hugo"},
		Categories:  []string{"blog"},
		Description: "A test post",
		Author:      "Test Author",
	}
	content := "This is test content for the blog post."
	filePath := "content/test-post.md"

	// Test CreateContent
	err = cm.CreateContent(filePath, frontMatter, content)
	if err != nil {
		t.Fatalf("Failed to create content: %v", err)
	}

	// Verify file exists
	fullPath := filepath.Join(tempDir, filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatalf("Content file was not created")
	}

	// Test GetContent
	contentItem, err := cm.GetContent(filePath)
	if err != nil {
		t.Fatalf("Failed to get content: %v", err)
	}

	// Verify content item
	if contentItem.Title != frontMatter.Title {
		t.Errorf("Expected title %s, got %s", frontMatter.Title, contentItem.Title)
	}
	if contentItem.Content != content {
		t.Errorf("Expected content %s, got %s", content, contentItem.Content)
	}
	if len(contentItem.Tags) != len(frontMatter.Tags) {
		t.Errorf("Expected %d tags, got %d", len(frontMatter.Tags), len(contentItem.Tags))
	}
}

func TestContentManager_ListContent(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create content directory
	contentDir := filepath.Join(tempDir, "content")
	err = os.MkdirAll(contentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create content dir: %v", err)
	}

	// Create content manager
	cm := NewContentManager(tempDir)

	// Create test content files
	testFiles := []struct {
		path    string
		title   string
		content string
	}{
		{"content/post1.md", "First Post", "Content of first post"},
		{"content/post2.md", "Second Post", "Content of second post"},
		{"content/blog/post3.md", "Third Post", "Content of third post"},
	}

	for _, tf := range testFiles {
		frontMatter := interfaces.FrontMatter{
			Title: tf.title,
			Date:  time.Now(),
			Draft: false,
		}
		
		// Create subdirectory if needed
		dir := filepath.Dir(filepath.Join(tempDir, tf.path))
		os.MkdirAll(dir, 0755)
		
		err = cm.CreateContent(tf.path, frontMatter, tf.content)
		if err != nil {
			t.Fatalf("Failed to create test content %s: %v", tf.path, err)
		}
	}

	// Test ListContent
	contentItems, err := cm.ListContent(tempDir)
	if err != nil {
		t.Fatalf("Failed to list content: %v", err)
	}

	// Verify we got all content items
	if len(contentItems) != len(testFiles) {
		t.Errorf("Expected %d content items, got %d", len(testFiles), len(contentItems))
	}

	// Verify content items have correct titles
	titleMap := make(map[string]bool)
	for _, item := range contentItems {
		titleMap[item.Title] = true
	}

	for _, tf := range testFiles {
		if !titleMap[tf.title] {
			t.Errorf("Expected to find content with title %s", tf.title)
		}
	}
}

func TestContentManager_SearchContent(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create content directory
	contentDir := filepath.Join(tempDir, "content")
	err = os.MkdirAll(contentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create content dir: %v", err)
	}

	// Create content manager
	cm := NewContentManager(tempDir)

	// Create test content with different tags and content
	testData := []struct {
		path    string
		title   string
		tags    []string
		content string
	}{
		{"content/go-post.md", "Learning Go", []string{"go", "programming"}, "This post is about Go programming language"},
		{"content/hugo-post.md", "Hugo Tutorial", []string{"hugo", "web"}, "This post explains how to use Hugo"},
		{"content/other-post.md", "Random Post", []string{"random"}, "This is some random content"},
	}

	for _, td := range testData {
		frontMatter := interfaces.FrontMatter{
			Title: td.title,
			Date:  time.Now(),
			Tags:  td.tags,
			Draft: false,
		}
		
		err = cm.CreateContent(td.path, frontMatter, td.content)
		if err != nil {
			t.Fatalf("Failed to create test content %s: %v", td.path, err)
		}
	}

	// Test search by tag
	results, err := cm.SearchContent("programming")
	if err != nil {
		t.Fatalf("Failed to search content: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'programming' search, got %d", len(results))
	}
	if len(results) > 0 && results[0].Title != "Learning Go" {
		t.Errorf("Expected 'Learning Go' in results, got %s", results[0].Title)
	}

	// Test search by content
	results, err = cm.SearchContent("Hugo")
	if err != nil {
		t.Fatalf("Failed to search content: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'Hugo' search, got %d", len(results))
	}
	if len(results) > 0 && results[0].Title != "Hugo Tutorial" {
		t.Errorf("Expected 'Hugo Tutorial' in results, got %s", results[0].Title)
	}

	// Test search with no results
	results, err = cm.SearchContent("nonexistent")
	if err != nil {
		t.Fatalf("Failed to search content: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent' search, got %d", len(results))
	}
}

func TestContentManager_ParseFrontMatter(t *testing.T) {
	cm := NewContentManager("")

	testContent := `---
title: Test Post
date: 2023-01-01T00:00:00Z
draft: false
tags:
  - test
  - hugo
categories:
  - blog
description: A test post
author: Test Author
---
This is the content of the post.

It has multiple lines.`

	frontMatter, content, err := cm.ParseFrontMatter(testContent)
	if err != nil {
		t.Fatalf("Failed to parse front matter: %v", err)
	}

	if frontMatter.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got %s", frontMatter.Title)
	}
	
	if len(frontMatter.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(frontMatter.Tags))
	}
	
	expectedContent := "This is the content of the post.\n\nIt has multiple lines."
	if content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content)
	}
}

func TestContentManager_SerializeFrontMatter(t *testing.T) {
	cm := NewContentManager("")

	frontMatter := interfaces.FrontMatter{
		Title:       "Test Post",
		Date:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Draft:       false,
		Tags:        []string{"test", "hugo"},
		Categories:  []string{"blog"},
		Description: "A test post",
		Author:      "Test Author",
	}

	yaml, err := cm.SerializeFrontMatter(frontMatter)
	if err != nil {
		t.Fatalf("Failed to serialize front matter: %v", err)
	}

	// Parse it back to verify
	parsedFM, _, err := cm.ParseFrontMatter("---\n" + yaml + "---\ntest content")
	if err != nil {
		t.Fatalf("Failed to parse serialized front matter: %v", err)
	}

	if parsedFM.Title != frontMatter.Title {
		t.Errorf("Expected title %s, got %s", frontMatter.Title, parsedFM.Title)
	}
	
	if len(parsedFM.Tags) != len(frontMatter.Tags) {
		t.Errorf("Expected %d tags, got %d", len(frontMatter.Tags), len(parsedFM.Tags))
	}
}



// **Feature: hugo-visual-client, Property 4: 内容文件创建完整性**
// **Validates: Requirements 2.1**
func TestContentFileCreationIntegrity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("content file creation should create valid markdown files with front matter", prop.ForAll(
		func(title string, content string) bool {
			// Skip empty titles as they're not valid
			if strings.TrimSpace(title) == "" {
				return true
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create content directory
			contentDir := filepath.Join(tempDir, "content")
			err = os.MkdirAll(contentDir, 0755)
			if err != nil {
				t.Logf("Failed to create content dir: %v", err)
				return false
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Create front matter with generated data
			frontMatter := interfaces.FrontMatter{
				Title:      title,
				Date:       time.Now(),
				Draft:      false,
				Tags:       []string{"test"},
				Categories: []string{"blog"},
				Author:     "Test Author",
			}

			// Use a simple, safe file path
			filePath := "content/test-post.md"

			// Test CreateContent
			err = cm.CreateContent(filePath, frontMatter, content)
			if err != nil {
				t.Logf("Failed to create content: %v", err)
				return false
			}

			// Verify file exists
			fullPath := filepath.Join(tempDir, filePath)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Logf("Content file was not created")
				return false
			}

			// Read the created file and verify it has valid front matter
			contentItem, err := cm.GetContent(filePath)
			if err != nil {
				t.Logf("Failed to read created content: %v", err)
				return false
			}

			// Verify the content item has the expected properties
			if contentItem.Title != title {
				t.Logf("Title mismatch: expected %s, got %s", title, contentItem.Title)
				return false
			}

			if contentItem.Content != content {
				t.Logf("Content mismatch: expected %s, got %s", content, contentItem.Content)
				return false
			}

			// Verify that the file has valid front matter structure
			if contentItem.FrontMatter == nil {
				t.Logf("Front matter is nil")
				return false
			}

			// Verify that the file can be parsed and re-serialized
			_, serializedContent, err := cm.ParseFrontMatter(fmt.Sprintf("---\ntitle: %s\ndate: %s\ndraft: false\ntags:\n  - test\ncategories:\n  - blog\nauthor: Test Author\n---\n%s", title, time.Now().Format(time.RFC3339), content))
			if err != nil {
				t.Logf("Failed to parse front matter: %v", err)
				return false
			}

			if serializedContent != content {
				t.Logf("Content serialization mismatch: expected %s, got %s", content, serializedContent)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 50 }), // title
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) <= 200 }), // content
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
// **Feature: hugo-visual-client, Property 6: 内容编辑保存一致性**
// **Validates: Requirements 2.3**
func TestContentEditingSaveConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("content editing should preserve all data after save and reload", prop.ForAll(
		func(newTitle, newContent string) bool {
			// Skip empty titles as they're not valid
			if strings.TrimSpace(newTitle) == "" {
				return true
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create content directory
			contentDir := filepath.Join(tempDir, "content")
			err = os.MkdirAll(contentDir, 0755)
			if err != nil {
				t.Logf("Failed to create content dir: %v", err)
				return false
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Create original content
			originalFrontMatter := interfaces.FrontMatter{
				Title:      "Original Title",
				Date:       time.Now(),
				Draft:      false,
				Tags:       []string{"original", "test"},
				Categories: []string{"blog"},
				Author:     "Original Author",
			}

			filePath := "content/test-edit.md"

			// Create the original content
			err = cm.CreateContent(filePath, originalFrontMatter, "Original content")
			if err != nil {
				t.Logf("Failed to create original content: %v", err)
				return false
			}

			// Update the content with new data
			updatedFrontMatter := interfaces.FrontMatter{
				Title:      newTitle,
				Date:       time.Now().Add(time.Hour), // Different date
				Draft:      true, // Different draft status
				Tags:       []string{"updated", "test", "modified"},
				Categories: []string{"blog", "updated"},
				Author:     "Updated Author",
			}

			err = cm.UpdateContent(filePath, updatedFrontMatter, newContent)
			if err != nil {
				t.Logf("Failed to update content: %v", err)
				return false
			}

			// Read the updated content back
			contentItem, err := cm.GetContent(filePath)
			if err != nil {
				t.Logf("Failed to read updated content: %v", err)
				return false
			}

			// Verify all the updated properties are preserved
			if contentItem.Title != newTitle {
				t.Logf("Title not preserved: expected %s, got %s", newTitle, contentItem.Title)
				return false
			}

			if contentItem.Content != newContent {
				t.Logf("Content not preserved: expected %s, got %s", newContent, contentItem.Content)
				return false
			}

			if contentItem.Draft != true {
				t.Logf("Draft status not preserved: expected true, got %v", contentItem.Draft)
				return false
			}

			// Verify author is preserved in front matter
			if author, ok := contentItem.FrontMatter["author"]; !ok || author != "Updated Author" {
				t.Logf("Author not preserved in front matter: expected 'Updated Author', got %v", author)
				return false
			}

			// Verify tags are preserved (order doesn't matter)
			expectedTags := []string{"updated", "test", "modified"}
			if len(contentItem.Tags) != len(expectedTags) {
				t.Logf("Tags length mismatch: expected %d, got %d", len(expectedTags), len(contentItem.Tags))
				return false
			}

			tagMap := make(map[string]bool)
			for _, tag := range contentItem.Tags {
				tagMap[tag] = true
			}
			for _, expectedTag := range expectedTags {
				if !tagMap[expectedTag] {
					t.Logf("Missing expected tag: %s", expectedTag)
					return false
				}
			}

			// Verify categories are preserved (order doesn't matter)
			expectedCategories := []string{"blog", "updated"}
			if len(contentItem.Categories) != len(expectedCategories) {
				t.Logf("Categories length mismatch: expected %d, got %d", len(expectedCategories), len(contentItem.Categories))
				return false
			}

			categoryMap := make(map[string]bool)
			for _, category := range contentItem.Categories {
				categoryMap[category] = true
			}
			for _, expectedCategory := range expectedCategories {
				if !categoryMap[expectedCategory] {
					t.Logf("Missing expected category: %s", expectedCategory)
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 50 }), // new title
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) <= 200 }), // new content
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
// **Feature: hugo-visual-client, Property 15: 内容列表完整性**
// **Validates: Requirements 5.1**
func TestContentListIntegrity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("content list should include all markdown files in content directory", prop.ForAll(
		func(numFiles int) bool {
			// Limit the number of files to a reasonable range
			if numFiles < 1 || numFiles > 10 {
				return true
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create content directory
			contentDir := filepath.Join(tempDir, "content")
			err = os.MkdirAll(contentDir, 0755)
			if err != nil {
				t.Logf("Failed to create content dir: %v", err)
				return false
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Create the specified number of content files
			expectedTitles := make(map[string]bool)
			for i := 0; i < numFiles; i++ {
				title := fmt.Sprintf("Test Post %d", i+1)
				frontMatter := interfaces.FrontMatter{
					Title:      title,
					Date:       time.Now().Add(time.Duration(i) * time.Hour),
					Draft:      i%2 == 0, // Alternate draft status
					Tags:       []string{fmt.Sprintf("tag%d", i), "test"},
					Categories: []string{"blog"},
					Author:     fmt.Sprintf("Author %d", i+1),
				}

				filePath := fmt.Sprintf("content/post-%d.md", i+1)
				content := fmt.Sprintf("This is the content of post %d", i+1)

				err = cm.CreateContent(filePath, frontMatter, content)
				if err != nil {
					t.Logf("Failed to create content file %d: %v", i+1, err)
					return false
				}

				expectedTitles[title] = true
			}

			// Also create some subdirectories with content
			subDir := filepath.Join(contentDir, "blog")
			err = os.MkdirAll(subDir, 0755)
			if err != nil {
				t.Logf("Failed to create subdirectory: %v", err)
				return false
			}

			// Add one more file in subdirectory
			subTitle := "Subdirectory Post"
			subFrontMatter := interfaces.FrontMatter{
				Title:      subTitle,
				Date:       time.Now(),
				Draft:      false,
				Tags:       []string{"subdirectory", "test"},
				Categories: []string{"blog"},
				Author:     "Sub Author",
			}

			subFilePath := "content/blog/sub-post.md"
			subContent := "This is content in a subdirectory"

			err = cm.CreateContent(subFilePath, subFrontMatter, subContent)
			if err != nil {
				t.Logf("Failed to create subdirectory content: %v", err)
				return false
			}

			expectedTitles[subTitle] = true
			totalExpectedFiles := numFiles + 1

			// Create a non-markdown file that should be ignored
			nonMdFile := filepath.Join(contentDir, "readme.txt")
			err = os.WriteFile(nonMdFile, []byte("This is not a markdown file"), 0644)
			if err != nil {
				t.Logf("Failed to create non-markdown file: %v", err)
				return false
			}

			// List all content
			contentItems, err := cm.ListContent(tempDir)
			if err != nil {
				t.Logf("Failed to list content: %v", err)
				return false
			}

			// Verify we got the expected number of content items (should exclude non-markdown files)
			if len(contentItems) != totalExpectedFiles {
				t.Logf("Expected %d content items, got %d", totalExpectedFiles, len(contentItems))
				return false
			}

			// Verify all expected titles are present
			foundTitles := make(map[string]bool)
			for _, item := range contentItems {
				foundTitles[item.Title] = true

				// Verify each item has required properties
				if item.Path == "" {
					t.Logf("Content item missing path")
					return false
				}

				if item.Title == "" {
					t.Logf("Content item missing title")
					return false
				}

				if item.FrontMatter == nil {
					t.Logf("Content item missing front matter")
					return false
				}

				// Verify the path is relative and points to a markdown file
				if !strings.HasSuffix(item.Path, ".md") {
					t.Logf("Content item path doesn't end with .md: %s", item.Path)
					return false
				}

				if strings.HasPrefix(item.Path, "/") || strings.Contains(item.Path, "..") {
					t.Logf("Content item path is not safe relative path: %s", item.Path)
					return false
				}
			}

			// Check that all expected titles were found
			for expectedTitle := range expectedTitles {
				if !foundTitles[expectedTitle] {
					t.Logf("Expected title not found in content list: %s", expectedTitle)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10), // Number of files to create
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
// **Feature: hugo-visual-client, Property 16: 搜索结果准确性**
// **Validates: Requirements 5.2**
func TestSearchResultAccuracy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("search results should only include content matching the query", prop.ForAll(
		func(searchTermIndex int) bool {
			// Use a predefined set of search terms to avoid generator issues
			searchTerms := []string{"golang", "hugo", "testing", "programming", "development", "tutorial", "guide", "example"}
			if searchTermIndex < 0 || searchTermIndex >= len(searchTerms) {
				return true
			}
			
			searchTerm := searchTerms[searchTermIndex]

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create content directory
			contentDir := filepath.Join(tempDir, "content")
			err = os.MkdirAll(contentDir, 0755)
			if err != nil {
				t.Logf("Failed to create content dir: %v", err)
				return false
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Create test content with known patterns
			testData := []struct {
				title    string
				tags     []string
				categories []string
				content  string
				shouldMatch bool
			}{
				{
					title:    fmt.Sprintf("Post about %s", searchTerm),
					tags:     []string{"general", "test"},
					categories: []string{"blog"},
					content:  "General content without the term",
					shouldMatch: true, // Title contains search term
				},
				{
					title:    "General Post",
					tags:     []string{searchTerm, "test"},
					categories: []string{"blog"},
					content:  "General content without the term",
					shouldMatch: true, // Tag contains search term
				},
				{
					title:    "Another Post",
					tags:     []string{"general", "test"},
					categories: []string{searchTerm, "blog"},
					content:  "General content without the term",
					shouldMatch: true, // Category contains search term
				},
				{
					title:    "Content Post",
					tags:     []string{"general", "test"},
					categories: []string{"blog"},
					content:  fmt.Sprintf("Content mentions %s in the body", searchTerm),
					shouldMatch: true, // Content contains search term
				},
				{
					title:    "Unrelated Post",
					tags:     []string{"unrelated", "different"},
					categories: []string{"other"},
					content:  "No match here at all",
					shouldMatch: false, // Should not match
				},
			}

			expectedMatches := 0
			for i, td := range testData {
				frontMatter := interfaces.FrontMatter{
					Title:      td.title,
					Date:       time.Now().Add(time.Duration(i) * time.Hour),
					Draft:      false,
					Tags:       td.tags,
					Categories: td.categories,
					Author:     "Test Author",
				}

				filePath := fmt.Sprintf("content/post-%d.md", i+1)

				err = cm.CreateContent(filePath, frontMatter, td.content)
				if err != nil {
					t.Logf("Failed to create test content %d: %v", i+1, err)
					return false
				}

				if td.shouldMatch {
					expectedMatches++
				}
			}

			// Perform the search
			searchResults, err := cm.SearchContent(searchTerm)
			if err != nil {
				t.Logf("Failed to search content: %v", err)
				return false
			}

			// Verify the number of results matches expectations
			if len(searchResults) != expectedMatches {
				t.Logf("Expected %d search results, got %d for term '%s'", expectedMatches, len(searchResults), searchTerm)
				return false
			}

			// Verify each result actually contains the search term
			searchTermLower := strings.ToLower(searchTerm)
			for _, result := range searchResults {
				found := false

				// Check title
				if strings.Contains(strings.ToLower(result.Title), searchTermLower) {
					found = true
				}

				// Check tags
				if !found {
					for _, tag := range result.Tags {
						if strings.Contains(strings.ToLower(tag), searchTermLower) {
							found = true
							break
						}
					}
				}

				// Check categories
				if !found {
					for _, category := range result.Categories {
						if strings.Contains(strings.ToLower(category), searchTermLower) {
							found = true
							break
						}
					}
				}

				// Check content
				if !found {
					if strings.Contains(strings.ToLower(result.Content), searchTermLower) {
						found = true
					}
				}

				if !found {
					t.Logf("Search result doesn't contain search term '%s': title=%s, tags=%v, categories=%v, content=%s", 
						searchTerm, result.Title, result.Tags, result.Categories, result.Content)
					return false
				}
			}

			// Test empty search returns all content
			allResults, err := cm.SearchContent("")
			if err != nil {
				t.Logf("Failed to search with empty term: %v", err)
				return false
			}

			if len(allResults) != len(testData) {
				t.Logf("Empty search should return all content: expected %d, got %d", len(testData), len(allResults))
				return false
			}

			return true
		},
		gen.IntRange(0, 7), // Index into predefined search terms
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
// Resource management tests

func TestContentManager_ResourceManagement(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create static directory
	staticDir := filepath.Join(tempDir, "static")
	err = os.MkdirAll(staticDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create static dir: %v", err)
	}

	// Create content manager
	cm := NewContentManager(tempDir)

	// Create a test source file
	sourceFile := filepath.Join(tempDir, "test-image.jpg")
	testData := []byte("fake image data")
	err = os.WriteFile(sourceFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test source file: %v", err)
	}

	// Test CopyResource
	destPath := "images/test-image.jpg"
	err = cm.CopyResource(sourceFile, destPath)
	if err != nil {
		t.Fatalf("Failed to copy resource: %v", err)
	}

	// Verify the file was copied
	copiedFile := filepath.Join(staticDir, destPath)
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Fatalf("Resource file was not copied")
	}

	// Verify the content is correct
	copiedData, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(copiedData) != string(testData) {
		t.Errorf("Copied file content doesn't match original")
	}

	// Test GetResourcePath
	resourcePath := cm.GetResourcePath("images/test-image.jpg")
	expectedPath := "/images/test-image.jpg"
	if resourcePath != expectedPath {
		t.Errorf("Expected resource path %s, got %s", expectedPath, resourcePath)
	}

	// Test ListResources
	resources, err := cm.ListResources()
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}

	if len(resources) > 0 {
		resource := resources[0]
		if resource.Name != "test-image.jpg" {
			t.Errorf("Expected resource name 'test-image.jpg', got %s", resource.Name)
		}
		if !resource.IsImage {
			t.Errorf("Expected resource to be identified as image")
		}
		if resource.Size != int64(len(testData)) {
			t.Errorf("Expected resource size %d, got %d", len(testData), resource.Size)
		}
	}

	// Test DeleteResource
	err = cm.DeleteResource(destPath)
	if err != nil {
		t.Fatalf("Failed to delete resource: %v", err)
	}

	// Verify the file was deleted
	if _, err := os.Stat(copiedFile); !os.IsNotExist(err) {
		t.Fatalf("Resource file was not deleted")
	}

	// Verify list is now empty
	resources, err = cm.ListResources()
	if err != nil {
		t.Fatalf("Failed to list resources after deletion: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("Expected 0 resources after deletion, got %d", len(resources))
	}
}

func TestContentManager_ResourcePathSecurity(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create static directory
	staticDir := filepath.Join(tempDir, "static")
	err = os.MkdirAll(staticDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create static dir: %v", err)
	}

	// Create content manager
	cm := NewContentManager(tempDir)

	// Create a test source file
	sourceFile := filepath.Join(tempDir, "test-file.txt")
	testData := []byte("test data")
	err = os.WriteFile(sourceFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test source file: %v", err)
	}

	// Test path traversal attack prevention
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/etc/passwd",
		"C:\\Windows\\System32\\config\\sam",
	}

	for _, maliciousPath := range maliciousPaths {
		err = cm.CopyResource(sourceFile, maliciousPath)
		if err == nil {
			t.Errorf("Expected error for malicious path %s, but got none", maliciousPath)
		}

		err = cm.DeleteResource(maliciousPath)
		if err == nil {
			t.Errorf("Expected error for malicious delete path %s, but got none", maliciousPath)
		}
	}
}
// **Feature: hugo-visual-client, Property 7: 资源路径处理正确性**
// **Validates: Requirements 2.5**
func TestResourcePathProcessingCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("resource path processing should correctly handle file paths and ensure security", prop.ForAll(
		func(resourceName string) bool {
			// Skip empty resource names
			if strings.TrimSpace(resourceName) == "" {
				return true
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create static directory
			staticDir := filepath.Join(tempDir, "static")
			err = os.MkdirAll(staticDir, 0755)
			if err != nil {
				t.Logf("Failed to create static dir: %v", err)
				return false
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Test GetResourcePath - should always return a path starting with /
			resourcePath := cm.GetResourcePath(resourceName)
			if !strings.HasPrefix(resourcePath, "/") {
				t.Logf("Resource path should start with /: got %s", resourcePath)
				return false
			}

			// The resource path should not contain the original leading slash if it had one
			expectedPath := "/" + strings.TrimPrefix(resourceName, "/")
			if resourcePath != expectedPath {
				t.Logf("Resource path mismatch: expected %s, got %s", expectedPath, resourcePath)
				return false
			}

			// Test safe resource names (no path traversal, no absolute paths)
			isSafeName := !cm.isUnsafePath(resourceName)
			
			if isSafeName {
				// For safe names, we should be able to create a test resource
				testData := []byte("test resource data")
				sourceFile := filepath.Join(tempDir, "temp-resource")
				err = os.WriteFile(sourceFile, testData, 0644)
				if err != nil {
					t.Logf("Failed to create temp source file: %v", err)
					return false
				}

				// Try to copy the resource
				err = cm.CopyResource(sourceFile, resourceName)
				if err != nil {
					t.Logf("Failed to copy safe resource %s: %v", resourceName, err)
					return false
				}

				// Verify the resource exists in the expected location
				expectedFile := filepath.Join(staticDir, resourceName)
				if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
					t.Logf("Resource file was not created at expected location: %s", expectedFile)
					return false
				}

				// Verify we can list the resource
				resources, err := cm.ListResources()
				if err != nil {
					t.Logf("Failed to list resources: %v", err)
					return false
				}

				found := false
				for _, resource := range resources {
					if resource.Path == resourceName {
						found = true
						// Verify resource properties
						if resource.Size != int64(len(testData)) {
							t.Logf("Resource size mismatch: expected %d, got %d", len(testData), resource.Size)
							return false
						}
						break
					}
				}

				if !found {
					t.Logf("Resource not found in list: %s", resourceName)
					return false
				}

				// Verify we can delete the resource
				err = cm.DeleteResource(resourceName)
				if err != nil {
					t.Logf("Failed to delete resource %s: %v", resourceName, err)
					return false
				}

				// Verify the resource is gone
				if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
					t.Logf("Resource file was not deleted: %s", expectedFile)
					return false
				}
			} else {
				// For unsafe names, operations should fail
				testData := []byte("test resource data")
				sourceFile := filepath.Join(tempDir, "temp-resource")
				err = os.WriteFile(sourceFile, testData, 0644)
				if err != nil {
					t.Logf("Failed to create temp source file: %v", err)
					return false
				}

				// Try to copy the resource - should fail
				err = cm.CopyResource(sourceFile, resourceName)
				if err == nil {
					t.Logf("Expected error for unsafe resource name %s, but got none", resourceName)
					return false
				}

				// Try to delete the resource - should fail
				err = cm.DeleteResource(resourceName)
				if err == nil {
					t.Logf("Expected error for unsafe delete path %s, but got none", resourceName)
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { 
			return len(s) > 0 && len(s) <= 50 && !strings.Contains(s, "..") && !strings.HasPrefix(s, "/") && !strings.Contains(s, ":")
		}), // Safe resource names only
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: hugo-visual-client, Property 17: 批量操作原子性**
// **Validates: Requirements 5.3**
func TestBatchOperationAtomicity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("batch operations should be atomic - either all succeed or all fail with rollback", prop.ForAll(
		func(numFiles int, shouldFail bool) bool {
			// Limit the number of files to a reasonable range
			if numFiles < 2 || numFiles > 5 {
				return true
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create content directory
			contentDir := filepath.Join(tempDir, "content")
			err = os.MkdirAll(contentDir, 0755)
			if err != nil {
				t.Logf("Failed to create content dir: %v", err)
				return false
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Create test content files
			var filePaths []string
			for i := 0; i < numFiles; i++ {
				frontMatter := interfaces.FrontMatter{
					Title:      fmt.Sprintf("Test Post %d", i+1),
					Date:       time.Now().Add(time.Duration(i) * time.Hour),
					Draft:      false,
					Tags:       []string{fmt.Sprintf("tag%d", i), "test"},
					Categories: []string{"blog"},
					Author:     fmt.Sprintf("Author %d", i+1),
				}

				filePath := fmt.Sprintf("content/post-%d.md", i+1)
				content := fmt.Sprintf("This is the content of post %d", i+1)

				err = cm.CreateContent(filePath, frontMatter, content)
				if err != nil {
					t.Logf("Failed to create content file %d: %v", i+1, err)
					return false
				}

				filePaths = append(filePaths, filePath)
			}

			// Verify all files exist before batch operation
			for _, filePath := range filePaths {
				fullPath := filepath.Join(tempDir, filePath)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Logf("File should exist before batch operation: %s", filePath)
					return false
				}
			}

			// Test batch delete operation
			var pathsToDelete []string
			if shouldFail {
				// Include a non-existent file to force failure
				pathsToDelete = append(filePaths, "content/non-existent.md")
			} else {
				pathsToDelete = filePaths
			}

			err = cm.BatchDeleteContent(pathsToDelete)

			if shouldFail {
				// Operation should fail
				if err == nil {
					t.Logf("Expected batch delete to fail, but it succeeded")
					return false
				}

				// All original files should still exist (rollback)
				for _, filePath := range filePaths {
					fullPath := filepath.Join(tempDir, filePath)
					if _, err := os.Stat(fullPath); os.IsNotExist(err) {
						t.Logf("File should still exist after failed batch delete: %s", filePath)
						return false
					}
				}
			} else {
				// Operation should succeed
				if err != nil {
					t.Logf("Expected batch delete to succeed, but it failed: %v", err)
					return false
				}

				// All files should be deleted
				for _, filePath := range filePaths {
					fullPath := filepath.Join(tempDir, filePath)
					if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
						t.Logf("File should be deleted after successful batch delete: %s", filePath)
						return false
					}
				}
			}

			// Test batch tag update atomicity
			if !shouldFail {
				// Recreate files for tag update test
				for i := 0; i < numFiles; i++ {
					frontMatter := interfaces.FrontMatter{
						Title:      fmt.Sprintf("Test Post %d", i+1),
						Date:       time.Now().Add(time.Duration(i) * time.Hour),
						Draft:      false,
						Tags:       []string{"original", "test"},
						Categories: []string{"blog"},
						Author:     fmt.Sprintf("Author %d", i+1),
					}

					filePath := fmt.Sprintf("content/post-%d.md", i+1)
					content := fmt.Sprintf("This is the content of post %d", i+1)

					err = cm.CreateContent(filePath, frontMatter, content)
					if err != nil {
						t.Logf("Failed to recreate content file %d: %v", i+1, err)
						return false
					}
				}

				// Test successful batch tag update
				var tagUpdates []interfaces.TagUpdate
				for _, filePath := range filePaths {
					tagUpdates = append(tagUpdates, interfaces.TagUpdate{
						Path: filePath,
						Tags: []string{"updated", "batch", "test"},
					})
				}

				err = cm.BatchUpdateTags(tagUpdates)
				if err != nil {
					t.Logf("Batch tag update should succeed: %v", err)
					return false
				}

				// Verify all files have updated tags
				for _, filePath := range filePaths {
					contentItem, err := cm.GetContent(filePath)
					if err != nil {
						t.Logf("Failed to get content after tag update: %v", err)
						return false
					}

					expectedTags := []string{"updated", "batch", "test"}
					if len(contentItem.Tags) != len(expectedTags) {
						t.Logf("Tag count mismatch for %s: expected %d, got %d", filePath, len(expectedTags), len(contentItem.Tags))
						return false
					}

					tagMap := make(map[string]bool)
					for _, tag := range contentItem.Tags {
						tagMap[tag] = true
					}
					for _, expectedTag := range expectedTags {
						if !tagMap[expectedTag] {
							t.Logf("Missing expected tag %s in file %s", expectedTag, filePath)
							return false
						}
					}
				}

				// Test failed batch tag update (with non-existent file)
				failingTagUpdates := append(tagUpdates, interfaces.TagUpdate{
					Path: "content/non-existent.md",
					Tags: []string{"should", "fail"},
				})

				err = cm.BatchUpdateTags(failingTagUpdates)
				if err == nil {
					t.Logf("Expected batch tag update to fail with non-existent file")
					return false
				}

				// Verify original tags are preserved (rollback)
				for _, filePath := range filePaths {
					contentItem, err := cm.GetContent(filePath)
					if err != nil {
						t.Logf("Failed to get content after failed tag update: %v", err)
						return false
					}

					// Tags should still be the updated ones from the successful operation
					expectedTags := []string{"updated", "batch", "test"}
					if len(contentItem.Tags) != len(expectedTags) {
						t.Logf("Tag count should be preserved after failed update for %s: expected %d, got %d", filePath, len(expectedTags), len(contentItem.Tags))
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(2, 5), // Number of files
		gen.Bool(),         // Should fail
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: hugo-visual-client, Property 18: 文件移动完整性**
// **Validates: Requirements 5.4**
func TestFileMoveIntegrity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("file move should remove source and create destination with identical content", prop.ForAll(
		func(numMoves int) bool {
			// Limit the number of moves to a reasonable range
			if numMoves < 1 || numMoves > 3 {
				return true
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create content directory
			contentDir := filepath.Join(tempDir, "content")
			err = os.MkdirAll(contentDir, 0755)
			if err != nil {
				t.Logf("Failed to create content dir: %v", err)
				return false
			}

			// Create subdirectories for moves
			subDirs := []string{"blog", "posts", "articles"}
			for _, subDir := range subDirs {
				err = os.MkdirAll(filepath.Join(contentDir, subDir), 0755)
				if err != nil {
					t.Logf("Failed to create subdirectory %s: %v", subDir, err)
					return false
				}
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Create test content files and prepare moves
			var moves []interfaces.ContentMove
			var originalContents []string
			
			for i := 0; i < numMoves; i++ {
				frontMatter := interfaces.FrontMatter{
					Title:      fmt.Sprintf("Test Post %d", i+1),
					Date:       time.Now().Add(time.Duration(i) * time.Hour),
					Draft:      i%2 == 0,
					Tags:       []string{fmt.Sprintf("tag%d", i), "test", "move"},
					Categories: []string{"blog", fmt.Sprintf("category%d", i)},
					Author:     fmt.Sprintf("Author %d", i+1),
					Description: fmt.Sprintf("Description for post %d", i+1),
				}

				sourcePath := fmt.Sprintf("content/original-post-%d.md", i+1)
				destPath := fmt.Sprintf("content/%s/moved-post-%d.md", subDirs[i%len(subDirs)], i+1)
				content := fmt.Sprintf("This is the original content of post %d.\n\nIt has multiple paragraphs and should be preserved exactly during the move operation.", i+1)

				// Create the original file
				err = cm.CreateContent(sourcePath, frontMatter, content)
				if err != nil {
					t.Logf("Failed to create source content file %d: %v", i+1, err)
					return false
				}

				// Store original content for verification
				originalContents = append(originalContents, content)

				// Add to moves
				moves = append(moves, interfaces.ContentMove{
					SourcePath: sourcePath,
					DestPath:   destPath,
				})
			}

			// Read original content items for comparison
			var originalItems []*interfaces.ContentItem
			for _, move := range moves {
				item, err := cm.GetContent(move.SourcePath)
				if err != nil {
					t.Logf("Failed to read original content: %v", err)
					return false
				}
				originalItems = append(originalItems, item)
			}

			// Perform batch move
			err = cm.BatchMoveContent(moves)
			if err != nil {
				t.Logf("Batch move failed: %v", err)
				return false
			}

			// Verify all source files are gone
			for _, move := range moves {
				sourcePath := filepath.Join(tempDir, move.SourcePath)
				if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
					t.Logf("Source file should be deleted after move: %s", move.SourcePath)
					return false
				}
			}

			// Verify all destination files exist and have correct content
			for i, move := range moves {
				destPath := filepath.Join(tempDir, move.DestPath)
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Logf("Destination file should exist after move: %s", move.DestPath)
					return false
				}

				// Read the moved content
				movedItem, err := cm.GetContent(move.DestPath)
				if err != nil {
					t.Logf("Failed to read moved content: %v", err)
					return false
				}

				originalItem := originalItems[i]

				// Verify content is identical
				if movedItem.Content != originalItem.Content {
					t.Logf("Content mismatch after move: expected %q, got %q", originalItem.Content, movedItem.Content)
					return false
				}

				// Verify front matter is preserved
				if movedItem.Title != originalItem.Title {
					t.Logf("Title mismatch after move: expected %s, got %s", originalItem.Title, movedItem.Title)
					return false
				}

				if movedItem.Draft != originalItem.Draft {
					t.Logf("Draft status mismatch after move: expected %v, got %v", originalItem.Draft, movedItem.Draft)
					return false
				}

				// Verify tags are preserved (order doesn't matter)
				if len(movedItem.Tags) != len(originalItem.Tags) {
					t.Logf("Tag count mismatch after move: expected %d, got %d", len(originalItem.Tags), len(movedItem.Tags))
					return false
				}

				tagMap := make(map[string]bool)
				for _, tag := range movedItem.Tags {
					tagMap[tag] = true
				}
				for _, originalTag := range originalItem.Tags {
					if !tagMap[originalTag] {
						t.Logf("Missing tag after move: %s", originalTag)
						return false
					}
				}

				// Verify categories are preserved (order doesn't matter)
				if len(movedItem.Categories) != len(originalItem.Categories) {
					t.Logf("Category count mismatch after move: expected %d, got %d", len(originalItem.Categories), len(movedItem.Categories))
					return false
				}

				categoryMap := make(map[string]bool)
				for _, category := range movedItem.Categories {
					categoryMap[category] = true
				}
				for _, originalCategory := range originalItem.Categories {
					if !categoryMap[originalCategory] {
						t.Logf("Missing category after move: %s", originalCategory)
						return false
					}
				}

				// Verify other front matter fields
				if author, ok := movedItem.FrontMatter["author"]; !ok || author != originalItem.FrontMatter["author"] {
					t.Logf("Author mismatch after move: expected %v, got %v", originalItem.FrontMatter["author"], author)
					return false
				}

				if description, ok := movedItem.FrontMatter["description"]; !ok || description != originalItem.FrontMatter["description"] {
					t.Logf("Description mismatch after move: expected %v, got %v", originalItem.FrontMatter["description"], description)
					return false
				}
			}

			// Test that the content list reflects the changes
			allContent, err := cm.ListContent(tempDir)
			if err != nil {
				t.Logf("Failed to list content after move: %v", err)
				return false
			}

			// Verify moved files are in the list with new paths
			movedPaths := make(map[string]bool)
			for _, item := range allContent {
				movedPaths[item.Path] = true
			}

			for _, move := range moves {
				if !movedPaths[move.DestPath] {
					t.Logf("Moved file not found in content list: %s", move.DestPath)
					return false
				}

				if movedPaths[move.SourcePath] {
					t.Logf("Source file should not be in content list after move: %s", move.SourcePath)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 3), // Number of moves
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: hugo-visual-client, Property 19: 文件系统同步一致性**
// **Validates: Requirements 5.5**
func TestFileSystemSyncConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("file system state should be consistent with interface display after operations", prop.ForAll(
		func(operationType int, numFiles int) bool {
			// Limit operation types and file counts
			if operationType < 0 || operationType > 3 || numFiles < 1 || numFiles > 4 {
				return true
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create content directory
			contentDir := filepath.Join(tempDir, "content")
			err = os.MkdirAll(contentDir, 0755)
			if err != nil {
				t.Logf("Failed to create content dir: %v", err)
				return false
			}

			// Create subdirectories
			subDirs := []string{"blog", "posts", "articles", "docs"}
			for _, subDir := range subDirs {
				err = os.MkdirAll(filepath.Join(contentDir, subDir), 0755)
				if err != nil {
					t.Logf("Failed to create subdirectory %s: %v", subDir, err)
					return false
				}
			}

			// Create content manager
			cm := NewContentManager(tempDir)

			// Create initial test files
			var initialPaths []string
			for i := 0; i < numFiles; i++ {
				frontMatter := interfaces.FrontMatter{
					Title:      fmt.Sprintf("Test Post %d", i+1),
					Date:       time.Now().Add(time.Duration(i) * time.Hour),
					Draft:      i%2 == 0,
					Tags:       []string{fmt.Sprintf("tag%d", i), "test"},
					Categories: []string{"blog", fmt.Sprintf("category%d", i)},
					Author:     fmt.Sprintf("Author %d", i+1),
				}

				filePath := fmt.Sprintf("content/%s/post-%d.md", subDirs[i%len(subDirs)], i+1)
				content := fmt.Sprintf("This is the content of post %d", i+1)

				err = cm.CreateContent(filePath, frontMatter, content)
				if err != nil {
					t.Logf("Failed to create initial content file %d: %v", i+1, err)
					return false
				}

				initialPaths = append(initialPaths, filePath)
			}

			// Get initial state from interface
			initialContent, err := cm.ListContent(tempDir)
			if err != nil {
				t.Logf("Failed to list initial content: %v", err)
				return false
			}

			// Verify initial consistency
			if !cm.verifyFileSystemConsistency(tempDir, initialContent) {
				t.Logf("Initial file system state is not consistent")
				return false
			}

			// Perform different operations based on operationType
			switch operationType {
			case 0: // Create operation
				newFrontMatter := interfaces.FrontMatter{
					Title:      "New Post",
					Date:       time.Now(),
					Draft:      false,
					Tags:       []string{"new", "test"},
					Categories: []string{"blog"},
					Author:     "New Author",
				}

				newPath := "content/blog/new-post.md"
				newContent := "This is new content"

				err = cm.CreateContent(newPath, newFrontMatter, newContent)
				if err != nil {
					t.Logf("Failed to create new content: %v", err)
					return false
				}

			case 1: // Update operation
				if len(initialPaths) > 0 {
					updatePath := initialPaths[0]
					
					// Get current content
					currentItem, err := cm.GetContent(updatePath)
					if err != nil {
						t.Logf("Failed to get current content for update: %v", err)
						return false
					}

					// Update with new data
					updatedFrontMatter := interfaces.FrontMatter{
						Title:      "Updated " + currentItem.Title,
						Date:       time.Now(),
						Draft:      !currentItem.Draft,
						Tags:       append(currentItem.Tags, "updated"),
						Categories: append(currentItem.Categories, "updated"),
						Author:     "Updated Author",
					}

					updatedContent := currentItem.Content + "\n\nThis content has been updated."

					err = cm.UpdateContent(updatePath, updatedFrontMatter, updatedContent)
					if err != nil {
						t.Logf("Failed to update content: %v", err)
						return false
					}
				}

			case 2: // Delete operation
				if len(initialPaths) > 0 {
					deletePath := initialPaths[len(initialPaths)-1]
					err = cm.DeleteContent(deletePath)
					if err != nil {
						t.Logf("Failed to delete content: %v", err)
						return false
					}
				}

			case 3: // Batch move operation
				if len(initialPaths) >= 2 {
					var moves []interfaces.ContentMove
					for i, path := range initialPaths[:2] {
						newPath := fmt.Sprintf("content/moved/moved-post-%d.md", i+1)
						moves = append(moves, interfaces.ContentMove{
							SourcePath: path,
							DestPath:   newPath,
						})
					}

					// Create moved directory
					err = os.MkdirAll(filepath.Join(contentDir, "moved"), 0755)
					if err != nil {
						t.Logf("Failed to create moved directory: %v", err)
						return false
					}

					err = cm.BatchMoveContent(moves)
					if err != nil {
						t.Logf("Failed to perform batch move: %v", err)
						return false
					}
				}
			}

			// Get updated state from interface
			updatedContent, err := cm.ListContent(tempDir)
			if err != nil {
				t.Logf("Failed to list updated content: %v", err)
				return false
			}

			// Verify file system consistency after operation
			if !cm.verifyFileSystemConsistency(tempDir, updatedContent) {
				t.Logf("File system state is not consistent after operation %d", operationType)
				return false
			}

			// Additional verification: ensure all files in interface actually exist
			for _, item := range updatedContent {
				fullPath := filepath.Join(tempDir, item.Path)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Logf("File listed in interface does not exist on file system: %s", item.Path)
					return false
				}

				// Verify the file can be read and parsed
				fileContent, err := os.ReadFile(fullPath)
				if err != nil {
					t.Logf("Failed to read file that should exist: %s", item.Path)
					return false
				}

				// Verify it's a valid markdown file with front matter
				_, _, err = cm.ParseFrontMatter(string(fileContent))
				if err != nil {
					t.Logf("File exists but has invalid front matter: %s", item.Path)
					return false
				}
			}

			// Verify no orphaned files exist (files on disk that aren't in the interface)
			err = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip directories and non-markdown files
				if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
					return nil
				}

				// Get relative path
				relPath, err := filepath.Rel(tempDir, path)
				if err != nil {
					return err
				}
				relPath = filepath.ToSlash(relPath)

				// Check if this file is in the interface list
				found := false
				for _, item := range updatedContent {
					if item.Path == relPath {
						found = true
						break
					}
				}

				if !found {
					t.Logf("Orphaned file found on file system but not in interface: %s", relPath)
					return fmt.Errorf("orphaned file: %s", relPath)
				}

				return nil
			})

			if err != nil {
				t.Logf("File system walk failed or found orphaned files: %v", err)
				return false
			}

			return true
		},
		gen.IntRange(0, 3), // Operation type
		gen.IntRange(1, 4), // Number of files
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper method to verify file system consistency
func (cm *ContentManagerService) verifyFileSystemConsistency(projectPath string, contentItems []interfaces.ContentItem) bool {
	// Check that every item in the interface corresponds to an actual file
	for _, item := range contentItems {
		fullPath := filepath.Join(projectPath, item.Path)
		
		// Check if file exists
		info, err := os.Stat(fullPath)
		if err != nil {
			return false
		}

		// Check if it's a regular file
		if !info.Mode().IsRegular() {
			return false
		}

		// Check if it's a markdown file
		if !strings.HasSuffix(strings.ToLower(item.Path), ".md") {
			return false
		}

		// Verify the file content matches the item
		fileContent, err := os.ReadFile(fullPath)
		if err != nil {
			return false
		}

		frontMatter, content, err := cm.ParseFrontMatter(string(fileContent))
		if err != nil {
			return false
		}

		// Verify basic properties match
		if frontMatter.Title != item.Title {
			return false
		}

		if content != item.Content {
			return false
		}

		if frontMatter.Draft != item.Draft {
			return false
		}
	}

	return true
}