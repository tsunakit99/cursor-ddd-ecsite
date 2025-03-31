# DDD ECサイト

このプロジェクトは、ドメイン駆動設計（DDD）を採用したECサイトのサンプル実装です。バックエンドはGo言語、フロントエンドはReact（TypeScript）で構築されています。

## 機能

- ユーザー認証（登録・ログイン）
- ユーザープロフィール管理
- 住所管理
- 商品一覧・詳細表示
- カート機能
- 注文処理
- 在庫管理
- 配送管理
- 決済処理

## アーキテクチャ

このプロジェクトは以下のアーキテクチャパターンとプラクティスを採用しています：

- ドメイン駆動設計（DDD）
- CQRS（コマンドクエリ責務分離）
- イベント駆動アーキテクチャ
- マイクロサービス

## 技術スタック

### バックエンド

- Go
- PostgreSQL
- Kafka
- gRPC
- Docker

### フロントエンド

- React
- TypeScript
- Material UI
- React Query
- React Hook Form
- Zustand

## 環境構築

### 必要条件

- Docker と Docker Compose
- Go 1.18以上
- Node.js 16以上
- npm または yarn

### セットアップ

1. リポジトリをクローン

```bash
git clone https://github.com/yourusername/cursor-ddd-ecsite.git
cd cursor-ddd-ecsite
```

2. Docker Composeでアプリケーションを起動

```bash
docker-compose up -d
```

これにより以下のサービスが起動します：
- API サーバー (http://localhost:50051)
- フロントエンド Web アプリケーション (http://localhost:3000)
- PostgreSQL データベース
- Kafka
- Zookeeper

### 開発環境

#### バックエンド開発

```bash
# APIサーバーを開発モードで起動
go mod tidy
go run cmd/api/main.go
```

#### フロントエンド開発

```bash
cd web
npm install
npm start
```

## ディレクトリ構造

```
.
├── cmd/               # アプリケーションのエントリーポイント
│   └── api/           # APIサーバー
├── internal/          # プライベートなアプリケーションコード
│   ├── app/           # アプリケーションサービス
│   ├── domain/        # ドメインモデル
│   ├── infra/         # インフラストラクチャ層
│   └── ports/         # 外部とのインターフェース
├── proto/             # Protocol Buffersの定義
├── web/               # フロントエンドアプリケーション
└── docker-compose.yml # Docker構成ファイル
```

## リソース

- DDD についての詳細情報： [Domain-Driven Design Reference](https://domainlanguage.com/ddd/reference/)
- CQRS パターン： [Martin Fowler's CQRS](https://martinfowler.com/bliki/CQRS.html)
- イベント駆動アーキテクチャ： [What is Event-Driven Architecture?](https://aws.amazon.com/event-driven-architecture/)

## ライセンス

MIT 