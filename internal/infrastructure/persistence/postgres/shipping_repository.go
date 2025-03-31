package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/shipping"
)

// ShippingRepository は配送リポジトリのPostgres実装
type ShippingRepository struct {
	db *sqlx.DB
}

// NewShippingRepository は新しい配送リポジトリを作成する
func NewShippingRepository(db *sqlx.DB) *ShippingRepository {
	return &ShippingRepository{
		db: db,
	}
}

// Save は配送を保存する
func (r *ShippingRepository) Save(ctx context.Context, shipping *shipping.Shipping) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	// 配送情報を保存
	query := `
		INSERT INTO shippings (
			id, order_id, recipient_name, street_line1, street_line2, city, state, postal_code, country, phone_number,
			method, status, tracking_number, carrier, estimated_delivery_date, actual_delivery_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	var estimatedDeliveryDate, actualDeliveryDate *time.Time
	if shipping.EstimatedDeliveryDate != nil {
		estimatedDeliveryDate = shipping.EstimatedDeliveryDate
	}
	if shipping.ActualDeliveryDate != nil {
		actualDeliveryDate = shipping.ActualDeliveryDate
	}

	_, err = tx.ExecContext(
		ctx,
		query,
		shipping.ID,
		shipping.OrderID,
		shipping.ShippingAddress.RecipientName,
		shipping.ShippingAddress.StreetLine1,
		shipping.ShippingAddress.StreetLine2,
		shipping.ShippingAddress.City,
		shipping.ShippingAddress.State,
		shipping.ShippingAddress.PostalCode,
		shipping.ShippingAddress.Country,
		shipping.ShippingAddress.PhoneNumber,
		string(shipping.Method),
		string(shipping.Status),
		shipping.TrackingNumber,
		shipping.Carrier,
		estimatedDeliveryDate,
		actualDeliveryDate,
		shipping.CreatedAt,
		shipping.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("配送の保存に失敗しました: %w", err)
	}

	// 追跡イベントを保存
	for _, event := range shipping.TrackingEvents {
		if err := r.saveTrackingEvent(ctx, tx, shipping.ID, event); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// FindByID はIDで配送を検索する
func (r *ShippingRepository) FindByID(ctx context.Context, id uuid.UUID) (*shipping.Shipping, error) {
	query := `
		SELECT 
			id, order_id, recipient_name, street_line1, street_line2, city, state, postal_code, country, phone_number,
			method, status, tracking_number, carrier, estimated_delivery_date, actual_delivery_date, created_at, updated_at
		FROM shippings
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	s, err := r.scanShipping(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("配送が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("配送の取得に失敗しました: %w", err)
	}

	// 追跡イベントを取得
	events, err := r.findTrackingEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	s.TrackingEvents = events

	return s, nil
}

