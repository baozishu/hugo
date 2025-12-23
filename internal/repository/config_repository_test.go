package repository

import (
	"os"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"hugo-visual-client/internal/models"
)

// **Feature: hugo-visual-client, Property 3: 数据持久化一致性**
// **Validates: Requirements 1.5**
func TestDataPersistenceConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("data persistence should maintain consistency after save and reload", prop.ForAll(
		func(config models.SiteConfig) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-persistence-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create config repository
			configRepo := NewConfigRepository(tempDir)
			configPath := "config.yaml"

			// Save configuration
			err = configRepo.SaveSiteConfig(configPath, &config)
			if err != nil {
				t.Logf("Failed to save config: %v", err)
				return false
			}

			// Load configuration back
			loadedConfig, err := configRepo.LoadSiteConfig(configPath)
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Compare configurations for consistency
			return compareSiteConfigForPersistence(config, *loadedConfig)
		},
		genValidSiteConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// compareSiteConfigForPersistence compares two SiteConfig structs for persistence consistency
func compareSiteConfigForPersistence(original, loaded models.SiteConfig) bool {
	// Compare basic fields
	if original.BaseURL != loaded.BaseURL {
		return false
	}
	if original.Title != loaded.Title {
		return false
	}
	if original.Description != loaded.Description {
		return false
	}
	if original.LanguageCode != loaded.LanguageCode {
		return false
	}
	if original.Theme != loaded.Theme {
		return false
	}

	// Compare params map
	if len(original.Params) != len(loaded.Params) {
		return false
	}
	for k, v1 := range original.Params {
		v2, exists := loaded.Params[k]
		if !exists {
			return false
		}
		// Handle type conversions that might occur during YAML serialization
		if !compareInterfaceValues(v1, v2) {
			return false
		}
	}

	// Compare taxonomies map
	if len(original.Taxonomies) != len(loaded.Taxonomies) {
		return false
	}
	for k, v1 := range original.Taxonomies {
		v2, exists := loaded.Taxonomies[k]
		if !exists || v1 != v2 {
			return false
		}
	}

	// Compare menu map
	if len(original.Menu) != len(loaded.Menu) {
		return false
	}
	for k, v1 := range original.Menu {
		v2, exists := loaded.Menu[k]
		if !exists || !equalMenuItemsForPersistence(v1, v2) {
			return false
		}
	}

	return true
}

// compareInterfaceValues compares interface{} values accounting for YAML type conversions
func compareInterfaceValues(v1, v2 interface{}) bool {
	// Handle common type conversions that occur during YAML serialization
	switch val1 := v1.(type) {
	case int:
		if val2, ok := v2.(int); ok {
			return val1 == val2
		}
		// YAML might convert int to int64 or float64
		if val2, ok := v2.(int64); ok {
			return int64(val1) == val2
		}
		if val2, ok := v2.(float64); ok {
			return float64(val1) == val2
		}
	case string:
		if val2, ok := v2.(string); ok {
			return val1 == val2
		}
	case bool:
		if val2, ok := v2.(bool); ok {
			return val1 == val2
		}
	case float64:
		if val2, ok := v2.(float64); ok {
			return val1 == val2
		}
		if val2, ok := v2.(int); ok {
			return val1 == float64(val2)
		}
	}
	
	// Fallback to direct comparison
	return v1 == v2
}

// equalMenuItemsForPersistence compares two slices of MenuItem for persistence
func equalMenuItemsForPersistence(items1, items2 []models.MenuItem) bool {
	if len(items1) != len(items2) {
		return false
	}
	for i, item1 := range items1 {
		item2 := items2[i]
		if item1.Name != item2.Name || item1.URL != item2.URL ||
			item1.Weight != item2.Weight || item1.Identifier != item2.Identifier ||
			item1.Parent != item2.Parent {
			return false
		}
	}
	return true
}

// genValidSiteConfig generates valid SiteConfig for testing
func genValidSiteConfig() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString(),
		gen.OneConstOf("en", "zh", "fr", "de", "es"),
		gen.AlphaString(),
		genValidStringMap(),
		genValidMenuMap(),
		genValidTaxonomyMap(),
		genValidStringMap(),
	).Map(func(values []interface{}) models.SiteConfig {
		return models.SiteConfig{
			BaseURL:       values[0].(string),
			Title:         values[1].(string),
			Description:   values[2].(string),
			LanguageCode:  values[3].(string),
			Theme:         values[4].(string),
			Params:        values[5].(map[string]interface{}),
			Menu:          values[6].(map[string][]models.MenuItem),
			Taxonomies:    values[7].(map[string]string),
			OutputFormats: values[8].(map[string]interface{}),
		}
	})
}

// genValidStringMap generates a valid map[string]interface{} for testing
func genValidStringMap() gopter.Gen {
	return gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { 
		return len(s) > 0 
	})).Map(func(keys []string) map[string]interface{} {
		result := make(map[string]interface{})
		for i, key := range keys {
			switch i % 4 {
			case 0:
				result[key] = "value_" + key
			case 1:
				result[key] = i + 1
			case 2:
				result[key] = true
			case 3:
				result[key] = float64(i) + 0.5
			}
		}
		return result
	})
}

