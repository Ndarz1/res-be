package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func ConnectDB() {
	var err error
	
	dsn := "postgres://postgres:hjesa@localhost:5432/pwd-res?sslmode=disable"
	
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)
	
	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Database Connection Successful")
}
