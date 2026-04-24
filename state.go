package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/phamphu232/go-atd-reminder/config"
	db "github.com/phamphu232/go-atd-reminder/database"
)

type Attendance struct {
	ID        int
	UserID    int
	DateCheck string
	CheckIn   sql.NullString
	CheckOut  sql.NullString
}

type State struct {
	mu                 sync.Mutex // Protects the state
	CurrentDate        string
	IsReminderCheckIn  bool
	IsReminderCheckOut bool
	IsWorking          bool
	LastInputTime      time.Time
}

var currentState State

func (s *State) ResetIfNewDay(today string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentDate != today {
		s.CurrentDate = today
		s.IsReminderCheckIn = false
		s.IsReminderCheckOut = false
	}
}

func GetLatestAttendance(userID int) (*Attendance, error) {
	var atd Attendance

	query := `SELECT id, user_id, date_check, check_in, check_out 
	          FROM attendance_time 
	          WHERE user_id = ? 
	          ORDER BY id DESC LIMIT 1`

	err := db.DB.QueryRow(query, userID).Scan(
		&atd.ID,
		&atd.UserID,
		&atd.DateCheck,
		&atd.CheckIn,
		&atd.CheckOut,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	atd.DateCheck = atd.DateCheck[:10]
	return &atd, nil
}

func checkAttendance() {
	now := time.Now()
	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	currentState.ResetIfNewDay(today)

	isWorking := IsWorking(config.GetConfig().UserPC)
	workEnd := config.GetConfig().WorkTimeEnd

	currentState.mu.Lock()
	defer currentState.mu.Unlock()

	duration := now.Sub(currentState.LastInputTime)

	if isWorking != currentState.IsWorking {
		log.Printf("UserIsWorking: %v, IsLocked: %v, LastInputTime: %s, Duration: %s", isWorking, IsScreenLocked(), currentState.LastInputTime, duration)
		currentState.IsWorking = isWorking
		currentState.LastInputTime = time.Now()
	}

	if config.GetConfig().ReminderCheckIn && isWorking && !currentState.IsReminderCheckIn && currentTime < workEnd && duration.Seconds() > float64(config.GetConfig().Delay) {
		atd, err := GetLatestAttendance(config.GetConfig().UserIdAtd)
		if err != nil {
			log.Printf("Database error: %v", err)
			return
		}

		hasCheckedInToday := atd != nil && atd.DateCheck == today

		if !hasCheckedInToday {
			log.Printf("ReminderCheckIn: %v, isWorking: %v, currentState.IsReminderCheckIn: %v, currentTime: %s, workEnd: %s, duration: %s, config.GetConfig().Delay: %d, atd.DateCheck: %s, today: %s", config.GetConfig().ReminderCheckIn, isWorking, currentState.IsReminderCheckIn, currentTime, workEnd, duration, config.GetConfig().Delay, atd.DateCheck, today)
			PostToGoogleChat(config.GetConfig().ReminderCheckInMessage, config.GetConfig().GoogleWebhook)
			currentState.IsReminderCheckIn = true

		}
	}

	if config.GetConfig().ReminderCheckOut && !isWorking && !currentState.IsReminderCheckOut && currentTime > workEnd && duration.Seconds() > float64(config.GetConfig().Delay) {
		atd, err := GetLatestAttendance(config.GetConfig().UserIdAtd)
		if err != nil {
			log.Printf("Database error: %v", err)
			return
		}

		hasCheckedInToday := atd != nil && atd.DateCheck == today
		needsCheckOut := !hasCheckedInToday || !atd.CheckOut.Valid || atd.CheckOut.String < workEnd

		if needsCheckOut {
			log.Printf("ReminderCheckOut: %v, isWorking: %v, currentState.IsReminderCheckOut: %v, currentTime: %s, atd.CheckOut: %s, 	workEnd: %s, duration: %s, config.GetConfig().Delay: %d, atd.DateCheck: %s, today: %s", config.GetConfig().ReminderCheckOut, isWorking, currentState.IsReminderCheckOut, currentTime, atd.CheckOut.String, workEnd, duration, config.GetConfig().Delay, atd.DateCheck, today)
			PostToGoogleChat(config.GetConfig().ReminderCheckOutMessage, config.GetConfig().GoogleWebhook)
			currentState.IsReminderCheckOut = true
		}
	}
}
