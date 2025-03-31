package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/application/commands"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/user"
	pb "github.com/tsunakit99/cursor-ddd-ecsite/proto/user"
)

// UserServer はUserサービスのgRPC実装
type UserServer struct {
	pb.UnimplementedUserServiceServer
	registerUserHandler *commands.RegisterUserHandler
	loginUserHandler    *commands.LoginUserHandler
	userRepository      user.Repository
	authService         *user.AuthService
}

// NewUserServer は新しいUserサーバーを作成する
func NewUserServer(
	registerUserHandler *commands.RegisterUserHandler,
	loginUserHandler *commands.LoginUserHandler,
	userRepository user.Repository,
	authService *user.AuthService,
) *UserServer {
	return &UserServer{
		registerUserHandler: registerUserHandler,
		loginUserHandler:    loginUserHandler,
		userRepository:      userRepository,
		authService:         authService,
	}
}

// Register はユーザー登録を処理する
func (s *UserServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
	// リクエストの検証
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "メールアドレスは必須です")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "パスワードは必須です")
	}
	if req.FirstName == "" || req.LastName == "" {
		return nil, status.Error(codes.InvalidArgument, "名前は必須です")
	}

	// コマンドの作成
	cmd := commands.RegisterUserCommand{
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
	}

	// コマンドの実行
	registeredUser, err := s.registerUserHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, user.ErrInvalidEmail) {
			return nil, status.Error(codes.AlreadyExists, "このメールアドレスは既に使用されています")
		}
		if errors.Is(err, user.ErrInvalidPassword) {
			return nil, status.Error(codes.InvalidArgument, "パスワードが無効です")
		}
		if errors.Is(err, user.ErrInvalidName) {
			return nil, status.Error(codes.InvalidArgument, "名前が無効です")
		}
		return nil, status.Errorf(codes.Internal, "ユーザー登録に失敗しました: %v", err)
	}

	// レスポンスの作成
	return userToProto(registeredUser), nil
}

// Login はユーザーログインを処理する
func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// リクエストの検証
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "メールアドレスは必須です")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "パスワードは必須です")
	}

	// コマンドの作成
	cmd := commands.LoginUserCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	// コマンドの実行
	result, err := s.loginUserHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, commands.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "認証情報が無効です")
		}
		return nil, status.Errorf(codes.Internal, "ログインに失敗しました: %v", err)
	}

	// レスポンスの作成
	return &pb.LoginResponse{
		UserId:        result.User.ID.String(),
		Email:         result.User.Email,
		FirstName:     result.User.FirstName,
		LastName:      result.User.LastName,
		AuthToken:     result.AccessToken,
		RefreshToken:  result.RefreshToken,
		TokenExpiresAt: uint64(result.ExpiresAt),
	}, nil
}

// GetProfile はユーザープロファイルを取得する
func (s *UserServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "ユーザーIDは必須です")
	}

	// UUIDの変換
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効なユーザーID形式: %v", err)
	}

	// ユーザー情報の取得
	userEntity, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ユーザーが見つかりません: %v", err)
	}

	return userToProto(userEntity), nil
}

// UpdateProfile はユーザープロファイルを更新する
func (s *UserServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "ユーザーIDは必須です")
	}
	if req.FirstName == "" || req.LastName == "" {
		return nil, status.Error(codes.InvalidArgument, "名前は必須です")
	}

	// UUIDの変換
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効なユーザーID形式: %v", err)
	}

	// ユーザー情報の取得
	userEntity, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ユーザーが見つかりません: %v", err)
	}

	// ユーザー情報の更新
	if err := userEntity.UpdateProfile(req.FirstName, req.LastName, req.PhoneNumber); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "プロファイルの更新に失敗しました: %v", err)
	}

	// ユーザー情報の保存
	if err := s.userRepository.Update(ctx, userEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "ユーザー情報の保存に失敗しました: %v", err)
	}

	return userToProto(userEntity), nil
}

