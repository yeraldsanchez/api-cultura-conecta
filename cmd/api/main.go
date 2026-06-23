package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
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

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("APP_ENV"),
		TracesSampleRate: 1.0,
	}); err != nil {
		log.Printf("sentry: init failed: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	runMigrations(dbURL)

	poolURL := strings.Replace(dbURL, "pgx5://", "postgres://", 1)
	// normalizar otros esquemas a postgres://
	poolURL = strings.Replace(poolURL, "postgresql://", "postgres://", 1)

	pool, err := pgxpool.New(context.Background(), poolURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}

	authSvc := service.NewAuthService(db.New(pool), jwtSecret)
	userSvc := service.NewUserProfileService(db.New(pool), pool)
	catalogSvc := service.NewCatalogService(db.New(pool))
	culturalWorkSvc := service.NewCulturalWorksService(db.New(pool))
	groupSvc := service.NewGroupService(db.New(pool), pool)
	eventSvc := service.NewEventService(pool)

	authHandler := transport.NewAuthHandler(authSvc)
	userHandler := transport.NewUserProfileHandler(userSvc)
	catalogHandler := transport.NewCatalogHandler(catalogSvc)
	culturalWorksHandler := transport.NewCulturalWorksHandler(culturalWorkSvc)
	groupHandler := transport.NewGroupHandler(groupSvc)
	eventHandler := transport.NewEventHandler(eventSvc)

	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")

	r := gin.Default()
	r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	transport.RegisterRoutes(r, authHandler, userHandler, catalogHandler, culturalWorksHandler, groupHandler, eventHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
