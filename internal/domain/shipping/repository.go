package shipping

import (
	"context"

	"github.com/google/uuid"
)

// Repository は配送リポジトリのインターフェース
type Repository interface {
	// Save は配送を保存する
	Save(ctx context.Context, shipping *Shipping) error
	
	// FindByID はIDで配送を検索する
	FindByID(ctx context.Context, id uuid.UUID) (*Shipping, error)
	
	// FindByOrderID は注文IDで配送を検索する
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*Shipping, error)
	
	// Update は配送を更新する
	Update(ctx context.Context, shipping *Shipping) error
	
	// AddTrackingEvent は追跡イベントを追加する
	AddTrackingEvent(ctx context.Context, shippingID uuid.UUID, event TrackingEvent) error
} 