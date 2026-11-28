package main

import (
	"errors"
	"os"
	"syscall"
)

func isReparsePoint(info os.FileInfo) (bool, error) {
	if info == nil {
		return false, errors.New("Отсутствует информация об объекте")
	}
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok || data == nil {
		return false, errors.New("Не удалось получить атрибуты Windows")
	}
	return data.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0, nil
}
