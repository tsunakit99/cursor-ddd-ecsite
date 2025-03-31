package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/inventory"
	pb "github.com/tsunakit99/cursor-ddd-ecsite/proto/inventory"
)

// InventoryServer はInventoryサービスのgRPC実装
type InventoryServer struct {
	pb.UnimplementedInventoryServiceServer
	inventoryRepository inventory.InventoryRepository
	reservationRepository inventory.ReservationRepository
	// TODO: 適切なコマンドハンドラを追加
}

// NewInventoryServer は新しいInventoryサーバーを作成する
func NewInventoryServer(
	inventoryRepository inventory.InventoryRepository,
	reservationRepository inventory.ReservationRepository,
) *InventoryServer {
	return &InventoryServer{
		inventoryRepository: inventoryRepository,
		reservationRepository: reservationRepository,
	}
}

// ReserveInventory は在庫を確保する
func (s *InventoryServer) ReserveInventory(ctx context.Context, req *pb.ReserveInventoryRequest) (*pb.ReserveInventoryResponse, error) {
	// リクエストの検証
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "アイテムは少なくとも1つ必要です")
	}

	// UUIDの変換
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// 商品IDのリストを作成
	productIDs := make([]uuid.UUID, len(req.Items))
	for i, item := range req.Items {
		productID, err := uuid.Parse(item.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "無効な商品ID形式: %v", err)
		}
		productIDs[i] = productID
	}

	// 在庫の確認
	inventoryItems, err := s.inventoryRepository.FindByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "在庫情報の取得に失敗しました: %v", err)
	}

	// 商品IDから在庫情報へのマップを作成
	inventoryMap := make(map[uuid.UUID]*inventory.InventoryItem)
	for _, item := range inventoryItems {
		inventoryMap[item.ProductID] = item
	}

	// 予約結果と予約アイテムのリストを作成
	results := make([]*pb.InventoryReservationResult, len(req.Items))
	reservationItems := make([]inventory.ReservationItem, len(req.Items))
	allAvailable := true

	for i, reqItem := range req.Items {
		productID, _ := uuid.Parse(reqItem.ProductId)
		invItem, exists := inventoryMap[productID]
		
		// 商品が存在しない場合のチェック
		if !exists {
			results[i] = &pb.InventoryReservationResult{
				ProductId:         reqItem.ProductId,
				Available:         false,
				RequestedQuantity: int32(reqItem.Quantity),
				AvailableQuantity: 0,
				ReservationId:     "",
			}
			allAvailable = false
			continue
		}

		// 在庫が十分かチェック
		available := invItem.CanReserve(int(reqItem.Quantity))
		
		results[i] = &pb.InventoryReservationResult{
			ProductId:         reqItem.ProductId,
			Available:         available,
			RequestedQuantity: int32(reqItem.Quantity),
			AvailableQuantity: int32(invItem.AvailableQuantity),
			ReservationId:     "",
		}

		reservationItems[i] = inventory.ReservationItem{
			ProductID:         productID,
			RequestedQuantity: int(reqItem.Quantity),
			AvailableQuantity: invItem.AvailableQuantity,
			Reserved:          available,
		}

		if !available {
			allAvailable = false
		}
	}

	// 在庫予約の作成
	reservation := inventory.NewInventoryReservation(orderID, reservationItems)

	// 予約IDを結果に追加
	for i := range results {
		if results[i].Available {
			results[i].ReservationId = reservation.ID.String()
		}
	}

	// 予約の保存
	if err := s.reservationRepository.Save(ctx, reservation); err != nil {
		return nil, status.Errorf(codes.Internal, "予約の保存に失敗しました: %v", err)
	}

	// 在庫の予約（実際の在庫減少）
	if allAvailable {
		for _, item := range reservationItems {
			if item.Reserved {
				invItem := inventoryMap[item.ProductID]
				if err := invItem.Reserve(item.RequestedQuantity); err != nil {
					return nil, status.Errorf(codes.Internal, "在庫の予約に失敗しました: %v", err)
				}
				// 在庫の更新
				if err := s.inventoryRepository.Update(ctx, invItem); err != nil {
					return nil, status.Errorf(codes.Internal, "在庫の更新に失敗しました: %v", err)
				}
			}
		}
	}

	return &pb.ReserveInventoryResponse{
		Success: allAvailable,
		Results: results,
	}, nil
}

