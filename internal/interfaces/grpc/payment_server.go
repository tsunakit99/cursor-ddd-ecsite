package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/payment"
	pb "github.com/tsunakit99/cursor-ddd-ecsite/internal/interfaces/grpc/payment"
)

// PaymentServer はPaymentサービスのgRPC実装
type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	paymentRepository payment.Repository
	// TODO: 適切なコマンドハンドラを追加
}

// NewPaymentServer は新しいPaymentサーバーを作成する
func NewPaymentServer(
	paymentRepository payment.Repository,
) *PaymentServer {
	return &PaymentServer{
		paymentRepository: paymentRepository,
	}
}

// CreatePayment は支払いを作成する
func (s *PaymentServer) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.PaymentResponse, error) {
	// リクエストの検証
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "金額は0より大きい必要があります")
	}
	if req.Currency == "" {
		return nil, status.Error(codes.InvalidArgument, "通貨は必須です")
	}

	// UUIDの変換
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// 支払い方法の変換
	var paymentMethod payment.PaymentMethod
	switch req.PaymentMethod {
	case pb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		paymentMethod = payment.PaymentMethodCreditCard
	case pb.PaymentMethod_PAYMENT_METHOD_PAYPAL:
		paymentMethod = payment.PaymentMethodPayPal
	case pb.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER:
		paymentMethod = payment.PaymentMethodBankTransfer
	default:
		return nil, status.Error(codes.InvalidArgument, "無効な支払い方法")
	}

	// 支払いの作成
	newPayment, err := payment.NewPayment(orderID, req.Amount, req.Currency, paymentMethod)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "支払いの作成に失敗しました: %v", err)
	}

	// 支払いの保存
	if err := s.paymentRepository.Save(ctx, newPayment); err != nil {
		return nil, status.Errorf(codes.Internal, "支払いの保存に失敗しました: %v", err)
	}

	// レスポンスの作成
	return paymentToProto(newPayment), nil
}

// ApprovePayment は支払いを承認する
func (s *PaymentServer) ApprovePayment(ctx context.Context, req *pb.ApprovePaymentRequest) (*pb.PaymentResponse, error) {
	if req.PaymentId == "" {
		return nil, status.Error(codes.InvalidArgument, "支払いIDは必須です")
	}

	paymentID, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な支払いID形式: %v", err)
	}

	// 支払い情報の取得
	paymentEntity, err := s.paymentRepository.FindByID(ctx, paymentID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "支払いが見つかりません: %v", err)
	}

	// トランザクションIDの取得（具体的な実装はここでは省略）
	var transactionID string
	switch {
	case req.PaymentDetails.GetCreditCard() != nil:
		// クレジットカード決済の処理
		// 実際の実装では、決済ゲートウェイとの連携が必要
		transactionID = uuid.New().String() // 仮のトランザクションID
	case req.PaymentDetails.GetPaypal() != nil:
		// PayPal決済の処理
		transactionID = uuid.New().String() // 仮のトランザクションID
	case req.PaymentDetails.GetBankTransfer() != nil:
		// 銀行振込の処理
		transactionID = uuid.New().String() // 仮のトランザクションID
	default:
		return nil, status.Error(codes.InvalidArgument, "支払い詳細が提供されていません")
	}

	// 支払いの承認
	if err := paymentEntity.Approve(transactionID); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "支払いの承認に失敗しました: %v", err)
	}

	// 支払い情報の更新
	if err := s.paymentRepository.Update(ctx, paymentEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "支払い情報の更新に失敗しました: %v", err)
	}

	return paymentToProto(paymentEntity), nil
}

// DeclinePayment は支払いを拒否する
func (s *PaymentServer) DeclinePayment(ctx context.Context, req *pb.DeclinePaymentRequest) (*pb.PaymentResponse, error) {
	if req.PaymentId == "" {
		return nil, status.Error(codes.InvalidArgument, "支払いIDは必須です")
	}

	paymentID, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な支払いID形式: %v", err)
	}

	// 支払い情報の取得
	paymentEntity, err := s.paymentRepository.FindByID(ctx, paymentID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "支払いが見つかりません: %v", err)
	}

	// 支払いの拒否
	if err := paymentEntity.Decline(req.Reason); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "支払いの拒否に失敗しました: %v", err)
	}

	// 支払い情報の更新
	if err := s.paymentRepository.Update(ctx, paymentEntity); err != nil {
		return nil, status.Errorf(codes.Internal, "支払い情報の更新に失敗しました: %v", err)
	}

	return paymentToProto(paymentEntity), nil
}

// GetPayment は支払い情報を取得する
func (s *PaymentServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.PaymentResponse, error) {
	if req.PaymentId == "" {
		return nil, status.Error(codes.InvalidArgument, "支払いIDは必須です")
	}

	paymentID, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な支払いID形式: %v", err)
	}

	// 支払い情報の取得
	paymentEntity, err := s.paymentRepository.FindByID(ctx, paymentID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "支払いが見つかりません: %v", err)
	}

	return paymentToProto(paymentEntity), nil
}

// GetPaymentByOrderId は注文IDに紐づく支払い情報を取得する
func (s *PaymentServer) GetPaymentByOrderId(ctx context.Context, req *pb.GetPaymentByOrderIdRequest) (*pb.PaymentResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// 支払い情報の取得
	paymentEntity, err := s.paymentRepository.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "支払いが見つかりません: %v", err)
	}

	return paymentToProto(paymentEntity), nil
}

// paymentToProto はドメインモデルをProtobufメッセージに変換する
func paymentToProto(p *payment.Payment) *pb.PaymentResponse {
	var status pb.PaymentStatus
	switch p.Status {
	case payment.PaymentStatusPending:
		status = pb.PaymentStatus_PAYMENT_STATUS_PENDING
	case payment.PaymentStatusApproved:
		status = pb.PaymentStatus_PAYMENT_STATUS_APPROVED
	case payment.PaymentStatusDeclined:
		status = pb.PaymentStatus_PAYMENT_STATUS_DECLINED
	case payment.PaymentStatusRefunded:
		status = pb.PaymentStatus_PAYMENT_STATUS_REFUNDED
	default:
		status = pb.PaymentStatus_PAYMENT_STATUS_UNKNOWN
	}

	var method pb.PaymentMethod
	switch p.Method {
	case payment.PaymentMethodCreditCard:
		method = pb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case payment.PaymentMethodPayPal:
		method = pb.PaymentMethod_PAYMENT_METHOD_PAYPAL
	case payment.PaymentMethodBankTransfer:
		method = pb.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER
	default:
		method = pb.PaymentMethod_PAYMENT_METHOD_UNKNOWN
	}

	return &pb.PaymentResponse{
		Id:            p.ID.String(),
		OrderId:       p.OrderID.String(),
		Amount:        p.Amount,
		Currency:      p.Currency,
		PaymentMethod: method,
		Status:        status,
		TransactionId: p.TransactionID,
		ErrorMessage:  p.ErrorMessage,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
} 