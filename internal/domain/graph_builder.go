package domain

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Helltale/process-mining/internal/infrastructure"
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
	Style       string  `json:"style"` // стиль линии (solid, dashed и т.д.)
}

type Event struct {
	ID        string
	SessionID string
	Timestamp time.Time
	Desc      string
}

type Session struct {
	Events []*Event
}

type GraphBuilder struct {
	graph         *Graph
	nodeMap       map[string]*Node
	edgeMap       map[string]*Edge
	csvReader     *infrastructure.CSVReader
	lastEvent     *Event
	lastSessionID string
	isFirstEvent  bool

	totalRecords int
}

func NewGraphBuilder(csvReader *infrastructure.CSVReader) *GraphBuilder {
	return &GraphBuilder{
		graph:        &Graph{},
		nodeMap:      make(map[string]*Node),
		edgeMap:      make(map[string]*Edge),
		csvReader:    csvReader,
		isFirstEvent: true,
	}
}

func (gb *GraphBuilder) BuildGraph(filePath string) error {
	err := gb.csvReader.ReadAndProcess(filePath, func(record []string) error {
		// Проверяем количество полей
		if len(record) != 3 {
			return fmt.Errorf("некорректная строка: %v", record)
		}

		// Парсим временную метку
		timestamp, err := time.Parse(time.RFC3339, record[1])
		if err != nil {
			return fmt.Errorf("ошибка парсинга времени '%s': %v", record[1], err)
		}

		// Создаем событие
		event := &Event{
			ID:        record[0],
			SessionID: record[0],
			Timestamp: timestamp,
			Desc:      record[2],
		}

		// Обрабатываем событие
		gb.ProcessEvent(event)
		return nil
	})

	if err != nil {
		return err
	}

	gb.finalizeGraph()
	return nil
}

func (gb *GraphBuilder) GetGraph() *Graph {
	return gb.graph
}

func (gb *GraphBuilder) ClearGraph() {
	gb.graph = &Graph{}
	gb.nodeMap = make(map[string]*Node)
	gb.edgeMap = make(map[string]*Edge)
}

func (gb *GraphBuilder) ProcessEvent(event *Event) {
	node := gb.getNode(event.Desc)
	node.Count++
	node.Total++

	if gb.isFirstEvent {
		gb.lastSessionID = event.SessionID
		gb.lastEvent = event
		gb.isFirstEvent = false

		// Добавляем связь от "start"
		gb.addStartEdge(event)
		return
	}

	// Новая сессия?
	if event.SessionID != gb.lastSessionID {
		// Завершаем старую сессию → "end"
		gb.addEndEdge(gb.lastEvent)

		// новая сессия → связь от "start"
		gb.addStartEdge(event)
	}

	// связь между событиями
	if gb.lastSessionID == event.SessionID {
		duration := event.Timestamp.Sub(gb.lastEvent.Timestamp).Seconds()
		key := gb.lastEvent.Desc + "_" + event.Desc
		edge := gb.getEdge(key, gb.lastEvent.Desc, event.Desc)
		edge.Count++
		edge.AvgDuration = (edge.AvgDuration*float64(edge.Count-1) + duration) / float64(edge.Count)
	}

	gb.lastEvent = event
	gb.lastSessionID = event.SessionID
}

func (gb *GraphBuilder) addStartEdge(event *Event) {
	startKey := "start_" + event.Desc
	edge := gb.getEdge(startKey, "start", event.Desc)
	edge.Count++
	edge.Style = "dashed"
}

func (gb *GraphBuilder) addEndEdge(event *Event) {
	endKey := event.Desc + "_end"
	edge := gb.getEdge(endKey, event.Desc, "end")
	edge.Count++
	edge.Style = "dashed"
}