// ReleaseInventory は在庫を解放する
func (s *InventoryServer) ReleaseInventory(ctx context.Context, req *pb.ReleaseInventoryRequest) (*pb.ReleaseInventoryResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// 予約の取得
	reservation, err := s.reservationRepository.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "予約が見つかりません: %v", err)
	}

	// 商品IDのリストを解析
	var productIDs []uuid.UUID
	if len(req.ProductIds) > 0 {
		productIDs = make([]uuid.UUID, len(req.ProductIds))
		for i, idStr := range req.ProductIds {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "無効な商品ID形式: %v", err)
			}
			productIDs[i] = id
		}
	} else {
		// 商品IDが指定されていない場合は、予約のすべての商品を使用
		productIDs = make([]uuid.UUID, len(reservation.Items))
		for i, item := range reservation.Items {
			productIDs[i] = item.ProductID
		}
	}

	// 予約をキャンセル
	if err := reservation.Cancel(); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "予約のキャンセルに失敗しました: %v", err)
	}

	// 予約の更新
	if err := s.reservationRepository.Update(ctx, reservation); err != nil {
		return nil, status.Errorf(codes.Internal, "予約の更新に失敗しました: %v", err)
	}

	// 在庫の解放
	failedProductIDs := []string{}
	for _, productID := range productIDs {
		// 予約アイテムを検索
		var reservedQuantity int
		var found bool
		for _, item := range reservation.Items {
			if item.ProductID == productID && item.Reserved {
				reservedQuantity = item.RequestedQuantity
				found = true
				break
			}
		}

		if !found {
			failedProductIDs = append(failedProductIDs, productID.String())
			continue
		}

		// 在庫を取得
		inventoryItem, err := s.inventoryRepository.FindByProductID(ctx, productID)
		if err != nil {
			failedProductIDs = append(failedProductIDs, productID.String())
			continue
		}

		// 在庫を解放
		if err := inventoryItem.Release(reservedQuantity); err != nil {
			failedProductIDs = append(failedProductIDs, productID.String())
			continue
		}

		// 在庫の更新
		if err := s.inventoryRepository.Update(ctx, inventoryItem); err != nil {
			failedProductIDs = append(failedProductIDs, productID.String())
			continue
		}
	}

	return &pb.ReleaseInventoryResponse{
		Success:          len(failedProductIDs) == 0,
		FailedProductIds: failedProductIDs,
	}, nil
}

// CommitInventory は在庫を確定する
func (s *InventoryServer) CommitInventory(ctx context.Context, req *pb.CommitInventoryRequest) (*pb.CommitInventoryResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "注文IDは必須です")
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な注文ID形式: %v", err)
	}

	// 予約の取得
	reservation, err := s.reservationRepository.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "予約が見つかりません: %v", err)
	}

	// 予約を確定
	if err := reservation.Confirm(); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "予約の確定に失敗しました: %v", err)
	}

	// 予約の更新
	if err := s.reservationRepository.Update(ctx, reservation); err != nil {
		return nil, status.Errorf(codes.Internal, "予約の更新に失敗しました: %v", err)
	}

	// 確定処理（実際の在庫減少）
	failedProductIDs := []string{}
	for _, item := range reservation.Items {
		if !item.Reserved {
			failedProductIDs = append(failedProductIDs, item.ProductID.String())
			continue
		}

		// 在庫を取得
		inventoryItem, err := s.inventoryRepository.FindByProductID(ctx, item.ProductID)
		if err != nil {
			failedProductIDs = append(failedProductIDs, item.ProductID.String())
			continue
		}

		// 予約を確定
		if err := inventoryItem.Commit(item.RequestedQuantity); err != nil {
			failedProductIDs = append(failedProductIDs, item.ProductID.String())
			continue
		}

		// 在庫の更新
		if err := s.inventoryRepository.Update(ctx, inventoryItem); err != nil {
			failedProductIDs = append(failedProductIDs, item.ProductID.String())
			continue
		}
	}

	return &pb.CommitInventoryResponse{
		Success:          len(failedProductIDs) == 0,
		FailedProductIds: failedProductIDs,
	}, nil
}

