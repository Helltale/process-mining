package presentation

import (
	"encoding/json"
	"net/http"

	"github.com/Helltale/process-mining/internal/service"
)

type GraphHandler struct {
	graphService *service.GraphService
}

func NewGraphHandler(graphService *service.GraphService) *GraphHandler {
	return &GraphHandler{graphService: graphService}
}

func (h *GraphHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Реализация загрузки файла через сервисный слой
}

func (h *GraphHandler) ServeGraphData(w http.ResponseWriter, r *http.Request) {
	graphData, err := h.graphService.GetGraphData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphData)
}
