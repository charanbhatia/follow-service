package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/pratilipi/follow-service/internal/database"
	"github.com/pratilipi/follow-service/internal/handler"
	"github.com/pratilipi/follow-service/internal/repository"
	pb "github.com/pratilipi/follow-service/proto/follow"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		logger.Fatal("DATABASE_URL is required")
	}

	db, err := database.NewConnection(databaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("database migrations completed successfully")

	repo := repository.New(db)
	followService := handler.NewFollowServiceServer(repo, logger)

	port := getEnv("GRPC_PORT", "50051")
	
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFollowServiceServer(grpcServer, followService)
	reflection.Register(grpcServer)

	logger.Info("starting gRPC server", zap.String("port", port))

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	grpcServer.GracefulStop()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
