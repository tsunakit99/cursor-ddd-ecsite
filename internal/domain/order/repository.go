package order

import (
	"context"

	"github.com/google/uuid"
)

// Repository は注文リポジトリのインターフェース
type Repository interface {
	// Save は注文を保存する
	Save(ctx context.Context, order *Order) error
	
	// FindByID は指定されたIDの注文を取得する
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	
	// FindByCustomerID は指定された顧客IDの注文リストを取得する
	FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*Order, error)
	
	// Update は注文を更新する
	Update(ctx context.Context, order *Order) error
	
	// Delete は注文を削除する（ソフトデリート）
	Delete(ctx context.Context, id uuid.UUID) error
} 