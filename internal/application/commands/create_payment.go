package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/payment"
	"github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus"
)

// CreatePaymentCommand は支払い作成コマンド
type CreatePaymentCommand struct {
	OrderID       uuid.UUID
	Amount        float64
	Currency      string
	PaymentMethod string
}

// CreatePaymentHandler は支払い作成コマンドのハンドラ
type CreatePaymentHandler struct {
	paymentRepository payment.Repository
	eventBus          eventbus.EventBus
}

// NewCreatePaymentHandler は新しい支払い作成ハンドラを作成する
func NewCreatePaymentHandler(repo payment.Repository, bus eventbus.EventBus) *CreatePaymentHandler {
	return &CreatePaymentHandler{
		paymentRepository: repo,
		eventBus:          bus,
	}
}

// Handle は支払い作成コマンドを処理する
func (h *CreatePaymentHandler) Handle(ctx context.Context, cmd CreatePaymentCommand) (*payment.Payment, error) {
	// 新しい支払いの作成
	newPayment, err := payment.NewPayment(
		cmd.OrderID,
		cmd.Amount,
		cmd.Currency,
		payment.PaymentMethod(cmd.PaymentMethod),
	)
	if err != nil {
		return nil, err
	}

	// 永続化
	if err := h.paymentRepository.Save(ctx, newPayment); err != nil {
		return nil, err
	}

	// イベント発行
	event := payment.PaymentCreatedEvent{
		BaseEvent:   payment.NewBaseEvent(newPayment.ID),
		OrderID:     newPayment.OrderID,
		Amount:      newPayment.Amount,
		Currency:    newPayment.Currency,
		Method:      newPayment.Method,
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		return newPayment, err
	}

	return newPayment, nil
} 