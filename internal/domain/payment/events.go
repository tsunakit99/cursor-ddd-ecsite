package payment

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
	ID        uuid.UUID
	PaymentID uuid.UUID
	EventTime time.Time
}

// NewBaseEvent は新しい基本イベントを作成する
func NewBaseEvent(paymentID uuid.UUID) BaseEvent {
	return BaseEvent{
		ID:        uuid.New(),
		PaymentID: paymentID,
		EventTime: time.Now(),
	}
}

// AggregateID は集約IDを返す
func (e BaseEvent) AggregateID() uuid.UUID {
	return e.PaymentID
}

// OccurredAt はイベントが発生した時間を返す
func (e BaseEvent) OccurredAt() time.Time {
	return e.EventTime
}

// PaymentCreatedEvent は支払い作成イベント
type PaymentCreatedEvent struct {
	BaseEvent
	OrderID    uuid.UUID
	Amount     float64
	Currency   string
	Method     PaymentMethod
}

// EventType はイベントタイプを返す
func (e PaymentCreatedEvent) EventType() string {
	return "payment.created"
}

// PaymentApprovedEvent は支払い承認イベント
type PaymentApprovedEvent struct {
	BaseEvent
	OrderID       uuid.UUID
	TransactionID string
}

// EventType はイベントタイプを返す
func (e PaymentApprovedEvent) EventType() string {
	return "payment.approved"
}

// PaymentDeclinedEvent は支払い拒否イベント
type PaymentDeclinedEvent struct {
	BaseEvent
	OrderID      uuid.UUID
	Reason       string
}

// EventType はイベントタイプを返す
func (e PaymentDeclinedEvent) EventType() string {
	return "payment.declined"
}

// PaymentRefundedEvent は支払い返金イベント
type PaymentRefundedEvent struct {
	BaseEvent
	OrderID      uuid.UUID
}

// EventType はイベントタイプを返す
func (e PaymentRefundedEvent) EventType() string {
	return "payment.refunded"
} 