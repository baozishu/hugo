package models

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gopkg.in/yaml.v3"
)

// **Feature: hugo-visual-client, Property 22: 部署配置持久化**
// **Validates: Requirements 6.4**
func TestDeploymentConfigPersistence(t *testing.T) {
	properties := gopter.NewProperties(nil)
	properties.Property("deployment config serialization round trip should preserve data", prop.ForAll(
		func(config DeploymentConfig) bool {
			// Debug: Print the config being tested
			t.Logf("Testing config: %+v", config)
			
			// Test JSON serialization round trip
			jsonData, err := json.Marshal(config)
			if err != nil {
				t.Logf("Failed to marshal deployment config to JSON: %v", err)
				return false
			}

			var deserializedJSONConfig DeploymentConfig
			err = json.Unmarshal(jsonData, &deserializedJSONConfig)
			if err != nil {
				t.Logf("Failed to unmarshal deployment config from JSON: %v", err)
				return false
			}

			if !compareDeploymentConfigDebug(config, deserializedJSONConfig, t) {
				t.Logf("JSON round trip failed: configs are not equal")
				t.Logf("Original: %+v", config)
				t.Logf("Deserialized: %+v", deserializedJSONConfig)
				return false
			}

			// Test YAML serialization round trip
			yamlData, err := yaml.Marshal(config)
			if err != nil {
				t.Logf("Failed to marshal deployment config to YAML: %v", err)
				return false
			}

			var deserializedYAMLConfig DeploymentConfig
			err = yaml.Unmarshal(yamlData, &deserializedYAMLConfig)
			if err != nil {
				t.Logf("Failed to unmarshal deployment config from YAML: %v", err)
				return false
			}

			if !compareDeploymentConfig(config, deserializedYAMLConfig) {
				t.Logf("YAML round trip failed: configs are not equal")
				t.Logf("Original: %+v", config)
				t.Logf("Deserialized: %+v", deserializedYAMLConfig)
				return false
			}

			return true
		},
		genDeploymentConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test individual deployment target serialization
func TestDeploymentTargetSerialization(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("deployment target serialization round trip should preserve data", prop.ForAll(
		func(target DeploymentTarget) bool {
			// Test JSON serialization round trip
			jsonData, err := json.Marshal(target)
			if err != nil {
				t.Logf("Failed to marshal deployment target to JSON: %v", err)
				return false
			}

			var deserializedJSONTarget DeploymentTarget
			err = json.Unmarshal(jsonData, &deserializedJSONTarget)
			if err != nil {
				t.Logf("Failed to unmarshal deployment target from JSON: %v", err)
				return false
			}

			if !compareDeploymentTarget(target, deserializedJSONTarget) {
				t.Logf("JSON round trip failed: targets are not equal")
				return false
			}

			// Test YAML serialization round trip
			yamlData, err := yaml.Marshal(target)
			if err != nil {
				t.Logf("Failed to marshal deployment target to YAML: %v", err)
				return false
			}

			var deserializedYAMLTarget DeploymentTarget
			err = yaml.Unmarshal(yamlData, &deserializedYAMLTarget)
			if err != nil {
				t.Logf("Failed to unmarshal deployment target from YAML: %v", err)
				return false
			}

			if !compareDeploymentTarget(target, deserializedYAMLTarget) {
				t.Logf("YAML round trip failed: targets are not equal")
				return false
			}

			return true
		},
		genDeploymentTarget(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test deployment status serialization
func TestDeploymentStatusSerialization(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("deployment status serialization round trip should preserve data", prop.ForAll(
		func(status DeploymentStatus) bool {
			// Test JSON serialization round trip
			jsonData, err := json.Marshal(status)
			if err != nil {
				t.Logf("Failed to marshal deployment status to JSON: %v", err)
				return false
			}

			var deserializedStatus DeploymentStatus
			err = json.Unmarshal(jsonData, &deserializedStatus)
			if err != nil {
				t.Logf("Failed to unmarshal deployment status from JSON: %v", err)
				return false
			}

			if !compareDeploymentStatus(status, deserializedStatus) {
				t.Logf("JSON round trip failed: statuses are not equal")
				return false
			}

			return true
		},
		genDeploymentStatus(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// compareDeploymentConfig compares two DeploymentConfig structs for equality
func compareDeploymentConfig(dc1, dc2 DeploymentConfig) bool {
	// Compare basic fields
	if dc1.DefaultTarget != dc2.DefaultTarget {
		return false
	}
	if dc1.BuildCommand != dc2.BuildCommand {
		return false
	}
	if dc1.OutputDir != dc2.OutputDir {
		return false
	}

	// Compare string slices (handle nil vs empty slice)
	if !equalStringSlicesNilSafe(dc1.ExcludeFiles, dc2.ExcludeFiles) {
		return false
	}
	if !equalStringSlicesNilSafe(dc1.IncludeFiles, dc2.IncludeFiles) {
		return false
	}

	// Compare targets
	if len(dc1.Targets) != len(dc2.Targets) {
		return false
	}
	for i, target1 := range dc1.Targets {
		if !compareDeploymentTarget(target1, dc2.Targets[i]) {
			return false
		}
	}

	return true
}

// compareDeploymentConfigDebug compares two DeploymentConfig structs with detailed logging
func compareDeploymentConfigDebug(dc1, dc2 DeploymentConfig, t *testing.T) bool {
	// Compare basic fields
	if dc1.DefaultTarget != dc2.DefaultTarget {
		t.Logf("DefaultTarget mismatch: '%s' vs '%s'", dc1.DefaultTarget, dc2.DefaultTarget)
		return false
	}
	if dc1.BuildCommand != dc2.BuildCommand {
		t.Logf("BuildCommand mismatch: '%s' vs '%s'", dc1.BuildCommand, dc2.BuildCommand)
		return false
	}
	if dc1.OutputDir != dc2.OutputDir {
		t.Logf("OutputDir mismatch: '%s' vs '%s'", dc1.OutputDir, dc2.OutputDir)
		return false
	}

	// Compare string slices (handle nil vs empty slice)
	if !equalStringSlicesNilSafe(dc1.ExcludeFiles, dc2.ExcludeFiles) {
		t.Logf("ExcludeFiles mismatch: %v vs %v", dc1.ExcludeFiles, dc2.ExcludeFiles)
		return false
	}
	if !equalStringSlicesNilSafe(dc1.IncludeFiles, dc2.IncludeFiles) {
		t.Logf("IncludeFiles mismatch: %v vs %v", dc1.IncludeFiles, dc2.IncludeFiles)
		return false
	}

	// Compare targets
	if len(dc1.Targets) != len(dc2.Targets) {
		t.Logf("Targets length mismatch: %d vs %d", len(dc1.Targets), len(dc2.Targets))
		return false
	}
	for i, target1 := range dc1.Targets {
		if !compareDeploymentTargetDebug(target1, dc2.Targets[i], t, i) {
			return false
		}
	}

	return true
}

// compareDeploymentTargetDebug compares two DeploymentTarget structs with detailed logging
func compareDeploymentTargetDebug(dt1, dt2 DeploymentTarget, t *testing.T, index int) bool {
	// Compare basic fields
	if dt1.Name != dt2.Name {
		t.Logf("Target[%d] Name mismatch: '%s' vs '%s'", index, dt1.Name, dt2.Name)
		return false
	}
	if dt1.Type != dt2.Type {
		t.Logf("Target[%d] Type mismatch: '%s' vs '%s'", index, dt1.Type, dt2.Type)
		return false
	}
	if dt1.URL != dt2.URL {
		t.Logf("Target[%d] URL mismatch: '%s' vs '%s'", index, dt1.URL, dt2.URL)
		return false
	}
	if dt1.Username != dt2.Username {
		t.Logf("Target[%d] Username mismatch: '%s' vs '%s'", index, dt1.Username, dt2.Username)
		return false
	}
	if dt1.Password != dt2.Password {
		t.Logf("Target[%d] Password mismatch: '%s' vs '%s'", index, dt1.Password, dt2.Password)
		return false
	}
	if dt1.Token != dt2.Token {
		t.Logf("Target[%d] Token mismatch: '%s' vs '%s'", index, dt1.Token, dt2.Token)
		return false
	}
	if dt1.Region != dt2.Region {
		t.Logf("Target[%d] Region mismatch: '%s' vs '%s'", index, dt1.Region, dt2.Region)
		return false
	}
	if dt1.Bucket != dt2.Bucket {
		t.Logf("Target[%d] Bucket mismatch: '%s' vs '%s'", index, dt1.Bucket, dt2.Bucket)
		return false
	}
	if dt1.Path != dt2.Path {
		t.Logf("Target[%d] Path mismatch: '%s' vs '%s'", index, dt1.Path, dt2.Path)
		return false
	}
	if dt1.Port != dt2.Port {
		t.Logf("Target[%d] Port mismatch: %d vs %d", index, dt1.Port, dt2.Port)
		return false
	}
	if dt1.IsDefault != dt2.IsDefault {
		t.Logf("Target[%d] IsDefault mismatch: %v vs %v", index, dt1.IsDefault, dt2.IsDefault)
		return false
	}

	// Compare time (allow for small differences due to serialization)
	if !dt1.LastDeploy.Truncate(time.Second).Equal(dt2.LastDeploy.Truncate(time.Second)) {
		t.Logf("Target[%d] LastDeploy mismatch: %v vs %v", index, dt1.LastDeploy, dt2.LastDeploy)
		return false
	}

	// Compare config map (handle nil vs empty map)
	if !equalConfigMaps(dt1.Config, dt2.Config) {
		t.Logf("Target[%d] Config mismatch: %v vs %v", index, dt1.Config, dt2.Config)
		return false
	}

	return true
}

// equalStringSlicesNilSafe compares two string slices handling nil vs empty slice
func equalStringSlicesNilSafe(a, b []string) bool {
	// Handle nil vs empty slice equivalence
	if len(a) == 0 && len(b) == 0 {
		return true
	}
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

// compareDeploymentTarget compares two DeploymentTarget structs for equality
func compareDeploymentTarget(dt1, dt2 DeploymentTarget) bool {
	// Compare basic fields
	if dt1.Name != dt2.Name {
		return false
	}
	if dt1.Type != dt2.Type {
		return false
	}
	if dt1.URL != dt2.URL {
		return false
	}
	if dt1.Username != dt2.Username {
		return false
	}
	if dt1.Password != dt2.Password {
		return false
	}
	if dt1.Token != dt2.Token {
		return false
	}
	if dt1.Region != dt2.Region {
		return false
	}
	if dt1.Bucket != dt2.Bucket {
		return false
	}
	if dt1.Path != dt2.Path {
		return false
	}
	if dt1.Port != dt2.Port {
		return false
	}
	if dt1.IsDefault != dt2.IsDefault {
		return false
	}

	// Compare time (allow for small differences due to serialization)
	if !dt1.LastDeploy.Truncate(time.Second).Equal(dt2.LastDeploy.Truncate(time.Second)) {
		return false
	}

	// Compare config map (handle nil vs empty map)
	if !equalConfigMaps(dt1.Config, dt2.Config) {
		return false
	}

	return true
}

// equalConfigMaps compares two config maps handling nil vs empty map and type conversions
func equalConfigMaps(m1, m2 map[string]interface{}) bool {
	// Handle nil vs empty map equivalence
	if len(m1) == 0 && len(m2) == 0 {
		return true
	}
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		v2, exists := m2[k]
		if !exists {
			return false
		}
		
		// Handle type conversions that happen during JSON marshaling/unmarshaling
		if !compareInterfaceValues(v1, v2) {
			return false
		}
	}
	return true
}

// compareInterfaceValues compares two interface{} values handling type conversions
func compareInterfaceValues(v1, v2 interface{}) bool {
	// Direct comparison first
	if v1 == v2 {
		return true
	}
	
	// Handle numeric type conversions (JSON unmarshaling converts numbers to float64)
	switch val1 := v1.(type) {
	case int:
		if val2, ok := v2.(float64); ok {
			return float64(val1) == val2
		}
	case int64:
		if val2, ok := v2.(float64); ok {
			return float64(val1) == val2
		}
	case float64:
		if val2, ok := v2.(int); ok {
			return val1 == float64(val2)
		} else if val2, ok := v2.(int64); ok {
			return val1 == float64(val2)
		}
	}
	
	return false
}

// compareDeploymentStatus compares two DeploymentStatus structs for equality
func compareDeploymentStatus(ds1, ds2 DeploymentStatus) bool {
	// Compare basic fields
	if ds1.Target != ds2.Target {
		return false
	}
	if ds1.Status != ds2.Status {
		return false
	}
	if ds1.Message != ds2.Message {
		return false
	}
	if ds1.Error != ds2.Error {
		return false
	}
	if ds1.FilesCount != ds2.FilesCount {
		return false
	}
	if ds1.BytesTotal != ds2.BytesTotal {
		return false
	}
	if ds1.Progress != ds2.Progress {
		return false
	}

	// Compare times (allow for small differences due to serialization)
	if !ds1.StartTime.Truncate(time.Second).Equal(ds2.StartTime.Truncate(time.Second)) {
		return false
	}
	if !ds1.EndTime.Truncate(time.Second).Equal(ds2.EndTime.Truncate(time.Second)) {
		return false
	}

	return true
}

// genDeploymentConfig generates random DeploymentConfig for testing
func genDeploymentConfig() gopter.Gen {
	return gopter.CombineGens(
		gen.SliceOfN(3, genDeploymentTarget()),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		genStringSlice(),
		genStringSlice(),
	).Map(func(values []interface{}) DeploymentConfig {
		targets := values[0].([]DeploymentTarget)
		defaultTarget := ""
		
		// Ensure only one target is marked as default
		for i := range targets {
			targets[i].IsDefault = false
		}
		
		if len(targets) > 0 {
			defaultTarget = targets[0].Name
			targets[0].IsDefault = true
		}

		return DeploymentConfig{
			Targets:       targets,
			DefaultTarget: defaultTarget,
			BuildCommand:  values[2].(string),
			OutputDir:     values[3].(string),
			ExcludeFiles:  values[4].([]string),
			IncludeFiles:  values[5].([]string),
		}
	})
}

// genDeploymentTarget generates random DeploymentTarget for testing
func genDeploymentTarget() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.OneConstOf("ftp", "sftp", "s3", "github", "netlify", "vercel"),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.IntRange(1, 65535),
		genStringMap(),
		genTime(),
		gen.Bool(),
	).Map(func(values []interface{}) DeploymentTarget {
		return DeploymentTarget{
			Name:       values[0].(string),
			Type:       values[1].(string),
			URL:        values[2].(string),
			Username:   values[3].(string),
			Password:   values[4].(string),
			Token:      values[5].(string),
			Region:     values[6].(string),
			Bucket:     values[7].(string),
			Path:       values[8].(string),
			Port:       values[9].(int),
			Config:     values[10].(map[string]interface{}),
			LastDeploy: values[11].(time.Time),
			IsDefault:  values[12].(bool),
		}
	})
}

// genDeploymentStatus generates random DeploymentStatus for testing
func genDeploymentStatus() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.OneConstOf("pending", "running", "success", "failed"),
		genTime(),
		genTime(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.IntRange(0, 10000),
		gen.Int64Range(0, 1000000000),
		gen.Float64Range(0.0, 1.0),
	).Map(func(values []interface{}) DeploymentStatus {
		startTime := values[2].(time.Time)
		endTime := values[3].(time.Time)
		
		// Ensure endTime is after startTime
		if endTime.Before(startTime) {
			endTime = startTime.Add(time.Hour)
		}

		return DeploymentStatus{
			Target:     values[0].(string),
			Status:     values[1].(string),
			StartTime:  startTime,
			EndTime:    endTime,
			Message:    values[4].(string),
			Error:      values[5].(string),
			FilesCount: values[6].(int),
			BytesTotal: values[7].(int64),
			Progress:   values[8].(float64),
		}
	})
}

// Test validation functions
func TestDeploymentTargetValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("valid deployment targets should pass validation", prop.ForAll(
		func(target DeploymentTarget) bool {
			// Ensure the target has required fields for its type
			switch target.Type {
			case "ftp", "sftp":
				if target.URL == "" || target.Username == "" {
					return true // Skip invalid targets for this test
				}
			case "s3":
				if target.Bucket == "" {
					return true // Skip invalid targets for this test
				}
			case "github", "netlify", "vercel":
				if target.Token == "" {
					return true // Skip invalid targets for this test
				}
			}

			err := target.Validate()
			return err == nil
		},
		genValidDeploymentTarget(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genValidDeploymentTarget generates valid DeploymentTarget for testing
func genValidDeploymentTarget() gopter.Gen {
	return gen.OneConstOf("ftp", "sftp", "s3", "github", "netlify", "vercel").FlatMap(func(targetType interface{}) gopter.Gen {
		switch targetType.(string) {
		case "ftp", "sftp":
			return gopter.CombineGens(
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.Const(targetType),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.AlphaString(),
				gen.AlphaString(),
				gen.IntRange(1, 65535),
				gen.AlphaString(),
				genTime(),
			).Map(func(values []interface{}) DeploymentTarget {
				return DeploymentTarget{
					Name:       values[0].(string),
					Type:       values[1].(string),
					URL:        values[2].(string),
					Username:   values[3].(string),
					Password:   values[4].(string),
					Path:       values[5].(string),
					Port:       values[6].(int),
					Config:     make(map[string]interface{}),
					LastDeploy: values[8].(time.Time),
				}
			})
		case "s3":
			return gopter.CombineGens(
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.Const(targetType),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.AlphaString(),
				genTime(),
			).Map(func(values []interface{}) DeploymentTarget {
				return DeploymentTarget{
					Name:       values[0].(string),
					Type:       values[1].(string),
					Bucket:     values[2].(string),
					Region:     values[3].(string),
					Path:       values[4].(string),
					Config:     make(map[string]interface{}),
					LastDeploy: values[5].(time.Time),
				}
			})
		case "github", "netlify", "vercel":
			return gopter.CombineGens(
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.Const(targetType),
				gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
				gen.AlphaString(),
				gen.AlphaString(),
				genTime(),
			).Map(func(values []interface{}) DeploymentTarget {
				return DeploymentTarget{
					Name:       values[0].(string),
					Type:       values[1].(string),
					Token:      values[2].(string),
					URL:        values[3].(string),
					Path:       values[4].(string),
					Config:     make(map[string]interface{}),
					LastDeploy: values[5].(time.Time),
				}
			})
		default:
			return gen.Fail(reflect.TypeOf(DeploymentTarget{}))
		}
	}, reflect.TypeOf(DeploymentTarget{}))
}

// Test deployment config operations
func TestDeploymentConfigOperations(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("adding and removing targets should maintain consistency", prop.ForAll(
		func(config DeploymentConfig, newTarget DeploymentTarget) bool {
			// Ensure new target has a unique name
			for _, existing := range config.Targets {
				if existing.Name == newTarget.Name {
					newTarget.Name = newTarget.Name + "_unique"
					break
				}
			}

			// Make sure the new target is valid
			if newTarget.Name == "" {
				newTarget.Name = "test_target"
			}
			if newTarget.Type == "" {
				newTarget.Type = "ftp"
			}
			if newTarget.URL == "" {
				newTarget.URL = "ftp.example.com"
			}
			if newTarget.Username == "" {
				newTarget.Username = "testuser"
			}

			originalCount := len(config.Targets)

			// Add target
			err := config.AddTarget(newTarget)
			if err != nil {
				t.Logf("Failed to add target: %v", err)
				return false
			}

			// Check that target was added
			if len(config.Targets) != originalCount+1 {
				t.Logf("Target count mismatch after add: expected %d, got %d", originalCount+1, len(config.Targets))
				return false
			}

			// Find the added target
			found := false
			for _, target := range config.Targets {
				if target.Name == newTarget.Name {
					found = true
					break
				}
			}
			if !found {
				t.Logf("Added target not found in targets list")
				return false
			}

			// Remove target
			err = config.RemoveTarget(newTarget.Name)
			if err != nil {
				t.Logf("Failed to remove target: %v", err)
				return false
			}

			// Check that target was removed
			if len(config.Targets) != originalCount {
				t.Logf("Target count mismatch after remove: expected %d, got %d", originalCount, len(config.Targets))
				return false
			}

			// Ensure target is no longer in the list
			for _, target := range config.Targets {
				if target.Name == newTarget.Name {
					t.Logf("Removed target still found in targets list")
					return false
				}
			}

			return true
		},
		genDeploymentConfig(),
		genValidDeploymentTarget(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}