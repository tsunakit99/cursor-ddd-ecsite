package user

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("無効なトークンです")
	ErrExpiredToken = errors.New("トークンの有効期限が切れています")
)

// AuthService は認証関連の処理を提供するサービス
type AuthService struct {
	jwtSecret         []byte
	tokenExpiry       time.Duration
	refreshTokenExpiry time.Duration
}

// NewAuthService は新しい認証サービスを作成する
func NewAuthService(jwtSecret string, tokenExpiry, refreshTokenExpiry time.Duration) *AuthService {
	return &AuthService{
		jwtSecret:         []byte(jwtSecret),
		tokenExpiry:       tokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// TokenClaims はJWTトークンの内容を表す
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

// AuthTokens は認証トークンのセットを表す
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

// GenerateTokens はユーザー情報からJWTトークンを生成する
func (s *AuthService) GenerateTokens(user *User) (*AuthTokens, error) {
	// トークンの有効期限を設定
	expiresAt := time.Now().Add(s.tokenExpiry)
	refreshExpiresAt := time.Now().Add(s.refreshTokenExpiry)
	
	// クレームの作成
	claims := TokenClaims{
		UserID:    user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}
	
	// アクセストークンの作成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	
	// リフレッシュトークンの作成（クレームは最小限にする）
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.ID.String(),
	}
	
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	
	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// ValidateToken はトークンを検証し、クレームを返す
func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	// トークンの解析
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	
	if err != nil {
		if _, ok := err.(*jwt.ValidationError); ok {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}
	
	// クレームの取り出し
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}

// RefreshToken はリフレッシュトークンを使用して新しいアクセストークンを生成する
func (s *AuthService) RefreshToken(refreshTokenString string, user *User) (*AuthTokens, error) {
	// リフレッシュトークンの検証
	token, err := jwt.ParseWithClaims(refreshTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	
	if err != nil {
		if _, ok := err.(*jwt.ValidationError); ok {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}
	
	// クレームの取り出し
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// ユーザーIDの検証
		userID, err := uuid.Parse(claims.Subject)
		if err != nil || userID != user.ID {
			return nil, ErrInvalidToken
		}
		
		// 新しいトークンの生成
		return s.GenerateTokens(user)
	}
	
	return nil, ErrInvalidToken
} 