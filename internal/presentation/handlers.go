package presentation

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/infrastructure"
	"github.com/Helltale/process-mining/internal/service"
)

type GraphHandler struct {
	graphService *service.GraphService
}

func NewGraphHandler(graphService *service.GraphService) *GraphHandler {
	return &GraphHandler{graphService: graphService}
}

func (h *GraphHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	log.Println("Начало обработки запроса на загрузку файла")

	if r.Method != http.MethodPost {
		log.Println("Метод не поддерживается")
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Ограничение размера тела запроса до 3 ГБ
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024*1024)

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("Ошибка получения файла: %v", err)
		http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpManager := infrastructure.NewTMPFileManager()
	tmpManager.DeleteTempFile()

	tempFile, err := tmpManager.CreateTempFile("uploaded-", "csv")
	if err != nil {
		log.Printf("Ошибка создания временного файла: %v", err)
		http.Error(w, "Ошибка создания временного файла", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Буферизированное копирование файла
	buf := make([]byte, 1024*1024)
	var totalBytes int64
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, writeErr := tempFile.Write(buf[:n]); writeErr != nil {
				log.Printf("Ошибка записи во временный файл: %v", writeErr)
				http.Error(w, "Ошибка записи во временный файл", http.StatusInternalServerError)
				return
			}
			totalBytes += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Ошибка чтения файла: %v", err)
			http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Файл успешно загружен. Размер: %.2f МБ", float64(totalBytes)/1024/1024)
	log.Printf("Путь к файлу: %s", tempFile.Name())

	// Построение графа
	err = h.graphService.BuildGraphFromCSV(tempFile.Name())
	if err != nil {
		log.Printf("Ошибка построения графа: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка построения графа: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Файл успешно загружен и граф построен"))
	log.Println("Обработка завершена успешно")
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

func (h *GraphHandler) ClearGraph(w http.ResponseWriter, r *http.Request) {

	tmpManager := infrastructure.NewTMPFileManager()
	tmpManager.DeleteTempFile()

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	h.graphService.ClearGraph()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Граф успешно очищен"))
}

func (h *GraphHandler) ListDatasets(w http.ResponseWriter, r *http.Request) {
	// проверим, что метод GET
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	const tmpDir = "./tmp"

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		http.Error(w, "Ошибка чтения временной директории", http.StatusInternalServerError)
		return
	}

	var datasets []map[string]interface{}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".csv" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		modTime := info.ModTime().UTC()

		datasets = append(datasets, map[string]interface{}{
			"id":         info.Name(),
			"name":       info.Name(),
			"createdAt":  modTime.Format(time.RFC3339),
			"uploadedAt": modTime.Format(time.RFC3339),
			"status":     "ready",
			"progress":   100,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(datasets)
}
