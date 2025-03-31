package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/order"
	"github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus"
)

// CancelOrderCommand は注文キャンセルコマンド
type CancelOrderCommand struct {
	OrderID uuid.UUID
	Reason  string
}

// CancelOrderHandler は注文キャンセルコマンドのハンドラ
type CancelOrderHandler struct {
	orderRepository order.Repository
	eventBus        eventbus.EventBus
}

// NewCancelOrderHandler は新しい注文キャンセルハンドラを作成する
func NewCancelOrderHandler(repo order.Repository, bus eventbus.EventBus) *CancelOrderHandler {
	return &CancelOrderHandler{
		orderRepository: repo,
		eventBus:        bus,
	}
}

// Handle は注文キャンセルコマンドを処理する
func (h *CancelOrderHandler) Handle(ctx context.Context, cmd CancelOrderCommand) (*order.Order, error) {
	// 注文を取得
	existingOrder, err := h.orderRepository.FindByID(ctx, cmd.OrderID)
	if err != nil {
		return nil, err
	}

	// 注文をキャンセル状態に変更
	if err := existingOrder.Cancel(); err != nil {
		return nil, err
	}

	// 永続化
	if err := h.orderRepository.Update(ctx, existingOrder); err != nil {
		return nil, err
	}

	// イベント発行
	event := order.OrderCanceledEvent{
		BaseEvent: order.NewBaseEvent(existingOrder.ID),
		Reason:    cmd.Reason,
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		return existingOrder, err
	}

	return existingOrder, nil
} 