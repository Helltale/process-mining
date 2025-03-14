package infrastructure

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type CSVReader struct{}

func NewCSVReader() *CSVReader {
	return &CSVReader{}
}

func (r *CSVReader) ReadAndProcess(filePath string, processFunc func([]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Пропускаем заголовок
	if scanner.Scan() {
		header := scanner.Text()
		log.Printf("Пропущен заголовок: %s", header)
	}

	// Обрабатываем остальные строки
	for scanner.Scan() {
		line := scanner.Text()
		record := strings.Split(line, ",") // Разделяем строку на поля
		if err := processFunc(record); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
