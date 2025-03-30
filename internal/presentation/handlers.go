package presentation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024*1024)

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка получения файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpManager := infrastructure.NewTMPFileManager()
	tempFile, err := tmpManager.CreateTempFile("uploaded-", "csv")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Ошибка записи файла", http.StatusInternalServerError)
		return
	}

	if err := validateCSVFile(tempFile.Name()); err != nil {
		http.Error(w, fmt.Sprintf("Файл не прошел валидацию: %v", err), http.StatusBadRequest)
		return
	}

	resp := map[string]interface{}{
		"file":   filepath.Base(tempFile.Name()),
		"status": "ready",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Println("Файл сохранён и свалидирован, готов к построению графа.")
}

func (h *GraphHandler) BuildGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "Не указано имя файла", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join("./tmp", file)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	if err := h.graphService.BuildGraphFromCSV(fullPath); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка построения графа: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Граф построен"))
}

func (h *GraphHandler) ServeGraphData(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "file param is required", http.StatusBadRequest)
		return
	}

	filePath := fmt.Sprintf("./tmp/%s", file)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	if err := h.graphService.BuildGraphFromCSV(filePath); err != nil {
		http.Error(w, fmt.Sprintf("ошибка построения графа: %v", err), http.StatusInternalServerError)
		return
	}

	graphData, err := h.graphService.GetGraphData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type CytoscapeNode struct {
		ID    string `json:"id"`
		Label string `json:"label"`
		Count int    `json:"count"`
		Total int    `json:"total"`
		Color string `json:"color"`
	}
	type CytoscapeEdge struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Label  string `json:"label"`
		Count  int    `json:"count"`
		Style  string `json:"style"`
	}

	cytoscapeData := struct {
		Nodes []map[string]CytoscapeNode `json:"nodes"`
		Edges []map[string]CytoscapeEdge `json:"edges"`
	}{
		Nodes: make([]map[string]CytoscapeNode, len(graphData.Nodes)),
		Edges: make([]map[string]CytoscapeEdge, len(graphData.Edges)),
	}

	for i, node := range graphData.Nodes {
		cytoscapeData.Nodes[i] = map[string]CytoscapeNode{"data": {
			ID:    strings.TrimSpace(node.ID),
			Label: strings.TrimSpace(node.Label),
			Count: node.Count,
			Total: node.Total,
			Color: node.Color,
		}}
	}

	for i, edge := range graphData.Edges {
		cytoscapeData.Edges[i] = map[string]CytoscapeEdge{"data": {
			Source: strings.TrimSpace(edge.From),
			Target: strings.TrimSpace(edge.To),
			Label:  edge.Label,
			Count:  edge.Count,
			Style:  edge.Style,
		}}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cytoscapeData)
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

func validateCSVFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		log.Printf("Заголовок: %s", scanner.Text())
	}

	lineNum := 1
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		fields := strings.Split(line, ",")

		if len(fields) != 3 {
			return fmt.Errorf("строка %d: ожидалось 3 поля, получено %d", lineNum, len(fields))
		}

		if _, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[1])); err != nil {
			return fmt.Errorf("строка %d: некорректная дата %q", lineNum, fields[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	return nil
}
