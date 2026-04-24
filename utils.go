package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/phamphu232/go-atd-reminder/config"
)

func requestParam(r *http.Request) map[string]interface{} {
	requestParams := make(map[string]interface{})

	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			requestParams[key] = values[0]
		} else {
			requestParams[key] = values
		}
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var jsonMap map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&jsonMap)
		if err == nil {
			for k, v := range jsonMap {
				requestParams[k] = v
			}
		}
	} else {
		r.ParseForm()

		for k, v := range r.Form {
			if len(v) > 0 {
				requestParams[k] = v[0]
			}
		}
	}

	return requestParams
}

func makeDownloadDir() {
	err := os.MkdirAll(filepath.Join(baseDir(), "downloads"), 0755)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
	}
}

func startCleanupWorker() {
	go func() {
		for {
			cleanOldFiles(filepath.Join(baseDir(), "logs"), config.GetConfig().LogRetentionDays)

			time.Sleep(24 * time.Hour)
		}
	}()
}

func normalizeText(s string) string {
	s = strings.TrimSpace(s)

	// remove control characters
	s = strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		return r
	}, s)

	return s
}

func todayDir() string {
	return filepath.Join(time.Now().Format("2006"), time.Now().Format("01"), time.Now().Format("02"))
}

func baseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir := filepath.Dir(exePath)

	return baseDir
}

func cleanOldFiles(rootPath string, days int) {
	cutoff := time.Now().AddDate(0, 0, -days)

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if !info.IsDir() {
			if info.ModTime().Before(cutoff) {
				os.Remove(path)
			}
		}

		return nil
	})

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() && path != rootPath {
			os.Remove(path)
		}

		return nil
	})
}
