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
	ID                uuid.UUID
	CustomerID        uuid.UUID
	Status            OrderStatus
	Items             []OrderItem
	TotalAmount       float64
	Currency          string
	BillingAddressID  uuid.UUID
	ShippingAddressID uuid.UUID
	PaymentID         uuid.UUID
	ShippingID        uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// OrderItem は注文アイテムを表す値オブジェクト
type OrderItem struct {
	ProductID   uuid.UUID
	ProductName string
	Quantity    int
	UnitPrice   float64
	TotalPrice  float64
}

// NewOrder は新しい注文を作成するファクトリメソッド
func NewOrder(customerID uuid.UUID, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrderItems
	}

	totalAmount := 0.0
	for _, item := range items {
		totalAmount += item.TotalPrice
	}

	return &Order{
		ID:                uuid.New(),
		CustomerID:        customerID,
		Status:            OrderStatusPending,
		Items:             items,
		TotalAmount:       totalAmount,
		Currency:          "JPY", // デフォルト通貨
		BillingAddressID:  uuid.Nil,
		ShippingAddressID: uuid.Nil,
		PaymentID:         uuid.Nil,
		ShippingID:        uuid.Nil,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
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
func (o *Order) MarkAsPaid(paymentID uuid.UUID) error {
	if o.Status != OrderStatusConfirmed && o.Status != OrderStatusBackordered {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusPaid
	o.PaymentID = paymentID
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
func (o *Order) MarkAsShipped(shippingID uuid.UUID) error {
	if o.Status != OrderStatusPaid {
		return ErrInvalidOrderState
	}
	
	o.Status = OrderStatusShipped
	o.ShippingID = shippingID
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

// SetBillingAddress は請求先住所を設定する
func (o *Order) SetBillingAddress(addressID uuid.UUID) {
	o.BillingAddressID = addressID
	o.UpdatedAt = time.Now()
}

// SetShippingAddress は配送先住所を設定する
func (o *Order) SetShippingAddress(addressID uuid.UUID) {
	o.ShippingAddressID = addressID
	o.UpdatedAt = time.Now()
}

// SetCurrency は通貨を設定する
func (o *Order) SetCurrency(currency string) {
	o.Currency = currency
	o.UpdatedAt = time.Now()
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