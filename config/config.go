package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	UserPC                  string `json:"user_pc"`                    // User PC name
	UserIdAtd               int    `json:"user_id_atd"`                // User ID in ATD system
	GoogleWebhook           string `json:"google_webhook"`             // Google Chat webhook URL
	ReminderCheckIn         bool   `json:"reminder_check_in"`          // Reminder check in
	ReminderCheckInMessage  string `json:"reminder_check_in_message"`  // Reminder check in message
	ReminderCheckOut        bool   `json:"reminder_check_out"`         // Reminder check out
	ReminderCheckOutMessage string `json:"reminder_check_out_message"` // Reminder check out message
	Interval                int    `json:"interval"`                   // Interval (seconds)
	Delay                   int    `json:"delay"`                      // Delay time (seconds) to send reminder after computer change status
	IdleThreshold           int    `json:"idle_threshold"`             // Idle threshold (seconds)
	WorkTimeEnd             string `json:"worktime_end"`               // Work time end (HH:mm:ss)
	LogRetentionDays        int    `json:"log_retention_days"`         // Log retention days
	DBHost                  string `json:"db_host"`                    // Database host
	DBPort                  int    `json:"db_port"`                    // Database port
	DBUser                  string `json:"db_user"`                    // Database user
	DBPassword              string `json:"db_password"`                // Database password
	DBName                  string `json:"db_name"`                    // Database name
}

var (
	AppConfig   Config
	configLock  sync.RWMutex
	lastModTime time.Time
)

func configFilePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir := filepath.Dir(exePath)
	return filepath.Join(baseDir, "config.json")
}

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return AppConfig
}

func Load() {
	path := configFilePath()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		configLock.Lock()
		AppConfig = Config{
			UserPC:                  "",
			UserIdAtd:               0,
			GoogleWebhook:           "",
			ReminderCheckIn:         true,
			ReminderCheckInMessage:  "⏰ Reminder: You haven't checked in yet!",
			ReminderCheckOut:        true,
			ReminderCheckOutMessage: "🌙 Reminder: Don't forget to check out before leaving!",
			Interval:                15,
			Delay:                   120,
			IdleThreshold:           360,
			WorkTimeEnd:             "17:00:00",
			LogRetentionDays:        14,
			DBHost:                  "127.0.0.1",
			DBPort:                  3306,
			DBUser:                  "root",
			DBPassword:              "root",
			DBName:                  "adt-reminder",
		}
		configLock.Unlock()

		data, _ := json.MarshalIndent(AppConfig, "", "    ")
		_ = os.WriteFile(path, data, 0644)

		if newInfo, err := os.Stat(path); err == nil {
			lastModTime = newInfo.ModTime()
		}
		return
	}

	file, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading config: %v", err)
		return
	}

	var tempConfig Config
	if err := json.Unmarshal(file, &tempConfig); err != nil {
		log.Printf("Invalid JSON in config file: %v. Keeping old config.", err)
		return
	}

	configLock.Lock()
	AppConfig = tempConfig
	lastModTime = info.ModTime()
	configLock.Unlock()

	log.Println("Config loaded successfully")
}

func WatchConfig(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			path := configFilePath()
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastModTime) {
				log.Println("Detected config change, reloading...")
				Load()
			}
		}
	}()
}
