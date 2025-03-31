package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tsunakit99/cursor-ddd-ecsite/internal/application/commands"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/infrastructure/persistence"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/infrastructure/persistence/postgres"
	grpcserver "github.com/tsunakit99/cursor-ddd-ecsite/internal/interfaces/grpc"
	kafkabus "github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus/kafka"
)

func main() {
	// 設定を読み込む
	if err := loadConfig(); err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// コンテキストの作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// データベース接続の設定
	db, err := persistence.NewPostgresDB()
	if err != nil {
		log.Fatalf("データベース接続の設定に失敗しました: %v", err)
	}
	defer db.Close()

	// イベントバスの設定
	kafkaBrokers := viper.GetStringSlice("kafka.brokers")
	eventBus := kafkabus.NewKafkaEventBus(kafkaBrokers, 5*time.Second)
	defer eventBus.Close()

	// リポジトリの設定
	orderRepo := postgres.NewOrderRepository(db)

	// コマンドハンドラの設定
	createOrderHandler := commands.NewCreateOrderHandler(orderRepo, eventBus)

	// gRPCサーバーの設定
	address := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// gRPCサーバーの作成
	server := grpc.NewServer()

	// サーバーインスタンスの作成
	orderServer := grpcserver.NewOrderServer(createOrderHandler, orderRepo)

	// サービスの登録（実装後にコメントを外す）
	// pb.RegisterOrderServiceServer(server, orderServer)
	// pb.RegisterPaymentServiceServer(server, paymentServer)
	// pb.RegisterInventoryServiceServer(server, inventoryServer)
	// pb.RegisterShippingServiceServer(server, shippingServer)

	// gRPCリフレクションの有効化（開発環境用）
	if viper.GetString("app.environment") == "development" {
		reflection.Register(server)
	}

	// サーバーの起動（非同期）
	log.Printf("gRPCサーバーを起動しています: %s", address)
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Kafka消費者の起動
	// TODO: 各種イベントハンドラを登録した後にコメントを外す
	// startKafkaConsumers(ctx, eventBus, kafkaBrokers)

	// グレースフルシャットダウンの設定
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("シャットダウンを開始します...")
	server.GracefulStop()
	log.Println("サーバーを停止しました")
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv() // 環境変数からの上書きを許可
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// デフォルト値の設定
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 50051)
	viper.SetDefault("app.environment", "development")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("設定ファイルが見つかりません。デフォルト値を使用します。")
			return nil
		}
		return err
	}

	log.Printf("設定ファイルを読み込みました: %s", viper.ConfigFileUsed())
	return nil
}

// startKafkaConsumers はKafkaコンシューマーを起動する
func startKafkaConsumers(ctx context.Context, eventBus *kafkabus.KafkaEventBus, brokers []string) {
	// トピックのリスト
	// TODO: 実際に使用するトピックに合わせて調整
	topics := []string{
		viper.GetString("kafka.topics.orderCreated"),
		viper.GetString("kafka.topics.orderConfirmed"),
		viper.GetString("kafka.topics.orderPaid"),
		viper.GetString("kafka.topics.paymentCreated"),
		viper.GetString("kafka.topics.paymentApproved"),
		viper.GetString("kafka.topics.inventoryReserved"),
	}

	// コンシューマーの起動
	groupID := viper.GetString("app.name") + "-consumer"
	if err := eventBus.StartConsumer(ctx, groupID, topics, brokers); err != nil {
		log.Fatalf("Kafkaコンシューマーの起動に失敗しました: %v", err)
	}
	log.Printf("Kafkaコンシューマーを起動しました（グループID: %s）", groupID)
} 