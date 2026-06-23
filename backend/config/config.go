package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DBDriver        string // "postgres" or "sqlite"
	DatabaseURL     string // PostgreSQL DSN (used when DBDriver=postgres)
	DBPath          string // SQLite file path (used when DBDriver=sqlite)
	JWTSecret       string
	JWTExpiration   int
	UploadPath      string
	FrontendURL     string
	AllowOrigins    []string
	SnapshotDir string
	ChromePath  string
}

func LoadConfig() *Config {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or error loading it:", err)
		log.Println("Using environment variables or default values")
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" || jwtSecret == "your-secret-key-change-in-production" {
		log.Fatal("ERROR: JWT_SECRET is not configured or still using default value! " +
			"Please set a strong, unique JWT_SECRET in your .env file.")
	}

	jwtExpiration, _ := strconv.Atoi(getEnv("JWT_EXPIRATION", "86400")) // 24小时
	
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DBDriver:        getEnv("DB_DRIVER", "postgres"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		DBPath:          getEnv("DB_PATH", "./data/pintree.db"),
		JWTSecret:       jwtSecret,
		JWTExpiration:   jwtExpiration,
		UploadPath:      getEnv("UPLOAD_PATH", "./uploads"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		AllowOrigins:    []string{getEnv("FRONTEND_URL", "http://localhost:3000")},
		SnapshotDir: getEnv("SNAPSHOT_DIR", "./uploads/snapshots"),
		ChromePath:  getEnv("CHROME_PATH", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
