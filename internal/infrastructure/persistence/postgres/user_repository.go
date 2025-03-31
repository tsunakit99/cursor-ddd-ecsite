package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/user"
)

// UserRepository はユーザーリポジトリのPostgres実装
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository は新しいユーザーリポジトリを作成する
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// UserDTO はユーザーのデータ転送オブジェクト
type UserDTO struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	PhoneNumber  string    `db:"phone_number"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// AddressDTO は住所のデータ転送オブジェクト
type AddressDTO struct {
	ID            string    `db:"id"`
	UserID        string    `db:"user_id"`
	RecipientName string    `db:"recipient_name"`
	StreetLine1   string    `db:"street_line1"`
	StreetLine2   string    `db:"street_line2"`
	City          string    `db:"city"`
	State         string    `db:"state"`
	PostalCode    string    `db:"postal_code"`
	Country       string    `db:"country"`
	PhoneNumber   string    `db:"phone_number"`
	IsDefault     bool      `db:"is_default"`
	AddressType   string    `db:"address_type"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// Save はユーザーを保存する
func (r *UserRepository) Save(ctx context.Context, user *user.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, phone_number, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID.String(),
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	// 住所の保存
	for _, addr := range user.Addresses {
		if err := r.SaveAddress(ctx, user.ID, addr); err != nil {
			return err
		}
	}

	return nil
}

// FindByID はIDでユーザーを検索する
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone_number, status, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	var dto UserDTO
	err := r.db.GetContext(ctx, &dto, query, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ユーザーが見つかりません: %w", err)
		}
		return nil, fmt.Errorf("ユーザーの検索に失敗しました: %w", err)
	}

	// ドメインモデルへの変換
	domainUser, err := r.toDomainUser(dto)
	if err != nil {
		return nil, err
	}

	// 住所の取得
	addresses, err := r.FindAddressesByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	domainUser.Addresses = addresses
	return domainUser, nil
}

// FindByEmail はメールアドレスでユーザーを検索する
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone_number, status, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var dto UserDTO
	err := r.db.GetContext(ctx, &dto, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ユーザーが見つかりません: %w", err)
		}
		return nil, fmt.Errorf("ユーザーの検索に失敗しました: %w", err)
	}

	// ドメインモデルへの変換
	domainUser, err := r.toDomainUser(dto)
	if err != nil {
		return nil, err
	}

	// ユーザーIDのパース
	userID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("無効なユーザーID: %w", err)
	}

	// 住所の取得
	addresses, err := r.FindAddressesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	domainUser.Addresses = addresses
	return domainUser, nil
}

// Update はユーザー情報を更新する
func (r *UserRepository) Update(ctx context.Context, user *user.User) error {
	query := `
		UPDATE users 
		SET email = $1, password_hash = $2, first_name = $3, last_name = $4, phone_number = $5, 
			status = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.Status,
		user.UpdatedAt,
		user.ID.String(),
	)

	if err != nil {
		return fmt.Errorf("ユーザーの更新に失敗しました: %w", err)
	}

	return nil
}

// Delete はユーザーを削除する（ソフトデリート）
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// ソフトデリートの実装（statusをINACTIVEに変更する）
	query := `
		UPDATE users 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.UserStatusInactive,
		time.Now(),
		id.String(),
	)

	if err != nil {
		return fmt.Errorf("ユーザーの削除に失敗しました: %w", err)
	}

	return nil
}

