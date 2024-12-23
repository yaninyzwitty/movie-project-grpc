package controllers

import (
	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/movie-project-grpc/pb"
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
