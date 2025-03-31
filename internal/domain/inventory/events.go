package inventory

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
	EventTime  time.Time
}

// NewBaseEvent は新しい基本イベントを作成する
func NewBaseEvent() BaseEvent {
	return BaseEvent{
		ID:        uuid.New(),
		EventTime: time.Now(),
	}
}

// OccurredAt はイベントが発生した時間を返す
func (e BaseEvent) OccurredAt() time.Time {
	return e.EventTime
}

// InventoryReservedEvent は在庫予約イベント
type InventoryReservedEvent struct {
	BaseEvent
	ReservationID uuid.UUID
	OrderID       uuid.UUID
	Items         []ReservationItem
}

// EventType はイベントタイプを返す
func (e InventoryReservedEvent) EventType() string {
	return "inventory.reserved"
}

// AggregateID は集約IDを返す
func (e InventoryReservedEvent) AggregateID() uuid.UUID {
	return e.ReservationID
}

// InventoryReleasedEvent は在庫解放イベント
type InventoryReleasedEvent struct {
	BaseEvent
	ReservationID uuid.UUID
	OrderID       uuid.UUID
	ProductIDs    []uuid.UUID
}

// EventType はイベントタイプを返す
func (e InventoryReleasedEvent) EventType() string {
	return "inventory.released"
}

// AggregateID は集約IDを返す
func (e InventoryReleasedEvent) AggregateID() uuid.UUID {
	return e.ReservationID
}

// InventoryCommittedEvent は在庫確定イベント
type InventoryCommittedEvent struct {
	BaseEvent
	ReservationID uuid.UUID
	OrderID       uuid.UUID
}

// EventType はイベントタイプを返す
func (e InventoryCommittedEvent) EventType() string {
	return "inventory.committed"
}

// AggregateID は集約IDを返す
func (e InventoryCommittedEvent) AggregateID() uuid.UUID {
	return e.ReservationID
}

// InventoryAddedEvent は在庫追加イベント
type InventoryAddedEvent struct {
	BaseEvent
	ProductID uuid.UUID
	Quantity  int
	LocationID string
}

// EventType はイベントタイプを返す
func (e InventoryAddedEvent) EventType() string {
	return "inventory.added"
}

// AggregateID は集約IDを返す
func (e InventoryAddedEvent) AggregateID() uuid.UUID {
	return e.ProductID
}

// StockDepletedEvent は在庫切れイベント
type StockDepletedEvent struct {
	BaseEvent
	ProductID uuid.UUID
}

// EventType はイベントタイプを返す
func (e StockDepletedEvent) EventType() string {
	return "inventory.stockDepleted"
}

// AggregateID は集約IDを返す
func (e StockDepletedEvent) AggregateID() uuid.UUID {
	return e.ProductID
} 