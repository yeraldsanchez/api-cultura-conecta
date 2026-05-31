package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	api "api-cultura-conecta"
	db "api-cultura-conecta/internal/db/generated"
	"api-cultura-conecta/internal/service"
	"api-cultura-conecta/internal/transport"
)

func runMigrations(dbURL string) {
	// golang-migrate con el driver pgx/v5 requiere el scheme "pgx5://"
	migrateURL := strings.Replace(dbURL, "postgresql://", "pgx5://", 1)
	migrateURL = strings.Replace(migrateURL, "postgres://", "pgx5://", 1)

	src, err := iofs.New(api.MigrationFS, "migrations")
	if err != nil {
		log.Fatalf("migrations: source error: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL)
	if err != nil {
		log.Fatalf("migrations: init error: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations: OK")
}

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	runMigrations(dbURL)

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer pool.Close()

	authSvc := service.NewAuthService(db.New(pool), jwtSecret)
	authHandler := transport.NewAuthHandler(authSvc)

	r := gin.Default()
	transport.RegisterRoutes(r, authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