// CheckInventory は商品の在庫状況を確認する
func (s *InventoryServer) CheckInventory(ctx context.Context, req *pb.CheckInventoryRequest) (*pb.InventoryItemResponse, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "商品IDは必須です")
	}

	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な商品ID形式: %v", err)
	}

	// 在庫を取得
	inventoryItem, err := s.inventoryRepository.FindByProductID(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "在庫が見つかりません: %v", err)
	}

	return inventoryItemToProto(inventoryItem), nil
}

// CheckInventoryBatch は複数商品の在庫状況を確認する
func (s *InventoryServer) CheckInventoryBatch(ctx context.Context, req *pb.CheckInventoryBatchRequest) (*pb.CheckInventoryBatchResponse, error) {
	if len(req.ProductIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "商品IDは少なくとも1つ必要です")
	}

	// 商品IDのリストを作成
	productIDs := make([]uuid.UUID, len(req.ProductIds))
	for i, idStr := range req.ProductIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "無効な商品ID形式: %v", err)
		}
		productIDs[i] = id
	}

	// 在庫の確認
	inventoryItems, err := s.inventoryRepository.FindByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "在庫情報の取得に失敗しました: %v", err)
	}

	// レスポンスの作成
	itemResponses := make([]*pb.InventoryItemResponse, len(inventoryItems))
	for i, item := range inventoryItems {
		itemResponses[i] = inventoryItemToProto(item)
	}

	return &pb.CheckInventoryBatchResponse{
		Items: itemResponses,
	}, nil
}

// AddInventory は在庫を増やす
func (s *InventoryServer) AddInventory(ctx context.Context, req *pb.AddInventoryRequest) (*pb.InventoryItemResponse, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "商品IDは必須です")
	}
	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "数量は0より大きい必要があります")
	}

	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "無効な商品ID形式: %v", err)
	}

	// 在庫を取得
	inventoryItem, err := s.inventoryRepository.FindByProductID(ctx, productID)
	if err != nil {
		// 在庫が見つからない場合は新規作成
		inventoryItem = inventory.NewInventoryItem(productID, "", int(req.Quantity))
		if err := s.inventoryRepository.Save(ctx, inventoryItem); err != nil {
			return nil, status.Errorf(codes.Internal, "在庫の保存に失敗しました: %v", err)
		}
	} else {
		// 在庫を追加
		if err := inventoryItem.Add(int(req.Quantity)); err != nil {
			return nil, status.Errorf(codes.Internal, "在庫の追加に失敗しました: %v", err)
		}
		// 在庫の更新
		if err := s.inventoryRepository.Update(ctx, inventoryItem); err != nil {
			return nil, status.Errorf(codes.Internal, "在庫の更新に失敗しました: %v", err)
		}
	}

	return inventoryItemToProto(inventoryItem), nil
}

// inventoryItemToProto はドメインモデルをProtobufメッセージに変換する
func inventoryItemToProto(item *inventory.InventoryItem) *pb.InventoryItemResponse {
	return &pb.InventoryItemResponse{
		ProductId:        item.ProductID.String(),
		ProductName:      item.ProductName,
		AvailableQuantity: int32(item.AvailableQuantity),
		ReservedQuantity: int32(item.ReservedQuantity),
		BackorderQuantity: int32(item.BackorderQuantity),
		InStock:          item.IsInStock(),
		LastUpdated:      timestamppb.New(item.LastUpdated),
	}
} 