package controllers

import (
	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
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
