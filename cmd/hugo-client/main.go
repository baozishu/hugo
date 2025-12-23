//go:build !js
// +build !js

package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	internalapp "hugo-visual-client/internal/app"
)

func main() {
	// Create Fyne application
	myApp := app.New()

	// Create main window
	myWindow := myApp.NewWindow("Hugo Visual Client")
	myWindow.Resize(fyne.NewSize(1200, 800))

	// Initialize application controller
	appController, err := internalapp.NewAppController(myWindow)
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