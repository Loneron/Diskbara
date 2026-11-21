package main

import "github.com/wailsapp/wails/v3/pkg/application"

type FolderService struct{}

func (s *FolderService) SelectFolder() (string, error) {
	app := application.Get()
	path, err := app.Dialog.OpenFile().
		SetTitle("Diskbara - выбери папочку").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return path, nil
}
