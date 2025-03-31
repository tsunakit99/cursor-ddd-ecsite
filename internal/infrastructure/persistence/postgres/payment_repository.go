package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/payment"
)

// PaymentRepository はPostgreSQLを使用した支払いリポジトリの実装
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository は新しい支払いリポジトリを作成する
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

// Save は支払いを保存する
func (r *PaymentRepository) Save(ctx context.Context, payment *payment.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, amount, currency, method, status, transaction_id, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		string(payment.Method),
		string(payment.Status),
		payment.TransactionID,
		payment.ErrorMessage,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("支払いの保存に失敗しました: %w", err)
	}
	return nil
}

// FindByID はIDで支払いを検索する
func (r *PaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, method, status, transaction_id, error_message, created_at, updated_at
		FROM payments
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var p payment.Payment
	var method, status string
	err := row.Scan(
		&p.ID,
		&p.OrderID,
		&p.Amount,
		&p.Currency,
		&method,
		&status,
		&p.TransactionID,
		&p.ErrorMessage,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("支払いが見つかりません: %w", err)
		}
		return nil, fmt.Errorf("支払いの取得に失敗しました: %w", err)
	}
	p.Method = payment.PaymentMethod(method)
	p.Status = payment.PaymentStatus(status)

	return &p, nil
}

// FindByOrderID は注文IDで支払いを検索する
func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID uuid.UUID) (*payment.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, method, status, transaction_id, error_message, created_at, updated_at
		FROM payments
		WHERE order_id = $1
	`
	row := r.db.QueryRowContext(ctx, query, orderID)

	var p payment.Payment
	var method, status string
	err := row.Scan(
		&p.ID,
		&p.OrderID,
		&p.Amount,
		&p.Currency,
		&method,
		&status,
		&p.TransactionID,
		&p.ErrorMessage,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("支払いが見つかりません: %w", err)
		}
		return nil, fmt.Errorf("支払いの取得に失敗しました: %w", err)
	}
	p.Method = payment.PaymentMethod(method)
	p.Status = payment.PaymentStatus(status)

	return &p, nil
}

// Update は支払いを更新する
func (r *PaymentRepository) Update(ctx context.Context, payment *payment.Payment) error {
	query := `
		UPDATE payments
		SET amount = $1, currency = $2, method = $3, status = $4, transaction_id = $5, error_message = $6, updated_at = $7
		WHERE id = $8
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		payment.Amount,
		payment.Currency,
		string(payment.Method),
		string(payment.Status),
		payment.TransactionID,
		payment.ErrorMessage,
		payment.UpdatedAt,
		payment.ID,
	)
	if err != nil {
		return fmt.Errorf("支払いの更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新された行数の取得に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("更新対象の支払いが見つかりません: %w", err)
	}

	return nil
} 