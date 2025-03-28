package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Helltale/process-mining/config"
	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/infrastructure"
	"github.com/Helltale/process-mining/internal/presentation"
	"github.com/Helltale/process-mining/internal/service"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Запуск HTTP-сервера",
	Long:  "Запускает HTTP-сервер для обработки запросов.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadEnv()
		if err != nil {
			log.Fatalln("can not load config", err)
		}

		csvReader := infrastructure.NewCSVReader()
		graphBuilder := domain.NewGraphBuilder(csvReader)
		graphService := service.NewGraphService(graphBuilder)
		graphHandler := presentation.NewGraphHandler(graphService)

		http.Handle("/", http.FileServer(http.Dir("./static"))) // Статические файлы
		http.HandleFunc("/upload", graphHandler.UploadFile)     // Загрузка CSV
		http.HandleFunc("/graph", graphHandler.ServeGraphData)  // Получение данных графа
		http.HandleFunc("/clear", graphHandler.ClearGraph)      // Очистка графа

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.APP_PORT),
			WriteTimeout: cfg.GetAppMaxWriteTime() * time.Minute, // Увеличенный таймаут для записи
			ReadTimeout:  cfg.GetAppMaxReadTime() * time.Minute,  // Увеличенный таймаут для чтения
			IdleTimeout:  60 * time.Second,                       // Таймаут бездействия
		}

		// Логирование запуска сервера
		log.Printf("Сервер запущен на порту %v", cfg.APP_PORT)

		// Запуск сервера
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
