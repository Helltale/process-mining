package cmd

import (
	"fmt"

	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/infrastructure"
	"github.com/Helltale/process-mining/internal/service"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Очистка данных графа",
	Long:  "Очищает данные графа на бэке.",
	Run: func(cmd *cobra.Command, args []string) {

		csvReader := infrastructure.NewCSVReader()
		graphBuilder := domain.NewGraphBuilder(csvReader)
		graphService := service.NewGraphService(graphBuilder)

		// clear graph
		graphService.ClearGraph()
		fmt.Println("Граф успешно очищен.")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
