package inventory

import (
	"context"

	"github.com/google/uuid"
)

// InventoryRepository は在庫リポジトリのインターフェース
type InventoryRepository interface {
	// FindByProductID は商品IDで在庫を検索する
	FindByProductID(ctx context.Context, productID uuid.UUID) (*InventoryItem, error)
	
	// FindByProductIDs は複数の商品IDで在庫を検索する
	FindByProductIDs(ctx context.Context, productIDs []uuid.UUID) ([]*InventoryItem, error)
	
	// Save は在庫を保存する
	Save(ctx context.Context, item *InventoryItem) error
	
	// Update は在庫を更新する
	Update(ctx context.Context, item *InventoryItem) error
}

// ReservationRepository は在庫予約リポジトリのインターフェース
type ReservationRepository interface {
	// Save は予約を保存する
	Save(ctx context.Context, reservation *InventoryReservation) error
	
	// FindByID はIDで予約を検索する
	FindByID(ctx context.Context, id uuid.UUID) (*InventoryReservation, error)
	
	// FindByOrderID は注文IDで予約を検索する
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*InventoryReservation, error)
	
	// Update は予約を更新する
	Update(ctx context.Context, reservation *InventoryReservation) error
} 