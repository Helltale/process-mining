package service

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/infrastructure"
)

// TODO: TMP TO CONFIG
const LargeFileSizeThreshold = 1.8 * 1024 * 1024 * 1024  // 1.8 ГБ в байтах
const LargeFileSizeThreshold2 = 3.1 * 1024 * 1024 * 1024 // 3.1 ГБ в байтах

type GraphService struct {
	graphBuilder *domain.GraphBuilder
}

func NewGraphService(graphBuilder *domain.GraphBuilder) *GraphService {
	return &GraphService{graphBuilder: graphBuilder}
}

func (s *GraphService) GetGraphData() (*domain.Graph, error) {
	return s.graphBuilder.GetGraph(), nil
}

func (s *GraphService) ClearGraph() {
	s.graphBuilder.ClearGraph()
}

func (s *GraphService) BuildGraphFromCSV(filePath string) error {
	fileSize, err := infrastructure.GetFileSize(filePath)
	if err != nil {
		return fmt.Errorf("ошибка получения размера файла: %v", err)
	}

	processFunc := func(record []string) error {
		if len(record) != 3 {
			return fmt.Errorf("некорректная строка: %v", record)
		}

		timestamp, err := time.Parse(time.RFC3339, record[1])
		if err != nil {
			return fmt.Errorf("ошибка парсинга времени '%s': %v", record[1], err)
		}

		event := &domain.Event{
			ID:        record[0],
			SessionID: record[0],
			Timestamp: timestamp,
			Desc:      record[2],
		}

		s.graphBuilder.ProcessEvent(event) // Вызываем экспортированный метод
		return nil
	}

	if fileSize > LargeFileSizeThreshold2 {
		numb := 18
		runtime.GOMAXPROCS(numb)
		slog.Info("Конкурентная обработка для больших файлов")
		slog.Info("Ограничили CPU", "значение", numb)

		return s.graphBuilder.BuildGraphSequential2(filePath, processFunc)
	}

	if fileSize > LargeFileSizeThreshold {

		//TODO: вынести runtime.GOMAXPROCS(numb) в конфиг
		// Ограничиваем использование CPU до numb ядер (или другого значения)
		numb := 18
		runtime.GOMAXPROCS(numb)
		// Конкурентная обработка для больших файлов
		slog.Info("Конкурентная обработка для больших файлов")
		slog.Info("Ограничили CPU", "значение", numb)

		// TODO: вынести в конфиг обработку больших файлов
		LargeFileMethod := true
		if LargeFileMethod {
			// последовательный большой файл обработка
			slog.Info("последовательная обработка")
			return s.graphBuilder.BuildGraphSequential(filePath, processFunc)
		}

		// конкурентная обработка, сомнительно потому что io ожидание большое получится
		slog.Info("конкурентная обработка")
		return s.graphBuilder.BuildGraphConcurrent(filePath, processFunc)
	}

	// Обычная обработка для маленьких файлов
	slog.Info("Обычная обработка для маленьких файлов")
	return s.graphBuilder.BuildGraphSequential(filePath, processFunc)
}
