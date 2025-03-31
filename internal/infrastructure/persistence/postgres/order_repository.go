package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
		INSERT INTO orders (id, customer_id, status, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(
		ctx,
		query,
		order.ID,
		order.CustomerID,
		string(order.Status),
		order.TotalPrice,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("注文の保存に失敗しました: %w", err)
	}

	// 注文アイテムを保存
	for _, item := range order.Items {
		query = `
			INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, total_price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.ExecContext(
			ctx,
			query,
			uuid.New(), // 注文アイテムのID
			order.ID,
			item.ProductID,
			item.Quantity,
			item.UnitPrice,
			item.TotalPrice,
		)
		if err != nil {
			return fmt.Errorf("注文アイテムの保存に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// FindByID は指定されたIDの注文を取得する
func (r *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	query := `
		SELECT id, customer_id, status, total_price, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var o order.Order
	var status string
	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&status,
		&o.TotalPrice,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("注文が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("注文の取得に失敗しました: %w", err)
	}
	o.Status = order.OrderStatus(status)

	// 注文アイテムを取得
	query = `
		SELECT product_id, quantity, unit_price, total_price
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("注文アイテムの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var items []order.OrderItem
	for rows.Next() {
		var item order.OrderItem
		if err := rows.Scan(
			&item.ProductID,
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
	o.Items = items

	return &o, nil
}

// FindByCustomerID は指定された顧客IDの注文リストを取得する
func (r *OrderRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID, offset, limit int) ([]*order.Order, int, error) {
	// 総件数を取得
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM orders
		WHERE customer_id = $1
	`
	err := r.db.QueryRowContext(ctx, countQuery, customerID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("注文数の取得に失敗しました: %w", err)
	}

	// 注文リストを取得
	query := `
		SELECT id, customer_id, status, total_price, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("注文リストの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var orders []*order.Order
	orderIDs := make([]uuid.UUID, 0)
	orderMap := make(map[uuid.UUID]*order.Order)

	for rows.Next() {
		var o order.Order
		var status string
		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&status,
			&o.TotalPrice,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("注文のスキャンに失敗しました: %w", err)
		}
		o.Status = order.OrderStatus(status)
		o.Items = []order.OrderItem{}
		
		orders = append(orders, &o)
		orderIDs = append(orderIDs, o.ID)
		orderMap[o.ID] = &o
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("注文の反復中にエラーが発生しました: %w", err)
	}

	// 注文アイテムを取得（N+1問題を回避するためにIN句を使用）
	if len(orderIDs) > 0 {
		query = `
			SELECT order_id, product_id, quantity, unit_price, total_price
			FROM order_items
			WHERE order_id = ANY($1)
		`
		itemRows, err := r.db.QueryContext(ctx, query, pq.Array(orderIDs))
		if err != nil {
			return nil, 0, fmt.Errorf("注文アイテムの取得に失敗しました: %w", err)
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var orderID uuid.UUID
			var item order.OrderItem
			if err := itemRows.Scan(
				&orderID,
				&item.ProductID,
				&item.Quantity,
				&item.UnitPrice,
				&item.TotalPrice,
			); err != nil {
				return nil, 0, fmt.Errorf("注文アイテムのスキャンに失敗しました: %w", err)
			}
			if order, ok := orderMap[orderID]; ok {
				order.Items = append(order.Items, item)
			}
		}
		if err := itemRows.Err(); err != nil {
			return nil, 0, fmt.Errorf("注文アイテムの反復中にエラーが発生しました: %w", err)
		}
	}

	return orders, totalCount, nil
}

// Update は注文を更新する
func (r *OrderRepository) Update(ctx context.Context, order *order.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	// 注文を更新
	query := `
		UPDATE orders
		SET status = $1, total_price = $2, updated_at = $3
		WHERE id = $4
	`
	result, err := tx.ExecContext(
		ctx,
		query,
		string(order.Status),
		order.TotalPrice,
		time.Now(),
		order.ID,
	)
	if err != nil {
		return fmt.Errorf("注文の更新に失敗しました: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("注文が見つかりません")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
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