package commands

import (
	"context"

	"github.com/google/uuid"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/inventory"
)

// ReserveInventoryCommand は在庫予約コマンド
type ReserveInventoryCommand struct {
	OrderID uuid.UUID
	Items   []ReserveInventoryItem
}

// ReserveInventoryItem は在庫予約アイテム
type ReserveInventoryItem struct {
	ProductID uuid.UUID
	Quantity  int
}

// ReserveInventoryHandler は在庫予約コマンドのハンドラ
type ReserveInventoryHandler struct {
	inventoryRepository   inventory.InventoryRepository
	reservationRepository inventory.ReservationRepository
}

// NewReserveInventoryHandler は新しい在庫予約ハンドラを作成する
func NewReserveInventoryHandler(
	inventoryRepo inventory.InventoryRepository,
	reservationRepo inventory.ReservationRepository,
) *ReserveInventoryHandler {
	return &ReserveInventoryHandler{
		inventoryRepository:   inventoryRepo,
		reservationRepository: reservationRepo,
	}
}

// Handle は在庫予約コマンドを処理する
func (h *ReserveInventoryHandler) Handle(ctx context.Context, cmd ReserveInventoryCommand) (*inventory.InventoryReservation, error) {
	// 商品IDのリストを作成
	productIDs := make([]uuid.UUID, len(cmd.Items))
	for i, item := range cmd.Items {
		productIDs[i] = item.ProductID
	}

	// 在庫の確認
	inventoryItems, err := h.inventoryRepository.FindByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	// 商品IDから在庫情報へのマップを作成
	inventoryMap := make(map[uuid.UUID]*inventory.InventoryItem)
	for _, item := range inventoryItems {
		inventoryMap[item.ProductID] = item
	}

	// 予約アイテムのリストを作成
	reservationItems := make([]inventory.ReservationItem, len(cmd.Items))
	
	// 各アイテムについて予約を試みる
	for i, item := range cmd.Items {
		invItem, exists := inventoryMap[item.ProductID]
		
		// 商品が存在しない場合は予約できない
		if !exists {
			reservationItems[i] = inventory.ReservationItem{
				ProductID:         item.ProductID,
				RequestedQuantity: item.Quantity,
				AvailableQuantity: 0,
				Reserved:          false,
			}
			continue
		}
		
		// 在庫数が足りるか確認
		canReserve := invItem.CanReserve(item.Quantity)
		
		reservationItems[i] = inventory.ReservationItem{
			ProductID:         item.ProductID,
			RequestedQuantity: item.Quantity,
			AvailableQuantity: invItem.AvailableQuantity,
			Reserved:          canReserve,
		}
		
		// 在庫が足りる場合は予約する
		if canReserve {
			if err := invItem.Reserve(item.Quantity); err != nil {
				reservationItems[i].Reserved = false
				continue
			}
			
			// 在庫の更新
			if err := h.inventoryRepository.Update(ctx, invItem); err != nil {
				reservationItems[i].Reserved = false
				continue
			}
		}
	}

	// 予約を作成
	reservation := inventory.NewInventoryReservation(cmd.OrderID, reservationItems)
	
	// 予約を保存
	if err := h.reservationRepository.Save(ctx, reservation); err != nil {
		// 注: 実際のアプリケーションでは、保存に失敗した場合に予約した在庫を戻す補償トランザクションを実装すべき
		return nil, err
	}
	
	return reservation, nil
} 