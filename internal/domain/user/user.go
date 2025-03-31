package user

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// 定義済みエラー
var (
	ErrInvalidEmail      = errors.New("無効なメールアドレスです")
	ErrInvalidPassword   = errors.New("無効なパスワードです")
	ErrInvalidName       = errors.New("無効な名前です")
	ErrPasswordMismatch  = errors.New("パスワードが一致しません")
)

// User はユーザー（顧客）のエンティティ
type User struct {
	ID          uuid.UUID
	Email       string
	PasswordHash string
	FirstName   string
	LastName    string
	PhoneNumber string
	Status      UserStatus
	Addresses   []Address
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserStatus はユーザーの状態を表す列挙型
type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusBanned   UserStatus = "BANNED"
)

// Address は住所を表す値オブジェクト
type Address struct {
	ID           uuid.UUID
	RecipientName string
	StreetLine1   string
	StreetLine2   string
	City          string
	State         string
	PostalCode    string
	Country       string
	PhoneNumber   string
	IsDefault     bool
	AddressType   AddressType
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AddressType は住所の種類を表す列挙型
type AddressType string

const (
	AddressTypeShipping AddressType = "SHIPPING"
	AddressTypeBilling  AddressType = "BILLING"
)

// NewUser は新しいユーザーを作成する
func NewUser(email, password, firstName, lastName, phoneNumber string) (*User, error) {
	// メールアドレスの検証
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	// パスワードの検証
	if !isValidPassword(password) {
		return nil, ErrInvalidPassword
	}

	// 名前の検証
	if firstName == "" || lastName == "" {
		return nil, ErrInvalidName
	}

	// パスワードのハッシュ化
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 新しいユーザーの作成
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
		FirstName:    firstName,
		LastName:     lastName,
		PhoneNumber:  phoneNumber,
		Status:       UserStatusActive,
		Addresses:    []Address{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// Authenticate はユーザーの認証を行う
func (u *User) Authenticate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ChangePassword はパスワードを変更する
func (u *User) ChangePassword(currentPassword, newPassword string) error {
	// 現在のパスワードの検証
	if !u.Authenticate(currentPassword) {
		return ErrPasswordMismatch
	}

	// 新しいパスワードの検証
	if !isValidPassword(newPassword) {
		return ErrInvalidPassword
	}

	// 新しいパスワードのハッシュ化
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// パスワードハッシュの更新
	u.PasswordHash = string(passwordHash)
	u.UpdatedAt = time.Now()
	return nil
}

// AddAddress は住所を追加する
func (u *User) AddAddress(address Address) {
	// 新しい住所を追加
	address.ID = uuid.New()
	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()
	
	// 初めての住所の場合はデフォルトに設定
	if len(u.Addresses) == 0 {
		address.IsDefault = true
	}
	
	// デフォルト設定の場合、他の同タイプの住所からデフォルト設定を解除
	if address.IsDefault {
		for i := range u.Addresses {
			if u.Addresses[i].AddressType == address.AddressType {
				u.Addresses[i].IsDefault = false
			}
		}
	}
	
	u.Addresses = append(u.Addresses, address)
	u.UpdatedAt = time.Now()
}

// UpdateAddress は住所を更新する
func (u *User) UpdateAddress(addressID uuid.UUID, updatedAddress Address) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			// IDと作成日時は保持
			updatedAddress.ID = addr.ID
			updatedAddress.CreatedAt = addr.CreatedAt
			updatedAddress.UpdatedAt = time.Now()
			
			// デフォルト設定の場合、他の同タイプの住所からデフォルト設定を解除
			if updatedAddress.IsDefault && !addr.IsDefault {
				for j := range u.Addresses {
					if u.Addresses[j].AddressType == updatedAddress.AddressType {
						u.Addresses[j].IsDefault = false
					}
				}
			}
			
			u.Addresses[i] = updatedAddress
			u.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return errors.New("住所が見つかりません")
}

// RemoveAddress は住所を削除する
func (u *User) RemoveAddress(addressID uuid.UUID) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			// 住所の削除
			u.Addresses = append(u.Addresses[:i], u.Addresses[i+1:]...)
			u.UpdatedAt = time.Now()
			
			// 削除した住所がデフォルトだった場合、同タイプの最初の住所をデフォルトに設定
			if addr.IsDefault {
				for j, a := range u.Addresses {
					if a.AddressType == addr.AddressType {
						u.Addresses[j].IsDefault = true
						break
					}
				}
			}
			
			return nil
		}
	}
	
	return errors.New("住所が見つかりません")
}

// GetDefaultAddress は指定されたタイプのデフォルト住所を取得する
func (u *User) GetDefaultAddress(addressType AddressType) *Address {
	for _, addr := range u.Addresses {
		if addr.AddressType == addressType && addr.IsDefault {
			return &addr
		}
	}
	return nil
}

// UpdateProfile はユーザープロフィールを更新する
func (u *User) UpdateProfile(firstName, lastName, phoneNumber string) error {
	// 名前の検証
	if firstName == "" || lastName == "" {
		return ErrInvalidName
	}
	
	u.FirstName = firstName
	u.LastName = lastName
	u.PhoneNumber = phoneNumber
	u.UpdatedAt = time.Now()
	return nil
}

// Ban はユーザーをバン状態にする
func (u *User) Ban() {
	u.Status = UserStatusBanned
	u.UpdatedAt = time.Now()
}

// Deactivate はユーザーを非アクティブ状態にする
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now()
}

// Activate はユーザーをアクティブ状態にする
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

// IsActive はユーザーがアクティブかどうかを確認する
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// isValidEmail はメールアドレスが有効かどうかを検証する
func isValidEmail(email string) bool {
	// 簡易的なメールアドレスの検証
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return re.MatchString(email)
}

// isValidPassword はパスワードが有効かどうかを検証する
func isValidPassword(password string) bool {
	// パスワードの長さのみ検証（実際の実装ではより複雑なルールを適用することができます）
	return len(password) >= 8
} 