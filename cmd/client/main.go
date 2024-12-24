package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/yaninyzwitty/movie-project-grpc/internal/database/pkg"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Person struct {
	Name      string `json:"name"`
	AliasName string `json:"alias_name"`
}

func main() {
	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Error("failed to open file", "error", err)
		os.Exit(1)
	}
	var cfg pkg.Config
	if err := cfg.LoadFile(file); err != nil {
		slog.Error("failed to load file", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	address := fmt.Sprintf(":%d", cfg.Server.Port)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	// user client service
	userClient := pb.NewUserServiceClient(conn)

	res, err := userClient.GetUser(ctx, &pb.GetUserRequest{Id: "420ebbc6-c208-11ef-8e78-54ee756d8952"})
	if err != nil {
		slog.Error("failed to get user", "error", err)
		os.Exit(1)
	}

	slog.Info("user", "val", res.Name)
	slog.Info("Alias name", "val", res.AliasName)

}
