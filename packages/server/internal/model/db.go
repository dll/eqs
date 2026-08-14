package model

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库句柄
var DB *gorm.DB

// RDB 全局 Redis 句柄
var RDB *redis.Client

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	if cfg.DBDriver == "sqlite" {
		// DEV/SQLite 模式：使用文件库 eqs.db，无需 MySQL
		name := cfg.DBName
		if name == "" || name == "eqs" || name == "eqs.db" {
			name = "eqs.db"
		}
		db, err = gorm.Open(sqlite.Open(name), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
	}
	if err != nil {
		return nil, err
	}

	if err := AutoMigrate(db); err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}

// AutoMigrate 迁移全部 V6 域模型
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Project{},
		&SupplierQualification{},
		&Bid{},
		&Order{},
		&Contract{},
		&PaymentMilestone{},
		&Deliverable{},
		&ProjectFile{},
		&FileAnnotation{},
		&PaymentTransaction{},
		&AttendanceRecord{},
		&DeliveryTemplate{},
		&ContractTemplate{},
		&Dispute{},
		&DisputeEvidence{},
		&DisputeExpertAssignment{},
		&Review{},
		&Message{},
		&Notification{},
		&AuditLog{},
		&SystemConfig{},
		&UserSetting{},
		&SystemVersion{},
		&CommissionRecord{},
		&CaseShowcase{},
		&EscrowLedger{},
	)
}

func InitRedis(cfg *config.Config) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       0,
	})
	return nil
}

// InitTestDB 创建内存 SQLite 数据库（仅供单元测试）
func InitTestDB() *gorm.DB {
	// 使用独立名称的共享内存库，避免测试间数据串扰
	name := fmt.Sprintf("file:eqs_test_%x?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}
	// 单连接串行使用，避免共享缓存并发问题
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := AutoMigrate(db); err != nil {
		panic(err)
	}
	DB = db
	return db
}