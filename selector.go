package main

import (
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var selectWg sync.WaitGroup

func selectFiles() (string, string) {
	var inputFilePath string
	var outputDirPath string

	myApp := app.New()
	myWindow := myApp.NewWindow("IO Selector")
	myWindow.Resize(fyne.NewSize(800, 600))

	inputLabel := widget.NewLabel("")
	outputLabel := widget.NewLabel("")

	//done button to close window
	finishButton := widget.NewButton("Finished", func() {
		myWindow.Close()
	})
	finishButton.Hide()

	// dialog that prompts user for input file
	selectDataDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		// collect selected file path.
		defer reader.Close()
		inputFilePath = reader.URI().Path()
		inputLabel.SetText(filepath.Base(inputFilePath))
		inputLabel.Refresh()

		// reveal done button if both paths are filled
		if outputDirPath != "" {
			finishButton.Show()
		}
		selectWg.Done()
	}, myWindow)
	selectDataDialog.Resize(fyne.NewSize(700, 500))

	// set the dialog's default directory
	uri, err := storage.ParseURI("file:///Users/samwalker/Library/Mobile Documents/com~apple~CloudDocs/Projects/Go/PP5 Image Migrator/resources")
	if err != nil {
		panic(err)
	}
	luri, err := storage.ListerForURI(uri)
	if err != nil {
		panic(err)
	}
	selectDataDialog.SetLocation(luri)

	// create button to activate input dialog
	inputPromptButton := widget.NewButton("Select input data", func() {
		selectWg.Add(1)
		selectDataDialog.Show()
	})

	selectOutputDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		// collect selected file path.
		outputDirPath = lu.Path()
		outputLabel.SetText(filepath.Base(outputDirPath))
		outputLabel.Refresh()

		// reveal done button if both paths are filled
		if inputFilePath != "" {
			finishButton.Show()
		}
		selectWg.Done()
	}, myWindow)
	selectOutputDialog.Resize(fyne.NewSize(700, 500))
	selectOutputDialog.SetLocation(luri)

	//create button to activate output dialog
	outputPromptButton := widget.NewButton("Select output folder", func() {
		selectWg.Add(1)
		selectOutputDialog.Show()
	})

	inputBox := container.NewHBox(inputPromptButton, inputLabel)
	outputBox := container.NewHBox(outputPromptButton, outputLabel)

	myWindow.SetContent(container.NewVBox(inputBox, outputBox, finishButton))
	myWindow.ShowAndRun()

	selectWg.Wait()

	return inputFilePath, outputDirPath
}
