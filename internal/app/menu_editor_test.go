package app

import (
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"hugo-visual-client/internal/models"
	"hugo-visual-client/internal/repository"
)

// **Feature: hugo-visual-client, Property 10: 菜单配置保存一致性**
// **Validates: Requirements 3.4**
func TestMenuConfigurationSaveConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("menu configuration should be saved and loaded consistently", prop.ForAll(
		func(menuConfig map[string][]models.MenuItem) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-menu-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create initial configuration with menu
			configPath := "config.yaml"
			initialConfig := &models.SiteConfig{
				BaseURL:       "https://example.com",
				Title:         "Test Site",
				Description:   "Test Description",
				LanguageCode:  "en",
				Theme:         "test-theme",
				Params:        make(map[string]interface{}),
				Menu:          menuConfig,
				Taxonomies:    make(map[string]string),
				OutputFormats: make(map[string]interface{}),
			}

			// Save initial configuration
			configRepo := repository.NewConfigRepository(tempDir)
			if err := configRepo.SaveSiteConfig(configPath, initialConfig); err != nil {
				t.Logf("Failed to save initial config: %v", err)
				return false
			}

			// Create menu editor
			menuEditor := NewMenuEditor(configRepo, configPath)

			// Load configuration
			if err := menuEditor.LoadConfig(); err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Save configuration through menu editor
			if err := menuEditor.SaveConfig(); err != nil {
				t.Logf("Failed to save config through menu editor: %v", err)
				return false
			}

			// Reload configuration to verify persistence
			reloadedConfig, err := configRepo.LoadSiteConfig(configPath)
			if err != nil {
				t.Logf("Failed to reload config: %v", err)
				return false
			}

			// Compare menu configurations for consistency
			return compareMenuConfig(menuConfig, reloadedConfig.Menu)
		},
		genValidMenuConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test menu item manipulation operations
func TestMenuItemOperations(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("menu item operations should maintain consistency", prop.ForAll(
		func(menuName string, items []models.MenuItem) bool {
			if len(items) == 0 {
				return true // Skip empty menu items
			}

			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-menu-ops-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create initial configuration
			configPath := "config.yaml"
			menuConfig := map[string][]models.MenuItem{
				menuName: items,
			}

			initialConfig := &models.SiteConfig{
				BaseURL:       "https://example.com",
				Title:         "Test Site",
				Description:   "Test Description",
				LanguageCode:  "en",
				Theme:         "test-theme",
				Params:        make(map[string]interface{}),
				Menu:          menuConfig,
				Taxonomies:    make(map[string]string),
				OutputFormats: make(map[string]interface{}),
			}

			// Save initial configuration
			configRepo := repository.NewConfigRepository(tempDir)
			if err := configRepo.SaveSiteConfig(configPath, initialConfig); err != nil {
				t.Logf("Failed to save initial config: %v", err)
				return false
			}

			// Create menu editor and load config
			menuEditor := NewMenuEditor(configRepo, configPath)
			if err := menuEditor.LoadConfig(); err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Verify menu was loaded correctly
			config := menuEditor.GetConfig()
			if config.Menu == nil {
				t.Logf("Menu config is nil after loading")
				return false
			}

			loadedItems, exists := config.Menu[menuName]
			if !exists {
				t.Logf("Menu %s not found after loading", menuName)
				return false
			}

			if len(loadedItems) != len(items) {
				t.Logf("Menu item count mismatch: expected %d, got %d", len(items), len(loadedItems))
				return false
			}

			// Verify each menu item
			for i, originalItem := range items {
				if !menuItemsEqual(originalItem, loadedItems[i]) {
					t.Logf("Menu item %d mismatch", i)
					return false
				}
			}

			// Save and reload to test persistence
			if err := menuEditor.SaveConfig(); err != nil {
				t.Logf("Failed to save config: %v", err)
				return false
			}

			reloadedConfig, err := configRepo.LoadSiteConfig(configPath)
			if err != nil {
				t.Logf("Failed to reload config: %v", err)
				return false
			}

			// Verify persistence
			reloadedItems, exists := reloadedConfig.Menu[menuName]
			if !exists {
				t.Logf("Menu %s not found after reload", menuName)
				return false
			}

			if len(reloadedItems) != len(items) {
				t.Logf("Menu item count mismatch after reload: expected %d, got %d", len(items), len(reloadedItems))
				return false
			}

			for i, originalItem := range items {
				if !menuItemsEqual(originalItem, reloadedItems[i]) {
					t.Logf("Menu item %d mismatch after reload", i)
					return false
				}
			}

			return true
		},
		genValidMenuName(),
		genValidMenuItems(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test menu validation
func TestMenuValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("valid menu configurations should pass validation", prop.ForAll(
		func(menuConfig map[string][]models.MenuItem) bool {
			// Create a site config with the menu
			config := &models.SiteConfig{
				BaseURL:       "https://example.com",
				Title:         "Test Site",
				Description:   "Test Description",
				LanguageCode:  "en",
				Theme:         "test-theme",
				Params:        make(map[string]interface{}),
				Menu:          menuConfig,
				Taxonomies:    make(map[string]string),
				OutputFormats: make(map[string]interface{}),
			}

			// Validation should pass for valid configurations
			err := config.Validate()
			if err != nil {
				t.Logf("Valid configuration failed validation: %v", err)
				return false
			}

			// Validate individual menu items
			for menuName, items := range menuConfig {
				if menuName == "" {
					t.Logf("Empty menu name should not be valid")
					return false
				}

				for _, item := range items {
					if err := item.Validate(); err != nil {
						t.Logf("Valid menu item failed validation: %v", err)
						return false
					}
				}
			}

			return true
		},
		genValidMenuConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper functions for comparison

// compareMenuConfig compares two menu configurations for equality
func compareMenuConfig(config1, config2 map[string][]models.MenuItem) bool {
	if len(config1) != len(config2) {
		return false
	}

	for menuName, items1 := range config1 {
		items2, exists := config2[menuName]
		if !exists {
			return false
		}

		if len(items1) != len(items2) {
			return false
		}

		for i, item1 := range items1 {
			if !menuItemsEqual(item1, items2[i]) {
				return false
			}
		}
	}

	return true
}

// menuItemsEqual compares two menu items for equality
func menuItemsEqual(item1, item2 models.MenuItem) bool {
	return item1.Name == item2.Name &&
		item1.URL == item2.URL &&
		item1.Weight == item2.Weight &&
		item1.Identifier == item2.Identifier &&
		item1.Parent == item2.Parent
}

// Generators for property-based testing

// genValidMenuConfig generates valid menu configurations
func genValidMenuConfig() gopter.Gen {
	return gen.MapOf(
		genValidMenuName(),
		genValidMenuItems(),
	).SuchThat(func(m map[string][]models.MenuItem) bool {
		return len(m) <= 5 // Limit to reasonable number of menus
	})
}

// genValidMenuName generates valid menu names
func genValidMenuName() gopter.Gen {
	menuNames := []string{"main", "footer", "sidebar", "header", "navigation"}
	return gen.OneConstOf(menuNames[0], menuNames[1], menuNames[2], menuNames[3], menuNames[4])
}

// genValidMenuItems generates valid menu items
func genValidMenuItems() gopter.Gen {
	return gen.SliceOfN(3, genValidMenuItem()).SuchThat(func(items []models.MenuItem) bool {
		// Ensure unique identifiers
		identifiers := make(map[string]bool)
		for _, item := range items {
			if item.Identifier != "" {
				if identifiers[item.Identifier] {
					return false // Duplicate identifier
				}
				identifiers[item.Identifier] = true
			}
		}
		return true
	})
}

// genValidMenuItem generates valid menu items
func genValidMenuItem() gopter.Gen {
	return gopter.CombineGens(
		genMenuItemName(),
		genMenuItemURL(),
		gen.IntRange(0, 100),
		genMenuItemIdentifier(),
		genMenuItemParent(),
	).Map(func(values []interface{}) models.MenuItem {
		return models.MenuItem{
			Name:       values[0].(string),
			URL:        values[1].(string),
			Weight:     values[2].(int),
			Identifier: values[3].(string),
			Parent:     values[4].(string),
		}
	})
}

// genMenuItemName generates menu item names
func genMenuItemName() gopter.Gen {
	names := []string{"Home", "About", "Blog", "Contact", "Services", "Products", "Portfolio"}
	return gen.OneConstOf(names[0], names[1], names[2], names[3], names[4], names[5], names[6])
}

// genMenuItemURL generates menu item URLs
func genMenuItemURL() gopter.Gen {
	urls := []string{"/", "/about", "/blog", "/contact", "/services", "/products", "/portfolio"}
	return gen.OneConstOf(urls[0], urls[1], urls[2], urls[3], urls[4], urls[5], urls[6])
}

// genMenuItemIdentifier generates menu item identifiers
func genMenuItemIdentifier() gopter.Gen {
	identifiers := []string{"", "home", "about", "blog", "contact", "services", "products"}
	return gen.OneConstOf(identifiers[0], identifiers[1], identifiers[2], identifiers[3], identifiers[4], identifiers[5], identifiers[6])
}

// genMenuItemParent generates menu item parent identifiers
func genMenuItemParent() gopter.Gen {
	parents := []string{"", "main", "services", "products"}
	return gen.OneConstOf(parents[0], parents[1], parents[2], parents[3])
}