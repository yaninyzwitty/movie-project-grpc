package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/yaninyzwitty/movie-project-grpc/internal/database/pkg"
	"github.com/yaninyzwitty/movie-project-grpc/internal/helpers"
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
	categoryClient := pb.NewCategoryServiceClient(conn)

	res, err := userClient.GetUser(ctx, &pb.GetUserRequest{Id: "889a8a7b-c2b1-11ef-9004-54ee756d8952"})
	if err != nil {
		slog.Error("failed to get user", "error", err)
		os.Exit(1)
	}

	slog.Info("user", "val", res.Name)
	slog.Info("Alias name", "val", res.AliasName)

	// read users.json file
	users, err := os.ReadFile("users.json")
	if err != nil {
		slog.Error("failed to read file", "error", err)
		os.Exit(1)
	}

	// create user stream
	stream, err := userClient.CreateUsers(ctx)
	if err != nil {
		slog.Error("failed to create user stream", "error", err)
		os.Exit(1)
	}

	// unmarshal users.json data into a slice of Person
	var persons []helpers.Person
	if err := json.Unmarshal(users, &persons); err != nil {
		slog.Error("failed to unmarshal users", "error", err)
		os.Exit(1)
	}

	// use a wait group to wait all go routines to finish
	var wg sync.WaitGroup
	// make a channel to send errors
	errChan := make(chan error, 1)
	// then we create users concurrently
	for index, person := range persons {
		wg.Add(1)
		go helpers.CreateUser(index, person, stream, &wg, errChan)

	}
	wg.Wait()
	// Check if there was an error while sending users
	select {
	case err := <-errChan:
		if err != nil {
			slog.Error("failed to send user", "error", err)
			os.Exit(1)
		}
	default:
		if err := stream.CloseSend(); err != nil {
			slog.Error("failed to close stream", "error", err)
			os.Exit(1)
		}
		// Wait for the server response
		response, err := stream.CloseAndRecv()
		if err != nil {
			slog.Error("failed to receive response", "error", err)
			os.Exit(1)
		}
		slog.Info("Users created successfully", "message", response.GetMessage())

	}

	categoryStream, err := categoryClient.CreateCategories(ctx)
	if err != nil {
		slog.Error("failed to create category stream", "error", err)
		os.Exit(1)
	}
	categories := []pb.CreateCategoryRequest{
		{Name: "Romance", Description: "Stories of passion and longing"},
		{Name: "Desire", Description: "Exploring intimate connections"},
		{Name: "Mystique", Description: "Unveiling secrets and allure"},
		{Name: "Euphoria", Description: "Moments of pure ecstasy and joy"},
		{Name: "Seduction", Description: "The art of charm and temptation"},
		{Name: "Whispers", Description: "Secrets shared in soft tones"},
		{Name: "Embrace", Description: "Warmth in closeness and affection"},
	}

	for i := range categories {
		slog.Info("creating category", "value", categories[i].Name)
		if err := categoryStream.Send(&categories[i]); err != nil {
			slog.Error("Error while sending category", "error", err)
		}
	}

	categoryResponse, err := categoryStream.CloseAndRecv()
	if err != nil {
		slog.Error("Error while receiving response", "error", err)
	}

	slog.Info("Categories created successfully", "message", categoryResponse.GetMessage())

}
