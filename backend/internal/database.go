package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"pintree-backend/config"
	"pintree-backend/models"
	"pintree-backend/utils"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) {
	var err error
	var dial gorm.Dialector

	switch cfg.DBDriver {
	case "sqlite":
		dir := filepath.Dir(cfg.DBPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create database directory %s: %v", dir, err)
		}
		dial = sqlite.Open(cfg.DBPath)
		log.Printf("Using SQLite database: %s", cfg.DBPath)
	case "postgres":
		dial = postgres.Open(cfg.DatabaseURL)
		log.Printf("Using PostgreSQL database")
	default:
		log.Fatalf("Unsupported database driver: %s (supported: postgres, sqlite)", cfg.DBDriver)
	}

	DB, err = gorm.Open(dial, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// SQLite 需要启用 WAL 模式和外键约束
	if cfg.DBDriver == "sqlite" {
		DB.Exec("PRAGMA journal_mode=WAL")
		DB.Exec("PRAGMA foreign_keys=ON")
	}

	// 自动迁移数据表
	err = DB.AutoMigrate(
		&models.User{},
		&models.Collection{},
		&models.Folder{},
		&models.Bookmark{},
		&models.Tag{},
		&models.Setting{},
		&models.Image{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建默认管理员账户
	createDefaultAdmin()

	log.Println(fmt.Sprintf("Database connected and migrated successfully (driver: %s)", cfg.DBDriver))
}

// createDefaultAdmin 如果不存在管理员账户则创建默认管理员
func createDefaultAdmin() {
	// 检查是否已有管理员
	var adminCount int64
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount > 0 {
		return
	}

	// 从环境变量读取默认管理员信息
	adminEmail := getEnvOrDefault("ADMIN_EMAIL", "admin@pintree.com")
	adminPassword := getEnvOrDefault("ADMIN_PASSWORD", "admin123")
	adminName := getEnvOrDefault("ADMIN_NAME", "管理员")

	// 检查邮箱是否已存在
	var existingUser models.User
	if err := DB.Where("email = ?", adminEmail).First(&existingUser).Error; err == nil {
		// 用户已存在，提升为管理员
		if existingUser.Role != "admin" {
			DB.Model(&existingUser).Update("role", "admin")
			log.Printf("已将用户 %s 提升为管理员", adminEmail)
		}
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		log.Printf("创建默认管理员失败：密码加密错误: %v", err)
		return
	}

	// 创建管理员
	admin := models.User{
		Email:    adminEmail,
		Password: hashedPassword,
		Name:     adminName,
		Role:     "admin",
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("创建默认管理员失败: %v", err)
		return
	}

	log.Printf("已创建默认管理员账户: %s", adminEmail)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
