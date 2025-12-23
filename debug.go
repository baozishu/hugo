package main

import (
	"encoding/json"
	"fmt"
	"hugo-visual-client/internal/models"
	"time"
)

func main() {
	// Create a simple deployment config
	config := models.DeploymentConfig{
		Targets: []models.DeploymentTarget{
			{
				Name:       "test",
				Type:       "ftp",
				URL:        "ftp.example.com",
				Username:   "user",
				Password:   "pass",
				Path:       "/",
				Port:       21,
				Config:     map[string]interface{}{"key": "value"},
				LastDeploy: time.Now().UTC().Truncate(time.Second),
				IsDefault:  true,
			},
		},
		DefaultTarget: "test",
		BuildCommand:  "hugo",
		OutputDir:     "public",
		ExcludeFiles:  []string{"*.log"},
		IncludeFiles:  []string{"*.html"},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(config)
	if err != nil {
		fmt.Printf("JSON marshal error: %v\n", err)
		return
	}

	var deserializedConfig models.DeploymentConfig
	err = json.Unmarshal(jsonData, &deserializedConfig)
	if err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return
	}

	fmt.Printf("Original: %+v\n", config)
	fmt.Printf("Deserialized: %+v\n", deserializedConfig)
	
	// Compare fields
	fmt.Printf("DefaultTarget equal: %v\n", config.DefaultTarget == deserializedConfig.DefaultTarget)
	fmt.Printf("BuildCommand equal: %v\n", config.BuildCommand == deserializedConfig.BuildCommand)
	fmt.Printf("OutputDir equal: %v\n", config.OutputDir == deserializedConfig.OutputDir)
	fmt.Printf("Targets length equal: %v\n", len(config.Targets) == len(deserializedConfig.Targets))
	
	if len(config.Targets) > 0 && len(deserializedConfig.Targets) > 0 {
		t1 := config.Targets[0]
		t2 := deserializedConfig.Targets[0]
		fmt.Printf("Target Name equal: %v ('%s' vs '%s')\n", t1.Name == t2.Name, t1.Name, t2.Name)
		fmt.Printf("Target Type equal: %v ('%s' vs '%s')\n", t1.Type == t2.Type, t1.Type, t2.Type)
		fmt.Printf("Target Config equal: %v (%+v vs %+v)\n", fmt.Sprintf("%v", t1.Config) == fmt.Sprintf("%v", t2.Config), t1.Config, t2.Config)
	}
}