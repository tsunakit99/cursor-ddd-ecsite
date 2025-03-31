package order

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// OrderStatus は注文の状態を表す列挙型
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusConfirmed  OrderStatus = "CONFIRMED"
	OrderStatusPaid       OrderStatus = "PAID"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCanceled   OrderStatus = "CANCELED"
	OrderStatusBackordered OrderStatus = "BACKORDERED"
)

// 定義済みエラー
var (
	ErrInvalidOrderState = errors.New("無効な注文状態です")
	ErrEmptyOrderItems   = errors.New("注文アイテムが空です")
)

// Order は注文の集約ルートエンティティ
type Order struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     OrderStatus
	Items      []OrderItem
	TotalPrice float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// OrderItem は注文アイテムを表す値オブジェクト
type OrderItem struct {
	ProductID  uuid.UUID
	Quantity   int
	UnitPrice  float64
	TotalPrice float64
}

// NewOrder は新しい注文を作成するファクトリメソッド
func NewOrder(customerID uuid.UUID, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrderItems
	}

	totalPrice := 0.0
	for _, item := range items {
		totalPrice += item.TotalPrice
	}

	return &Order{
		ID:         uuid.New(),
		CustomerID: customerID,
		Status:     OrderStatusPending,
		Items:      items,
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// Confirm は注文を確認済み状態に変更する
func (o *Order) Confirm() error {
	if o.Status != OrderStatusPending {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()
	return nil
}

// MarkAsPaid は注文を支払い済み状態に変更する
func (o *Order) MarkAsPaid() error {
	if o.Status != OrderStatusConfirmed && o.Status != OrderStatusBackordered {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

// Backorder は注文を在庫不足状態に変更する
func (o *Order) Backorder() error {
	if o.Status != OrderStatusConfirmed {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusBackordered
	o.UpdatedAt = time.Now()
	return nil
}

// MarkAsShipped は注文を発送済み状態に変更する
func (o *Order) MarkAsShipped() error {
	if o.Status != OrderStatusPaid {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusShipped
	o.UpdatedAt = time.Now()
	return nil
}

// MarkAsDelivered は注文を配送完了状態に変更する
func (o *Order) MarkAsDelivered() error {
	if o.Status != OrderStatusShipped {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusDelivered
	o.UpdatedAt = time.Now()
	return nil
}

// Cancel は注文をキャンセル状態に変更する
func (o *Order) Cancel() error {
	if o.Status == OrderStatusShipped || o.Status == OrderStatusDelivered || o.Status == OrderStatusCanceled {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusCanceled
	o.UpdatedAt = time.Now()
	return nil
}

// CanBePaid は注文が支払い可能かどうかを確認する
func (o *Order) CanBePaid() bool {
	return o.Status == OrderStatusConfirmed || o.Status == OrderStatusBackordered
}

// CanBeShipped は注文が発送可能かどうかを確認する
func (o *Order) CanBeShipped() bool {
	return o.Status == OrderStatusPaid
}

// CanBeCanceled は注文がキャンセル可能かどうかを確認する
func (o *Order) CanBeCanceled() bool {
	return o.Status != OrderStatusShipped && o.Status != OrderStatusDelivered && o.Status != OrderStatusCanceled
} 