func (gb *GraphBuilder) finalizeGraph() {
	// Добавим специальные узлы
	startNode := &Node{
		ID:    "start",
		Label: "Начало процесса",
		Count: gb.totalRecords,
		Total: gb.totalRecords,
		Color: "green",
	}
	gb.graph.Nodes = append(gb.graph.Nodes, startNode)

	endNode := &Node{
		ID:    "end",
		Label: "Конец процесса",
		Count: gb.totalRecords,
		Total: gb.totalRecords,
		Color: "red",
	}
	gb.graph.Nodes = append(gb.graph.Nodes, endNode)

	// обычные узлы
	for _, node := range gb.nodeMap {
		gb.graph.Nodes = append(gb.graph.Nodes, node)
	}

	// ребра
	for _, edge := range gb.edgeMap {
		edge.Label = fmt.Sprintf("%d\n%.2f sec avg", edge.Count, edge.AvgDuration)
		gb.graph.Edges = append(gb.graph.Edges, edge)
	}
}

func (gb *GraphBuilder) processSession(session *Session) {
	events := session.Events
	if len(events) == 0 {
		return
	}

	for _, event := range events {
		node := gb.getNode(event.Desc)
		node.Count++
		node.Total++
	}

	if len(events) > 1 {
		prevEvent := events[0]
		for i := 1; i < len(events); i++ {
			currEvent := events[i]

			duration := currEvent.Timestamp.Sub(prevEvent.Timestamp).Seconds()
			key := prevEvent.Desc + "_" + currEvent.Desc

			edge := gb.getEdge(key, prevEvent.Desc, currEvent.Desc)
			edge.Count++
			edge.AvgDuration = (edge.AvgDuration*float64(edge.Count-1) + duration) / float64(edge.Count)

			prevEvent = currEvent
		}
	}
}

func (gb *GraphBuilder) getNode(desc string) *Node {
	node := gb.nodeMap[desc]
	if node == nil {
		node = &Node{
			ID:    desc,
			Label: desc,
			Color: "blue", // Устанавливаем значение по умолчанию
		}
		gb.nodeMap[desc] = node
	}
	return node
}

func (gb *GraphBuilder) getEdge(key, from, to string) *Edge {
	edge := gb.edgeMap[key]
	if edge == nil {
		edge = &Edge{
			From: from,
			To:   to,
		}
		gb.edgeMap[key] = edge
	}
	return edge
}