// ChangePassword はパスワードを変更する
func (s *UserServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*emptypb.Empty, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "ユーザーIDは必須です")
	}
	if req.CurrentPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "現在のパスワードは必須です")
	}
	if req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "新しいパスワードは必須です")
	}

	// UUIDの変換
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効なユーザーID形式: %v", err)
	}

	// ユーザー情報の取得
	userEntity, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ユーザーが見つかりません: %v", err)
	}

	// パスワードの変更
	if err := userEntity.ChangePassword(req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, user.ErrPasswordMismatch) {
			return nil, status.Error(codes.PermissionDenied, "現在のパスワードが一致しません")
		}
		if errors.Is(err, user.ErrInvalidPassword) {
			return nil, status.Error(codes.InvalidArgument, "新しいパスワードが無効です")
		}
		return nil, status.Errorf(codes.Internal, "パスワードの変更に失敗しました: %v", err)
	}

	// ユーザー情報の保存
	if err := s.userRepository.Update(ctx, userEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "ユーザー情報の保存に失敗しました: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// AddAddress は住所を追加する
func (s *UserServer) AddAddress(ctx context.Context, req *pb.AddAddressRequest) (*pb.AddressResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "ユーザーIDは必須です")
	}
	if req.Address == nil {
		return nil, status.Error(codes.InvalidArgument, "住所情報は必須です")
	}

	// UUIDの変換
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効なユーザーID形式: %v", err)
	}

	// ユーザー情報の取得
	userEntity, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ユーザーが見つかりません: %v", err)
	}

	// アドレスタイプの変換
	var addressType user.AddressType
	switch req.Address.AddressType {
	case "SHIPPING":
		addressType = user.AddressTypeShipping
	case "BILLING":
		addressType = user.AddressTypeBilling
	default:
		addressType = user.AddressTypeShipping
	}

	// 住所の作成
	address := user.Address{
		RecipientName: req.Address.RecipientName,
		StreetLine1:   req.Address.StreetLine1,
		StreetLine2:   req.Address.StreetLine2,
		City:          req.Address.City,
		State:         req.Address.State,
		PostalCode:    req.Address.PostalCode,
		Country:       req.Address.Country,
		PhoneNumber:   req.Address.PhoneNumber,
		IsDefault:     req.Address.IsDefault,
		AddressType:   addressType,
	}

	// 住所の追加
	userEntity.AddAddress(address)

	// 最後に追加された住所を取得
	newAddress := userEntity.Addresses[len(userEntity.Addresses)-1]

	// ユーザー情報の保存
	if err := s.userRepository.Update(ctx, userEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "ユーザー情報の保存に失敗しました: %v", err)
	}

	// 住所の保存
	if err := s.userRepository.SaveAddress(ctx, userID, newAddress); err != nil {
		return nil, status.Errorf(codes.Internal, "住所の保存に失敗しました: %v", err)
	}

	return addressToProto(userID, newAddress), nil
}

// UpdateAddress は住所を更新する
func (s *UserServer) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.AddressResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "ユーザーIDは必須です")
	}
	if req.AddressId == "" {
		return nil, status.Error(codes.InvalidArgument, "住所IDは必須です")
	}
	if req.Address == nil {
		return nil, status.Error(codes.InvalidArgument, "住所情報は必須です")
	}

	// UUIDの変換
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効なユーザーID形式: %v", err)
	}

	addressID, err := uuid.Parse(req.AddressId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な住所ID形式: %v", err)
	}

	// ユーザー情報の取得
	userEntity, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ユーザーが見つかりません: %v", err)
	}

	// アドレスタイプの変換
	var addressType user.AddressType
	switch req.Address.AddressType {
	case "SHIPPING":
		addressType = user.AddressTypeShipping
	case "BILLING":
		addressType = user.AddressTypeBilling
	default:
		addressType = user.AddressTypeShipping
	}

	// 更新する住所の作成
	updatedAddress := user.Address{
		RecipientName: req.Address.RecipientName,
		StreetLine1:   req.Address.StreetLine1,
		StreetLine2:   req.Address.StreetLine2,
		City:          req.Address.City,
		State:         req.Address.State,
		PostalCode:    req.Address.PostalCode,
		Country:       req.Address.Country,
		PhoneNumber:   req.Address.PhoneNumber,
		IsDefault:     req.Address.IsDefault,
		AddressType:   addressType,
	}

	// 住所の更新
	if err := userEntity.UpdateAddress(addressID, updatedAddress); err != nil {
		return nil, status.Errorf(codes.NotFound, "住所の更新に失敗しました: %v", err)
	}

	// 更新後の住所を取得
	var updatedAddressEntity *user.Address
	for i, addr := range userEntity.Addresses {
		if addr.ID == addressID {
			updatedAddressEntity = &userEntity.Addresses[i]
			break
		}
	}

	if updatedAddressEntity == nil {
		return nil, status.Error(codes.Internal, "更新された住所が見つかりません")
	}

	// ユーザー情報の保存
	if err := s.userRepository.Update(ctx, userEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "ユーザー情報の保存に失敗しました: %v", err)
	}

	// 住所の更新
	if err := s.userRepository.UpdateAddress(ctx, userID, *updatedAddressEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "住所の更新に失敗しました: %v", err)
	}

	return addressToProto(userID, *updatedAddressEntity), nil
}

