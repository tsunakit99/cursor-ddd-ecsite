package payment

import (
	"context"

	"github.com/google/uuid"
)

// Repository は支払いリポジトリのインターフェース
type Repository interface {
	// Save は支払いを保存する
	Save(ctx context.Context, payment *Payment) error
	
	// FindByID はIDで支払いを検索する
	FindByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	
	// FindByOrderID は注文IDで支払いを検索する
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*Payment, error)
	
	// Update は支払いを更新する
	Update(ctx context.Context, payment *Payment) error
} 