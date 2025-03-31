-- 拡張機能
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 住所テーブル
CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    recipient_name VARCHAR(200) NOT NULL,
    street_line1 VARCHAR(255) NOT NULL,
    street_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    address_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 注文テーブル
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL,
    total_amount DECIMAL(19, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    billing_address_id UUID NOT NULL,
    shipping_address_id UUID NOT NULL,
    payment_id UUID,
    shipping_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 注文アイテムテーブル
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id),
    product_id UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(19, 4) NOT NULL,
    total_price DECIMAL(19, 4) NOT NULL
);

-- 支払いテーブル
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id),
    amount DECIMAL(19, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    method VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    transaction_id VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 在庫テーブル
CREATE TABLE IF NOT EXISTS inventory_items (
    product_id UUID PRIMARY KEY,
    product_name VARCHAR(255) NOT NULL,
    available_quantity INT NOT NULL DEFAULT 0,
    reserved_quantity INT NOT NULL DEFAULT 0,
    backorder_quantity INT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 在庫予約テーブル
CREATE TABLE IF NOT EXISTS inventory_reservations (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 在庫予約アイテムテーブル
CREATE TABLE IF NOT EXISTS inventory_reservation_items (
    reservation_id UUID NOT NULL REFERENCES inventory_reservations(id),
    product_id UUID NOT NULL,
    requested_quantity INT NOT NULL,
    available_quantity INT NOT NULL,
    reserved BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (reservation_id, product_id)
);

-- 配送テーブル
CREATE TABLE IF NOT EXISTS shippings (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id),
    recipient_name VARCHAR(200) NOT NULL,
    street_line1 VARCHAR(255) NOT NULL,
    street_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    method VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    tracking_number VARCHAR(100),
    carrier VARCHAR(100),
    estimated_delivery_date TIMESTAMP WITH TIME ZONE,
    actual_delivery_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 配送追跡イベントテーブル
CREATE TABLE IF NOT EXISTS shipping_tracking_events (
    id UUID PRIMARY KEY,
    shipping_id UUID NOT NULL REFERENCES shippings(id),
    description TEXT NOT NULL,
    location VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);

-- インデックス
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_inventory_reservations_order_id ON inventory_reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_shippings_order_id ON shippings(order_id); 