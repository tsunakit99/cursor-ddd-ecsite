-- 注文テーブルに新しいカラムを追加
ALTER TABLE orders 
  ADD COLUMN currency VARCHAR(3) DEFAULT 'JPY' NOT NULL,
  ADD COLUMN billing_address_id UUID,
  ADD COLUMN shipping_address_id UUID,
  ADD COLUMN payment_id UUID,
  ADD COLUMN shipping_id UUID;

-- 既存の列名を変更
ALTER TABLE orders RENAME COLUMN total_price TO total_amount;

-- order_itemsテーブルに商品名カラムを追加
ALTER TABLE order_items
  ADD COLUMN product_name VARCHAR(255) NOT NULL DEFAULT 'Unknown Product';

-- インデックスを追加
CREATE INDEX idx_orders_payment_id ON orders(payment_id);
CREATE INDEX idx_orders_shipping_id ON orders(shipping_id); 