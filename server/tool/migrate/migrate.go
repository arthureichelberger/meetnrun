package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file" // for migration file driver
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type MigrateHandler struct {
	db      *sqlx.DB
	dbname  string
	migrate *migrate.Migrate
}

func NewMigrateHandler(db *sqlx.DB, dbname string) (*MigrateHandler, error) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Error().Err(err).Caller().Msg("Error creating postgres migrate driver")
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", os.Getenv("MIGRATION_DIR")),
		dbname,
		driver,
	)
	if err != nil {
		log.Error().Err(err).Caller().Msg("Error creating postgres migrate instance")
		return nil, err
	}

	return &MigrateHandler{
		db:      db,
		dbname:  dbname,
		migrate: m,
	}, nil
}

func (mh MigrateHandler) getVersion() (uint, bool, error) {
	version, dirty, err := mh.migrate.Version()
	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

func (mh MigrateHandler) Up() {
	version, dirty, err := mh.getVersion()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			err := mh.migrate.Steps(1)
			if err != nil {
				log.Error().Err(err).Caller().Msg("Error initializing first migration")
				return
			}
		} else {
			log.Error().Err(err).Caller().Msg("Error getting migration version")
			return
		}
	}

	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Got current migration state")

	err = mh.migrate.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error().Err(err).Caller().Msg("Error migrating up")
		return
	}

	newVersion, dirty, err := mh.getVersion()
	if err != nil {
		log.Error().Err(err).Caller().Msg("Error getting new migration version")
		return
	}

	if newVersion > version {
		log.Info().Uint("version", newVersion).Bool("dirty", dirty).Msg("Successfully migrated")
	} else {
		log.Info().Msg("Nothing to migrate")
	}
}

func (mh MigrateHandler) Down() {
	version, dirty, err := mh.getVersion()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			err := mh.migrate.Steps(1)
			if err != nil {
				log.Error().Err(err).Caller().Msg("Error initializing first migration")
				return
			}
		} else {
			log.Error().Err(err).Caller().Msg("Error getting migration version")
			return
		}
	}

	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Got current migration state")

	err = mh.migrate.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error().Err(err).Caller().Msg("Error migrating down")
		return
	}

	newVersion, dirty, err := mh.getVersion()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Error().Err(err).Caller().Msg("Error getting new migration version")
		return
	}

	if newVersion < version {
		log.Info().Uint("version", newVersion).Bool("dirty", dirty).Msg("Successfully migrated")
	} else {
		log.Info().Msg("Nothing to migrate")
	}
}

type App struct {
	MeetNRunDatabaseUser     string `env:"MEETNRUN_DATABASE_USER"`
	MeetNRunDatabasePassword string `env:"MEETNRUN_DATABASE_PASSWORD"`
	MeetNRunDatabaseDatabase string `env:"MEETNRUN_DATABASE_DATABASE"`
	MeetNRunDatabaseHost     string `env:"MEETNRUN_DATABASE_HOST"`
	MeetNRunDatabasePort     string `env:"MEETNRUN_DATABASE_PORT"`
}

var app = App{}

var isMigrationDown *bool

func init() {
	isMigrationDown = flag.Bool("d", false, "Is migration going down")
}

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

	mh, err := NewMigrateHandler(db, app.MeetNRunDatabaseDatabase)
	if err != nil {
		return
	}

	flag.Parse()
	if *isMigrationDown {
		mh.Down()
	} else {
		mh.Up()
	}
}
