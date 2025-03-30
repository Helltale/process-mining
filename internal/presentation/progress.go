package presentation

import "sync"

var (
	progressMu     sync.RWMutex
	progressByFile = make(map[string]int) // от 0 до 100
)

func setProgress(file string, percent int) {
	progressMu.Lock()
	defer progressMu.Unlock()
	progressByFile[file] = percent
}

func getProgress(file string) int {
	progressMu.RLock()
	defer progressMu.RUnlock()
	return progressByFile[file]
}

func clearProgress(file string) {
	progressMu.Lock()
	defer progressMu.Unlock()
	delete(progressByFile, file)
}
