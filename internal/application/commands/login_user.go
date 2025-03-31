package commands

import (
	"context"
	"errors"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/user"
	"github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus"
)

var (
	ErrInvalidCredentials = errors.New("認証情報が無効です")
)

// LoginUserCommand はユーザーログインコマンド
type LoginUserCommand struct {
	Email    string
	Password string
}

// LoginUserResult はログイン結果
type LoginUserResult struct {
	User       *user.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

// LoginUserHandler はユーザーログインコマンドハンドラ
type LoginUserHandler struct {
	userRepository user.Repository
	authService    *user.AuthService
	eventBus       eventbus.EventBus
}

// NewLoginUserHandler は新しいユーザーログインハンドラを作成する
func NewLoginUserHandler(repo user.Repository, auth *user.AuthService, bus eventbus.EventBus) *LoginUserHandler {
	return &LoginUserHandler{
		userRepository: repo,
		authService:    auth,
		eventBus:       bus,
	}
}

// Handle はユーザーログインコマンドを処理する
func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (*LoginUserResult, error) {
	// メールアドレスでユーザーを検索
	existingUser, err := h.userRepository.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// ユーザーが非アクティブまたはバンされている場合
	if !existingUser.IsActive() {
		return nil, ErrInvalidCredentials
	}

	// パスワードの検証
	if !existingUser.Authenticate(cmd.Password) {
		return nil, ErrInvalidCredentials
	}

	// トークンの生成
	tokens, err := h.authService.GenerateTokens(existingUser)
	if err != nil {
		return nil, err
	}

	// イベント発行
	event := user.UserLoggedInEvent{
		BaseEvent: user.NewBaseEvent(existingUser.ID),
		Email:     existingUser.Email,
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		// イベント発行の失敗はログに記録するが、ログイン自体は成功とする
	}

	return &LoginUserResult{
		User:         existingUser,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
} 