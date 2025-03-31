package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/application/commands"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/order"
	pb "github.com/tsunakit99/cursor-ddd-ecsite/internal/interfaces/grpc/order"
)

// OrderServer はOrderサービスのgRPC実装
type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	createOrderHandler *commands.CreateOrderHandler
	orderRepository    order.Repository
}

// NewOrderServer は新しいOrderサーバーを作成する
func NewOrderServer(
	createOrderHandler *commands.CreateOrderHandler,
	orderRepository order.Repository,
) *OrderServer {
	return &OrderServer{
		createOrderHandler: createOrderHandler,
		orderRepository:    orderRepository,
	}
}

// CreateOrder は注文を作成する
func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	// リクエストの検証
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "顧客IDは必須です")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "注文アイテムは少なくとも1つ必要です")
	}

	// UUIDの変換
	customerID, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な顧客ID形式: %v", err)
	}

	// コマンドの作成
	items := make([]commands.CreateOrderItem, len(req.Items))
	for i, item := range req.Items {
		productID, err := uuid.Parse(item.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "無効な商品ID形式: %v", err)
		}
		
		items[i] = commands.CreateOrderItem{
			ProductID: productID,
			Quantity:  int(item.Quantity),
			UnitPrice: item.UnitPrice,
		}
	}

	cmd := commands.CreateOrderCommand{
		CustomerID: customerID,
		Items:      items,
	}

	// コマンドの実行
	createdOrder, err := s.createOrderHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "注文の作成に失敗しました: %v", err)
	}

	// レスポンスの作成
	return orderToProto(createdOrder), nil
}

// ConfirmOrder は注文を確認する
func (s *OrderServer) ConfirmOrder(ctx context.Context, req *pb.ConfirmOrderRequest) (*pb.OrderResponse, error) {
	// TODO: 注文確認コマンドハンドラを使用して実装
	return nil, status.Error(codes.Unimplemented, "未実装")
}

// CancelOrder は注文をキャンセルする
func (s *OrderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	// TODO: 注文キャンセルコマンドハンドラを使用して実装
	return nil, status.Error(codes.Unimplemented, "未実装")
}

// GetOrder は注文詳細を取得する
func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// リポジトリから注文を取得
	orderEntity, err := s.orderRepository.FindByID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "注文が見つかりません: %v", err)
	}

	return orderToProto(orderEntity), nil
}

// GetCustomerOrders は顧客の注文リストを取得する
func (s *OrderServer) GetCustomerOrders(ctx context.Context, req *pb.GetCustomerOrdersRequest) (*pb.OrderListResponse, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "顧客IDは必須です")
	}

	customerID, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な顧客ID形式: %v", err)
	}

	// ページネーションパラメータの準備
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	
	pageSize := int(req.PageSize)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // デフォルトページサイズ
	}
	
	offset := (page - 1) * pageSize

	// リポジトリから注文リストを取得
	orders, totalCount, err := s.orderRepository.FindByCustomerID(ctx, customerID, offset, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "注文リストの取得に失敗しました: %v", err)
	}

	// レスポンスの作成
	orderResponses := make([]*pb.OrderResponse, len(orders))
	for i, o := range orders {
		orderResponses[i] = orderToProto(o)
	}

	return &pb.OrderListResponse{
		Orders:     orderResponses,
		TotalCount: int32(totalCount),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}

// orderToProto はドメインモデルをProtobufメッセージに変換する
func orderToProto(o *order.Order) *pb.OrderResponse {
	items := make([]*pb.OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItemResponse{
			ProductId:  item.ProductID.String(),
			ProductName: "", // 注: 商品名はこの実装では取得できないため空欄
			Quantity:   int32(item.Quantity),
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		}
	}

	var status pb.OrderStatus
	switch o.Status {
	case order.OrderStatusPending:
		status = pb.OrderStatus_ORDER_STATUS_PENDING
	case order.OrderStatusConfirmed:
		status = pb.OrderStatus_ORDER_STATUS_CONFIRMED
	case order.OrderStatusPaid:
		status = pb.OrderStatus_ORDER_STATUS_PAID
	case order.OrderStatusShipped:
		status = pb.OrderStatus_ORDER_STATUS_SHIPPED
	case order.OrderStatusDelivered:
		status = pb.OrderStatus_ORDER_STATUS_DELIVERED
	case order.OrderStatusCanceled:
		status = pb.OrderStatus_ORDER_STATUS_CANCELED
	case order.OrderStatusBackordered:
		status = pb.OrderStatus_ORDER_STATUS_BACKORDERED
	default:
		status = pb.OrderStatus_ORDER_STATUS_UNKNOWN
	}

	return &pb.OrderResponse{
		Id:         o.ID.String(),
		CustomerId: o.CustomerID.String(),
		Status:     status,
		Items:      items,
		TotalPrice: o.TotalPrice,
		CreatedAt:  timestamppb.New(o.CreatedAt),
		UpdatedAt:  timestamppb.New(o.UpdatedAt),
	}
} 