.PHONY: all build clean proto test run docker-up docker-down

# 変数定義
APP_NAME=cursor-ddd-ecsite
PROTO_DIR=./proto
OUT_DIR=./internal/interfaces/grpc

# すべてのターゲットを実行
all: proto build

# Go アプリケーションをビルド
build:
	go build -o bin/$(APP_NAME) ./cmd/api

# コンパイルされたバイナリとプロトバッファ生成物を削除
clean:
	rm -f bin/$(APP_NAME)
	find $(OUT_DIR) -name "*.pb.go" -type f -delete

# Protocol Buffers をコンパイル
proto:
	@echo "Generating Protocol Buffers code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/order/*.proto \
		$(PROTO_DIR)/payment/*.proto \
		$(PROTO_DIR)/inventory/*.proto \
		$(PROTO_DIR)/shipping/*.proto

# テストを実行
test:
	go test -v ./...

# アプリケーションを実行
run:
	go run ./cmd/api

# Docker環境を起動
docker-up:
	docker-compose up -d

# Docker環境を停止・削除
docker-down:
	docker-compose down 