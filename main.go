package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/kardianos/service"
	"github.com/phamphu232/go-atd-reminder/config"
	db "github.com/phamphu232/go-atd-reminder/database"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	initLogger()
	startCleanupWorker()
	config.Load()
	db.Connect()

	config.WatchConfig(3 * time.Second)

	for {
		time.Sleep(time.Duration(config.GetConfig().Interval) * time.Second)
		checkAttendance()
	}
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	exePath, _ := os.Executable()

	svcConfig := &service.Config{
		Name:        "ATDReminder",
		DisplayName: "ATDReminder",
		Description: "Attendance Reminder",

		WorkingDirectory: filepath.Dir(exePath),
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		service.Control(s, os.Args[1])
		return
	}

	lockPath := filepath.Join(filepath.Dir(exePath), ".pid.lock")
	fileLock := flock.New(lockPath)
	locked, err := fileLock.TryLock()
	if err != nil || !locked {
		return
	}
	defer fileLock.Unlock()

	err = s.Run()
	if err != nil {
		log.Println(err)
	}
}
