//go:build !js
// +build !js

package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"hugo-visual-client/internal/app"
)

func main() {
	// Create Fyne application
	myApp := app.New()
	myApp.SetMetadata(&app.Metadata{
		Name:        "Hugo Visual Client",
		Description: "A visual client for Hugo static site generator",
		Version:     "1.0.0",
	})

	// Create main window
	myWindow := myApp.NewWindow("Hugo Visual Client")
	myWindow.Resize(fyne.NewSize(1200, 800))

	// Initialize application controller
	appController, err := app.NewAppController(myWindow)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Set up the main content
	content := widget.NewLabel("Hugo Visual Client - Loading...")
	myWindow.SetContent(content)

	// Initialize the UI
	appController.InitializeUI()

	// Show window and run
	myWindow.ShowAndRun()
}