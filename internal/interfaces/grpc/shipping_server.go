package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/shipping"
	pb "github.com/tsunakit99/cursor-ddd-ecsite/internal/interfaces/grpc/shipping"
)

// ShippingServer はShippingサービスのgRPC実装
type ShippingServer struct {
	pb.UnimplementedShippingServiceServer
	shippingRepository shipping.Repository
	// TODO: 適切なコマンドハンドラを追加
}

// NewShippingServer は新しいShippingサーバーを作成する
func NewShippingServer(
	shippingRepository shipping.Repository,
) *ShippingServer {
	return &ShippingServer{
		shippingRepository: shippingRepository,
	}
}

// CreateShipping は配送を作成する
func (s *ShippingServer) CreateShipping(ctx context.Context, req *pb.CreateShippingRequest) (*pb.ShippingResponse, error) {
	// リクエストの検証
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}
	if req.ShippingAddress == nil {
		return nil, status.Error(codes.InvalidArgument, "配送先住所は必須です")
	}

	// UUIDの変換
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// 住所の変換
	address := shipping.Address{
		RecipientName: req.ShippingAddress.RecipientName,
		StreetLine1:   req.ShippingAddress.StreetLine1,
		StreetLine2:   req.ShippingAddress.StreetLine2,
		City:          req.ShippingAddress.City,
		State:         req.ShippingAddress.State,
		PostalCode:    req.ShippingAddress.PostalCode,
		Country:       req.ShippingAddress.Country,
		PhoneNumber:   req.ShippingAddress.PhoneNumber,
	}

	// 配送方法の変換
	var method shipping.ShippingMethod
	switch req.Method {
	case pb.ShippingMethod_SHIPPING_METHOD_STANDARD:
		method = shipping.ShippingMethodStandard
	case pb.ShippingMethod_SHIPPING_METHOD_EXPRESS:
		method = shipping.ShippingMethodExpress
	case pb.ShippingMethod_SHIPPING_METHOD_OVERNIGHT:
		method = shipping.ShippingMethodOvernight
	default:
		return nil, status.Error(codes.InvalidArgument, "無効な配送方法")
	}

	// 配送の作成
	newShipping, err := shipping.NewShipping(orderID, address, method)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "配送の作成に失敗しました: %v", err)
	}

	// 配送の保存
	if err := s.shippingRepository.Save(ctx, newShipping); err != nil {
		return nil, status.Errorf(codes.Internal, "配送の保存に失敗しました: %v", err)
	}

	// レスポンスの作成
	return shippingToProto(newShipping), nil
}

// MarkAsReady は配送を準備完了にする
func (s *ShippingServer) MarkAsReady(ctx context.Context, req *pb.MarkAsReadyRequest) (*pb.ShippingResponse, error) {
	if req.ShippingId == "" {
		return nil, status.Error(codes.InvalidArgument, "配送IDは必須です")
	}

	shippingID, err := uuid.Parse(req.ShippingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な配送ID形式: %v", err)
	}

	// 配送情報の取得
	shippingEntity, err := s.shippingRepository.FindByID(ctx, shippingID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配送が見つかりません: %v", err)
	}

	// 配送を準備完了にする
	if err := shippingEntity.MarkAsReady(); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "配送の準備完了設定に失敗しました: %v", err)
	}

	// 配送情報の更新
	if err := s.shippingRepository.Update(ctx, shippingEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "配送情報の更新に失敗しました: %v", err)
	}

	return shippingToProto(shippingEntity), nil
}

// MarkAsShipped は配送を発送済みにする
func (s *ShippingServer) MarkAsShipped(ctx context.Context, req *pb.MarkAsShippedRequest) (*pb.ShippingResponse, error) {
	if req.ShippingId == "" {
		return nil, status.Error(codes.InvalidArgument, "配送IDは必須です")
	}
	if req.TrackingNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "追跡番号は必須です")
	}
	if req.Carrier == "" {
		return nil, status.Error(codes.InvalidArgument, "配送業者は必須です")
	}

	shippingID, err := uuid.Parse(req.ShippingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な配送ID形式: %v", err)
	}

	// 配送情報の取得
	shippingEntity, err := s.shippingRepository.FindByID(ctx, shippingID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配送が見つかりません: %v", err)
	}

	// 配達予定日の変換
	var estimatedDeliveryDate *time.Time
	if req.EstimatedDeliveryDate != nil {
		t := req.EstimatedDeliveryDate.AsTime()
		estimatedDeliveryDate = &t
	}

	// 配送を発送済みにする
	if err := shippingEntity.MarkAsShipped(req.TrackingNumber, req.Carrier, estimatedDeliveryDate); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "配送の発送済み設定に失敗しました: %v", err)
	}

	// 配送情報の更新
	if err := s.shippingRepository.Update(ctx, shippingEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "配送情報の更新に失敗しました: %v", err)
	}

	return shippingToProto(shippingEntity), nil
}

