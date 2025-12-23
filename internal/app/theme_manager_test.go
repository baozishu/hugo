package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"hugo-visual-client/internal/models"
	"hugo-visual-client/internal/repository"
)

// **Feature: hugo-visual-client, Property 9: 主题切换状态一致性**
// **Validates: Requirements 3.3**
func TestThemeSwitchingConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("theme switching should update configuration consistently", prop.ForAll(
		func(initialTheme, newTheme string) bool {
			// Skip if themes are the same
			if initialTheme == newTheme {
				return true
			}
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-theme-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create themes directory and theme folders
			themesDir := filepath.Join(tempDir, "themes")
			if err := os.MkdirAll(themesDir, 0755); err != nil {
				t.Logf("Failed to create themes dir: %v", err)
				return false
			}

			// Create theme directories
			initialThemeDir := filepath.Join(themesDir, initialTheme)
			newThemeDir := filepath.Join(themesDir, newTheme)
			
			if err := os.MkdirAll(initialThemeDir, 0755); err != nil {
				t.Logf("Failed to create initial theme dir: %v", err)
				return false
			}
			
			if err := os.MkdirAll(newThemeDir, 0755); err != nil {
				t.Logf("Failed to create new theme dir: %v", err)
				return false
			}

			// Create initial configuration with initial theme
			configPath := "config.yaml"  // Use relative path
			initialConfig := &models.SiteConfig{
				BaseURL:      "https://example.com",
				Title:        "Test Site",
				Description:  "Test Description",
				LanguageCode: "en",
				Theme:        initialTheme,
				Params:       make(map[string]interface{}),
				Menu:         make(map[string][]models.MenuItem),
				Taxonomies:   make(map[string]string),
				OutputFormats: make(map[string]interface{}),
			}

			// Save initial configuration
			configRepo := repository.NewConfigRepository(tempDir)
			if err := configRepo.SaveSiteConfig(configPath, initialConfig); err != nil {
				t.Logf("Failed to save initial config: %v", err)
				return false
			}

			// Create theme manager
			themeManager := NewThemeManager(tempDir, configRepo, configPath)

			// Load themes
			if err := themeManager.LoadThemes(); err != nil {
				t.Logf("Failed to load themes: %v", err)
				return false
			}

			// Verify initial theme is active
			activeTheme := themeManager.GetActiveTheme()
			if activeTheme == nil || activeTheme.Name != initialTheme {
				t.Logf("Initial theme not correctly identified as active: expected %s, got %v", initialTheme, activeTheme)
				return false
			}

			// Switch to new theme
			if err := themeManager.SetActiveTheme(newTheme); err != nil {
				t.Logf("Failed to set active theme: %v", err)
				return false
			}

			// Reload configuration to verify persistence
			reloadedConfig, err := configRepo.LoadSiteConfig(configPath)
			if err != nil {
				t.Logf("Failed to reload config: %v", err)
				return false
			}

			// Verify theme was updated in configuration
			if reloadedConfig.Theme != newTheme {
				t.Logf("Theme not updated in config: expected %s, got %s", newTheme, reloadedConfig.Theme)
				return false
			}

			// Reload themes to verify active status
			if err := themeManager.LoadThemes(); err != nil {
				t.Logf("Failed to reload themes: %v", err)
				return false
			}

			// Verify new theme is now active
			newActiveTheme := themeManager.GetActiveTheme()
			if newActiveTheme == nil || newActiveTheme.Name != newTheme {
				t.Logf("New theme not correctly identified as active: expected %s, got %v", newTheme, newActiveTheme)
				return false
			}

			// Verify only one theme is active
			activeCount := 0
			for _, theme := range themeManager.GetThemes() {
				if theme.IsActive {
					activeCount++
				}
			}

			if activeCount != 1 {
				t.Logf("Expected exactly 1 active theme, got %d", activeCount)
				return false
			}

			return true
		},
		genValidThemeName(),
		genValidThemeName(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test theme loading and scanning functionality
func TestThemeScanning(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("theme scanning should correctly identify all themes", prop.ForAll(
		func(themeNames []string) bool {
			if len(themeNames) == 0 {
				return true // Skip empty theme lists
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-theme-scan-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create themes directory
			themesDir := filepath.Join(tempDir, "themes")
			if err := os.MkdirAll(themesDir, 0755); err != nil {
				t.Logf("Failed to create themes dir: %v", err)
				return false
			}

			// Create theme directories with theme.toml files
			for _, themeName := range themeNames {
				themeDir := filepath.Join(themesDir, themeName)
				if err := os.MkdirAll(themeDir, 0755); err != nil {
					t.Logf("Failed to create theme dir %s: %v", themeName, err)
					return false
				}

				// Create theme.toml with metadata
				themeConfig := filepath.Join(themeDir, "theme.toml")
				configContent := `name = "` + themeName + `"
description = "Test theme for ` + themeName + `"
version = "1.0.0"
author = "Test Author"
license = "MIT"
`
				if err := os.WriteFile(themeConfig, []byte(configContent), 0644); err != nil {
					t.Logf("Failed to create theme config for %s: %v", themeName, err)
					return false
				}
			}

			// Create configuration
			configPath := "config.yaml"  // Use relative path
			config := &models.SiteConfig{
				BaseURL:      "https://example.com",
				Title:        "Test Site",
				Description:  "Test Description",
				LanguageCode: "en",
				Theme:        themeNames[0], // Set first theme as active
				Params:       make(map[string]interface{}),
				Menu:         make(map[string][]models.MenuItem),
				Taxonomies:   make(map[string]string),
				OutputFormats: make(map[string]interface{}),
			}

			configRepo := repository.NewConfigRepository(tempDir)
			if err := configRepo.SaveSiteConfig(configPath, config); err != nil {
				t.Logf("Failed to save config: %v", err)
				return false
			}

			// Create theme manager and load themes
			themeManager := NewThemeManager(tempDir, configRepo, configPath)
			if err := themeManager.LoadThemes(); err != nil {
				t.Logf("Failed to load themes: %v", err)
				return false
			}

			// Verify all themes were found
			foundThemes := themeManager.GetThemes()
			if len(foundThemes) != len(themeNames) {
				t.Logf("Expected %d themes, found %d", len(themeNames), len(foundThemes))
				return false
			}

			// Verify theme names match
			foundNames := make(map[string]bool)
			for _, theme := range foundThemes {
				foundNames[theme.Name] = true
			}

			for _, expectedName := range themeNames {
				if !foundNames[expectedName] {
					t.Logf("Expected theme %s not found", expectedName)
					return false
				}
			}

			// Verify exactly one theme is active (the first one)
			activeCount := 0
			var activeTheme string
			for _, theme := range foundThemes {
				if theme.IsActive {
					activeCount++
					activeTheme = theme.Name
				}
			}

			if activeCount != 1 {
				t.Logf("Expected exactly 1 active theme, got %d", activeCount)
				return false
			}

			if activeTheme != themeNames[0] {
				t.Logf("Expected active theme %s, got %s", themeNames[0], activeTheme)
				return false
			}

			return true
		},
		genUniqueThemeNames(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test theme metadata extraction
func TestThemeMetadataExtraction(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("theme metadata should be correctly extracted", prop.ForAll(
		func(themeName, description, version, author, license string) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-theme-metadata-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create themes directory and theme folder
			themesDir := filepath.Join(tempDir, "themes")
			themeDir := filepath.Join(themesDir, themeName)
			if err := os.MkdirAll(themeDir, 0755); err != nil {
				t.Logf("Failed to create theme dir: %v", err)
				return false
			}

			// Create theme.toml with metadata
			themeConfig := filepath.Join(themeDir, "theme.toml")
			configContent := `name = "` + themeName + `"
description = "` + description + `"
version = "` + version + `"
author = "` + author + `"
license = "` + license + `"
`
			if err := os.WriteFile(themeConfig, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to create theme config: %v", err)
				return false
			}

			// Create configuration
			configPath := "config.yaml"  // Use relative path
			config := &models.SiteConfig{
				BaseURL:      "https://example.com",
				Title:        "Test Site",
				Description:  "Test Description",
				LanguageCode: "en",
				Theme:        themeName,
				Params:       make(map[string]interface{}),
				Menu:         make(map[string][]models.MenuItem),
				Taxonomies:   make(map[string]string),
				OutputFormats: make(map[string]interface{}),
			}

			configRepo := repository.NewConfigRepository(tempDir)
			if err := configRepo.SaveSiteConfig(configPath, config); err != nil {
				t.Logf("Failed to save config: %v", err)
				return false
			}

			// Create theme manager and load themes
			themeManager := NewThemeManager(tempDir, configRepo, configPath)
			if err := themeManager.LoadThemes(); err != nil {
				t.Logf("Failed to load themes: %v", err)
				return false
			}

			// Find the theme
			themes := themeManager.GetThemes()
			if len(themes) != 1 {
				t.Logf("Expected 1 theme, found %d", len(themes))
				return false
			}

			theme := themes[0]

			// Verify metadata was extracted correctly
			if theme.Name != themeName {
				t.Logf("Theme name mismatch: expected %s, got %s", themeName, theme.Name)
				return false
			}

			if theme.Description != description {
				t.Logf("Theme description mismatch: expected %s, got %s", description, theme.Description)
				return false
			}

			if theme.Version != version {
				t.Logf("Theme version mismatch: expected %s, got %s", version, theme.Version)
				return false
			}

			if theme.Author != author {
				t.Logf("Theme author mismatch: expected %s, got %s", author, theme.Author)
				return false
			}

			if theme.License != license {
				t.Logf("Theme license mismatch: expected %s, got %s", license, theme.License)
				return false
			}

			return true
		},
		genValidThemeName(),
		genNonEmptyString(),
		genVersionString(),
		genNonEmptyString(),
		genLicenseString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Generators for property-based testing

// genValidThemeName generates valid theme names
func genValidThemeName() gopter.Gen {
	themes := []string{"theme1", "theme2", "theme3", "bootstrap", "minimal", "academic", "hugo-theme", "clean", "simple"}
	return gen.OneConstOf(themes[0], themes[1], themes[2], themes[3], themes[4], themes[5], themes[6], themes[7], themes[8])
}

// genUniqueThemeNames generates a slice of unique theme names
func genUniqueThemeNames() gopter.Gen {
	return gen.SliceOfN(5, genValidThemeName()).SuchThat(func(names []string) bool {
		if len(names) == 0 {
			return false
		}
		
		// Check for uniqueness
		seen := make(map[string]bool)
		for _, name := range names {
			if seen[name] {
				return false
			}
			seen[name] = true
		}
		return true
	})
}

// genNonEmptyString generates non-empty strings
func genNonEmptyString() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 100
	})
}

// genVersionString generates version strings
func genVersionString() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("1.0.0"),
		gen.Const("2.1.3"),
		gen.Const("0.9.1"),
		gen.Const("1.2.0-beta"),
		gen.Const("Unknown"),
	)
}

// genLicenseString generates license strings
func genLicenseString() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("MIT"),
		gen.Const("Apache-2.0"),
		gen.Const("GPL-3.0"),
		gen.Const("BSD-3-Clause"),
		gen.Const("Unknown"),
	)
}