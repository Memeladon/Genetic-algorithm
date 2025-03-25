package main

import (
	"Genetic-algorithm/frontend"
	"fyne.io/fyne/v2/app"
)

func main() {
	myApp := app.New()
	mainWindow := frontend.NewMainWindow(myApp)
	mainWindow.Window.ShowAndRun()
}