// MarkAsDelivered は配送を配達済みにする
func (s *ShippingServer) MarkAsDelivered(ctx context.Context, req *pb.MarkAsDeliveredRequest) (*pb.ShippingResponse, error) {
	if req.ShippingId == "" {
		return nil, status.Error(codes.InvalidArgument, "配送IDは必須です")
	}
	if req.DeliveryDate == nil {
		return nil, status.Error(codes.InvalidArgument, "配達日は必須です")
	}

	shippingID, err := uuid.Parse(req.ShippingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な配送ID形式: %v", err)
	}

	// 配送情報の取得
	shippingEntity, err := s.shippingRepository.FindByID(ctx, shippingID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配送が見つかりません: %v", err)
	}

	// 配送を配達済みにする
	if err := shippingEntity.MarkAsDelivered(req.DeliveryDate.AsTime(), req.ProofOfDelivery); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "配送の配達済み設定に失敗しました: %v", err)
	}

	// 配送情報の更新
	if err := s.shippingRepository.Update(ctx, shippingEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "配送情報の更新に失敗しました: %v", err)
	}

	return shippingToProto(shippingEntity), nil
}

// MarkAsFailed は配送を失敗にする
func (s *ShippingServer) MarkAsFailed(ctx context.Context, req *pb.MarkAsFailedRequest) (*pb.ShippingResponse, error) {
	if req.ShippingId == "" {
		return nil, status.Error(codes.InvalidArgument, "配送IDは必須です")
	}
	if req.Reason == "" {
		return nil, status.Error(codes.InvalidArgument, "失敗理由は必須です")
	}

	shippingID, err := uuid.Parse(req.ShippingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な配送ID形式: %v", err)
	}

	// 配送情報の取得
	shippingEntity, err := s.shippingRepository.FindByID(ctx, shippingID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配送が見つかりません: %v", err)
	}

	// 配送を失敗にする
	if err := shippingEntity.MarkAsFailed(req.Reason); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "配送の失敗設定に失敗しました: %v", err)
	}

	// 配送情報の更新
	if err := s.shippingRepository.Update(ctx, shippingEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "配送情報の更新に失敗しました: %v", err)
	}

	return shippingToProto(shippingEntity), nil
}

// GetShipping は配送情報を取得する
func (s *ShippingServer) GetShipping(ctx context.Context, req *pb.GetShippingRequest) (*pb.ShippingResponse, error) {
	if req.ShippingId == "" {
		return nil, status.Error(codes.InvalidArgument, "配送IDは必須です")
	}

	shippingID, err := uuid.Parse(req.ShippingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な配送ID形式: %v", err)
	}

	// 配送情報の取得
	shippingEntity, err := s.shippingRepository.FindByID(ctx, shippingID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配送が見つかりません: %v", err)
	}

	return shippingToProto(shippingEntity), nil
}

// GetShippingByOrderId は注文IDに紐づく配送情報を取得する
func (s *ShippingServer) GetShippingByOrderId(ctx context.Context, req *pb.GetShippingByOrderIdRequest) (*pb.ShippingResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// 配送情報の取得
	shippingEntity, err := s.shippingRepository.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配送が見つかりません: %v", err)
	}

	return shippingToProto(shippingEntity), nil
}

