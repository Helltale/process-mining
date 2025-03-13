package domain

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
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
	sessionMap map[string]*Session
}

func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		graph:      &Graph{},
		nodeMap:    make(map[string]*Node),
		edgeMap:    make(map[string]*Edge),
		sessionMap: make(map[string]*Session),
	}
}

func (gb *GraphBuilder) BuildGraph(filePath string) error {
	err := gb.readAndProcessCSV(filePath, func(record []string) error {
		timestamp, err := time.Parse(time.RFC3339, record[1])
		if err != nil {
			return fmt.Errorf("ошибка парсинга времени: %v", err)
		}

		event := &Event{
			ID:        record[0],
			SessionID: record[0],
			Timestamp: timestamp,
			Desc:      record[2],
		}

		gb.processEvent(event)
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

func (gb *GraphBuilder) processEvent(event *Event) {
	session := gb.sessionMap[event.SessionID]
	if session == nil {
		session = &Session{}
		gb.sessionMap[event.SessionID] = session
	}
	session.Events = append(session.Events, event)
}

func (gb *GraphBuilder) finalizeGraph() {
	for _, session := range gb.sessionMap {
		gb.processSession(session)
	}

	for _, node := range gb.nodeMap {
		gb.graph.Nodes = append(gb.graph.Nodes, node)
	}

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
			Color: "blue",
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

func (gb *GraphBuilder) readAndProcessCSV(filePath string, processFunc func([]string) error) error {
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

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := processFunc(record); err != nil {
			return err
		}
	}

	return nil
}
