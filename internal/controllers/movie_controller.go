package controllers

import (
	"io"
	"time"

	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MovieController struct {
	session *gocql.Session
	pb.UnimplementedMovieServiceServer
}

func NewMovieController(session *gocql.Session) *MovieController {
	return &MovieController{
		session: session,
	}
}

func (c *MovieController) CreateMovies(stream pb.MovieService_CreateMoviesServer) error {
	// Prepare Cassandra statement for inserting movies
	stmt := `INSERT INTO movie_db.movies_by_user (user_id, movie_id, category_id, name, banner_url, movie_url, description, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var createdMovies []*pb.MovieResponse

	batch := c.session.NewBatch(gocql.UnloggedBatch)
	for {
		// Receive request from the stream
		req, err := stream.Recv()
		if err == io.EOF {
			// If we reach the end of the stream, break out of the loop
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "cannot receive stream request: %v", err)
		}

		// Validate incoming request
		if req.UserId == "" || req.CategoryId == "" || req.Name == "" || req.BannerUrl == "" || req.MovieUrl == "" || req.Description == "" {
			return status.Errorf(codes.InvalidArgument, "invalid request")
		}

		// Generate movie ID and timestamps
		movieID := gocql.TimeUUID()
		createdAt := time.Now()
		updatedAt := time.Now()

		// Add insert query to the batch
		batch.Query(stmt, req.UserId, movieID, req.CategoryId, req.Name, req.BannerUrl, req.MovieUrl, req.Description, createdAt, updatedAt)

		// Create MovieResponse object to send back in the response
		createdMovies = append(createdMovies, &pb.MovieResponse{
			MovieId:     movieID.String(),
			UserId:      req.UserId,
			CategoryId:  req.CategoryId,
			Name:        req.Name,
			BannerUrl:   req.BannerUrl,
			MovieUrl:    req.MovieUrl,
			Description: req.Description,
		})

		// If the batch size reaches 100, execute it and create a new batch
		if batch.Size() >= 100 {
			if err := c.session.ExecuteBatch(batch); err != nil {
				return status.Errorf(codes.Internal, "Error while executing batch: %v", err)
			}
			// Reset the batch after execution
			batch = c.session.NewBatch(gocql.UnloggedBatch)
		}
	}

	// After finishing the stream, if there are any remaining queries in the batch, execute them
	if len(batch.Entries) > 0 {
		if err := c.session.ExecuteBatch(batch); err != nil {
			return status.Errorf(codes.Internal, "Error inserting final batch: %v", err)
		}
	}

	// Send the success response with the created movies
	return stream.SendAndClose(&pb.CreateMoviesResponse{
		Message: "Movies created successfully",
		Movies:  createdMovies,
	})
}
