package controllers

import (
	"context"
	"io"

	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserController struct {
	session *gocql.Session
	pb.UnimplementedUserServiceServer
}

func NewUserController(session *gocql.Session) *UserController {
	return &UserController{
		session: session,
	}
}

func (c *UserController) CreateUsers(stream pb.UserService_CreateUsersServer) error {
	// Prepare the Cassandra statement
	insertStmt := `INSERT INTO users (id, name, alias_name) VALUES (?, ?, ?)`
	batch := c.session.NewBatch(gocql.UnloggedBatch) // Use unlogged batch for efficiency

	totalUsersCreated := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Error while reading client stream: %v", err)
		}

		// Validate input
		if req.Name == "" || req.AliasName == "" {
			return status.Errorf(codes.InvalidArgument, "Name and AliasName are required")
		}

		// Generate a unique ID for the user
		userID := gocql.TimeUUID()

		// Add the query to the batch
		batch.Query(insertStmt, userID, req.Name, req.AliasName)
		totalUsersCreated++

		// Flush the batch if it exceeds a threshold (e.g., 100 queries)
		if batch.Size() >= 100 {
			if err := c.session.ExecuteBatch(batch); err != nil {
				return status.Errorf(codes.Internal, "Error while inserting users: %v", err)
			}
			batch = c.session.NewBatch(gocql.UnloggedBatch) // Reset batch
		}
	}

	// Insert any remaining users in the batch
	if len(batch.Entries) > 0 {
		if err := c.session.ExecuteBatch(batch); err != nil {
			return status.Errorf(codes.Internal, "Error inserting batch: %v", err)
		}
	}

	// Send response to the client
	return stream.SendAndClose(&pb.CreateUsersResponse{
		Message:        "Users created successfully",
		NoCreatedUsers: int32(totalUsersCreated),
	})
}

func (c *UserController) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userID, err := gocql.ParseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error inserting batch: %v", err)
	}

	stmt := `SELECT name, alias_name FROM users WHERE id = ?`
	var name, aliasName string
	if err := c.session.Query(stmt, userID).Scan(&name, &aliasName); err != nil {

		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.Internal, "user with id %s not found: %v", userID, err)
		}
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	return &pb.GetUserResponse{
		Id:        userID.String(),
		Name:      name,
		AliasName: aliasName,
	}, nil

}
