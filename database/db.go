package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/phamphu232/go-atd-reminder/config"
)

var DB *sql.DB

func Connect() {
	dsn := config.GetConfig().DBUser + ":" + config.GetConfig().DBPassword + "@tcp(" + config.GetConfig().DBHost + ":" + strconv.Itoa(config.GetConfig().DBPort) + ")/" + config.GetConfig().DBName + "?parseTime=true"

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error when config database:", err)
	}

	DB.SetMaxOpenConns(2)
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(5 * time.Minute)
	if err := DB.Ping(); err != nil {
		log.Fatal("Error when connect to database:", err)
	}

	fmt.Println("Database is ready!")
}
