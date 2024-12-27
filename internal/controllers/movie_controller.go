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

func (c *MovieController) GetMoviesByUserIDAndCategoryID(req *pb.GetMoviesRequest, stream pb.MovieService_GetMoviesByUserIDAndCategoryIDServer) error {
	// Validate input parameters
	if req.CategoryId == "" || req.UserId == "" {
		return status.Errorf(codes.InvalidArgument, "categoryId or userId cannot be empty")
	}

	categoryID, err := gocql.ParseUUID(req.CategoryId)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid categoryId format: %v", err)
	}

	userID, err := gocql.ParseUUID(req.UserId)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid userId format: %v", err)
	}

	// Query Cassandra for movies
	movies, err := c.fetchMovies(userID, categoryID)
	if err != nil {
		return err
	}

	// Handle empty results
	if len(movies) == 0 {
		return status.Errorf(codes.NotFound, "no movies found for the specified userId and categoryId")
	}

	// Send the response
	if err := stream.Send(&pb.GetMoviesResponse{
		Movies:  movies,
		Message: "Movies retrieved successfully",
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to send response: %v", err)
	}

	return nil
}

func (c *MovieController) fetchMovies(userID, categoryID gocql.UUID) ([]*pb.MovieResponse, error) {
	stmt := `SELECT movie_id, category_id, name, banner_url, movie_url, description, created_at, updated_at FROM movie_db.movies_by_user WHERE user_id = ? AND category_id = ?`
	query := c.session.Query(stmt, userID, categoryID)
	iter := query.Iter()

	var movies []*pb.MovieResponse
	var (
		movieID, catID                         gocql.UUID
		name, bannerUrl, movieUrl, description string
		createdAt, updatedAt                   time.Time
	)

	for iter.Scan(&movieID, &catID, &name, &bannerUrl, &movieUrl, &description, &createdAt, &updatedAt) {
		movies = append(movies, &pb.MovieResponse{
			MovieId:     movieID.String(),
			UserId:      userID.String(),
			CategoryId:  categoryID.String(),
			Name:        name,
			BannerUrl:   bannerUrl,
			MovieUrl:    movieUrl,
			Description: description,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to close iterator: %v", err)
	}

	return movies, nil
}
