package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/shipping"
	"github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus"
)

// CreateShippingCommand は配送作成コマンド
type CreateShippingCommand struct {
	OrderID         uuid.UUID
	ShippingAddress shipping.Address
	ShippingMethod  shipping.ShippingMethod
}

// CreateShippingHandler は配送作成コマンドのハンドラ
type CreateShippingHandler struct {
	shippingRepository shipping.Repository
	eventBus           eventbus.EventBus
}

// NewCreateShippingHandler は新しい配送作成ハンドラを作成する
func NewCreateShippingHandler(repo shipping.Repository, bus eventbus.EventBus) *CreateShippingHandler {
	return &CreateShippingHandler{
		shippingRepository: repo,
		eventBus:           bus,
	}
}

// Handle は配送作成コマンドを処理する
func (h *CreateShippingHandler) Handle(ctx context.Context, cmd CreateShippingCommand) (*shipping.Shipping, error) {
	// 新しい配送の作成
	newShipping, err := shipping.NewShipping(
		cmd.OrderID,
		cmd.ShippingAddress,
		cmd.ShippingMethod,
	)
	if err != nil {
		return nil, err
	}

	// 永続化
	if err := h.shippingRepository.Save(ctx, newShipping); err != nil {
		return nil, err
	}

	// イベント発行
	event := shipping.ShippingCreatedEvent{
		BaseEvent: shipping.NewBaseEvent(newShipping.ID),
		OrderID:   newShipping.OrderID,
		Address:   newShipping.ShippingAddress,
		Method:    newShipping.Method,
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		return newShipping, err
	}

	return newShipping, nil
} 