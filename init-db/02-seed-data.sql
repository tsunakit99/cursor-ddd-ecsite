-- テストユーザー（パスワード: password123）
INSERT INTO users (id, email, password_hash, first_name, last_name, phone_number, status, created_at, updated_at)
VALUES 
    (uuid_generate_v4(), 'test@example.com', '$2a$10$UOkKQDJKZK7QdT3TJJMkWusdkw0qPsWZ0Z3VjCjwxR1BRvdkXUjxi', 'テスト', 'ユーザー', '090-1234-5678', 'ACTIVE', NOW(), NOW()),
    (uuid_generate_v4(), 'admin@example.com', '$2a$10$UOkKQDJKZK7QdT3TJJMkWusdkw0qPsWZ0Z3VjCjwxR1BRvdkXUjxi', '管理者', 'ユーザー', '090-8765-4321', 'ACTIVE', NOW(), NOW());

-- テスト商品
INSERT INTO inventory_items (product_id, product_name, available_quantity, reserved_quantity, backorder_quantity, last_updated)
VALUES 
    (uuid_generate_v4(), 'テスト商品1', 100, 0, 0, NOW()),
    (uuid_generate_v4(), 'テスト商品2', 50, 0, 0, NOW()),
    (uuid_generate_v4(), 'テスト商品3', 20, 0, 0, NOW()),
    (uuid_generate_v4(), 'テスト商品4', 0, 0, 10, NOW()); 