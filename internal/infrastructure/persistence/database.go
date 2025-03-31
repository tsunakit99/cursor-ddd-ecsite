package persistence

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// NewPostgresDB はPostgreSQLデータベース接続を作成する
func NewPostgresDB() (*sqlx.DB, error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.command.host"),
		viper.GetInt("database.command.port"),
		viper.GetString("database.command.user"),
		viper.GetString("database.command.password"),
		viper.GetString("database.command.dbname"),
		viper.GetString("database.command.sslmode"),
	)

	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("データベース接続のオープンに失敗しました: %w", err)
	}

	// 接続プールの設定
	db.SetMaxIdleConns(viper.GetInt("database.command.maxIdleConns"))
	db.SetMaxOpenConns(viper.GetInt("database.command.maxOpenConns"))
	db.SetConnMaxLifetime(time.Hour)

	log.Println("PostgreSQLデータベースに接続しました")
	return db, nil
} 