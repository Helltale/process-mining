package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Helltale/process-mining/config"
	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/infrastructure"
	"github.com/Helltale/process-mining/internal/presentation"
	"github.com/Helltale/process-mining/internal/service"
)

func main() {
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
		Addr:         cfg.APP_PORT,
		WriteTimeout: cfg.GetAppMaxWriteTime() * time.Minute, // Увеличенный таймаут для записи
		ReadTimeout:  cfg.GetAppMaxReadTime() * time.Minute,  // Увеличенный таймаут для чтения
	}

	// Логирование запуска сервера
	log.Printf("Сервер запущен на порту %v", cfg.APP_PORT)

	// Запуск сервера
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
