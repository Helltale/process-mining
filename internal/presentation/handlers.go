package presentation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/service"
)

type GraphHandler struct {
	graphService *service.GraphService
}

func NewGraphHandler(graphService *service.GraphService) *GraphHandler {
	return &GraphHandler{graphService: graphService}
}

func (h *GraphHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Установка лимита на размер тела запроса (например, 3 ГБ)
	r.Body = http.MaxBytesReader(w, r.Body, 3*1024*1024*1024) // 3 ГБ

	// Проверка наличия файла
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Создаем временный файл
	tempFile, err := os.CreateTemp("", "uploaded-*.csv")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Копируем содержимое загруженного файла во временный файл по частям
	buf := make([]byte, 1024*1024) // Буфер размером 1 МБ
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, writeErr := tempFile.Write(buf[:n]); writeErr != nil {
				http.Error(w, "Ошибка записи во временный файл", http.StatusInternalServerError)
				return
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
			return
		}
	}

	// Вызов сервисного слоя для обработки файла
	err = h.graphService.BuildGraphFromCSV(tempFile.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка построения графа: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Файл успешно загружен и граф построен"))
}

func (h *GraphHandler) ServeGraphData(w http.ResponseWriter, r *http.Request) {
	graphData, err := h.graphService.GetGraphData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// // Логирование данных для отладки
	// for _, edge := range graphData.Edges {
	// 	fmt.Printf("Edge: %s -> %s, Style: %s\n", edge.From, edge.To, edge.Style)
	// }

	// Преобразуем данные в формат, понятный фронтенду
	cytoscapeData := struct {
		Nodes []map[string]*domain.Node `json:"nodes"`
		Edges []map[string]*domain.Edge `json:"edges"`
	}{
		Nodes: make([]map[string]*domain.Node, len(graphData.Nodes)),
		Edges: make([]map[string]*domain.Edge, len(graphData.Edges)),
	}

	for i, node := range graphData.Nodes {
		cytoscapeData.Nodes[i] = map[string]*domain.Node{"data": node}
	}

	for i, edge := range graphData.Edges {
		edge.Label = fmt.Sprintf("%d\n%.2f sec avg", edge.Count, edge.AvgDuration)
		cytoscapeData.Edges[i] = map[string]*domain.Edge{"data": edge}
	}

	// Отправляем данные клиенту
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cytoscapeData); err != nil {
		http.Error(w, "Ошибка сериализации", http.StatusInternalServerError)
		return
	}
}
