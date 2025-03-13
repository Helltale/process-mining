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
		// Инициализация инфраструктурного слоя
		csvReader := infrastructure.NewCSVReader()

		// Инициализация доменного слоя
		graphBuilder := domain.NewGraphBuilder(csvReader)

		// Инициализация сервисного слоя
		graphService := service.NewGraphService(graphBuilder)

		// Очистка графа
		graphService.ClearGraph()
		fmt.Println("Граф успешно очищен.")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
