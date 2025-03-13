package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Helltale/process-mining/internal/domain"
	"github.com/Helltale/process-mining/internal/infrastructure"
	"github.com/Helltale/process-mining/internal/presentation"
	"github.com/Helltale/process-mining/internal/service"
)

func main() {
	// Инициализация инфраструктурного слоя
	csvReader := infrastructure.NewCSVReader()

	// Инициализация доменного слоя
	graphBuilder := domain.NewGraphBuilder(csvReader)

	// Инициализация сервисного слоя
	graphService := service.NewGraphService(graphBuilder)

	// Инициализация слоя представления
	graphHandler := presentation.NewGraphHandler(graphService)

	// Настройка маршрутов
	http.Handle("/", http.FileServer(http.Dir("./static"))) // Статические файлы
	http.HandleFunc("/upload", graphHandler.UploadFile)     // Загрузка CSV
	http.HandleFunc("/graph", graphHandler.ServeGraphData)  // Получение данных графа
	http.HandleFunc("/clear", graphHandler.ClearGraph)      // Очистка графа

	// Настройка сервера с увеличенными таймаутами
	srv := &http.Server{
		Addr:         ":8085",
		WriteTimeout: 15 * time.Minute, // Увеличенный таймаут для записи
		ReadTimeout:  15 * time.Minute, // Увеличенный таймаут для чтения
	}

	// Логирование запуска сервера
	log.Printf("Сервер запущен на порту %v", srv.Addr)

	// Запуск сервера
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