// FindByOrderID は注文IDで配送を検索する
func (r *ShippingRepository) FindByOrderID(ctx context.Context, orderID uuid.UUID) (*shipping.Shipping, error) {
	query := `
		SELECT 
			id, order_id, recipient_name, street_line1, street_line2, city, state, postal_code, country, phone_number,
			method, status, tracking_number, carrier, estimated_delivery_date, actual_delivery_date, created_at, updated_at
		FROM shippings
		WHERE order_id = $1
	`
	row := r.db.QueryRowContext(ctx, query, orderID)

	s, err := r.scanShipping(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("配送が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("配送の取得に失敗しました: %w", err)
	}

	// 追跡イベントを取得
	events, err := r.findTrackingEvents(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.TrackingEvents = events

	return s, nil
}

// Update は配送を更新する
func (r *ShippingRepository) Update(ctx context.Context, shipping *shipping.Shipping) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE shippings
		SET 
			recipient_name = $1, street_line1 = $2, street_line2 = $3, city = $4, state = $5, postal_code = $6, country = $7, phone_number = $8,
			method = $9, status = $10, tracking_number = $11, carrier = $12, estimated_delivery_date = $13, actual_delivery_date = $14, updated_at = $15
		WHERE id = $16
	`

	var estimatedDeliveryDate, actualDeliveryDate *time.Time
	if shipping.EstimatedDeliveryDate != nil {
		estimatedDeliveryDate = shipping.EstimatedDeliveryDate
	}
	if shipping.ActualDeliveryDate != nil {
		actualDeliveryDate = shipping.ActualDeliveryDate
	}

	result, err := tx.ExecContext(
		ctx,
		query,
		shipping.ShippingAddress.RecipientName,
		shipping.ShippingAddress.StreetLine1,
		shipping.ShippingAddress.StreetLine2,
		shipping.ShippingAddress.City,
		shipping.ShippingAddress.State,
		shipping.ShippingAddress.PostalCode,
		shipping.ShippingAddress.Country,
		shipping.ShippingAddress.PhoneNumber,
		string(shipping.Method),
		string(shipping.Status),
		shipping.TrackingNumber,
		shipping.Carrier,
		estimatedDeliveryDate,
		actualDeliveryDate,
		shipping.UpdatedAt,
		shipping.ID,
	)
	if err != nil {
		return fmt.Errorf("配送の更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新された行数の取得に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("更新対象の配送が見つかりません: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// AddTrackingEvent は追跡イベントを追加する
func (r *ShippingRepository) AddTrackingEvent(ctx context.Context, shippingID uuid.UUID, event shipping.TrackingEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	if err := r.saveTrackingEvent(ctx, tx, shippingID, event); err != nil {
		return err
	}

	// 配送の更新日時を更新
	query := `
		UPDATE shippings
		SET updated_at = $1
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, query, time.Now(), shippingID)
	if err != nil {
		return fmt.Errorf("配送の更新に失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}

// saveTrackingEvent は追跡イベントを保存する
func (r *ShippingRepository) saveTrackingEvent(ctx context.Context, tx *sql.Tx, shippingID uuid.UUID, event shipping.TrackingEvent) error {
	query := `
		INSERT INTO shipping_tracking_events (id, shipping_id, description, location, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(
		ctx,
		query,
		uuid.New(), // イベントのID
		shippingID,
		event.Description,
		event.Location,
		event.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("追跡イベントの保存に失敗しました: %w", err)
	}
	return nil
}

// findTrackingEvents は配送IDに関連する追跡イベントを取得する
func (r *ShippingRepository) findTrackingEvents(ctx context.Context, shippingID uuid.UUID) ([]shipping.TrackingEvent, error) {
	query := `
		SELECT description, location, timestamp
		FROM shipping_tracking_events
		WHERE shipping_id = $1
		ORDER BY timestamp
	`
	rows, err := r.db.QueryContext(ctx, query, shippingID)
	if err != nil {
		return nil, fmt.Errorf("追跡イベントの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var events []shipping.TrackingEvent
	for rows.Next() {
		var event shipping.TrackingEvent
		if err := rows.Scan(&event.Description, &event.Location, &event.Timestamp); err != nil {
			return nil, fmt.Errorf("追跡イベントのスキャンに失敗しました: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("追跡イベントの反復中にエラーが発生しました: %w", err)
	}

	return events, nil
}

// scanShipping はSQL行から配送エンティティを作成する
func (r *ShippingRepository) scanShipping(row *sql.Row) (*shipping.Shipping, error) {
	var s shipping.Shipping
	var method, status string
	var recipientName, streetLine1, streetLine2, city, state, postalCode, country, phoneNumber string
	var estimatedDeliveryDate, actualDeliveryDate sql.NullTime

	err := row.Scan(
		&s.ID,
		&s.OrderID,
		&recipientName,
		&streetLine1,
		&streetLine2,
		&city,
		&state,
		&postalCode,
		&country,
		&phoneNumber,
		&method,
		&status,
		&s.TrackingNumber,
		&s.Carrier,
		&estimatedDeliveryDate,
		&actualDeliveryDate,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	s.ShippingAddress = shipping.Address{
		RecipientName: recipientName,
		StreetLine1:   streetLine1,
		StreetLine2:   streetLine2,
		City:          city,
		State:         state,
		PostalCode:    postalCode,
		Country:       country,
		PhoneNumber:   phoneNumber,
	}

	s.Method = shipping.ShippingMethod(method)
	s.Status = shipping.ShippingStatus(status)

	if estimatedDeliveryDate.Valid {
		t := estimatedDeliveryDate.Time
		s.EstimatedDeliveryDate = &t
	}

	if actualDeliveryDate.Valid {
		t := actualDeliveryDate.Time
		s.ActualDeliveryDate = &t
	}

	s.TrackingEvents = []shipping.TrackingEvent{}

	return &s, nil
} 