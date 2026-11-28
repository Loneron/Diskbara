package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
)

type unknownFileInfo struct {
	os.FileInfo
}

func (unknownFileInfo) Sys() any {
	return nil
}

type nilAttributesFileInfo struct {
	os.FileInfo
}

func (nilAttributesFileInfo) Sys() any {
	return (*syscall.Win32FileAttributeData)(nil)
}
func TestIsReparsePointJunction(t *testing.T) {
	root := t.TempDir()
	target := t.TempDir()
	junction := filepath.Join(root, "junction")
	cmd := exec.Command("cmd", "/c", "mklink", "/J", junction, target)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Не удалось создать junction: %v\n%s", err, output)
	}
	info, err := os.Lstat(junction)
	if err != nil {
		t.Fatal(err)
	}
	isReparse, err := isReparsePoint(info)
	if err != nil {
		t.Fatal(err)
	}
	if !isReparse {
		t.Error("Windows junction должен определяться как reparse point")
	}
}
func TestIsReparsePointSymlink(t *testing.T) {
	root := t.TempDir()
	target := t.TempDir()
	link := filepath.Join(root, "symlink")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Не удалось создать символическую ссылку: %v", err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	isReparse, err := isReparsePoint(info)
	if err != nil {
		t.Fatal(err)
	}
	if !isReparse {
		t.Error("Символическая ссылка должна определяться как reparse point")
	}
}
func TestIsReparsePointRegularFile(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "file.txt")
	if err := os.WriteFile(filePath, []byte("Diskbara"), 0644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filePath)
	if err != nil {
		t.Fatal(err)
	}
	isReparse, err := isReparsePoint(info)
	if err != nil {
		t.Fatal(err)
	}
	if isReparse {
		t.Error("Обычный файл не должен определяться как reparse point")
	}
}
func TestIsReparsePointRegularDirectory(t *testing.T) {
	root := t.TempDir()
	info, err := os.Lstat(root)
	if err != nil {
		t.Fatal(err)
	}
	isReparse, err := isReparsePoint(info)
	if err != nil {
		t.Fatal(err)
	}
	if isReparse {
		t.Error("Обычная папочка не должна определяться как reparse point")
	}
}
func TestIsReparsePointNilInfo(t *testing.T) {
	isReparse, err := isReparsePoint(nil)
	if err == nil {
		t.Fatal("При отсутствии информации должна возвращаться ошибка")
	}
	if isReparse {
		t.Error("При отсутствии информации нельзя утверждать, что объект является reparse point")
	}
}
func TestIsReparsePointUnknownAttributes(t *testing.T) {
	info := unknownFileInfo{}
	isReparse, err := isReparsePoint(info)
	if err == nil {
		t.Fatal("При неизвестных атрибутах должна возвращаться ошибка")
	}
	if isReparse {
		t.Error("Нельзя определить reparse point без атрибутов Windows")
	}
}
func TestIsReparsePointNilAttributes(t *testing.T) {
	info := nilAttributesFileInfo{}

	isReparse, err := isReparsePoint(info)

	if err == nil {
		t.Fatal("При nil-атрибутах должна возвращаться ошибка")
	}

	if isReparse {
		t.Error("Нельзя определить reparse point при nil-атрибутах")
	}
}
func TestScanMixedStructure(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(root, "file.txt"),
		[]byte("Hello"),
		0644,
	); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(nested, "data.txt"),
		[]byte("Diskbara"),
		0644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(outside, "secret.txt"),
		[]byte("Secret data"),
		0644,
	); err != nil {
		t.Fatal(err)
	}
	junction := filepath.Join(root, "junction")
	cmd := exec.Command("cmd", "/c", "mklink", "/J", junction, outside)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Не удалось создать junction: %v\n%s", err, output)
	}
	symlink := filepath.Join(root, "symlink")
	if err := os.Symlink(outside, symlink); err != nil {
		t.Fatal(err)
	}
	service := ScanService{}
	result := service.Scan(root)
	if result.RootPath != root {
		t.Errorf("Некорректный корневой путь: %s", result.RootPath)
	}
	if result.FileCount != 2 {
		t.Errorf("Ожидалось 2 файла, получено %d", result.FileCount)
	}
	if result.DirectoryCount != 1 {
		t.Errorf("Ожидался 1 каталог, получено %d", result.DirectoryCount)
	}
	if result.TotalSize != 13 {
		t.Errorf("Ожидалось 13 байт, получено %d", result.TotalSize)
	}
	if result.ErrorCount != 0 {
		t.Errorf("Обнаружены ошибки: %d", result.ErrorCount)
	}
	if !result.Completed {
		t.Error("Сканирование должно завершиться успешно")
	}
	if result.ErrorMessage != "" {
		t.Errorf("Неожиданное сообщение об ошибке: %s", result.ErrorMessage)
	}
}
