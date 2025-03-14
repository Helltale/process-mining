package infrastructure

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type CSVReader struct{}
type TMPCleaner struct{}

func NewCSVReader() *CSVReader {
	return &CSVReader{}
}

func NewTMPCleaner() *TMPCleaner {
	return &TMPCleaner{}
}

func (r *CSVReader) ReadAndProcess(filePath string, processFunc func([]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read() // without file header
	if err != nil && err != io.EOF {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := processFunc(record); err != nil {
			return err
		}
	}

	return nil
}

func (c *TMPCleaner) ClearTempFiles() error {
	// Указываем путь к директории /tmp
	dir := "/tmp"

	// Читаем содержимое директории
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("ошибка чтения директории %s: %v", dir, err)
	}

	// Удаляем все файлы в директории
	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		if err := os.Remove(filePath); err != nil {
			log.Printf("Ошибка удаления файла %s: %v", filePath, err)
		} else {
			log.Printf("Файл удален: %s", filePath)
		}
	}

	return nil
}
