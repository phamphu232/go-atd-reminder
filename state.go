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
	mu                  sync.Mutex // Protects the state
	CurrentDate         string
	hasNotifiedCheckIn  bool
	hasNotifiedCheckOut bool
	IsWorking           bool
	LastChangeTime      time.Time
}

var currentState State

func (s *State) ResetIfNewDay(today string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentDate != today {
		s.CurrentDate = today
		s.hasNotifiedCheckIn = false
		s.hasNotifiedCheckOut = false
	}
}

func GetLatestAttendance(userID int) (*Attendance, error) {
	if db.DB == nil {
		db.Connect()
	}

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

	if currentState.LastChangeTime.IsZero() {
		currentState.LastChangeTime = time.Now()
	}

	activeState := IsActive(config.GetConfig().UserPC)
	screenLockedState := IsScreenLocked(config.GetConfig().UserPC)

	isWorking := activeState && !screenLockedState

	workEnd := config.GetConfig().WorkTimeEnd

	duration := now.Sub(currentState.LastChangeTime)

	if isWorking != currentState.IsWorking {
		log.Printf("isWorking: %v, isScreenLocked: %v, hasNotifiedCheckIn: %v, hasNotifiedCheckOut: %v, LastChangeTime: %s, Duration: %v, Config Delay: %d", isWorking, screenLockedState, currentState.hasNotifiedCheckIn, currentState.hasNotifiedCheckOut, currentState.LastChangeTime.Format("2006-01-02 15:04:05"), duration, config.GetConfig().Delay)
		currentState.IsWorking = isWorking
		currentState.LastChangeTime = time.Now()
		duration = 0
	}

	if config.GetConfig().ReminderCheckIn && isWorking && !currentState.hasNotifiedCheckIn && currentTime < workEnd && duration.Seconds() > float64(config.GetConfig().Delay) {
		atd, err := GetLatestAttendance(config.GetConfig().UserIdAtd)
		if err != nil {
			log.Printf("Database error: %v", err)
			return
		}

		hasCheckInToday := atd != nil && atd.DateCheck == today
		if !currentState.hasNotifiedCheckIn && !hasCheckInToday {
			log.Printf("PostToGoogleChat Notify Check In, DateCheck: %s CheckIn: %s", atd.DateCheck, atd.CheckIn.String)
			PostToGoogleChat(config.GetConfig().ReminderCheckInMessage, config.GetConfig().GoogleWebhook)
			currentState.hasNotifiedCheckIn = true
		}
	}

	if config.GetConfig().ReminderCheckOut && !isWorking && !currentState.hasNotifiedCheckOut && currentTime > workEnd && duration.Seconds() > float64(config.GetConfig().Delay) {
		atd, err := GetLatestAttendance(config.GetConfig().UserIdAtd)
		if err != nil {
			log.Printf("Database error: %v", err)
			return
		}

		hasCheckInToday := atd != nil && atd.DateCheck == today
		needsCheckOut := !currentState.hasNotifiedCheckOut && hasCheckInToday && (!atd.CheckOut.Valid || atd.CheckOut.String < workEnd)

		if needsCheckOut {
			log.Printf("PostToGoogleChat Notify Check Out, DateCheck: %s CheckOut: %s", atd.DateCheck, atd.CheckOut.String)
			PostToGoogleChat(config.GetConfig().ReminderCheckOutMessage, config.GetConfig().GoogleWebhook)
			currentState.hasNotifiedCheckOut = true
		}
	}
}