// TODO: TMP
func (gb *GraphBuilder) BuildGraphSequential(filePath string, processFunc func([]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	totalLines, err := countLines(filePath)
	if err != nil {
		return err
	}
	gb.totalRecords = totalLines

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		header := scanner.Text()
		log.Printf("Пропущен заголовок: %s", header)
	}

	progress := NewProgressLogger(totalLines, "BuildGraphSequential", true)
	defer progress.Done()

	for scanner.Scan() {
		line := scanner.Text()
		record := strings.Split(line, ",")
		if err := processFunc(record); err != nil {
			return err
		}
		progress.Inc()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	gb.finalizeGraph()
	return nil
}

// TODO: TMP2
func (gb *GraphBuilder) BuildGraphSequential2(filePath string, processFunc func([]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Подсчет общего количества строк в файле
	totalLines, err := countLines(filePath)
	if err != nil {
		return err
	}
	gb.totalRecords = totalLines

	scanner := bufio.NewScanner(file)

	// Пропускаем заголовок
	if scanner.Scan() {
		header := scanner.Text()
		log.Printf("Пропущен заголовок: %s", header)
	}

	progress := NewProgressLogger(totalLines, "BuildGraphSequential2", true)
	defer progress.Done()

	for scanner.Scan() {
		line := scanner.Text()
		record := strings.Split(line, ",")
		if err := processFunc(record); err != nil {
			return err
		}
		progress.Inc()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	gb.finalizeGraph()
	return nil
}

// Вспомогательная функция для подсчета строк в файле
func countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lineCount, nil
}

// TODO: TMP
func (gb *GraphBuilder) BuildGraphConcurrent(filePath string, processFunc func([]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	totalLines, err := countLines(filePath)
	if err != nil {
		return err
	}
	gb.totalRecords = totalLines

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		header := scanner.Text()
		log.Printf("Пропущен заголовок: %s", header)
	}

	lines := make(chan []string, 100)
	errChan := make(chan error, 1)
	progress := NewProgressLogger(totalLines, "BuildGraphConcurrent", true)
	defer progress.Done()

	go func() {
		defer close(lines)
		batch := []string{}
		for scanner.Scan() {
			line := scanner.Text()
			batch = append(batch, line)
			if len(batch) >= 100 {
				lines <- batch
				batch = []string{}
			}
		}
		if len(batch) > 0 {
			lines <- batch
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	numWorkers := 4
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for batch := range lines {
				for _, line := range batch {
					record := strings.Split(line, ",")
					if err := processFunc(record); err != nil {
						errChan <- err
						return
					}
					progress.Inc()
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return err
	}

	gb.finalizeGraph()
	return nil
}

// TODO: TMP BIG FILES Sequential
func (gb *GraphBuilder) BuildGraphSequentialLargeFile(filePath string, processFunc func([]string) error) error {
	gb.ClearGraph() // ✅ очищаем граф перед новой загрузкой

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	totalLines, err := countLines(filePath)
	if err != nil {
		return fmt.Errorf("ошибка подсчета строк: %v", err)
	}
	gb.totalRecords = totalLines

	// fileInfo, err := file.Stat()
	// if err != nil {
	// 	return fmt.Errorf("ошибка получения информации о файле: %v", err)
	// }
	// totalSize := fileInfo.Size()

	const blockSize = 1024 * 1024
	buffer := make([]byte, blockSize)

	var carryOver string
	var bytesRead int64

	progress := NewProgressLogger(totalLines, "BuildGraphSequentialLargeFile", true)
	defer progress.Done()

	headerSkipped := false

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ошибка чтения файла: %v", err)
		}

		bytesRead += int64(n)
		data := carryOver + string(buffer[:n])
		lines := strings.Split(data, "\n")

		if len(lines) == 0 {
			continue
		}

		// Последняя строка, возможно, обрезана
		carryOver = lines[len(lines)-1]
		lines = lines[:len(lines)-1]

		for _, line := range lines {
			if line == "" {
				continue
			}

			if !headerSkipped {
				log.Printf("Пропущен заголовок: %s", line)
				headerSkipped = true
				continue
			}

			record := strings.Split(line, ",")
			if err := processFunc(record); err != nil {
				return err
			}

			progress.Inc()
		}
	}

	// Обрабатываем оставшуюся строку
	if carryOver != "" {
		if !headerSkipped {
			log.Printf("Пропущен заголовок: %s", carryOver)
		} else {
			record := strings.Split(carryOver, ",")
			if err := processFunc(record); err != nil {
				return err
			}
			progress.Inc()
		}
	}

	gb.finalizeGraph()
	return nil
}

func (gb *GraphBuilder) ProcessFileInChunks(filePath string, chunkSize int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	chunk := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		chunk = append(chunk, line)

		// Обрабатываем чанк, если достигнут нужный размер
		if len(chunk) >= chunkSize {
			if err := gb.processChunk(chunk); err != nil {
				return err
			}
			chunk = []string{} // Очищаем чанк
			runtime.GC()       // Вызываем сборщик мусора
		}
	}

	// Обрабатываем оставшиеся строки
	if len(chunk) > 0 {
		if err := gb.processChunk(chunk); err != nil {
			return err
		}
	}

	return nil
}

func (gb *GraphBuilder) processChunk(chunk []string) error {
	for _, line := range chunk {
		record := strings.Split(line, ",")
		event := &Event{
			ID:        record[0],
			SessionID: record[0],
			Timestamp: parseTimestamp(record[1]),
			Desc:      record[2],
		}
		gb.ProcessEvent(event)
	}
	return nil
}

func parseTimestamp(timestampStr string) time.Time {
	timestamp, _ := time.Parse(time.RFC3339, timestampStr)
	return timestamp
}
