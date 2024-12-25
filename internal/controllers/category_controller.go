package controllers

import (
	"io"

	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CategoryController struct {
	session *gocql.Session
	pb.UnimplementedCategoryServiceServer
}

func NewCategoryController(session *gocql.Session) *CategoryController {
	return &CategoryController{
		session: session,
	}
}

func (c *CategoryController) CreateCategories(stream pb.CategoryService_CreateCategoriesServer) error {
	// Prepare Cassandra statement
	stmt := `INSERT INTO movie_db.categories (id, name, description) VALUES(?, ?, ?)`
	batch := c.session.NewBatch(gocql.UnloggedBatch)
	totalCategoriesCreated := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Error while reading client stream: %v", err)
		}

		if req.Name == "" || req.Description == "" {
			return status.Errorf(codes.InvalidArgument, "Name and description cannot be empty")
		}

		// Generate unique category ID
		categoryID := gocql.TimeUUID()

		// Add query to the batch
		batch.Query(stmt, categoryID, req.Name, req.Description)
		totalCategoriesCreated++

		// Flush batch if it exceeds a threshold (e.g., 100 queries)
		if batch.Size() >= 100 {
			if err := c.session.ExecuteBatch(batch); err != nil {
				return status.Errorf(codes.Internal, "Error while executing batch: %v", err)
			}
			batch = c.session.NewBatch(gocql.UnloggedBatch)
		}
	}

	// Execute any remaining queries in the batch
	if len(batch.Entries) > 0 {
		if err := c.session.ExecuteBatch(batch); err != nil {
			return status.Errorf(codes.Internal, "Error inserting batch: %v", err)
		}
	}

	// Send response to the client
	return stream.SendAndClose(&pb.CreateCategoriesResponse{
		Message:             "Categories created successfully",
		NoCreatedCategories: int32(totalCategoriesCreated),
	})
}
