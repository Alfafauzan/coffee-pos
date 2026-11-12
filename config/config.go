package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
    AppPort  string
    AppEnv   string

    DBHost     string
    DBPort     string
    DBName     string
    DBUser     string
    DBPassword string

    RedisHost     string
    RedisPort     string
    RedisPassword string

    JWTSecret      string
    JWTExpireHours int
    JWTExpiryHours int

    MidtransServerKey string
    MidtransClientKey string
    MidtransEnv       string
}

func Load() (*Config, error) {
    // Tidak error kalau .env tidak ada — production inject env lewat sistem
    _ = godotenv.Load()

    cfg := &Config{
        AppPort: getEnv("APP_PORT", "8080"),
        AppEnv:  getEnv("APP_ENV", "development"),

        DBHost:     os.Getenv("DB_HOST"),
        DBPort:     getEnv("DB_PORT", "3306"),
        DBName:     os.Getenv("DB_NAME"),
        DBUser:     os.Getenv("DB_USER"),
        DBPassword: os.Getenv("DB_PASSWORD"),

        RedisHost:     getEnv("REDIS_HOST", "localhost"),
        RedisPort:     getEnv("REDIS_PORT", "6379"),
        RedisPassword: os.Getenv("REDIS_PASSWORD"),

        JWTSecret: os.Getenv("JWT_SECRET"),

        MidtransServerKey: os.Getenv("MIDTRANS_SERVER_KEY"),
        MidtransClientKey: os.Getenv("MIDTRANS_CLIENT_KEY"),
        MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
    }

    // Parse JWT expire hours
    cfg.JWTExpireHours = 24
    expireEnv := os.Getenv("JWT_EXPIRE_HOURS")
    if expireEnv == "" {
        expireEnv = os.Getenv("JWT_EXPIRY_HOURS")
    }
    if expireEnv != "" {
        if hours, err := strconv.Atoi(expireEnv); err == nil {
            cfg.JWTExpireHours = hours
        }
    }
    cfg.JWTExpiryHours = cfg.JWTExpireHours

    // Validasi field wajib
    required := map[string]string{
        "DB_HOST":     cfg.DBHost,
        "DB_NAME":     cfg.DBName,
        "DB_USER":     cfg.DBUser,
        "DB_PASSWORD": cfg.DBPassword,
        "JWT_SECRET":  cfg.JWTSecret,
    }
    for key, val := range required {
        if val == "" {
            return nil, fmt.Errorf("environment variable %s wajib diisi", key)
        }
    }

    return cfg, nil
}

func (c *Config) MysqlDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
        c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) RedisAddr() string {
    return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
