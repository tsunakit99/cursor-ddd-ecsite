-- インデックスを削除
DROP INDEX IF EXISTS idx_orders_payment_id;
DROP INDEX IF EXISTS idx_orders_shipping_id;

-- order_itemsテーブルから商品名カラムを削除
ALTER TABLE order_items DROP COLUMN product_name;

-- 列名を元に戻す
ALTER TABLE orders RENAME COLUMN total_amount TO total_price;

-- 追加したカラムを削除
ALTER TABLE orders 
  DROP COLUMN currency,
  DROP COLUMN billing_address_id,
  DROP COLUMN shipping_address_id,
  DROP COLUMN payment_id,
  DROP COLUMN shipping_id; 