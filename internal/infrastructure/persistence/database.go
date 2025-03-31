package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// NewPostgresDB はPostgreSQLデータベース接続を作成する
func NewPostgresDB() (*sql.DB, error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.command.host"),
		viper.GetInt("database.command.port"),
		viper.GetString("database.command.user"),
		viper.GetString("database.command.password"),
		viper.GetString("database.command.dbname"),
		viper.GetString("database.command.sslmode"),
	)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("データベース接続のオープンに失敗しました: %w", err)
	}

	// 接続プールの設定
	db.SetMaxIdleConns(viper.GetInt("database.command.maxIdleConns"))
	db.SetMaxOpenConns(viper.GetInt("database.command.maxOpenConns"))
	db.SetConnMaxLifetime(time.Hour)

	// 接続テスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("データベース接続のテストに失敗しました: %w", err)
	}

	log.Println("PostgreSQLデータベースに接続しました")
	return db, nil
} 