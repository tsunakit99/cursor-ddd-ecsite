package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/inventory"
)

// InventoryRepository は在庫リポジトリのPostgres実装
type InventoryRepository struct {
	db *sqlx.DB
}

// NewInventoryRepository は新しい在庫リポジトリを作成する
func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{
		db: db,
	}
}

// FindByProductID は商品IDで在庫を検索する
func (r *InventoryRepository) FindByProductID(ctx context.Context, productID uuid.UUID) (*inventory.InventoryItem, error) {
	query := `
		SELECT product_id, product_name, available_quantity, reserved_quantity, backorder_quantity, last_updated
		FROM inventory_items
		WHERE product_id = $1
	`
	row := r.db.QueryRowxContext(ctx, query, productID)

	var item inventory.InventoryItem
	err := row.Scan(
		&item.ProductID,
		&item.ProductName,
		&item.AvailableQuantity,
		&item.ReservedQuantity,
		&item.BackorderQuantity,
		&item.LastUpdated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("在庫が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("在庫の取得に失敗しました: %w", err)
	}

	return &item, nil
}

// FindByProductIDs は複数の商品IDで在庫を検索する
func (r *InventoryRepository) FindByProductIDs(ctx context.Context, productIDs []uuid.UUID) ([]*inventory.InventoryItem, error) {
	if len(productIDs) == 0 {
		return []*inventory.InventoryItem{}, nil
	}

	// UUIDの配列をプレースホルダーの形式に変換
	args := make([]interface{}, len(productIDs))
	placeholders := make([]string, len(productIDs))
	for i, id := range productIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT product_id, product_name, available_quantity, reserved_quantity, backorder_quantity, last_updated
		FROM inventory_items
		WHERE product_id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("在庫の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var items []*inventory.InventoryItem
	for rows.Next() {
		var item inventory.InventoryItem
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.AvailableQuantity,
			&item.ReservedQuantity,
			&item.BackorderQuantity,
			&item.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("在庫のスキャンに失敗しました: %w", err)
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("在庫の反復中にエラーが発生しました: %w", err)
	}

	return items, nil
}

// Save は在庫を保存する
func (r *InventoryRepository) Save(ctx context.Context, item *inventory.InventoryItem) error {
	query := `
		INSERT INTO inventory_items (product_id, product_name, available_quantity, reserved_quantity, backorder_quantity, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ProductID,
		item.ProductName,
		item.AvailableQuantity,
		item.ReservedQuantity,
		item.BackorderQuantity,
		item.LastUpdated,
	)
	if err != nil {
		return fmt.Errorf("在庫の保存に失敗しました: %w", err)
	}
	return nil
}

// Update は在庫を更新する
func (r *InventoryRepository) Update(ctx context.Context, item *inventory.InventoryItem) error {
	query := `
		UPDATE inventory_items
		SET product_name = $1, available_quantity = $2, reserved_quantity = $3, backorder_quantity = $4, last_updated = $5
		WHERE product_id = $6
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		item.ProductName,
		item.AvailableQuantity,
		item.ReservedQuantity,
		item.BackorderQuantity,
		item.LastUpdated,
		item.ProductID,
	)
	if err != nil {
		return fmt.Errorf("在庫の更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新された行数の取得に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("更新対象の在庫が見つかりません: %w", err)
	}

	return nil
}

// ReservationRepository は在庫予約リポジトリのPostgres実装
type ReservationRepository struct {
	db *sqlx.DB
}

// NewReservationRepository は新しい予約リポジトリを作成する
func NewReservationRepository(db *sqlx.DB) *ReservationRepository {
	return &ReservationRepository{
		db: db,
	}
}

// Save は予約を保存する
func (r *ReservationRepository) Save(ctx context.Context, reservation *inventory.InventoryReservation) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	// 予約を保存
	query := `
		INSERT INTO inventory_reservations (id, order_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(
		ctx,
		query,
		reservation.ID,
		reservation.OrderID,
		string(reservation.Status),
		reservation.CreatedAt,
		reservation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("予約の保存に失敗しました: %w", err)
	}

	// 予約アイテムを保存
	for _, item := range reservation.Items {
		query = `
			INSERT INTO inventory_reservation_items (reservation_id, product_id, requested_quantity, available_quantity, reserved)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err = tx.ExecContext(
			ctx,
			query,
			reservation.ID,
			item.ProductID,
			item.RequestedQuantity,
			item.AvailableQuantity,
			item.Reserved,
		)
		if err != nil {
			return fmt.Errorf("予約アイテムの保存に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// FindByID はIDで予約を検索する
func (r *ReservationRepository) FindByID(ctx context.Context, id uuid.UUID) (*inventory.InventoryReservation, error) {
	// 予約データを取得
	query := `
		SELECT id, order_id, status, created_at, updated_at
		FROM inventory_reservations
		WHERE id = $1
	`
	row := r.db.QueryRowxContext(ctx, query, id)

	var reservation inventory.InventoryReservation
	var status string
	err := row.Scan(
		&reservation.ID,
		&reservation.OrderID,
		&status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("予約が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("予約の取得に失敗しました: %w", err)
	}
	reservation.Status = inventory.ReservationStatus(status)

	// 予約アイテムを取得
	query = `
		SELECT product_id, requested_quantity, available_quantity, reserved
		FROM inventory_reservation_items
		WHERE reservation_id = $1
	`
	rows, err := r.db.QueryxContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("予約アイテムの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var items []inventory.ReservationItem
	for rows.Next() {
		var item inventory.ReservationItem
		if err := rows.Scan(
			&item.ProductID,
			&item.RequestedQuantity,
			&item.AvailableQuantity,
			&item.Reserved,
		); err != nil {
			return nil, fmt.Errorf("予約アイテムのスキャンに失敗しました: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("予約アイテムの反復中にエラーが発生しました: %w", err)
	}
	reservation.Items = items

	return &reservation, nil
}

// FindByOrderID は注文IDで予約を検索する
func (r *ReservationRepository) FindByOrderID(ctx context.Context, orderID uuid.UUID) (*inventory.InventoryReservation, error) {
	// 予約データを取得
	query := `
		SELECT id, order_id, status, created_at, updated_at
		FROM inventory_reservations
		WHERE order_id = $1
	`
	row := r.db.QueryRowxContext(ctx, query, orderID)

	var reservation inventory.InventoryReservation
	var status string
	err := row.Scan(
		&reservation.ID,
		&reservation.OrderID,
		&status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("予約が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("予約の取得に失敗しました: %w", err)
	}
	reservation.Status = inventory.ReservationStatus(status)

	// 予約アイテムを取得
	query = `
		SELECT product_id, requested_quantity, available_quantity, reserved
		FROM inventory_reservation_items
		WHERE reservation_id = $1
	`
	rows, err := r.db.QueryxContext(ctx, query, reservation.ID)
	if err != nil {
		return nil, fmt.Errorf("予約アイテムの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var items []inventory.ReservationItem
	for rows.Next() {
		var item inventory.ReservationItem
		if err := rows.Scan(
			&item.ProductID,
			&item.RequestedQuantity,
			&item.AvailableQuantity,
			&item.Reserved,
		); err != nil {
			return nil, fmt.Errorf("予約アイテムのスキャンに失敗しました: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("予約アイテムの反復中にエラーが発生しました: %w", err)
	}
	reservation.Items = items

	return &reservation, nil
}

// Update は予約を更新する
func (r *ReservationRepository) Update(ctx context.Context, reservation *inventory.InventoryReservation) error {
	query := `
		UPDATE inventory_reservations
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		string(reservation.Status),
		reservation.UpdatedAt,
		reservation.ID,
	)
	if err != nil {
		return fmt.Errorf("予約の更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新された行数の取得に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("更新対象の予約が見つかりません: %w", err)
	}

	return nil
} 