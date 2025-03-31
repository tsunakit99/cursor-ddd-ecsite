package shipping

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ShippingStatus は配送の状態を表す列挙型
type ShippingStatus string

const (
	ShippingStatusCreated       ShippingStatus = "CREATED"
	ShippingStatusReady         ShippingStatus = "READY"
	ShippingStatusShipped       ShippingStatus = "SHIPPED"
	ShippingStatusInTransit     ShippingStatus = "IN_TRANSIT"
	ShippingStatusOutForDelivery ShippingStatus = "OUT_FOR_DELIVERY"
	ShippingStatusDelivered     ShippingStatus = "DELIVERED"
	ShippingStatusFailed        ShippingStatus = "FAILED"
)

// ShippingMethod は配送方法を表す列挙型
type ShippingMethod string

const (
	ShippingMethodStandard  ShippingMethod = "STANDARD"
	ShippingMethodExpress   ShippingMethod = "EXPRESS"
	ShippingMethodOvernight ShippingMethod = "OVERNIGHT"
)

// 配送関連のエラー
var (
	ErrInvalidShippingState = errors.New("invalid shipping state")
	ErrInvalidAddress       = errors.New("invalid address")
	ErrInvalidTrackingInfo  = errors.New("invalid tracking information")
)

// Address は住所を表す値オブジェクト
type Address struct {
	RecipientName string
	StreetLine1   string
	StreetLine2   string
	City          string
	State         string
	PostalCode    string
	Country       string
	PhoneNumber   string
}

// Validate は住所が有効かどうかを検証する
func (a Address) Validate() error {
	if a.RecipientName == "" || a.StreetLine1 == "" || a.City == "" || 
	   a.State == "" || a.PostalCode == "" || a.Country == "" {
		return ErrInvalidAddress
	}
	return nil
}

// TrackingEvent は配送追跡イベントを表す値オブジェクト
type TrackingEvent struct {
	Description string
	Location    string
	Timestamp   time.Time
}

// Shipping は配送集約のルートエンティティ
type Shipping struct {
	ID                  uuid.UUID
	OrderID             uuid.UUID
	ShippingAddress     Address
	Method              ShippingMethod
	Status              ShippingStatus
	TrackingNumber      string
	Carrier             string
	EstimatedDeliveryDate *time.Time
	ActualDeliveryDate    *time.Time
	TrackingEvents      []TrackingEvent
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewShipping は新しい配送を作成する
func NewShipping(orderID uuid.UUID, address Address, method ShippingMethod) (*Shipping, error) {
	if err := address.Validate(); err != nil {
		return nil, err
	}
	
	return &Shipping{
		ID:                  uuid.New(),
		OrderID:             orderID,
		ShippingAddress:     address,
		Method:              method,
		Status:              ShippingStatusCreated,
		TrackingNumber:      "",
		Carrier:             "",
		EstimatedDeliveryDate: nil,
		ActualDeliveryDate:    nil,
		TrackingEvents:      []TrackingEvent{},
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
}

// MarkAsReady は配送を準備完了状態にする
func (s *Shipping) MarkAsReady() error {
	if s.Status != ShippingStatusCreated {
		return ErrInvalidShippingState
	}
	
	s.Status = ShippingStatusReady
	s.UpdatedAt = time.Now()
	return nil
}

// MarkAsShipped は配送を発送済み状態にする
func (s *Shipping) MarkAsShipped(trackingNumber string, carrier string, estimatedDeliveryDate *time.Time) error {
	if s.Status != ShippingStatusReady {
		return ErrInvalidShippingState
	}
	
	if trackingNumber == "" || carrier == "" {
		return ErrInvalidTrackingInfo
	}
	
	s.Status = ShippingStatusShipped
	s.TrackingNumber = trackingNumber
	s.Carrier = carrier
	s.EstimatedDeliveryDate = estimatedDeliveryDate
	
	// 発送イベントを追加
	s.AddTrackingEvent("Package has been shipped", carrier + " shipping facility", time.Now())
	
	s.UpdatedAt = time.Now()
	return nil
}

// MarkAsDelivered は配送を配達済み状態にする
func (s *Shipping) MarkAsDelivered(deliveryDate time.Time, proofOfDelivery string) error {
	if s.Status != ShippingStatusShipped && 
	   s.Status != ShippingStatusInTransit && 
	   s.Status != ShippingStatusOutForDelivery {
		return ErrInvalidShippingState
	}
	
	s.Status = ShippingStatusDelivered
	actualDelivery := deliveryDate
	s.ActualDeliveryDate = &actualDelivery
	
	// 配達イベントを追加
	s.AddTrackingEvent("Package has been delivered. " + proofOfDelivery, 
					   s.ShippingAddress.City + ", " + s.ShippingAddress.State, 
					   deliveryDate)
	
	s.UpdatedAt = time.Now()
	return nil
}

// MarkAsFailed は配送を失敗状態にする
func (s *Shipping) MarkAsFailed(reason string) error {
	if s.Status == ShippingStatusDelivered || s.Status == ShippingStatusFailed {
		return ErrInvalidShippingState
	}
	
	s.Status = ShippingStatusFailed
	
	// 失敗イベントを追加
	s.AddTrackingEvent("Delivery failed: " + reason, "", time.Now())
	
	s.UpdatedAt = time.Now()
	return nil
}

// AddTrackingEvent は追跡イベントを追加する
func (s *Shipping) AddTrackingEvent(description string, location string, timestamp time.Time) {
	event := TrackingEvent{
		Description: description,
		Location:    location,
		Timestamp:   timestamp,
	}
	
	s.TrackingEvents = append(s.TrackingEvents, event)
	s.UpdatedAt = time.Now()
}

// UpdateShippingStatus は配送状況を更新する
func (s *Shipping) UpdateShippingStatus(status ShippingStatus, description string, location string) error {
	// ステータス変更の妥当性チェック
	switch status {
	case ShippingStatusInTransit:
		if s.Status != ShippingStatusShipped && s.Status != ShippingStatusInTransit {
			return ErrInvalidShippingState
		}
	case ShippingStatusOutForDelivery:
		if s.Status != ShippingStatusInTransit && s.Status != ShippingStatusShipped {
			return ErrInvalidShippingState
		}
	default:
		return ErrInvalidShippingState
	}
	
	s.Status = status
	s.AddTrackingEvent(description, location, time.Now())
	
	return nil
} 