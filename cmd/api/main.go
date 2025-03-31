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
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/domain/user"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/infrastructure/persistence"
	"github.com/tsunakit99/cursor-ddd-ecsite/internal/infrastructure/persistence/postgres"
	grpcserver "github.com/tsunakit99/cursor-ddd-ecsite/internal/interfaces/grpc"
	kafkabus "github.com/tsunakit99/cursor-ddd-ecsite/pkg/eventbus/kafka"
	inventorypb "github.com/tsunakit99/cursor-ddd-ecsite/proto/inventory"
	orderpb "github.com/tsunakit99/cursor-ddd-ecsite/proto/order"
	paymentpb "github.com/tsunakit99/cursor-ddd-ecsite/proto/payment"
	shippingpb "github.com/tsunakit99/cursor-ddd-ecsite/proto/shipping"
	userpb "github.com/tsunakit99/cursor-ddd-ecsite/proto/user"
)

func main() {
	// 設定を読み込む
	if err := loadConfig(); err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// コンテキストの作成
	_, cancel := context.WithCancel(context.Background())
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

	// Order リポジトリの設定
	orderRepo := postgres.NewOrderRepository(db)

	// Payment リポジトリの設定
	paymentRepo := postgres.NewPaymentRepository(db)

	// Inventory リポジトリの設定
	inventoryRepo := postgres.NewInventoryRepository(db)
	reservationRepo := postgres.NewReservationRepository(db)

	// Shipping リポジトリの設定
	shippingRepo := postgres.NewShippingRepository(db)

	// User リポジトリとサービスの設定
	userRepo := postgres.NewUserRepository(db)
	authService := user.NewAuthService(
		viper.GetString("auth.jwtSecret"),
		time.Duration(viper.GetInt("auth.accessTokenExpiry"))*time.Minute,
		time.Duration(viper.GetInt("auth.refreshTokenExpiry"))*time.Hour,
	)

	// コマンドハンドラの設定
	createOrderHandler := commands.NewCreateOrderHandler(orderRepo, eventBus)
	confirmOrderHandler := commands.NewConfirmOrderHandler(orderRepo, eventBus)
	cancelOrderHandler := commands.NewCancelOrderHandler(orderRepo, eventBus)
	registerUserHandler := commands.NewRegisterUserHandler(userRepo, eventBus)
	loginUserHandler := commands.NewLoginUserHandler(userRepo, authService, eventBus)

	// gRPCサーバーの設定
	address := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// gRPCサーバーの作成
	server := grpc.NewServer()

	// サーバーインスタンスの作成
	orderServer := grpcserver.NewOrderServer(
		createOrderHandler,
		confirmOrderHandler,
		cancelOrderHandler,
		orderRepo,
	)
	paymentServer := grpcserver.NewPaymentServer(paymentRepo)
	inventoryServer := grpcserver.NewInventoryServer(inventoryRepo, reservationRepo)
	shippingServer := grpcserver.NewShippingServer(shippingRepo)
	userServer := grpcserver.NewUserServer(
		registerUserHandler,
		loginUserHandler,
		userRepo,
		authService,
	)

	// サービスの登録
	orderpb.RegisterOrderServiceServer(server, orderServer)
	paymentpb.RegisterPaymentServiceServer(server, paymentServer)
	inventorypb.RegisterInventoryServiceServer(server, inventoryServer)
	shippingpb.RegisterShippingServiceServer(server, shippingServer)
	userpb.RegisterUserServiceServer(server, userServer)

	// gRPCリフレクションの有効化（開発環境用）
	if viper.GetBool("server.enableReflection") {
		reflection.Register(server)
		log.Println("gRPC Reflection enabled")
	}

	// シグナルハンドリングの設定
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// サーバーの起動
	log.Printf("Starting gRPC server on %s", address)
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// シグナルを待機
	sig := <-sigCh
	log.Printf("Received signal: %v", sig)

	// グレースフルシャットダウン
	log.Println("Gracefully stopping server...")
	server.GracefulStop()
	log.Println("Server stopped")
}

// loadConfig は設定ファイルを読み込む
func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 環境変数の設定
	viper.SetEnvPrefix("ECSITE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// デフォルト値の設定
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 50051)
	viper.SetDefault("server.enableReflection", true)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "ecsite")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("auth.jwtSecret", "your-jwt-secret-key")
	viper.SetDefault("auth.accessTokenExpiry", 15) // 15分
	viper.SetDefault("auth.refreshTokenExpiry", 24) // 24時間

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("設定ファイルが見つかりません。デフォルト値を使用します")
			return nil
		}
		return err
	}

	log.Println("設定ファイルを読み込みました")
	return nil
} 