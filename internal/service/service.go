package service

import (
	"fmt"
	"time"

	"github.com/Helltale/process-mining/internal/domain"
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
	s.graphBuilder.ClearGraph()

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
		s.graphBuilder.ProcessEvent(event)
		return nil
	}

	return s.graphBuilder.BuildGraphSequentialLargeFile(filePath, processFunc)
}
