package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/order"
)

// OrderRepository はPostgreSQLを使用した注文リポジトリの実装
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository は新しい注文リポジトリを作成する
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// Save は注文を保存する
func (r *OrderRepository) Save(ctx context.Context, order *order.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	// 注文を保存
	query := `
		INSERT INTO orders (
			id, customer_id, status, total_amount, currency, billing_address_id, shipping_address_id,
			payment_id, shipping_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	var paymentID, shippingID *uuid.UUID
	if order.PaymentID != uuid.Nil {
		paymentID = &order.PaymentID
	}
	if order.ShippingID != uuid.Nil {
		shippingID = &order.ShippingID
	}

	_, err = tx.ExecContext(
		ctx,
		query,
		order.ID,
		order.CustomerID,
		string(order.Status),
		order.TotalAmount,
		order.Currency,
		order.BillingAddressID,
		order.ShippingAddressID,
		paymentID,
		shippingID,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("注文の保存に失敗しました: %w", err)
	}

	// 注文アイテムを保存
	for _, item := range order.Items {
		err = r.saveOrderItem(ctx, tx, order.ID, item)
		if err != nil {
			return err
		}
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// saveOrderItem は注文アイテムを保存する
func (r *OrderRepository) saveOrderItem(ctx context.Context, tx *sql.Tx, orderID uuid.UUID, item order.OrderItem) error {
	query := `
		INSERT INTO order_items (
			id, order_id, product_id, product_name, quantity, unit_price, total_price
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.ExecContext(
		ctx,
		query,
		uuid.New(),
		orderID,
		item.ProductID,
		item.ProductName,
		item.Quantity,
		item.UnitPrice,
		item.TotalPrice,
	)
	if err != nil {
		return fmt.Errorf("注文アイテムの保存に失敗しました: %w", err)
	}
	return nil
}

// FindByID はIDで注文を検索する
func (r *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	// 注文情報を取得
	query := `
		SELECT 
			id, customer_id, status, total_amount, currency, billing_address_id, shipping_address_id,
			payment_id, shipping_id, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	o, err := r.scanOrder(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("注文が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("注文の取得に失敗しました: %w", err)
	}

	// 注文アイテムを取得
	items, err := r.findOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

// FindByCustomerID は顧客IDで注文を検索する
func (r *OrderRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*order.Order, error) {
	// 注文情報を取得
	query := `
		SELECT 
			id, customer_id, status, total_amount, currency, billing_address_id, shipping_address_id,
			payment_id, shipping_id, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("注文の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var orders []*order.Order
	for rows.Next() {
		var o order.Order
		var status string
		var paymentID, shippingID sql.NullString

		err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&status,
			&o.TotalAmount,
			&o.Currency,
			&o.BillingAddressID,
			&o.ShippingAddressID,
			&paymentID,
			&shippingID,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("注文のスキャンに失敗しました: %w", err)
		}

		o.Status = order.OrderStatus(status)
		
		if paymentID.Valid {
			pid, err := uuid.Parse(paymentID.String)
			if err == nil {
				o.PaymentID = pid
			}
		}
		
		if shippingID.Valid {
			sid, err := uuid.Parse(shippingID.String)
			if err == nil {
				o.ShippingID = sid
			}
		}

		// 注文アイテムを取得
		items, err := r.findOrderItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items

		orders = append(orders, &o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("注文の反復中にエラーが発生しました: %w", err)
	}

	return orders, nil
}

// Update は注文を更新する
func (r *OrderRepository) Update(ctx context.Context, order *order.Order) error {
	query := `
		UPDATE orders
		SET 
			status = $1, 
			total_amount = $2,
			payment_id = $3,
			shipping_id = $4,
			updated_at = $5
		WHERE id = $6
	`

	var paymentID, shippingID *uuid.UUID
	if order.PaymentID != uuid.Nil {
		paymentID = &order.PaymentID
	}
	if order.ShippingID != uuid.Nil {
		shippingID = &order.ShippingID
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		string(order.Status),
		order.TotalAmount,
		paymentID,
		shippingID,
		time.Now(),
		order.ID,
	)
	if err != nil {
		return fmt.Errorf("注文の更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新された行数の取得に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("更新対象の注文が見つかりません: %w", err)
	}

	return nil
}

// scanOrder はSQL行から注文エンティティを作成する
func (r *OrderRepository) scanOrder(row *sql.Row) (*order.Order, error) {
	var o order.Order
	var status string
	var paymentID, shippingID sql.NullString

	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&status,
		&o.TotalAmount,
		&o.Currency,
		&o.BillingAddressID,
		&o.ShippingAddressID,
		&paymentID,
		&shippingID,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	o.Status = order.OrderStatus(status)
	
	if paymentID.Valid {
		pid, err := uuid.Parse(paymentID.String)
		if err == nil {
			o.PaymentID = pid
		}
	}
	
	if shippingID.Valid {
		sid, err := uuid.Parse(shippingID.String)
		if err == nil {
			o.ShippingID = sid
		}
	}

	o.Items = []order.OrderItem{}

	return &o, nil
}

// findOrderItems は注文IDに関連する注文アイテムを取得する
func (r *OrderRepository) findOrderItems(ctx context.Context, orderID uuid.UUID) ([]order.OrderItem, error) {
	query := `
		SELECT product_id, product_name, quantity, unit_price, total_price
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("注文アイテムの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var items []order.OrderItem
	for rows.Next() {
		var item order.OrderItem
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
		); err != nil {
			return nil, fmt.Errorf("注文アイテムのスキャンに失敗しました: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("注文アイテムの反復中にエラーが発生しました: %w", err)
	}

	return items, nil
}

// Delete は注文を削除する（ソフトデリート）
func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE orders
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("注文の削除に失敗しました: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("注文が見つかりません、または既に削除されています")
	}
	
	return nil
} 