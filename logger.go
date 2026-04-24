package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type dailyLogger struct {
	logFile *os.File
	date    string
}

func (l *dailyLogger) Write(p []byte) (n int, err error) {
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	// Rotate log file if date changes
	if l.date != currentDate {
		if l.logFile != nil {
			l.logFile.Close()
		}

		year := now.Format("2006")
		month := now.Format("01")
		logDir := filepath.Join(baseDir(), "logs", year, month)

		if err := os.MkdirAll(logDir, 0755); err == nil {
			logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", currentDate))
			if file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
				l.logFile = file
				l.date = currentDate
			} else {
				log.Printf("Failed to open log file: %v\n", err)
			}
		} else {
			log.Printf("Failed to create log directory: %v\n", err)
		}
	}

	// Write to file if it exists
	if l.logFile != nil {
		l.logFile.Write(p)
	}

	// Also write to standard output so console still shows logs
	return os.Stdout.Write(p)
}

func initLogger() {
	// Setup standard logger format
	log.SetFlags(log.Ldate | log.Ltime)

	// Set log output to our custom daily logger
	log.SetOutput(&dailyLogger{})
}
