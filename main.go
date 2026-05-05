package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	"github.com/kardianos/service"
	"github.com/phamphu232/go-atd-reminder/config"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	log.Printf("ATDReminder: start service")
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	log.Printf("ATDReminder: stop service")
	return nil
}

func (p *program) Restart(s service.Service) error {
	log.Printf("ATDReminder: restart service")
	return nil
}

func (p *program) Install(s service.Service) error {
	log.Printf("ATDReminder: install service")
	return nil
}

func (p *program) Uninstall(s service.Service) error {
	log.Printf("ATDReminder: uninstall service")
	return nil
}

func (p *program) run() {
	exePath, _ := os.Executable()
	lockPath := filepath.Join(filepath.Dir(exePath), ".pid.lock")
	fileLock := flock.New(lockPath)
	locked, err := fileLock.TryLock()
	os.Chmod(lockPath, 0666)
	if err != nil || !locked {
		log.Println("ATDReminder is already running...")
		os.Exit(1)
	}

	releaseLock := func() {
		fileLock.Unlock()
		os.Remove(lockPath)
	}

	defer releaseLock()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		releaseLock()
		os.Exit(0)
	}()

	config.WatchConfig(3 * time.Second)

	for {
		time.Sleep(time.Duration(config.GetConfig().Interval) * time.Second)
		checkAttendance()
	}
}

func main() {
	config.Load()
	initLogger()
	startCleanupWorker()

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
		if os.Args[1] != "reinstall" {
			service.Control(s, os.Args[1])
		}

		return
	}

	if service.Interactive() {
		runAsAdmin(exePath, "reinstall")
		time.Sleep(3 * time.Second)
		log.Println("Service ATDReminder reinstalled successfully")
	}

	err = s.Run()
	if err != nil {
		log.Println(err)
	}
}
