package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/order"
	"github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus"
)

// ConfirmOrderCommand は注文確認コマンド
type ConfirmOrderCommand struct {
	OrderID uuid.UUID
}

// ConfirmOrderHandler は注文確認コマンドのハンドラ
type ConfirmOrderHandler struct {
	orderRepository order.Repository
	eventBus        eventbus.EventBus
}

// NewConfirmOrderHandler は新しい注文確認ハンドラを作成する
func NewConfirmOrderHandler(repo order.Repository, bus eventbus.EventBus) *ConfirmOrderHandler {
	return &ConfirmOrderHandler{
		orderRepository: repo,
		eventBus:        bus,
	}
}

// Handle は注文確認コマンドを処理する
func (h *ConfirmOrderHandler) Handle(ctx context.Context, cmd ConfirmOrderCommand) (*order.Order, error) {
	// 注文を取得
	existingOrder, err := h.orderRepository.FindByID(ctx, cmd.OrderID)
	if err != nil {
		return nil, err
	}

	// 注文を確認状態に変更
	if err := existingOrder.Confirm(); err != nil {
		return nil, err
	}

	// 永続化
	if err := h.orderRepository.Update(ctx, existingOrder); err != nil {
		return nil, err
	}

	// イベント発行
	event := order.OrderConfirmedEvent{
		BaseEvent: order.NewBaseEvent(existingOrder.ID),
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		return existingOrder, err
	}

	return existingOrder, nil
} 