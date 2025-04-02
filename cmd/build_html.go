package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/infrastructure"
	"github.com/Helltale/process-mining/internal/service"
	"github.com/spf13/cobra"
)

var filePath string
var outPath string

var buildHTMLCmd = &cobra.Command{
	Use:   "build-html",
	Short: "Построение графа и генерация HTML",
	Long:  "Считывает CSV, строит граф и сохраняет интерактивную HTML-страницу с визуализацией.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if filePath == "" {
			return fmt.Errorf("нужно указать путь до CSV через --file")
		}

		if outPath == "" {
			outPath = "graph_output.html"
		}

		csvReader := infrastructure.NewCSVReader()
		graphBuilder := domain.NewGraphBuilder(csvReader)
		graphService := service.NewGraphService(graphBuilder)

		err := graphService.BuildGraphFromCSV(filePath)
		if err != nil {
			return fmt.Errorf("ошибка построения графа: %v", err)
		}

		graph := graphBuilder.GetGraph()

		data := map[string]interface{}{
			"nodes": graph.Nodes,
			"edges": graph.Edges,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("ошибка сериализации JSON: %v", err)
		}

		html := generateHTML(string(jsonData))
		err = os.WriteFile(outPath, []byte(html), 0644)
		if err != nil {
			return fmt.Errorf("ошибка записи HTML: %v", err)
		}

		fmt.Printf("✅ HTML-граф успешно сгенерирован: %s\n", outPath)
		return nil
	},
}

func init() {
	buildHTMLCmd.Flags().StringVarP(&filePath, "file", "f", "", "Путь до CSV-файла (обязателен)")
	buildHTMLCmd.Flags().StringVarP(&outPath, "output", "o", "graph_output.html", "Путь для сохранения HTML")
	rootCmd.AddCommand(buildHTMLCmd)
}

func generateHTML(graphJSON string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Process Graph</title>
  <style>
    html, body, #cy { width: 100%%; height: 100%%; margin: 0; padding: 0; }
  </style>
  <script src="./libs/dagre.min.js"></script>
  <script src="./libs/cytoscape.min.js"></script>
  <script src="./libs/cytoscape-dagre.js"></script>
</head>
<body>
<div id="cy"></div>
<script>
  const graph = %s;
  const cy = cytoscape({
    container: document.getElementById('cy'),
    elements: [
      ...graph.nodes.map(n => ({ data: { id: n.id, label: n.label, color: n.color } })),
      ...graph.edges.map(e => ({
        data: {
          source: e.from,
          target: e.to,
          label: e.label,
        },
        classes: e.style === 'dashed' ? 'dashed' : ''
      })),
    ],
    style: [
      {
        selector: 'node',
        style: {
          'background-color': 'data(color)',
          'label': 'data(label)',
          'text-wrap': 'wrap',
          'text-valign': 'center',
          'text-halign': 'center',
          'shape': 'round-rectangle',
          'font-size': '10px',
          'padding': '10px',
          'min-width': '80px'
        }
      },
      {
        selector: 'edge',
        style: {
          'width': 1.5,
          'label': 'data(label)',
          'curve-style': 'bezier',
          'target-arrow-shape': 'triangle',
          'line-style': 'solid',
          'font-size': '8px',
          'text-rotation': 'autorotate',
          'line-color': '#999',
          'target-arrow-color': '#999'
        }
      },
      {
        selector: '.dashed',
        style: {
          'line-style': 'dashed'
        }
      }
    ],
    layout: {
      name: 'dagre',
      rankDir: 'LR',
      nodeSep: 80,
      edgeSep: 20,
      rankSep: 100
    }
  });
</script>
</body>
</html>
`, graphJSON)
}