// genValidMenuMap generates a valid map[string][]MenuItem for testing
func genValidMenuMap() gopter.Gen {
	return gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { 
		return len(s) > 0 
	})).Map(func(keys []string) map[string][]models.MenuItem {
		result := make(map[string][]models.MenuItem)
		for i, key := range keys {
			result[key] = []models.MenuItem{
				{
					Name:       "Menu " + key,
					URL:        "/" + key,
					Weight:     i + 1,
					Identifier: key,
					Parent:     "",
				},
			}
		}
		return result
	})
}

// genValidTaxonomyMap generates a valid map[string]string for taxonomies
func genValidTaxonomyMap() gopter.Gen {
	return gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { 
		return len(s) > 0 
	})).Map(func(keys []string) map[string]string {
		result := make(map[string]string)
		for _, key := range keys {
			result[key] = key + "s" // e.g., "tag" -> "tags"
		}
		return result
	})
}

// Test for front matter persistence consistency
func TestFrontMatterPersistenceConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("front matter persistence should maintain consistency", prop.ForAll(
		func(frontMatter models.FrontMatter, content string) bool {
			// Create temporary directory for testing
			tempDir, err := os.MkdirTemp("", "hugo-frontmatter-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// Create config repository
			configRepo := NewConfigRepository(tempDir)
			filePath := "test-content.md"

			// Save content with front matter
			err = configRepo.SaveContentWithFrontMatter(filePath, &frontMatter, content)
			if err != nil {
				t.Logf("Failed to save content: %v", err)
				return false
			}

			// Load content and front matter back
			loadedFrontMatter, loadedContent, err := configRepo.LoadFrontMatter(filePath)
			if err != nil {
				t.Logf("Failed to load content: %v", err)
				return false
			}

			// Compare for consistency
			return compareFrontMatterForPersistence(frontMatter, *loadedFrontMatter) &&
				content == loadedContent
		},
		genValidFrontMatter(),
		gen.AlphaString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// compareFrontMatterForPersistence compares two FrontMatter structs for persistence
func compareFrontMatterForPersistence(original, loaded models.FrontMatter) bool {
	// Compare basic fields
	if original.Title != loaded.Title {
		return false
	}
	if original.Draft != loaded.Draft {
		return false
	}
	if original.Description != loaded.Description {
		return false
	}
	if original.Author != loaded.Author {
		return false
	}
	if original.Image != loaded.Image {
		return false
	}

	// Compare time (allow for small differences due to serialization)
	if !original.Date.Truncate(time.Second).Equal(loaded.Date.Truncate(time.Second)) {
		return false
	}

	// Compare slices
	if !equalStringSlicesForPersistence(original.Tags, loaded.Tags) {
		return false
	}
	if !equalStringSlicesForPersistence(original.Categories, loaded.Categories) {
		return false
	}

	// Compare custom fields
	if len(original.Custom) != len(loaded.Custom) {
		return false
	}
	for k, v1 := range original.Custom {
		v2, exists := loaded.Custom[k]
		if !exists || !compareInterfaceValues(v1, v2) {
			return false
		}
	}

	return true
}

// equalStringSlicesForPersistence compares two string slices for persistence
func equalStringSlicesForPersistence(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// genValidFrontMatter generates valid FrontMatter for testing
func genValidFrontMatter() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		genValidTime(),
		gen.Bool(),
		genValidStringSlice(),
		genValidStringSlice(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		genValidCustomFields(),
	).Map(func(values []interface{}) models.FrontMatter {
		return models.FrontMatter{
			Title:       values[0].(string),
			Date:        values[1].(time.Time),
			Draft:       values[2].(bool),
			Tags:        values[3].([]string),
			Categories:  values[4].([]string),
			Description: values[5].(string),
			Author:      values[6].(string),
			Image:       values[7].(string),
			Custom:      values[8].(map[string]interface{}),
		}
	})
}

// genValidTime generates a valid time for testing
func genValidTime() gopter.Gen {
	return gen.Int64Range(
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		time.Now().Unix(),
	).Map(func(timestamp int64) time.Time {
		return time.Unix(timestamp, 0).UTC()
	})
}

// genValidStringSlice generates a valid slice of strings
func genValidStringSlice() gopter.Gen {
	return gen.SliceOfN(2, gen.AlphaString())
}

// genValidCustomFields generates valid custom fields
func genValidCustomFields() gopter.Gen {
	return gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { 
		return len(s) > 0 
	})).Map(func(keys []string) map[string]interface{} {
		result := make(map[string]interface{})
		for i, key := range keys {
			switch i % 3 {
			case 0:
				result[key] = "test_value"
			case 1:
				result[key] = 42
			case 2:
				result[key] = true
			}
		}
		return result
	})
}