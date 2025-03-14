package infrastructure

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type TMPFileManager struct{}

func NewTMPFileManager() *TMPFileManager {
	return &TMPFileManager{}
}

// Функция для создания временного файла
func (t *TMPFileManager) CreateTempFile(prefix, suffix string) (*os.File, error) {
	tempDir := "./tmp"                                 // Пользовательская директория для временных файлов // TODO ВЫНЕСТИ В CONFIG
	if err := os.MkdirAll(tempDir, 0755); err != nil { // TODO ВЫНЕСТИ В CONFIG
		return nil, fmt.Errorf("ошибка создания директории %s: %v", tempDir, err)
	}

	tempFile, err := os.CreateTemp(tempDir, fmt.Sprintf("%s*.%s", prefix, suffix))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания временного файла: %v", err)
	}

	log.Printf("Создан временный файл: %s", tempFile.Name())
	return tempFile, nil
}

// DeleteTempFile удаляет все временные файлы в пользовательской директории
func (m *TMPFileManager) DeleteTempFile() error {
	tempDir := "./tmp" // Пользовательская директория для временных файлов // TODO ВЫНЕСТИ В CONFIG

	// Читаем содержимое директории
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	// Удаляем все файлы в директории
	for _, file := range files {
		filePath := filepath.Join(tempDir, file.Name())
		if err := os.Remove(filePath); err != nil {
			log.Printf("Ошибка удаления файла %s: %v", filePath, err)
		} else {
			log.Printf("Файл удален: %s", filePath)
		}
	}

	return nil
}
