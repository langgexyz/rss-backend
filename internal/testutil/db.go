// backend/internal/testutil/db.go
package testutil

import (
	"context"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"rss-backend/internal/model"
)

// SetupMySQL 启动一个 MySQL Docker 容器，返回 *gorm.DB。
// 测试结束后容器自动销毁。
func SetupMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("rss_reader_test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("start mysql container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "charset=utf8mb4", "parseTime=True", "loc=UTC")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	if err := db.AutoMigrate(&model.Feed{}, &model.Article{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() {
		db.Exec("DELETE FROM articles")
		db.Exec("DELETE FROM feeds")
	})

	return db
}
