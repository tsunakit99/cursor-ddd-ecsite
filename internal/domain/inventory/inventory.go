package inventory

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// 在庫関連のエラー
var (
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrReservationNotFound   = errors.New("reservation not found")
	ErrProductNotFound       = errors.New("product not found")
	ErrInvalidReservationState = errors.New("invalid reservation state")
)

// ReservationStatus は在庫予約の状態を表す列挙型
type ReservationStatus string

const (
	ReservationStatusPending  ReservationStatus = "PENDING"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
	ReservationStatusCancelled ReservationStatus = "CANCELLED"
)

// InventoryItem は在庫アイテムを表すエンティティ
type InventoryItem struct {
	ProductID         uuid.UUID
	ProductName       string
	AvailableQuantity int
	ReservedQuantity  int
	BackorderQuantity int
	LastUpdated       time.Time
}

// NewInventoryItem は新しい在庫アイテムを作成する
func NewInventoryItem(productID uuid.UUID, productName string, quantity int) *InventoryItem {
	return &InventoryItem{
		ProductID:         productID,
		ProductName:       productName,
		AvailableQuantity: quantity,
		ReservedQuantity:  0,
		BackorderQuantity: 0,
		LastUpdated:       time.Now(),
	}
}

// CanReserve は指定された数量を予約できるかどうかを確認する
func (i *InventoryItem) CanReserve(quantity int) bool {
	return i.AvailableQuantity >= quantity
}

// Reserve は在庫を予約する
func (i *InventoryItem) Reserve(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	
	if i.AvailableQuantity < quantity {
		return ErrInsufficientStock
	}
	
	i.AvailableQuantity -= quantity
	i.ReservedQuantity += quantity
	i.LastUpdated = time.Now()
	return nil
}

// Commit は予約済みの在庫を確定する
func (i *InventoryItem) Commit(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	
	if i.ReservedQuantity < quantity {
		return ErrInsufficientStock
	}
	
	i.ReservedQuantity -= quantity
	i.LastUpdated = time.Now()
	return nil
}

// Release は予約をキャンセルし、在庫を戻す
func (i *InventoryItem) Release(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	
	if i.ReservedQuantity < quantity {
		return ErrInsufficientStock
	}
	
	i.ReservedQuantity -= quantity
	i.AvailableQuantity += quantity
	i.LastUpdated = time.Now()
	return nil
}

// Add は在庫を追加する
func (i *InventoryItem) Add(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	
	i.AvailableQuantity += quantity
	i.LastUpdated = time.Now()
	return nil
}

// IsInStock は在庫があるかどうかを確認する
func (i *InventoryItem) IsInStock() bool {
	return i.AvailableQuantity > 0
}

// InventoryReservation は在庫予約を表すエンティティ
type InventoryReservation struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	Items       []ReservationItem
	Status      ReservationStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ReservationItem は予約アイテムを表す値オブジェクト
type ReservationItem struct {
	ProductID         uuid.UUID
	RequestedQuantity int
	AvailableQuantity int
	Reserved          bool
}

// NewInventoryReservation は新しい在庫予約を作成する
func NewInventoryReservation(orderID uuid.UUID, items []ReservationItem) *InventoryReservation {
	return &InventoryReservation{
		ID:          uuid.New(),
		OrderID:     orderID,
		Items:       items,
		Status:      ReservationStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Confirm は予約を確定する
func (r *InventoryReservation) Confirm() error {
	if r.Status != ReservationStatusPending {
		return ErrInvalidReservationState
	}
	
	r.Status = ReservationStatusConfirmed
	r.UpdatedAt = time.Now()
	return nil
}

// Cancel は予約をキャンセルする
func (r *InventoryReservation) Cancel() error {
	if r.Status != ReservationStatusPending {
		return ErrInvalidReservationState
	}
	
	r.Status = ReservationStatusCancelled
	r.UpdatedAt = time.Now()
	return nil
}

// IsFullyReserved はすべてのアイテムが予約されているかどうかを確認する
func (r *InventoryReservation) IsFullyReserved() bool {
	for _, item := range r.Items {
		if !item.Reserved {
			return false
		}
	}
	return true
} 