package controllers

import (
	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
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
