package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus は支払いの状態を表す列挙型
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusApproved PaymentStatus = "APPROVED"
	PaymentStatusDeclined PaymentStatus = "DECLINED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
)

// PaymentMethod は支払い方法を表す列挙型
type PaymentMethod string

const (
	PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentMethodPayPal       PaymentMethod = "PAYPAL"
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
)

// 支払い関連のエラー
var (
	ErrInvalidPaymentState = errors.New("invalid payment state")
	ErrInvalidAmount       = errors.New("invalid payment amount")
	ErrInvalidCurrency     = errors.New("invalid currency")
)

// Payment は支払い集約のルートエンティティ
type Payment struct {
	ID           uuid.UUID
	OrderID      uuid.UUID
	Amount       float64
	Currency     string
	Method       PaymentMethod
	Status       PaymentStatus
	TransactionID string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewPayment は新しい支払いを作成する
func NewPayment(orderID uuid.UUID, amount float64, currency string, method PaymentMethod) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// 通貨の検証 (実際にはもっと詳細なバリデーションが必要)
	if currency == "" {
		return nil, ErrInvalidCurrency
	}

	return &Payment{
		ID:           uuid.New(),
		OrderID:      orderID,
		Amount:       amount,
		Currency:     currency,
		Method:       method,
		Status:       PaymentStatusPending,
		TransactionID: "",
		ErrorMessage: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// Approve は支払いを承認済み状態にする
func (p *Payment) Approve(transactionID string) error {
	if p.Status != PaymentStatusPending {
		return ErrInvalidPaymentState
	}
	
	p.Status = PaymentStatusApproved
	p.TransactionID = transactionID
	p.UpdatedAt = time.Now()
	return nil
}

// Decline は支払いを拒否状態にする
func (p *Payment) Decline(reason string) error {
	if p.Status != PaymentStatusPending {
		return ErrInvalidPaymentState
	}
	
	p.Status = PaymentStatusDeclined
	p.ErrorMessage = reason
	p.UpdatedAt = time.Now()
	return nil
}

// Refund は支払いを返金状態にする
func (p *Payment) Refund() error {
	if p.Status != PaymentStatusApproved {
		return ErrInvalidPaymentState
	}
	
	p.Status = PaymentStatusRefunded
	p.UpdatedAt = time.Now()
	return nil
} 