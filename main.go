package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Graph struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Count int    `json:"count"`
	Total int    `json:"total"`
	Color string `json:"color"`
}

type Edge struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Count       int     `json:"count"`
	AvgDuration float64 `json:"-"`
	Label       string  `json:"label"`
}

type Event struct {
	ID        string // unique ID event
	SessionID string
	Timestamp time.Time
	Desc      string
}

type Session struct {
	Events []*Event
}

var graph *Graph // for graph

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static"))) // Статические файлы
	http.HandleFunc("/upload", uploadFile)                  // Загрузка CSV
	http.HandleFunc("/graph", serveGraphData)               // Получение данных графа

	// Настройка сервера с увеличенными таймаутами
	srv := &http.Server{
		Addr:         ":8085",
		WriteTimeout: 15 * time.Minute, // Увеличенный таймаут для записи
		ReadTimeout:  15 * time.Minute, // Увеличенный таймаут для чтения
	}
	fmt.Printf("Сервер запущен на порту %v", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

// Обработчик загрузки файла
func uploadFile(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("Файл успешно загружен: %s", tempFile.Name())

	// Строим граф из загруженного файла
	graph, err = buildGraph(tempFile.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка построения графа: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Файл успешно загружен и граф построен"))
}

// Обработчик получения данных графа
func serveGraphData(w http.ResponseWriter, r *http.Request) {
	if graph == nil {
		http.Error(w, "Граф еще не построен. Загрузите CSV-файл.", http.StatusNotFound)
		return
	}

	cytoscapeData := struct {
		Nodes []map[string]*Node `json:"nodes"`
		Edges []map[string]*Edge `json:"edges"`
	}{
		Nodes: make([]map[string]*Node, len(graph.Nodes)),
		Edges: make([]map[string]*Edge, len(graph.Edges)),
	}

	for i, node := range graph.Nodes {
		cytoscapeData.Nodes[i] = map[string]*Node{"data": node}
	}

	for i, edge := range graph.Edges {
		cytoscapeData.Edges[i] = map[string]*Edge{"data": edge}
	}

	jsonData, err := json.Marshal(cytoscapeData)
	if err != nil {
		http.Error(w, "Ошибка сериализации", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// Функция для построения графа
func buildGraph(filePath string) (*Graph, error) {
	graph := &Graph{
		Nodes: make([]*Node, 0),
		Edges: make([]*Edge, 0),
	}

	nodeMap := make(map[string]*Node)
	edgeMap := make(map[string]*Edge)
	sessionMap := make(map[string]*Session)

	var totalRecords int
	var validSessions int

	err := readAndProcessCSV(filePath, func(record []string) error {
		totalRecords++
		if len(record) < 3 {
			log.Printf("Пропущена некорректная запись: %v", record)
			return nil
		}

		timestamp, err := time.Parse(time.RFC3339, record[1])
		if err != nil {
			log.Printf("Ошибка парсинга времени: %v для записи %v", err, record)
			return nil
		}

		// Создание события
		event := &Event{
			ID:        generateEventID(record[0], 0), // Начальное значение индекса
			SessionID: record[0],
			Timestamp: timestamp,
			Desc:      record[2],
		}

		sessionID := event.SessionID
		session := sessionMap[sessionID]

		// Если сессия ещё не существует, создаём её
		if session == nil {
			validSessions++
			session = &Session{}
			sessionMap[sessionID] = session
		}

		// Добавляем событие в сессию
		session.Events = append(session.Events, event)

		// Обновляем ID события (с учётом текущего количества событий)
		event.ID = generateEventID(record[0], len(session.Events)-1)

		return nil
	})

	log.Printf("Обработано записей: %d, уникальных сессий: %d", totalRecords, validSessions)

	if err != nil {
		return nil, err
	}

	for _, session := range sessionMap {
		processSession(session, nodeMap, edgeMap)
	}

	// Преобразование карт в слайсы
	for _, node := range nodeMap {
		graph.Nodes = append(graph.Nodes, node)
	}

	for _, edge := range edgeMap {
		edge.Label = fmt.Sprintf("%d\n%.2f sec avg", edge.Count, edge.AvgDuration)
		graph.Edges = append(graph.Edges, edge)
	}

	addStartEndNodes(graph, sessionMap, nodeMap)

	log.Printf("Сформирован граф: узлов=%d, ребер=%d", len(graph.Nodes), len(graph.Edges))

	return graph, nil
}

func processSession(session *Session, nodeMap map[string]*Node, edgeMap map[string]*Edge) {
	events := session.Events
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		node := getNode(nodeMap, event.Desc)
		node.Count++
		node.Total++
	}

	if len(events) > 1 {
		prevEvent := events[0]
		for i := 1; i < len(events); i++ {
			currEvent := events[i]

			duration := currEvent.Timestamp.Sub(prevEvent.Timestamp).Seconds()
			key := prevEvent.Desc + "_" + currEvent.Desc

			edge := getEdge(edgeMap, key, prevEvent.Desc, currEvent.Desc)
			edge.Count++
			edge.AvgDuration = (edge.AvgDuration*float64(edge.Count-1) + duration) / float64(edge.Count)

			prevEvent = currEvent
		}
	}
}

func readAndProcessCSV(filePath string, processFunc func([]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read() // Пропуск заголовка
	if err != nil && err != io.EOF {
		return err
	}

	const maxRecords = 10_000_000 // Максимальное количество записей
	var totalRecords int

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if totalRecords >= maxRecords {
			log.Printf("Достигнуто ограничение на количество записей: %d", maxRecords)
			break
		}

		if err := processFunc(record); err != nil {
			return err
		}
		totalRecords++
	}

	return nil
}

func getNode(nodeMap map[string]*Node, desc string) *Node {
	node := nodeMap[desc]
	if node == nil {
		node = &Node{
			ID:    desc,
			Label: desc,
			Color: "blue",
		}
		nodeMap[desc] = node
	}
	return node
}

func getEdge(edgeMap map[string]*Edge, key, from, to string) *Edge {
	edge := edgeMap[key]
	if edge == nil {
		edge = &Edge{
			From: from,
			To:   to,
		}
		edgeMap[key] = edge
	}
	return edge
}

func addStartEndNodes(graph *Graph, sessionMap map[string]*Session, _ map[string]*Node) {
	totalSessions := len(sessionMap)
	startNode := &Node{
		ID:    "start",
		Label: "Начало процесса",
		Count: totalSessions,
		Total: totalSessions,
		Color: "green",
	}
	graph.Nodes = append(graph.Nodes, startNode)

	endNode := &Node{
		ID:    "end",
		Label: "Конец",
		Count: totalSessions,
		Total: totalSessions,
		Color: "red",
	}
	graph.Nodes = append(graph.Nodes, endNode)

	startEdgeMap := make(map[string]*Edge)
	endEdgeMap := make(map[string]*Edge)

	for _, session := range sessionMap {
		if len(session.Events) == 0 {
			continue
		}

		firstDesc := session.Events[0].Desc
		lastDesc := session.Events[len(session.Events)-1].Desc

		startKey := "start_" + firstDesc
		endKey := lastDesc + "_end"

		getEdge(startEdgeMap, startKey, "start", firstDesc).Count++
		getEdge(endEdgeMap, endKey, lastDesc, "end").Count++
	}

	for _, edge := range startEdgeMap {
		edge.Label = strconv.Itoa(edge.Count)
		graph.Edges = append(graph.Edges, edge)
	}

	for _, edge := range endEdgeMap {
		edge.Label = strconv.Itoa(edge.Count)
		graph.Edges = append(graph.Edges, edge)
	}
}

func generateEventID(sessionID string, index int) string {
	return fmt.Sprintf("%s_%d", sessionID, index)
}
