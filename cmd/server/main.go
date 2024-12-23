package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/yaninyzwitty/movie-project-grpc/internal/controllers"
	"github.com/yaninyzwitty/movie-project-grpc/internal/database"
	"github.com/yaninyzwitty/movie-project-grpc/internal/database/pkg"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
	"google.golang.org/grpc"
)

var (
	s string
)

func main() {
	var cfg pkg.Config
	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Error("failed to open config.yaml")
		os.Exit(1)
	}
	defer file.Close()
	if err := cfg.LoadFile(file); err != nil {
		slog.Error("failed to load config file", "error", err)
		os.Exit(1)
	}
	// you dont require godotenv.Load() here when using docker and docker compose
	err = godotenv.Load()
	if err != nil {
		slog.Error("failed to load .env file", "value", err)

	}

	if password := os.Getenv("ASTRA_DB_PASSWORD"); password != "" {
		s = password

	}

	_, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
	defer cancel()
	astraConn := database.NewAstraDb()

	config := &database.AstraConfig{
		Path:     cfg.Database.Path,
		Username: cfg.Database.Username,
		Password: s,
		Timeout:  cfg.Server.Timeout,
	}
	session, err := astraConn.CreateDBConn(config)
	if err != nil {
		slog.Error("failed to connect to Astra DB", "error", err)
		os.Exit(1)
	}

	defer session.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	userController := controllers.NewUserController(session)
	categoryController := controllers.NewCategoryController(session)
	movieController := controllers.NewMovieController(session)

	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, userController)
	pb.RegisterCategoryServiceServer(server, categoryController)
	pb.RegisterMovieServiceServer(server, movieController)

	// handle graceful stop, signals etc.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)
		slog.Info("Shutting down gRPC server...")

		// Gracefully stop the gRPC server
		server.GracefulStop()
		cancel()

		slog.Info("gRPC server has been stopped gracefully")
	}()

	slog.Info("Starting gRPC server", "port", cfg.Server.Port)
	if err := server.Serve(lis); err != nil {
		slog.Error("gRPC server encountered an error while serving", "error", err)
		os.Exit(1)
	}

}
