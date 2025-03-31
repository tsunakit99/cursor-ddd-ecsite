package user

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
	UserID    uuid.UUID
	EventTime time.Time
}

// NewBaseEvent は新しい基本イベントを作成する
func NewBaseEvent(userID uuid.UUID) BaseEvent {
	return BaseEvent{
		ID:        uuid.New(),
		UserID:    userID,
		EventTime: time.Now(),
	}
}

// AggregateID は集約IDを返す
func (e BaseEvent) AggregateID() uuid.UUID {
	return e.UserID
}

// OccurredAt はイベントが発生した時間を返す
func (e BaseEvent) OccurredAt() time.Time {
	return e.EventTime
}

// UserRegisteredEvent はユーザー登録イベント
type UserRegisteredEvent struct {
	BaseEvent
	Email     string
	FirstName string
	LastName  string
}

// EventType はイベントタイプを返す
func (e UserRegisteredEvent) EventType() string {
	return "user.registered"
}

// UserLoggedInEvent はユーザーログインイベント
type UserLoggedInEvent struct {
	BaseEvent
	Email string
}

// EventType はイベントタイプを返す
func (e UserLoggedInEvent) EventType() string {
	return "user.logged_in"
}

// UserUpdatedEvent はユーザー更新イベント
type UserUpdatedEvent struct {
	BaseEvent
	FirstName string
	LastName  string
}

// EventType はイベントタイプを返す
func (e UserUpdatedEvent) EventType() string {
	return "user.updated"
}

// UserPasswordChangedEvent はパスワード変更イベント
type UserPasswordChangedEvent struct {
	BaseEvent
}

// EventType はイベントタイプを返す
func (e UserPasswordChangedEvent) EventType() string {
	return "user.password_changed"
}

// UserAddressAddedEvent は住所追加イベント
type UserAddressAddedEvent struct {
	BaseEvent
	AddressID uuid.UUID
}

// EventType はイベントタイプを返す
func (e UserAddressAddedEvent) EventType() string {
	return "user.address_added"
}

// UserAddressUpdatedEvent は住所更新イベント
type UserAddressUpdatedEvent struct {
	BaseEvent
	AddressID uuid.UUID
}

// EventType はイベントタイプを返す
func (e UserAddressUpdatedEvent) EventType() string {
	return "user.address_updated"
}

// UserAddressDeletedEvent は住所削除イベント
type UserAddressDeletedEvent struct {
	BaseEvent
	AddressID uuid.UUID
}

// EventType はイベントタイプを返す
func (e UserAddressDeletedEvent) EventType() string {
	return "user.address_deleted"
} 