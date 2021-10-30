package main

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Postgres support.
	"github.com/rs/zerolog/log"
)

type App struct {
	MeetNRunDatabaseUser     string `env:"MEETNRUN_DATABASE_USER"`
	MeetNRunDatabasePassword string `env:"MEETNRUN_DATABASE_PASSWORD"`
	MeetNRunDatabaseDatabase string `env:"MEETNRUN_DATABASE_DATABASE"`
	MeetNRunDatabaseHost     string `env:"MEETNRUN_DATABASE_HOST"`
	MeetNRunDatabasePort     string `env:"MEETNRUN_DATABASE_PORT"`
}

var app = App{}

func main() {
	godotenv.Load()
	if err := env.Parse(&app); err != nil {
		log.Fatal().Err(err).Msg("could not parse environment variables")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		app.MeetNRunDatabaseHost,
		app.MeetNRunDatabasePort,
		app.MeetNRunDatabaseUser,
		app.MeetNRunDatabasePassword,
		app.MeetNRunDatabaseDatabase,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("could not open database connection")
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("could not reach database")
	}
}
