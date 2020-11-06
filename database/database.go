package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/nadirbasalamah/go-simple-grpc/config"
)

// DB represents database
var DB *sql.DB

// Connect func to connect to the database, if failed returns error
func Connect() error {
	var err error
	p := config.Config("DB_PORT")

	port, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		log.Fatalf("Error parsing string to int\n")
	}

	DB, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Config("DB_HOST"), port, config.Config("DB_USER"), config.Config("DB_PASSWORD"), config.Config("DB_NAME")))
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	DB.Query(`CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		amount integer,
		name text UNIQUE,
		description text,
		category text NOT NULL
	)
	`)

	fmt.Println("Connected to the database")
	return nil
}