// TrackShipping は配送を追跡する
func (s *ShippingServer) TrackShipping(ctx context.Context, req *pb.TrackShippingRequest) (*pb.TrackingResponse, error) {
	if req.ShippingId == "" {
		return nil, status.Error(codes.InvalidArgument, "配送IDは必須です")
	}

	shippingID, err := uuid.Parse(req.ShippingId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な配送ID形式: %v", err)
	}

	// 配送情報の取得
	shippingEntity, err := s.shippingRepository.FindByID(ctx, shippingID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "配送が見つかりません: %v", err)
	}

	// 追跡イベントの変換
	events := make([]*pb.TrackingEvent, len(shippingEntity.TrackingEvents))
	for i, event := range shippingEntity.TrackingEvents {
		events[i] = &pb.TrackingEvent{
			Description: event.Description,
			Location:    event.Location,
			Timestamp:   timestamppb.New(event.Timestamp),
		}
	}

	// ステータスの変換
	var status pb.ShippingStatus
	var statusDescription string
	switch shippingEntity.Status {
	case shipping.ShippingStatusCreated:
		status = pb.ShippingStatus_SHIPPING_STATUS_CREATED
		statusDescription = "配送が作成されました"
	case shipping.ShippingStatusReady:
		status = pb.ShippingStatus_SHIPPING_STATUS_READY
		statusDescription = "配送の準備ができました"
	case shipping.ShippingStatusShipped:
		status = pb.ShippingStatus_SHIPPING_STATUS_SHIPPED
		statusDescription = "配送が発送されました"
	case shipping.ShippingStatusInTransit:
		status = pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT
		statusDescription = "配送が輸送中です"
	case shipping.ShippingStatusOutForDelivery:
		status = pb.ShippingStatus_SHIPPING_STATUS_OUT_FOR_DELIVERY
		statusDescription = "配送が配達中です"
	case shipping.ShippingStatusDelivered:
		status = pb.ShippingStatus_SHIPPING_STATUS_DELIVERED
		statusDescription = "配送が配達されました"
	case shipping.ShippingStatusFailed:
		status = pb.ShippingStatus_SHIPPING_STATUS_FAILED
		statusDescription = "配送が失敗しました"
	default:
		status = pb.ShippingStatus_SHIPPING_STATUS_UNKNOWN
		statusDescription = "不明な配送状態"
	}

	// 追跡レスポンスの作成
	return &pb.TrackingResponse{
		ShippingId:          shippingEntity.ID.String(),
		TrackingNumber:      shippingEntity.TrackingNumber,
		Carrier:             shippingEntity.Carrier,
		Status:              status,
		StatusDescription:   statusDescription,
		EstimatedDeliveryDate: timestamppb.New(time.Time{}),
		Events:                events,
	}, nil
}

// shippingToProto はドメインモデルをProtobufメッセージに変換する
func shippingToProto(s *shipping.Shipping) *pb.ShippingResponse {
	// 住所の変換
	address := &pb.Address{
		RecipientName: s.ShippingAddress.RecipientName,
		StreetLine1:   s.ShippingAddress.StreetLine1,
		StreetLine2:   s.ShippingAddress.StreetLine2,
		City:          s.ShippingAddress.City,
		State:         s.ShippingAddress.State,
		PostalCode:    s.ShippingAddress.PostalCode,
		Country:       s.ShippingAddress.Country,
		PhoneNumber:   s.ShippingAddress.PhoneNumber,
	}

	// 配送方法の変換
	var method pb.ShippingMethod
	switch s.Method {
	case shipping.ShippingMethodStandard:
		method = pb.ShippingMethod_SHIPPING_METHOD_STANDARD
	case shipping.ShippingMethodExpress:
		method = pb.ShippingMethod_SHIPPING_METHOD_EXPRESS
	case shipping.ShippingMethodOvernight:
		method = pb.ShippingMethod_SHIPPING_METHOD_OVERNIGHT
	default:
		method = pb.ShippingMethod_SHIPPING_METHOD_UNKNOWN
	}

	// 配送状態の変換
	var status pb.ShippingStatus
	switch s.Status {
	case shipping.ShippingStatusCreated:
		status = pb.ShippingStatus_SHIPPING_STATUS_CREATED
	case shipping.ShippingStatusReady:
		status = pb.ShippingStatus_SHIPPING_STATUS_READY
	case shipping.ShippingStatusShipped:
		status = pb.ShippingStatus_SHIPPING_STATUS_SHIPPED
	case shipping.ShippingStatusInTransit:
		status = pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT
	case shipping.ShippingStatusOutForDelivery:
		status = pb.ShippingStatus_SHIPPING_STATUS_OUT_FOR_DELIVERY
	case shipping.ShippingStatusDelivered:
		status = pb.ShippingStatus_SHIPPING_STATUS_DELIVERED
	case shipping.ShippingStatusFailed:
		status = pb.ShippingStatus_SHIPPING_STATUS_FAILED
	default:
		status = pb.ShippingStatus_SHIPPING_STATUS_UNKNOWN
	}

	// 日付の変換
	var estimatedDeliveryDate, actualDeliveryDate *timestamppb.Timestamp
	if s.EstimatedDeliveryDate != nil {
		estimatedDeliveryDate = timestamppb.New(*s.EstimatedDeliveryDate)
	}
	if s.ActualDeliveryDate != nil {
		actualDeliveryDate = timestamppb.New(*s.ActualDeliveryDate)
	}

	return &pb.ShippingResponse{
		Id:                    s.ID.String(),
		OrderId:               s.OrderID.String(),
		ShippingAddress:       address,
		Method:                method,
		Status:                status,
		TrackingNumber:        s.TrackingNumber,
		Carrier:               s.Carrier,
		EstimatedDeliveryDate: estimatedDeliveryDate,
		ActualDeliveryDate:    actualDeliveryDate,
		CreatedAt:             timestamppb.New(s.CreatedAt),
		UpdatedAt:             timestamppb.New(s.UpdatedAt),
	}
} 