package infrastructure

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
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

// работает даже дольше
func (r *CSVReader) ReadAndProcessConcurrent(filePath string, processFunc func([]string) error) error {
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

	// Создаем канал для передачи строк
	lines := make(chan string, 100) // Буферизованный канал для строк
	errChan := make(chan error, 1)  // Канал для ошибок

	// Горутина для чтения строк из файла
	go func() {
		defer close(lines)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	// Горутины для обработки строк
	numWorkers := runtime.NumCPU() // Количество ядер процессора
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for line := range lines {
				record := strings.Split(line, ",")
				if err := processFunc(record); err != nil {
					errChan <- err
					return
				}
			}
		}()
	}

	// Ожидаем завершения всех горутин
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Проверяем наличие ошибок
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}
