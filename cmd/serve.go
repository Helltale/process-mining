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

		mux := http.NewServeMux()

		mux.Handle("/upload", presentation.WithCORS(presentation.LogRequest(http.HandlerFunc(graphHandler.UploadFile))))
		mux.Handle("/build", presentation.WithCORS(presentation.LogRequest(http.HandlerFunc(graphHandler.BuildGraph))))
		mux.Handle("/graph", presentation.WithCORS(presentation.LogRequest(http.HandlerFunc(graphHandler.ServeGraphData))))
		mux.Handle("/clear", presentation.WithCORS(presentation.LogRequest(http.HandlerFunc(graphHandler.ClearGraph))))
		mux.Handle("/api/tmp/", presentation.WithCORS(presentation.LogRequest(http.HandlerFunc(graphHandler.DeleteDataset))))
		mux.Handle("/api/datasets", presentation.WithCORS(presentation.LogRequest(http.HandlerFunc(graphHandler.ListDatasets))))

		// Обработчик для неизвестных путей с CORS
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			presentation.WithCORS(http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
		})

		srv := &http.Server{
			Handler:      mux,
			Addr:         fmt.Sprintf(":%s", cfg.APP_PORT),
			WriteTimeout: cfg.GetAppMaxWriteTime() * time.Minute,
			ReadTimeout:  cfg.GetAppMaxReadTime() * time.Minute,
			IdleTimeout:  60 * time.Second,
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
