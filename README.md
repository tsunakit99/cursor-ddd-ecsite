# DDD ECサイト (cursor-ddd-ecsite)

ドメイン駆動設計 (DDD) を使用した ECサイトのサンプルプロジェクト。CQRS、イベントソーシング、レイヤードアーキテクチャを実装しています。

## アーキテクチャ

このプロジェクトでは以下のアーキテクチャパターンを採用しています：

- **ドメイン駆動設計 (DDD)**: 複雑なビジネスロジックをドメインモデルとして表現
- **コマンドクエリ責務分離 (CQRS)**: 書き込みと読み取りの責務を分離
- **イベントソーシング**: 状態変更を一連のイベントとして保存
- **レイヤードアーキテクチャ**: 関心事の分離によるスケーラブルな設計
- **gRPC**: サービス間通信のためのプロトコル

## バウンデッドコンテキスト

プロジェクトは以下のバウンデッドコンテキストに分かれています：

1. **Order**: 注文の作成と管理
2. **Payment**: 支払い処理
3. **Inventory**: 在庫管理
4. **Shipping**: 配送管理

## 技術スタック

- **バックエンド**: Go
- **API**: gRPC
- **データベース**: 
  - コマンド側: PostgreSQL
  - クエリ側: MongoDB
- **メッセージング**: Kafka
- **コンテナ化**: Docker, Docker Compose

## 開発環境のセットアップ

### 前提条件

- Docker と Docker Compose
- Go 1.20以上
- Protocol Buffers コンパイラ

### 環境構築

```bash
# リポジトリのクローン
git clone https://github.com/tsunakit99/cursor-ddd-ecsite.git
cd cursor-ddd-ecsite

# 依存関係のインストール
go mod download

# Docker環境の起動
docker-compose up -d

# プロトコルバッファのコンパイル
make proto

# アプリケーションの実行
go run cmd/api/main.go
```

## プロジェクト構造

```
cursor-ddd-ecsite/
├── cmd/                      # エントリーポイント
├── internal/                 # 非公開コード
│   ├── domain/               # ドメインモデル
│   │   ├── order/            # 注文コンテキスト
│   │   ├── payment/          # 決済コンテキスト
│   │   ├── inventory/        # 在庫コンテキスト
│   │   └── shipping/         # 配送コンテキスト
│   ├── application/          # アプリケーションサービス
│   │   ├── commands/         # コマンドハンドラー
│   │   ├── queries/          # クエリハンドラー
│   │   └── events/           # イベントハンドラー
│   ├── infrastructure/       # インフラ層
│   │   ├── persistence/      # データベースアダプター
│   │   ├── messaging/        # メッセージングアダプター
│   │   └── api/              # API関連
│   └── interfaces/           # インターフェース層
│       └── grpc/             # gRPCサービス
├── pkg/                      # 公開コード
│   ├── eventbus/             # イベントバス実装
│   └── common/               # 共通ユーティリティ
├── proto/                    # Protocol Buffers定義
├── config/                   # 設定ファイル
└── migrations/               # データベースマイグレーション
```

## ライセンス

MITライセンス 