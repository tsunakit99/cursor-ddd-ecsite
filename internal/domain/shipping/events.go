package shipping

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
	ID         uuid.UUID
	ShippingID uuid.UUID
	EventTime  time.Time
}

// NewBaseEvent は新しい基本イベントを作成する
func NewBaseEvent(shippingID uuid.UUID) BaseEvent {
	return BaseEvent{
		ID:         uuid.New(),
		ShippingID: shippingID,
		EventTime:  time.Now(),
	}
}

// AggregateID は集約IDを返す
func (e BaseEvent) AggregateID() uuid.UUID {
	return e.ShippingID
}

// OccurredAt はイベントが発生した時間を返す
func (e BaseEvent) OccurredAt() time.Time {
	return e.EventTime
}

// ShippingCreatedEvent は配送作成イベント
type ShippingCreatedEvent struct {
	BaseEvent
	OrderID     uuid.UUID
	Address     Address
	Method      ShippingMethod
}

// EventType はイベントタイプを返す
func (e ShippingCreatedEvent) EventType() string {
	return "shipping.created"
}

// ShippingReadyEvent は配送準備完了イベント
type ShippingReadyEvent struct {
	BaseEvent
	OrderID     uuid.UUID
}

// EventType はイベントタイプを返す
func (e ShippingReadyEvent) EventType() string {
	return "shipping.ready"
}

// ShippingShippedEvent は配送発送済みイベント
type ShippingShippedEvent struct {
	BaseEvent
	OrderID                uuid.UUID
	TrackingNumber         string
	Carrier                string
	EstimatedDeliveryDate  *time.Time
}

// EventType はイベントタイプを返す
func (e ShippingShippedEvent) EventType() string {
	return "shipping.shipped"
}

// ShippingDeliveredEvent は配送配達済みイベント
type ShippingDeliveredEvent struct {
	BaseEvent
	OrderID           uuid.UUID
	ActualDeliveryDate time.Time
	ProofOfDelivery    string
}

// EventType はイベントタイプを返す
func (e ShippingDeliveredEvent) EventType() string {
	return "shipping.delivered"
}

// ShippingFailedEvent は配送失敗イベント
type ShippingFailedEvent struct {
	BaseEvent
	OrderID uuid.UUID
	Reason  string
}

// EventType はイベントタイプを返す
func (e ShippingFailedEvent) EventType() string {
	return "shipping.failed"
}

// ShippingStatusUpdatedEvent は配送状況更新イベント
type ShippingStatusUpdatedEvent struct {
	BaseEvent
	OrderID     uuid.UUID
	Status      ShippingStatus
	Description string
	Location    string
}

// EventType はイベントタイプを返す
func (e ShippingStatusUpdatedEvent) EventType() string {
	return "shipping.statusUpdated"
} 