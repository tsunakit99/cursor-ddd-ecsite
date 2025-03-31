package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository はユーザーリポジトリのインターフェース
type Repository interface {
	// Save はユーザーを保存する
	Save(ctx context.Context, user *User) error
	
	// FindByID はIDでユーザーを検索する
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	
	// FindByEmail はメールアドレスでユーザーを検索する
	FindByEmail(ctx context.Context, email string) (*User, error)
	
	// Update はユーザー情報を更新する
	Update(ctx context.Context, user *User) error
	
	// Delete はユーザーを削除する（ソフトデリート）
	Delete(ctx context.Context, id uuid.UUID) error
	
	// SaveAddress はユーザーの住所を保存する
	SaveAddress(ctx context.Context, userID uuid.UUID, address Address) error
	
	// UpdateAddress はユーザーの住所を更新する
	UpdateAddress(ctx context.Context, userID uuid.UUID, address Address) error
	
	// DeleteAddress はユーザーの住所を削除する
	DeleteAddress(ctx context.Context, userID uuid.UUID, addressID uuid.UUID) error
	
	// FindAddressesByUserID はユーザーIDに紐づく住所リストを取得する
	FindAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]Address, error)
} 