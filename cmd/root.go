package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "process-mining",
	Short: "Process Mining Service",
	Long:  "Сервис для анализа процессов и построения графов.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Используйте 'serve' для запуска сервера.")
		fmt.Println("Используйте 'clear' для очистки данных графа.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}
