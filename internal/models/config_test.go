package models

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gopkg.in/yaml.v3"
)

// **Feature: hugo-visual-client, Property 5: Front Matter序列化往返一致性**
// **Validates: Requirements 2.2, 2.4**
func TestFrontMatterSerializationRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("front matter serialization round trip should preserve data", prop.ForAll(
		func(fm FrontMatter) bool {
			// Serialize to YAML
			yamlData, err := yaml.Marshal(fm)
			if err != nil {
				t.Logf("Failed to marshal front matter to YAML: %v", err)
				return false
			}

			// Deserialize from YAML
			var deserializedFM FrontMatter
			err = yaml.Unmarshal(yamlData, &deserializedFM)
			if err != nil {
				t.Logf("Failed to unmarshal front matter from YAML: %v", err)
				return false
			}

			// Compare essential fields (ignoring time precision differences)
			return compareFrontMatter(fm, deserializedFM)
		},
		genFrontMatter(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: hugo-visual-client, Property 8: 配置序列化往返一致性**
// **Validates: Requirements 3.2, 3.5**
func TestSiteConfigSerializationRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("site config serialization round trip should preserve data", prop.ForAll(
		func(config SiteConfig) bool {
			// Serialize to YAML
			yamlData, err := yaml.Marshal(config)
			if err != nil {
				t.Logf("Failed to marshal site config to YAML: %v", err)
				return false
			}

			// Deserialize from YAML
			var deserializedConfig SiteConfig
			err = yaml.Unmarshal(yamlData, &deserializedConfig)
			if err != nil {
				t.Logf("Failed to unmarshal site config from YAML: %v", err)
				return false
			}

			// Compare configurations
			return compareSiteConfig(config, deserializedConfig)
		},
		genSiteConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// compareSiteConfig compares two SiteConfig structs for equality
func compareSiteConfig(sc1, sc2 SiteConfig) bool {
	// Compare basic fields
	if sc1.BaseURL != sc2.BaseURL {
		return false
	}
	if sc1.Title != sc2.Title {
		return false
	}
	if sc1.Description != sc2.Description {
		return false
	}
	if sc1.LanguageCode != sc2.LanguageCode {
		return false
	}
	if sc1.Theme != sc2.Theme {
		return false
	}

	// Compare params map
	if len(sc1.Params) != len(sc2.Params) {
		return false
	}
	for k, v1 := range sc1.Params {
		v2, exists := sc2.Params[k]
		if !exists || v1 != v2 {
			return false
		}
	}

	// Compare taxonomies map
	if len(sc1.Taxonomies) != len(sc2.Taxonomies) {
		return false
	}
	for k, v1 := range sc1.Taxonomies {
		v2, exists := sc2.Taxonomies[k]
		if !exists || v1 != v2 {
			return false
		}
	}

	// Compare menu map
	if len(sc1.Menu) != len(sc2.Menu) {
		return false
	}
	for k, v1 := range sc1.Menu {
		v2, exists := sc2.Menu[k]
		if !exists || !equalMenuItems(v1, v2) {
			return false
		}
	}

	return true
}

// equalMenuItems compares two slices of MenuItem
func equalMenuItems(items1, items2 []MenuItem) bool {
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

// genSiteConfig generates random SiteConfig for testing
func genSiteConfig() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString(),
		gen.OneConstOf("en", "zh", "fr", "de", "es"),
		gen.AlphaString(),
		genStringMap(),
		genMenuMap(),
		genTaxonomyMap(),
		genStringMap(),
	).Map(func(values []interface{}) SiteConfig {
		return SiteConfig{
			BaseURL:       values[0].(string),
			Title:         values[1].(string),
			Description:   values[2].(string),
			LanguageCode:  values[3].(string),
			Theme:         values[4].(string),
			Params:        values[5].(map[string]interface{}),
			Menu:          values[6].(map[string][]MenuItem),
			Taxonomies:    values[7].(map[string]string),
			OutputFormats: values[8].(map[string]interface{}),
		}
	})
}

// genTaxonomyMap generates a map[string]string for taxonomies
func genTaxonomyMap() gopter.Gen {
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

// genStringMap generates a map[string]interface{} for testing
func genStringMap() gopter.Gen {
	return gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { 
		return len(s) > 0 
	})).Map(func(keys []string) map[string]interface{} {
		result := make(map[string]interface{})
		for i, key := range keys {
			switch i % 3 {
			case 0:
				result[key] = "value_" + key
			case 1:
				result[key] = i + 1
			case 2:
				result[key] = true
			}
		}
		return result
	})
}

// genMenuMap generates a map[string][]MenuItem for testing
func genMenuMap() gopter.Gen {
	return gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { 
		return len(s) > 0 
	})).Map(func(keys []string) map[string][]MenuItem {
		result := make(map[string][]MenuItem)
		for _, key := range keys {
			result[key] = []MenuItem{
				{
					Name:       "Menu " + key,
					URL:        "/" + key,
					Weight:     1,
					Identifier: key,
					Parent:     "",
				},
			}
		}
		return result
	})
}

// compareFrontMatter compares two FrontMatter structs for equality
func compareFrontMatter(fm1, fm2 FrontMatter) bool {
	// Compare basic fields
	if fm1.Title != fm2.Title {
		return false
	}
	if fm1.Draft != fm2.Draft {
		return false
	}
	if fm1.Description != fm2.Description {
		return false
	}
	if fm1.Author != fm2.Author {
		return false
	}
	if fm1.Image != fm2.Image {
		return false
	}

	// Compare time (allow for small differences due to serialization)
	if !fm1.Date.Truncate(time.Second).Equal(fm2.Date.Truncate(time.Second)) {
		return false
	}

	// Compare slices
	if !equalStringSlices(fm1.Tags, fm2.Tags) {
		return false
	}
	if !equalStringSlices(fm1.Categories, fm2.Categories) {
		return false
	}

	// Compare custom fields
	if len(fm1.Custom) != len(fm2.Custom) {
		return false
	}
	for k, v1 := range fm1.Custom {
		v2, exists := fm2.Custom[k]
		if !exists {
			return false
		}
		// Simple comparison for basic types
		if v1 != v2 {
			return false
		}
	}

	return true
}

// equalStringSlices compares two string slices for equality
func equalStringSlices(a, b []string) bool {
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

// genFrontMatter generates random FrontMatter for testing
func genFrontMatter() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		genTime(),
		gen.Bool(),
		genStringSlice(),
		genStringSlice(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		genCustomFields(),
	).Map(func(values []interface{}) FrontMatter {
		return FrontMatter{
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

// genTime generates a random time for testing
func genTime() gopter.Gen {
	return gen.Int64Range(
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		time.Now().Unix(),
	).Map(func(timestamp int64) time.Time {
		return time.Unix(timestamp, 0).UTC()
	})
}

// genStringSlice generates a slice of strings
func genStringSlice() gopter.Gen {
	return gen.SliceOfN(3, gen.AlphaString().SuchThat(func(s string) bool { 
		return len(s) > 0 // Ensure no empty strings
	}))
}

// genCustomFields generates a map of custom fields
func genCustomFields() gopter.Gen {
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