package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Alfafauzan/coffee-pos/config"
	"github.com/Alfafauzan/coffee-pos/internal/handler"
	"github.com/Alfafauzan/coffee-pos/pkg/database"
	"github.com/Alfafauzan/coffee-pos/pkg/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 1. Inisialisasi koneksi MySQL
	db, err := database.NewMySQL(cfg.MysqlDSN())
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()
	fmt.Println("MySQL connected.")

	// 2. Inisialisasi koneksi Redis
	rdb, err := redis.NewRedis(cfg.RedisAddr(), cfg.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()
	fmt.Println("Redis connected.")

	// 3. Inisialisasi Gin router
	r := handler.NewRouter(cfg.AppEnv)

	// 4. Jalankan server HTTP
	log.Printf("Starting server on port :%s", cfg.AppPort)
  if err := r.Run(":" + cfg.AppPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
    log.Fatalf("Failed to start server: %v", err)
  }
}
