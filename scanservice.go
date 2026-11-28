package main

import (
	"os"
	"path/filepath"
)

type ScanService struct{}
type ScanResult struct {
	RootPath       string `json:"rootPath"`
	FileCount      int64  `json:"fileCount"`
	DirectoryCount int64  `json:"directoryCount"`
	TotalSize      int64  `json:"totalSize"`
	ErrorCount     int64  `json:"errorCount"`
	Completed      bool   `json:"completed"`
	ErrorMessage   string `json:"errorMessage"`
}

func (s *ScanService) Scan(rootPath string) ScanResult {
	result := ScanResult{
		RootPath: rootPath,
	}
	info, err := os.Lstat(rootPath)
	if err != nil {
		result.ErrorCount++
		result.ErrorMessage = "Не удалось открыть папочку: " + err.Error()
		return result
	}
	isReparse, err := isReparsePoint(info)
	if err != nil {
		result.ErrorCount++
		result.ErrorMessage = "Не удалось проверить безопасность папочки: " + err.Error()
		return result
	}
	if isReparse {
		result.ErrorCount++
		result.ErrorMessage = "Сканирование reparse points запрещено"
		return result
	}
	if !info.IsDir() {
		result.ErrorCount++
		result.ErrorMessage = "Выбранный путь не является обычной папочкой"
		return result
	}
	walkErr := filepath.WalkDir(rootPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			result.ErrorCount++
			if result.ErrorMessage == "" {
				result.ErrorMessage = "Не удалось прочитать папочку или файл: " + err.Error()
			}
			if path == rootPath {
				return err
			}
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			result.ErrorCount++
			if result.ErrorMessage == "" {
				result.ErrorMessage = "Не удалось получить информацию об объекте: " + infoErr.Error()
			}
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		isReparse, reparseErr := isReparsePoint(info)
		if reparseErr != nil {
			result.ErrorCount++
			if result.ErrorMessage == "" {
				result.ErrorMessage = "Не удалось проверить безопасность объекта: " + reparseErr.Error()
			}
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if isReparse {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if path != rootPath {
				result.DirectoryCount++
			}
		} else if info.Mode().IsRegular() {
			result.FileCount++
			result.TotalSize += info.Size()
		}
		return nil
	})
	if walkErr != nil && result.ErrorCount == 0 {
		result.ErrorCount++
		result.ErrorMessage = "Сканирование прервано: " + walkErr.Error()
	}
	result.Completed = result.ErrorCount == 0
	return result
}
