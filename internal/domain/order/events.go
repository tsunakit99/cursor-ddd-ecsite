package order

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent はドメインイベントの基本インターフェース
type DomainEvent interface {
	EventType() string
	AggregateID() uuid.UUID
	OccurredAt() time.Time
}

// BaseEvent はすべてのイベントの基本構造
type BaseEvent struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	EventTime   time.Time
}

// NewBaseEvent は新しい基本イベントを作成する
func NewBaseEvent(orderID uuid.UUID) BaseEvent {
	return BaseEvent{
		ID:          uuid.New(),
		OrderID:     orderID,
		EventTime:   time.Now(),
	}
}

// AggregateID は集約IDを返す
func (e BaseEvent) AggregateID() uuid.UUID {
	return e.OrderID
}

// OccurredAt はイベントが発生した時間を返す
func (e BaseEvent) OccurredAt() time.Time {
	return e.EventTime
}

// OrderCreatedEvent は注文作成イベント
type OrderCreatedEvent struct {
	BaseEvent
	CustomerID  uuid.UUID
	Items       []OrderItem
	TotalPrice  float64
}

// EventType はイベントタイプを返す
func (e OrderCreatedEvent) EventType() string {
	return "order.created"
}

// OrderConfirmedEvent は注文確認イベント
type OrderConfirmedEvent struct {
	BaseEvent
}

// EventType はイベントタイプを返す
func (e OrderConfirmedEvent) EventType() string {
	return "order.confirmed"
}

// OrderPaidEvent は注文支払い完了イベント
type OrderPaidEvent struct {
	BaseEvent
	PaymentID uuid.UUID
}

// EventType はイベントタイプを返す
func (e OrderPaidEvent) EventType() string {
	return "order.paid"
}

// OrderBackorderedEvent は注文在庫不足イベント
type OrderBackorderedEvent struct {
	BaseEvent
	BackorderedItems []OrderItem
}

// EventType はイベントタイプを返す
func (e OrderBackorderedEvent) EventType() string {
	return "order.backordered"
}

// OrderShippedEvent は注文発送済みイベント
type OrderShippedEvent struct {
	BaseEvent
	ShippingID uuid.UUID
}

// EventType はイベントタイプを返す
func (e OrderShippedEvent) EventType() string {
	return "order.shipped"
}

// OrderDeliveredEvent は注文配送完了イベント
type OrderDeliveredEvent struct {
	BaseEvent
	DeliveryDate time.Time
}

// EventType はイベントタイプを返す
func (e OrderDeliveredEvent) EventType() string {
	return "order.delivered"
}

// OrderCanceledEvent は注文キャンセルイベント
type OrderCanceledEvent struct {
	BaseEvent
	Reason string
}

// EventType はイベントタイプを返す
func (e OrderCanceledEvent) EventType() string {
	return "order.canceled"
} 