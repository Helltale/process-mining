package domain

import (
	"log"
	"time"
)

type FileProgressLogger struct {
	totalLines  int
	processed   int
	lastLogged  time.Time
	interval    time.Duration
	enable      bool
	description string
}

func NewProgressLogger(total int, description string, enable bool) *FileProgressLogger {
	return &FileProgressLogger{
		totalLines:  total,
		interval:    2 * time.Second,
		lastLogged:  time.Now(),
		enable:      enable,
		description: description,
	}
}

func (p *FileProgressLogger) Inc() {
	if !p.enable {
		return
	}
	p.processed++
	now := time.Now()
	if now.Sub(p.lastLogged) >= p.interval {
		p.Log()
		p.lastLogged = now
	}
}

func (p *FileProgressLogger) Log() {
	if p.totalLines == 0 {
		log.Printf("[%s] Обработано строк: %d", p.description, p.processed)
	} else {
		progress := float64(p.processed) / float64(p.totalLines) * 100
		log.Printf("[%s] Прогресс: %.2f%% (%d/%d)", p.description, progress, p.processed, p.totalLines)
	}
}

func (p *FileProgressLogger) Done() {
	if !p.enable {
		return
	}
	p.Log()
	log.Printf("[%s] Завершено. Обработано строк: %d", p.description, p.processed)
}
