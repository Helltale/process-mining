package domain

import (
	"bufio"
	"fmt"
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
	graph      *Graph
	nodeMap    map[string]*Node
	edgeMap    map[string]*Edge
	sessionMap sync.Map
	csvReader  *infrastructure.CSVReader
}

func NewGraphBuilder(csvReader *infrastructure.CSVReader) *GraphBuilder {
	return &GraphBuilder{
		graph:      &Graph{},
		nodeMap:    make(map[string]*Node),
		edgeMap:    make(map[string]*Edge),
		sessionMap: sync.Map{},
		csvReader:  csvReader,
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
	gb.sessionMap = sync.Map{}
}

func (gb *GraphBuilder) ProcessEvent(event *Event) {
	// Проверяем, существует ли сессия
	if _, ok := gb.sessionMap.Load(event.SessionID); !ok {
		gb.sessionMap.Store(event.SessionID, &Session{})
	}

	// Получаем сессию
	session, _ := gb.sessionMap.Load(event.SessionID)
	session.(*Session).Events = append(session.(*Session).Events, event)
}

func (gb *GraphBuilder) finalizeGraph() {
	// Перебираем все сессии
	gb.sessionMap.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		gb.processSession(session)
		return true // Продолжаем перебор
	})

	// Добавляем узлы
	for _, node := range gb.nodeMap {
		gb.graph.Nodes = append(gb.graph.Nodes, node)
	}

	// Добавляем ребра
	for _, edge := range gb.edgeMap {
		edge.Label = fmt.Sprintf("%d\n%.2f sec avg", edge.Count, edge.AvgDuration)
		gb.graph.Edges = append(gb.graph.Edges, edge)
	}

	// Добавляем специальные узлы "Начало" и "Конец"
	startNode := &Node{
		ID:    "start",
		Label: "Начало процесса",
		Count: gb.getSessionCount(),
		Total: gb.getSessionCount(),
		Color: "green",
	}
	gb.graph.Nodes = append(gb.graph.Nodes, startNode)

	endNode := &Node{
		ID:    "end",
		Label: "Конец",
		Count: gb.getSessionCount(),
		Total: gb.getSessionCount(),
		Color: "red",
	}
	gb.graph.Nodes = append(gb.graph.Nodes, endNode)

	// Добавляем связи между "Начало" -> первый узел и последний узел -> "Конец"
	gb.sessionMap.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		events := session.Events
		if len(events) == 0 {
			return true
		}

		// Связь "Начало" -> первый узел
		firstEvent := events[0]
		startKey := "start_" + firstEvent.Desc
		startEdge := gb.getEdge(startKey, "start", firstEvent.Desc)
		startEdge.Count++
		startEdge.Style = "dashed"
		if startEdge.Count == 1 {
			gb.graph.Edges = append(gb.graph.Edges, startEdge)
		}

		// Связь последний узел -> "Конец"
		lastEvent := events[len(events)-1]
		endKey := lastEvent.Desc + "_end"
		endEdge := gb.getEdge(endKey, lastEvent.Desc, "end")
		endEdge.Count++
		endEdge.Style = "dashed"
		if endEdge.Count == 1 {
			gb.graph.Edges = append(gb.graph.Edges, endEdge)
		}

		return true
	})
}

// Метод для подсчета количества сессий
func (gb *GraphBuilder) getSessionCount() int {
	count := 0
	gb.sessionMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
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

	scanner := bufio.NewScanner(file)

	// Пропускаем заголовок
	if scanner.Scan() {
		header := scanner.Text()
		log.Printf("Пропущен заголовок: %s", header)
	}

	for scanner.Scan() {
		line := scanner.Text()
		record := strings.Split(line, ",")
		if err := processFunc(record); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	gb.finalizeGraph()
	return nil
}

// TODO: TMP
func (gb *GraphBuilder) BuildGraphConcurrent(filePath string, processFunc func([]string) error) error {
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
	lines := make(chan []string, 1000) // Увеличенный буфер
	errChan := make(chan error, 1)     // Канал для ошибок

	// Горутина для чтения строк из файла
	go func() {
		defer close(lines)
		batch := []string{}
		for scanner.Scan() {
			line := scanner.Text()
			batch = append(batch, line)
			if len(batch) >= 100 { // Размер батча
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

	// Горутины для обработки строк
	//TODO: протестировать больше воркеров
	numWorkers := 20 // Ограниченное количество горутин
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
				}
				// Добавляем паузу после обработки батча
				time.Sleep(1 * time.Millisecond) // Пауза 1 мс
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
