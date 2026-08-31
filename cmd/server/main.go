package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"toll_plaza/internal/cache"
	"toll_plaza/internal/config"
	"toll_plaza/internal/handler"
	"toll_plaza/internal/repository/postgres"
	"toll_plaza/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("==================================================")
	log.Println("     Toll Plazas Between Indian Pincodes API      ")
	log.Println("==================================================")

	// 1. Load Configuration
	cfg := config.Load()

	if cfg.ServerMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. Connect to PostgreSQL Database
	log.Printf("Connecting to PostgreSQL at %s...", cfg.DatabaseURL)
	db, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}
	defer db.Close()
	log.Println("PostgreSQL connection established and schema initialized.")

	// 3. Connect to Redis Cache
	log.Printf("Connecting to Redis at %s...", cfg.RedisURL)
	redisCache := cache.NewRedisCache(cfg.RedisURL)
	defer redisCache.Close()

	// 4. Initialize Repositories
	tollRepo := postgres.NewTollRepository(db)
	pincodeRepo := postgres.NewPincodeRepository(db)

	// 5. Auto-seed Toll Data from CSV if table is empty
	seeder := service.NewSeederService(tollRepo)
	ctxSeed, cancelSeed := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := seeder.AutoSeedIfEmpty(ctxSeed, cfg.CSVFilePath); err != nil {
		log.Printf("Warning: Auto-seeding encountered an issue: %v", err)
	}
	cancelSeed()

	// 6. Initialize Services
	geocoder := service.NewGeocodingService(pincodeRepo, cfg.NominatimBaseURL)
	router := service.NewRoutingService(cfg.OSRMBaseURL)
	spatial := service.NewSpatialService()
	tollService := service.NewTollService(cfg, tollRepo, geocoder, router, spatial, redisCache)

	// 7. Initialize Handlers
	tollHandler := handler.NewTollHandler(tollService)
	healthHandler := handler.NewHealthHandler(db, redisCache)

	// 8. Setup Router & Middlewares
	r := gin.New()
	r.Use(handler.RequestLoggerMiddleware())
	r.Use(handler.RecoveryMiddleware())
	r.Use(handler.CORSMiddleware())

	// Route definitions
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/api/v1/health", healthHandler.HealthCheck)
	r.POST("/api/v1/toll-plazas", tollHandler.GetTollPlazas)

	// 9. Start Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server listening on port :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to listen: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully.")
}
