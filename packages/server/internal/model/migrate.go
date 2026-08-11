package model

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// P2-08：版本化数据库迁移
// 启动时读取 migrations/*.sql，按文件名版本顺序应用未执行的迁移，记录到 schema_migrations 表。
// 相比 AutoMigrate，提供：版本记录、顺序控制、可回滚基点。

// migrationDir 迁移 SQL 目录（相对服务器运行目录 packages/server）
const migrationDir = "migrations"

// SchemaMigration 迁移版本记录（CreatedAt 由 gorm 自动填充，作为应用时间）
type SchemaMigration struct {
	Version   string    `gorm:"primaryKey;size:50" json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

// ApplyMigrations 应用未执行的迁移（幂等）
func ApplyMigrations(db *gorm.DB) error {
	// 确保版本表存在
	if err := db.AutoMigrate(&SchemaMigration{}); err != nil {
		return err
	}

	// 读取所有 .sql 文件并排序
	files, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, f := range files {
		// 版本号 = 文件名去 .sql 后缀，如 001_init.sql → 001_init
		version := strings.TrimSuffix(filepath.Base(f), ".sql")

		// 检查是否已应用
		var cnt int64
		if err := db.Model(&SchemaMigration{}).Where("version = ?", version).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			continue
		}

		// 执行迁移 SQL（拆分多条语句）
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		log.Printf("[migration] 应用迁移 %s", version)
		if err := db.Exec(string(sqlBytes)).Error; err != nil {
			return err
		}
		// 记录版本
		if err := db.Create(&SchemaMigration{Version: version}).Error; err != nil {
			return err
		}
	}
	return nil
}
