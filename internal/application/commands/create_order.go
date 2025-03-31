package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/order"
	"github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus"
)

// CreateOrderCommand は注文作成コマンド
type CreateOrderCommand struct {
	CustomerID uuid.UUID
	Items      []CreateOrderItem
}

// CreateOrderItem は注文アイテム作成用のDTO
type CreateOrderItem struct {
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
}

// CreateOrderHandler は注文作成コマンドのハンドラ
type CreateOrderHandler struct {
	orderRepository order.Repository
	eventBus        eventbus.EventBus
}

// NewCreateOrderHandler は新しい注文作成ハンドラを作成する
func NewCreateOrderHandler(repo order.Repository, bus eventbus.EventBus) *CreateOrderHandler {
	return &CreateOrderHandler{
		orderRepository: repo,
		eventBus:        bus,
	}
}

// Handle は注文作成コマンドを処理する
func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (*order.Order, error) {
	// コマンドからドメインモデルの注文アイテムに変換
	items := make([]order.OrderItem, len(cmd.Items))
	for i, item := range cmd.Items {
		items[i] = order.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: float64(item.Quantity) * item.UnitPrice,
		}
	}

	// 新しい注文の作成（ドメインロジックを使用）
	newOrder, err := order.NewOrder(cmd.CustomerID, items)
	if err != nil {
		return nil, err
	}
	
	// 永続化
	if err := h.orderRepository.Save(ctx, newOrder); err != nil {
		return nil, err
	}
	
	// イベント発行
	event := order.OrderCreatedEvent{
		BaseEvent: order.NewBaseEvent(newOrder.ID),
		CustomerID: newOrder.CustomerID,
		Items: newOrder.Items,
		TotalPrice: newOrder.TotalPrice,
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		// 注: イベント発行の失敗を処理する方法はプロジェクトの要件によって異なる
		// 例: ログに記録、リトライ、または補償トランザクションの開始など
		return newOrder, err
	}
	
	return newOrder, nil
} 