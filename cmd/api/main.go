package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"URL-Shotener/internal/api"
	"URL-Shotener/internal/cache"
	"URL-Shotener/internal/config"
	"URL-Shotener/internal/repository"
	"URL-Shotener/internal/service"
)

func main() {
	cfg := config.Load()

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully")

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS urls (
		id BIGSERIAL PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE
	);
	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	redisCache := cache.NewRedisCache(cfg.Redis)
	defer redisCache.Close()

	urlRepo := repository.NewPostgresURLRepository(db)
	urlService := service.NewURLService(urlRepo, redisCache, cfg)
	urlHandler := api.NewURLHandler(urlService)

	router := api.SetupRouter(urlHandler)

	log.Printf("Server is running on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
