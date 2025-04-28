package main

import (
	"encoding/csv"
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
	ID    string
	Count int
}

type Edge struct {
	From  string
	To    string
	Count int
}

type Graph struct {
	Nodes []*Node
	Edges []*Edge
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
  --file=events.csv --output=graph.graphml [--autoparse]

Флаги:
`)
		flag.PrintDefaults()
	}

	fmt.Println("✅ Запуск скрипта на GO")
	inFile := flag.String("file", "", "Путь до CSV-файла")
	outFile := flag.String("output", "graph.graphml", "Путь для GraphML-выхода")
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

	graphml := generateGraphML(graph)

	if err := os.WriteFile(*outFile, []byte(graphml), 0644); err != nil {
		fmt.Println("❌ Ошибка записи GraphML:", err)
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
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05.000",
		"2006/01/02 15:04:05.000",
		"02-01-2006 15:04:05.000",
		"02/01/2006 15:04:05.000",
		"02.01.2006 15:04:05.000",
		"2006-01-02 15:04:05",
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
		"20060102T150405.000",
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

	addNode := func(id string) {
		if _, ok := nodeMap[id]; !ok {
			nodeMap[id] = &Node{ID: id, Count: 1}
		} else {
			nodeMap[id].Count++
		}
	}

	addEdge := func(from, to string) {
		key := from + "->" + to
		if e, ok := edgeMap[key]; ok {
			e.Count++
		} else {
			edgeMap[key] = &Edge{From: from, To: to, Count: 1}
		}
	}

	lastEvent := map[string]*Event{}

	for _, event := range events {
		addNode(event.Desc)
		if _, ok := lastEvent[event.ID]; !ok {
			addNode("start")
			addEdge("start", event.Desc)
		} else {
			prev := lastEvent[event.ID]
			addEdge(prev.Desc, event.Desc)
		}
		lastEvent[event.ID] = &event
	}

	for _, prev := range lastEvent {
		addNode("end")
		addEdge(prev.Desc, "end")
	}

	for _, node := range nodeMap {
		graph.Nodes = append(graph.Nodes, node)
	}
	for _, edge := range edgeMap {
		graph.Edges = append(graph.Edges, edge)
	}

	return &graph
}

func generateGraphML(graph *Graph) string {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<graphml xmlns="http://graphml.graphdrawing.org/xmlns" ` +
		`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" ` +
		`xsi:schemaLocation="http://graphml.graphdrawing.org/xmlns ` +
		`http://graphml.graphdrawing.org/xmlns/1.0/graphml.xsd">` + "\n")
	sb.WriteString(`<graph edgedefault="directed">` + "\n")

	sort.Slice(graph.Nodes, func(i, j int) bool {
		return graph.Nodes[i].ID < graph.Nodes[j].ID
	})
	sort.Slice(graph.Edges, func(i, j int) bool {
		if graph.Edges[i].From == graph.Edges[j].From {
			return graph.Edges[i].To < graph.Edges[j].To
		}
		return graph.Edges[i].From < graph.Edges[j].From
	})

	for _, node := range graph.Nodes {
		sb.WriteString(fmt.Sprintf(`  <node id="%s" />`+"\n", escapeXML(node.ID)))
	}

	for _, edge := range graph.Edges {
		sb.WriteString(fmt.Sprintf(`  <edge source="%s" target="%s" />`+"\n", escapeXML(edge.From), escapeXML(edge.To)))
	}

	sb.WriteString(`</graph>` + "\n")
	sb.WriteString(`</graphml>` + "\n")

	return sb.String()
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
