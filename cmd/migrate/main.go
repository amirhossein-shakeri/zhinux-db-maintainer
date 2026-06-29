package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: migrate [up|down|version|force]")
	}

	cmd := os.Args[1]

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize postgres driver: %v", err)
	}

	src, err := (&file.File{}).Open("db/postgres/migrations")
	if err != nil {
		log.Fatalf("Failed to open migration file: %v", err)
	}

	m, err := migrate.NewWithInstance("file", src, "postgres", driver)
	if err != nil {
		log.Fatalf("Failed to initialize migration instance: %v", err)
	}

	switch cmd {
	case "up":
		err := m.Up()
		if err != nil {
			if err != migrate.ErrNoChange {
				log.Fatalf("Migration UP failed: %v", err)
			}
			fmt.Printf("OK: No change")
		}

	case "down":
		err := m.Steps(-1)
		if err != nil {
			log.Fatalf("Migration step back -1 (DOWN) failed: %v", err)
		}

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		fmt.Printf("version=%d dirty:%v\n", v, dirty)

	case "force":
		if len(os.Args) < 3 {
			log.Fatalf("force requires version")
		}

		var version int
		fmt.Sscanf(os.Args[2], "%d", &version)

		err := m.Force(version)
		if err != nil {
			log.Fatalf("Failed to force migration to version %d: %v", version, err)
		}

	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}
