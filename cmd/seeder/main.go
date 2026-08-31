package main

import (
	"context"
	"log"
	"time"
	"toll_plaza/internal/config"
	"toll_plaza/internal/repository/postgres"
	"toll_plaza/internal/service"
)

func main() {
	cfg := config.Load()

	log.Println("Connecting to PostgreSQL database...")
	db, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Fatal: Database connection failed: %v", err)
	}
	defer db.Close()

	tollRepo := postgres.NewTollRepository(db)
	seeder := service.NewSeederService(tollRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Printf("Starting manual toll plaza seeding from: %s", cfg.CSVFilePath)
	if err := seeder.SeedFromCSV(ctx, cfg.CSVFilePath); err != nil {
		log.Fatalf("Fatal: Seeding failed: %v", err)
	}

	count, _ := tollRepo.Count(ctx)
	log.Printf("Seeding complete! Total toll plazas in database: %d", count)
}
