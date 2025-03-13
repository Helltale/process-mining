package service

import (
	"github.com/Helltale/process-mining/internal/domain"
)

type GraphService struct {
	graphBuilder *domain.GraphBuilder
}

func NewGraphService(graphBuilder *domain.GraphBuilder) *GraphService {
	return &GraphService{graphBuilder: graphBuilder}
}

func (s *GraphService) BuildGraphFromCSV(filePath string) error {
	return s.graphBuilder.BuildGraph(filePath)
}

func (s *GraphService) GetGraphData() (*domain.Graph, error) {
	return s.graphBuilder.GetGraph(), nil
}

func (s *GraphService) ClearGraph() {
	s.graphBuilder.ClearGraph()
}
