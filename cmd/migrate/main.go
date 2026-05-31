package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"

	"api-cultura-conecta"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("ERROR: La variable de entorno DATABASE_URL no está configurada")
	}

	sourceDriver, err := iofs.New(api_cultura_conecta.MigrationFS, "migrations")
	if err != nil {
		log.Fatalf("Error al inicializar el driver de iofs: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dbURL)
	if err != nil {
		log.Fatalf("Error al inicializar golang-migrate: %v", err)
	}
	defer m.Close()

	comando := "up"
	if len(os.Args) > 1 {
		comando = os.Args[1]
	}

	switch comando {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Error crítico ejecutando Up: %v", err)
		}
		log.Println("¡Migraciones procesadas con éxito!")
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Error crítico ejecutando Down: %v", err)
		}
		log.Println("¡Down ejecutado con éxito!")
	}
}
