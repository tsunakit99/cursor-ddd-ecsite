package commands

import (
	"context"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/user"
	"github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus"
)

// RegisterUserCommand はユーザー登録コマンド
type RegisterUserCommand struct {
	Email       string
	Password    string
	FirstName   string
	LastName    string
	PhoneNumber string
}

// RegisterUserHandler はユーザー登録コマンドハンドラ
type RegisterUserHandler struct {
	userRepository user.Repository
	eventBus       eventbus.EventBus
}

// NewRegisterUserHandler は新しいユーザー登録ハンドラを作成する
func NewRegisterUserHandler(repo user.Repository, bus eventbus.EventBus) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepository: repo,
		eventBus:       bus,
	}
}

// Handle はユーザー登録コマンドを処理する
func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*user.User, error) {
	// 既存ユーザーのチェック
	existingUser, err := h.userRepository.FindByEmail(ctx, cmd.Email)
	if err == nil && existingUser != nil {
		return nil, user.ErrInvalidEmail // すでに存在するメールアドレス
	}

	// 新しいユーザーの作成
	newUser, err := user.NewUser(
		cmd.Email,
		cmd.Password,
		cmd.FirstName,
		cmd.LastName,
		cmd.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}

	// ユーザーの保存
	if err := h.userRepository.Save(ctx, newUser); err != nil {
		return nil, err
	}

	// イベント発行
	event := user.UserRegisteredEvent{
		BaseEvent: user.NewBaseEvent(newUser.ID),
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		// イベント発行の失敗はログに記録するが、ユーザー登録自体は成功とする
		return newUser, err
	}

	return newUser, nil
} 