// DeleteAddress は住所を削除する
func (s *UserServer) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "ユーザーIDは必須です")
	}
	if req.AddressId == "" {
		return nil, status.Error(codes.InvalidArgument, "住所IDは必須です")
	}

	// UUIDの変換
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効なユーザーID形式: %v", err)
	}

	addressID, err := uuid.Parse(req.AddressId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な住所ID形式: %v", err)
	}

	// ユーザー情報の取得
	userEntity, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ユーザーが見つかりません: %v", err)
	}

	// 住所の削除
	if err := userEntity.RemoveAddress(addressID); err != nil {
		return nil, status.Errorf(codes.NotFound, "住所の削除に失敗しました: %v", err)
	}

	// ユーザー情報の保存
	if err := s.userRepository.Update(ctx, userEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "ユーザー情報の保存に失敗しました: %v", err)
	}

	// 住所の削除
	if err := s.userRepository.DeleteAddress(ctx, userID, addressID); err != nil {
		return nil, status.Errorf(codes.Internal, "住所の削除に失敗しました: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// GetAddresses は住所一覧を取得する
func (s *UserServer) GetAddresses(ctx context.Context, req *pb.GetAddressesRequest) (*pb.GetAddressesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "ユーザーIDは必須です")
	}

	// UUIDの変換
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効なユーザーID形式: %v", err)
	}

	// 住所の取得
	addresses, err := s.userRepository.FindAddressesByUserID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "住所の取得に失敗しました: %v", err)
	}

	// レスポンスの作成
	addressResponses := make([]*pb.AddressResponse, len(addresses))
	for i, addr := range addresses {
		addressResponses[i] = addressToProto(userID, addr)
	}

	return &pb.GetAddressesResponse{
		Addresses: addressResponses,
	}, nil
}

// userToProto はドメインユーザーをProtobufメッセージに変換する
func userToProto(u *user.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:          u.ID.String(),
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		PhoneNumber: u.PhoneNumber,
		CreatedAt:   timestamppb.New(u.CreatedAt),
		UpdatedAt:   timestamppb.New(u.UpdatedAt),
	}
}

// addressToProto はドメイン住所をProtobufメッセージに変換する
func addressToProto(userID uuid.UUID, addr user.Address) *pb.AddressResponse {
	addressType := "SHIPPING"
	if addr.AddressType == user.AddressTypeBilling {
		addressType = "BILLING"
	}

	return &pb.AddressResponse{
		Id:     addr.ID.String(),
		UserId: userID.String(),
		Address: &pb.Address{
			RecipientName: addr.RecipientName,
			StreetLine1:   addr.StreetLine1,
			StreetLine2:   addr.StreetLine2,
			City:          addr.City,
			State:         addr.State,
			PostalCode:    addr.PostalCode,
			Country:       addr.Country,
			PhoneNumber:   addr.PhoneNumber,
			IsDefault:     addr.IsDefault,
			AddressType:   addressType,
		},
		CreatedAt: timestamppb.New(addr.CreatedAt),
		UpdatedAt: timestamppb.New(addr.UpdatedAt),
	}
} 