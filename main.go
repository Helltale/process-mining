package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"`
	Count int    `json:"count"`
}

type Edge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
	Style string `json:"style,omitempty"`
	Count int    `json:"count"`
}

type Graph struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

type Event struct {
	ID        string
	Timestamp time.Time
	Desc      string
}

var (
	autoParseDate = flag.Bool("autoparse", false, "Попытаться автоматически распознать формат даты")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Пример использования:
  --file=events.csv --output=graph.html [--autoparse]

CSV файл должен содержать строки вида:
  id,timestamp,desc
  123,2023-10-20T12:00:00Z,StartEvent
  123,2023-10-20T12:05:00Z,EndEvent

Флаги:
`)
		flag.PrintDefaults()
	}

	fmt.Println("✅ Запуск скрипта на GO")
	inFile := flag.String("file", "", "Путь до CSV-файла")
	outFile := flag.String("output", "graph.html", "Путь для HTML-выхода")
	flag.Parse()

	if *inFile == "" {
		fmt.Println("❌ Укажите путь до файла через --file")
		os.Exit(1)
	}

	fileInfo, err := os.Stat(*inFile)
	if err != nil {
		fmt.Println("❌ Ошибка доступа к файлу:", err)
		os.Exit(1)
	}

	totalLines, err := countLines(*inFile)
	if err != nil {
		fmt.Println("❌ Ошибка подсчета строк:", err)
		os.Exit(1)
	}
	fmt.Printf("📦 Размер файла: %.2f MB\tСтрок в файле: %d\n", float64(fileInfo.Size())/1024/1024, totalLines)

	start := time.Now()
	progressChan := make(chan int)
	go logProgress(start, totalLines, progressChan)

	events, err := readCSV(*inFile, progressChan)
	if err != nil {
		fmt.Println("❌ Ошибка чтения CSV:", err)
		os.Exit(1)
	}

	graph := buildGraph(events)

	data := map[string]interface{}{
		"nodes": graph.Nodes,
		"edges": graph.Edges,
	}

	jsonData, _ := json.Marshal(data)
	html := generateHTML(string(jsonData))

	if err := os.WriteFile(*outFile, []byte(html), 0644); err != nil {
		fmt.Println("❌ Ошибка записи HTML:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Граф сохранён в:", *outFile)
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := csv.NewReader(f)
	records, err := scanner.ReadAll()
	if err != nil {
		return 0, err
	}
	return len(records), nil
}

func logProgress(start time.Time, total int, ch <-chan int) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var processed int
	var mem runtime.MemStats

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(start).Seconds()
			runtime.ReadMemStats(&mem)
			fmt.Printf("[Думаю...]\t%.0fs\t%.2f%%\tCPU: %d\tRAM: %.2f MB\n", elapsed, float64(processed)/float64(total)*100, runtime.NumCPU(), float64(mem.Alloc)/1024/1024)
		case count, ok := <-ch:
			if !ok {
				return
			}
			processed = count
		}
	}
}

func tryParseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	formats := []string{
		time.RFC3339,     // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano, // 2006-01-02T15:04:05.999999999Z07:00

		"2006-01-02 15:04:05.000", // 2023-12-31 23:59:59.123
		"2006/01/02 15:04:05.000", // 2023/12/31 23:59:59.123
		"02-01-2006 15:04:05.000", // 31-12-2023 23:59:59.123
		"02/01/2006 15:04:05.000", // 31/12/2023 23:59:59.123
		"02.01.2006 15:04:05.000", // 31.12.2023 23:59:59.123

		"2006-01-02 15:04:05", // без миллисекунд
		"2006/01/02 15:04:05",
		"02-01-2006 15:04:05",
		"02/01/2006 15:04:05",
		"02.01.2006 15:04:05",

		"2006-01-02",
		"02.01.2006",
		"02/01/2006",
		"02-01-2006",
		"2006/01/02",

		"January 2, 2006",
		"2 Jan 2006",
		"2 January 2006",

		"02 Jan 2006 15:04",
		"02 Jan 2006 15:04:05",
		"Mon Jan 2 15:04:05 2006",
		"Mon Jan 2 15:04:05 MST 2006",

		"20060102",
		"20060102T150405",
		"20060102T150405.000", // 20231231T235959.123
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("не удалось распознать дату: %s", s)
}

func readCSV(path string, ch chan<- int) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, errors.New("файл пустой или без заголовков")
	}

	var events []Event
	for i, rec := range records[1:] {
		if len(rec) != 3 {
			return nil, fmt.Errorf("строка %d: ожидается 3 столбца", i+2)
		}

		s := strings.TrimSpace(rec[1])
		t, err := time.Parse(time.RFC3339, s)
		if err != nil && *autoParseDate {
			t, err = tryParseDate(s)
		}
		if err != nil {
			return nil, fmt.Errorf("строка %d: неверный формат даты: %v", i+2, err)
		}

		events = append(events, Event{
			ID:        strings.TrimSpace(rec[0]),
			Timestamp: t,
			Desc:      strings.TrimSpace(rec[2]),
		})

		ch <- i + 1
	}
	close(ch)

	sort.Slice(events, func(i, j int) bool {
		if events[i].ID == events[j].ID {
			return events[i].Timestamp.Before(events[j].Timestamp)
		}
		return events[i].ID < events[j].ID
	})

	return events, nil
}

func buildGraph(events []Event) *Graph {
	nodeMap := map[string]*Node{}
	edgeMap := map[string]*Edge{}
	var graph Graph

	addNode := func(desc string, color string) {
		if _, ok := nodeMap[desc]; !ok {
			nodeMap[desc] = &Node{ID: desc, Label: desc, Color: color, Count: 1}
		} else {
			nodeMap[desc].Count++
		}
	}

	addEdge := func(from, to, style string) {
		key := from + "->" + to
		if e, ok := edgeMap[key]; ok {
			e.Count++
		} else {
			edgeMap[key] = &Edge{
				From:  from,
				To:    to,
				Label: "",
				Style: style,
				Count: 1,
			}
		}
	}

	lastEvent := map[string]*Event{}

	for _, event := range events {
		addNode(event.Desc, "blue")
		if _, ok := lastEvent[event.ID]; !ok {
			addNode("start", "green")
			addEdge("start", event.Desc, "dashed")
		} else {
			prev := lastEvent[event.ID]
			addEdge(prev.Desc, event.Desc, "")
		}
		lastEvent[event.ID] = &event
	}

	for _, prev := range lastEvent {
		addNode("end", "red")
		addEdge(prev.Desc, "end", "dashed")
	}

	for _, node := range nodeMap {
		graph.Nodes = append(graph.Nodes, node)
	}
	for _, edge := range edgeMap {
		edge.Label = fmt.Sprintf("%d", edge.Count)
		graph.Edges = append(graph.Edges, edge)
	}

	return &graph
}

func generateHTML(graphJSON string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Process Graph</title>
  <style>
    html, body, #cy { width: 100%%; height: 100%%; margin: 0; padding: 0; }
  </style>
</head>
<body>
<div id="cy"></div>
<script>
  %s
</script>
<script>
  const graph = %s;
  const cy = cytoscape({
    container: document.getElementById('cy'),
    elements: [
      ...graph.nodes.map(n => ({ data: n })),
      ...graph.edges.map(e => ({
        data: {
          source: e.from,
          target: e.to,
          label: e.label,
        },
        classes: e.style === 'dashed' ? 'dashed' : ''
      }))
    ],
    style: [
      { selector: 'node', style: {
          'background-color': 'data(color)',
          'label': 'data(label)',
          'text-valign': 'center',
          'text-halign': 'center',
          'shape': 'round-rectangle',
          'font-size': 10,
          'padding': 10
      }},
      { selector: 'edge', style: {
          'width': 2,
          'label': 'data(label)',
          'curve-style': 'bezier',
          'target-arrow-shape': 'triangle',
          'line-color': '#999',
          'target-arrow-color': '#999',
          'font-size': 8,
          'text-rotation': 'autorotate'
      }},
      { selector: '.dashed', style: { 'line-style': 'dashed' } }
    ],
    layout: {
      name: 'grid',
      fit: true,
      padding: 30
    }
  });
</script>
</body>
</html>`, cytoscapeJS(), graphJSON)
}

func cytoscapeJS() string {
	data, err := os.ReadFile("cytoscape.min.js")
	if err != nil {
		fmt.Println("❌ Не удалось загрузить cytoscape.min.js:", err)
		os.Exit(1)
	}
	return string(data)
}
