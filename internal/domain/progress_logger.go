package domain

import (
	"log"
	"runtime"
	"runtime/debug"
	"time"
)

type FileProgressLogger struct {
	totalLines  int
	processed   int
	lastLogged  time.Time
	description string
	enable      bool
	gcEveryN    int
	logTicker   *time.Ticker
	doneChan    chan struct{}
}

func NewProgressLogger(total int, description string, enable bool) *FileProgressLogger {
	p := &FileProgressLogger{
		totalLines:  total,
		lastLogged:  time.Now(),
		description: description,
		enable:      enable,
		gcEveryN:    10000,
		logTicker:   time.NewTicker(1 * time.Second),
		doneChan:    make(chan struct{}),
	}

	if enable {
		go p.runLogger()
	}
	return p
}

func (p *FileProgressLogger) Inc() {
	if !p.enable {
		return
	}
	p.processed++

	if p.processed%p.gcEveryN == 0 {
		runtime.GC()
		debug.FreeOSMemory()
	}
}

func (p *FileProgressLogger) Done() {
	if !p.enable {
		return
	}
	close(p.doneChan)
	p.logNow()
	log.Printf("[%s] Завершено. Всего строк: %d", p.description, p.processed)
}

func (p *FileProgressLogger) runLogger() {
	for {
		select {
		case <-p.logTicker.C:
			p.logNow()
		case <-p.doneChan:
			p.logTicker.Stop()
			return
		}
	}
}

func (p *FileProgressLogger) logNow() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	cpuCount := runtime.NumCPU()
	usedMB := float64(m.Alloc) / 1024 / 1024
	totalMB := float64(m.Sys) / 1024 / 1024

	progress := 0.0
	if p.totalLines > 0 {
		progress = float64(p.processed) / float64(p.totalLines) * 100
	}

	log.Printf("[%s] Прогресс: %.2f%% (%d/%d), RAM: %.2f MB / %.2f MB, CPU: %d", p.description, progress, p.processed, p.totalLines, usedMB, totalMB, cpuCount)
}
