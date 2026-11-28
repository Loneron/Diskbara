package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestScanBasic(t *testing.T) {
	root := t.TempDir()
	subfolder := filepath.Join(root, "documents")
	if err := os.Mkdir(subfolder, 0755); err != nil {
		t.Fatal(err)
	}
	file1 := filepath.Join(root, "hello.txt")
	file2 := filepath.Join(subfolder, "notes.txt")
	if err := os.WriteFile(file1, []byte("Hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("Diskbara"), 0644); err != nil {
		t.Fatal(err)
	}
	scanner := &ScanService{}
	result := scanner.Scan(root)
	if result.FileCount != 2 {
		t.Errorf("Ожидалось 2 файла, получено %d", result.FileCount)
	}
	if result.DirectoryCount != 1 {
		t.Errorf("Ожидалась 1 папка, получено %d", result.DirectoryCount)
	}
	if result.TotalSize != 13 {
		t.Errorf("Ожидалось 13 байт, получено %d", result.TotalSize)
	}
	if result.ErrorCount != 0 {
		t.Errorf("Ожидалось 0 ошибок, получено %d", result.ErrorCount)
	}
	if !result.Completed {
		t.Error("Успешное сканирование должно иметь Complited = true")
	}
}
func TestScanNonexistentFolder(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nonexistent")
	scanner := &ScanService{}
	result := scanner.Scan(root)
	if result.ErrorCount == 0 {
		t.Error("Ожидалась ошибка для несуществующей папки")
	}
	if result.Completed {
		t.Error("Сканирование не должно считаться завершённым")
	}
	if result.FileCount != 0 || result.DirectoryCount != 0 {
		t.Error("Несуществующая папка не должна содержать файлов и каталогов")
	}
	if result.ErrorMessage == "" {
		t.Error("Ожидалось описание ошибки для несуществующей папочки")
	}
}
func TestScanDoesNotFollowSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secretFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "external-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("Не удалось создать символическую ссылку: %v", err)
	}
	scanner := &ScanService{}
	result := scanner.Scan(root)
	if result.FileCount != 0 {
		t.Errorf("Ожидалось 0 файлов, получено %d", result.FileCount)
	}
	if result.DirectoryCount != 0 {
		t.Errorf("Ожидалось 0 каталогов, получено %d", result.DirectoryCount)
	}
	if result.TotalSize != 0 {
		t.Errorf("Сканер вышел за пределы папки: размер %d", result.TotalSize)
	}
}
func TestScanDoesNotFollowJunction(t *testing.T) {
	if os.PathSeparator != '\\' {
		t.Skip("Тест предназначен только для Windows")
	}

	root := t.TempDir()
	outside := t.TempDir()

	secretFile := filepath.Join(outside, "secret.txt")

	if err := os.WriteFile(secretFile, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	junction := filepath.Join(root, "external-junction")
	cmd := exec.Command("cmd", "/c", "mklink", "/J", junction, outside)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("Не удалось создать junction: %v, %s", err, output)
	}
	scanner := &ScanService{}
	result := scanner.Scan(root)
	if result.FileCount != 0 {
		t.Errorf("Ожидалось 0 файлов, получено %d", result.FileCount)
	}
	if result.DirectoryCount != 0 {
		t.Errorf("Ожидалось 0 каталогов, получено %d", result.DirectoryCount)
	}
	if result.TotalSize != 0 {
		t.Errorf("Сканер вышел за пределы папки: %d байт", result.TotalSize)
	}
}
func TestScanRejectsRootSymlink(t *testing.T) {
	target := t.TempDir()
	root := filepath.Join(t.TempDir(), "linked-folder")
	if err := os.Symlink(target, root); err != nil {
		t.Skipf("Не удалось создать символическую ссылку: %v", err)
	}
	scanner := &ScanService{}
	result := scanner.Scan(root)
	if result.ErrorCount == 0 {
		t.Error("Сканер должен отклонить корневую символическую ссылку")
	}
	if result.Completed {
		t.Error("Отклонённое сканирование не должно иметь Completed = true")
	}
	if result.FileCount != 0 || result.DirectoryCount != 0 || result.TotalSize != 0 {
		t.Error("Сканер не должен анализировать содержимое корневой ссылки")
	}
}
func TestScanRejectsRootJunction(t *testing.T) {
	if os.PathSeparator != '\\' {
		t.Skip("Тест предназначен только для Windows")
	}
	target := t.TempDir()
	root := filepath.Join(t.TempDir(), "junction-folder")
	secretFile := filepath.Join(target, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cmd", "/c", "mklink", "/J", root, target)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("Не удалось создать junction: %v, %s", err, output)
	}
	scanner := &ScanService{}
	result := scanner.Scan(root)
	if result.ErrorCount == 0 {
		t.Error("Сканер должен отклонить корневой junction point")
	}
	if result.Completed {
		t.Error("Отклонённое сканирование не должно иметь Completed = true")
	}
	if result.FileCount != 0 || result.DirectoryCount != 0 || result.TotalSize != 0 {
		t.Errorf(
			"Сканер не должен обходить корневой junction: files=%d, dirs=%d, size=%d",
			result.FileCount,
			result.DirectoryCount,
			result.TotalSize,
		)
	}
}
func TestScanEmptyFolder(t *testing.T) {
	root := t.TempDir()
	scanner := &ScanService{}
	result := scanner.Scan(root)
	if result.RootPath != root {
		t.Errorf("Ожидался путь %q, получено %q", root, result.RootPath)
	}
	if result.FileCount != 0 {
		t.Errorf("Ожидалось 0 файлов, получено %d", result.FileCount)
	}
	if result.DirectoryCount != 0 {
		t.Errorf("Ожидалось 0 каталогов, получено %d", result.DirectoryCount)
	}
	if result.TotalSize != 0 {
		t.Errorf("Ожидалось 0 байт, получено %d", result.TotalSize)
	}
	if result.ErrorCount != 0 {
		t.Errorf("Ожидалось 0 ошибок, получено %d", result.ErrorCount)
	}
	if !result.Completed {
		t.Error("Сканирование пустой папочки должно завершиться успешно")
	}
}
func TestScanRejectsRootFile(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "document.txt")
	if err := os.WriteFile(filePath, []byte("Diskbara"), 0644); err != nil {
		t.Fatal(err)
	}
	scanner := &ScanService{}
	result := scanner.Scan(filePath)
	if result.ErrorCount == 0 {
		t.Error("Сканер должен отклонить обычный файл вместо папочки")
	}
	if result.Completed {
		t.Error("Отклонённое сканирование не должно иметь Completed = true")
	}
	if result.FileCount != 0 ||
		result.DirectoryCount != 0 ||
		result.TotalSize != 0 {
		t.Errorf(
			"Ожидались нулевые счётчики, получено: files=%d, dirs=%d, size=%d",
			result.FileCount,
			result.DirectoryCount,
			result.TotalSize,
		)
	}
}
func TestScanRejectsEmptyPath(t *testing.T) {
	scanner := &ScanService{}
	result := scanner.Scan("")
	if result.ErrorCount == 0 {
		t.Error("Сканер должен отклонить пустой путь")
	}
	if result.Completed {
		t.Error("Сканирование пустого пути не должно считаться успешным")
	}
	if result.FileCount != 0 ||
		result.DirectoryCount != 0 ||
		result.TotalSize != 0 {
		t.Errorf(
			"Ожидались нулевые счётчики, получено: files=%d, dirs=%d, size=%d",
			result.FileCount,
			result.DirectoryCount,
			result.TotalSize,
		)
	}
	if result.ErrorMessage == "" {
		t.Error("Ожидалось описание ошибки при выборе файла вместо папочки")
	}
}