// SaveAddress はユーザーの住所を保存する
func (r *UserRepository) SaveAddress(ctx context.Context, userID uuid.UUID, address user.Address) error {
	query := `
		INSERT INTO addresses (
			id, user_id, recipient_name, street_line1, street_line2, city, state, postal_code, 
			country, phone_number, is_default, address_type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		address.ID.String(),
		userID.String(),
		address.RecipientName,
		address.StreetLine1,
		address.StreetLine2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.PhoneNumber,
		address.IsDefault,
		address.AddressType,
		address.CreatedAt,
		address.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("住所の保存に失敗しました: %w", err)
	}

	return nil
}

// UpdateAddress はユーザーの住所を更新する
func (r *UserRepository) UpdateAddress(ctx context.Context, userID uuid.UUID, address user.Address) error {
	query := `
		UPDATE addresses 
		SET recipient_name = $1, street_line1 = $2, street_line2 = $3, city = $4, state = $5, 
			postal_code = $6, country = $7, phone_number = $8, is_default = $9, address_type = $10, 
			updated_at = $11
		WHERE id = $12 AND user_id = $13
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		address.RecipientName,
		address.StreetLine1,
		address.StreetLine2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.PhoneNumber,
		address.IsDefault,
		address.AddressType,
		address.UpdatedAt,
		address.ID.String(),
		userID.String(),
	)

	if err != nil {
		return fmt.Errorf("住所の更新に失敗しました: %w", err)
	}

	// 更新された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新結果の確認に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("住所が見つかりません")
	}

	return nil
}

// DeleteAddress はユーザーの住所を削除する
func (r *UserRepository) DeleteAddress(ctx context.Context, userID uuid.UUID, addressID uuid.UUID) error {
	query := `
		DELETE FROM addresses 
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, addressID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("住所の削除に失敗しました: %w", err)
	}

	// 削除された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("削除結果の確認に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("住所が見つかりません")
	}

	return nil
}

// FindAddressesByUserID はユーザーIDに紐づく住所リストを取得する
func (r *UserRepository) FindAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]user.Address, error) {
	query := `
		SELECT id, user_id, recipient_name, street_line1, street_line2, city, state, postal_code, 
			country, phone_number, is_default, address_type, created_at, updated_at
		FROM addresses 
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	var dtos []AddressDTO
	err := r.db.SelectContext(ctx, &dtos, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("住所の検索に失敗しました: %w", err)
	}

	// ドメインモデルへの変換
	addresses := make([]user.Address, len(dtos))
	for i, dto := range dtos {
		address, err := r.toDomainAddress(dto)
		if err != nil {
			return nil, err
		}
		addresses[i] = *address
	}

	return addresses, nil
}

// toDomainUser はDTOからドメインユーザーへ変換する
func (r *UserRepository) toDomainUser(dto UserDTO) (*user.User, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("無効なユーザーID: %w", err)
	}

	var status user.UserStatus
	switch dto.Status {
	case string(user.UserStatusActive):
		status = user.UserStatusActive
	case string(user.UserStatusInactive):
		status = user.UserStatusInactive
	case string(user.UserStatusBanned):
		status = user.UserStatusBanned
	default:
		status = user.UserStatusInactive
	}

	return &user.User{
		ID:           id,
		Email:        dto.Email,
		PasswordHash: dto.PasswordHash,
		FirstName:    dto.FirstName,
		LastName:     dto.LastName,
		PhoneNumber:  dto.PhoneNumber,
		Status:       status,
		Addresses:    []user.Address{},
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	}, nil
}

// toDomainAddress はDTOからドメイン住所へ変換する
func (r *UserRepository) toDomainAddress(dto AddressDTO) (*user.Address, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("無効な住所ID: %w", err)
	}

	var addressType user.AddressType
	switch dto.AddressType {
	case string(user.AddressTypeShipping):
		addressType = user.AddressTypeShipping
	case string(user.AddressTypeBilling):
		addressType = user.AddressTypeBilling
	default:
		addressType = user.AddressTypeShipping
	}

	return &user.Address{
		ID:            id,
		RecipientName: dto.RecipientName,
		StreetLine1:   dto.StreetLine1,
		StreetLine2:   dto.StreetLine2,
		City:          dto.City,
		State:         dto.State,
		PostalCode:    dto.PostalCode,
		Country:       dto.Country,
		PhoneNumber:   dto.PhoneNumber,
		IsDefault:     dto.IsDefault,
		AddressType:   addressType,
		CreatedAt:     dto.CreatedAt,
		UpdatedAt:     dto.UpdatedAt,
	}, nil
} 