package helpers

import (
	"fmt"
	"sync"

	"github.com/yaninyzwitty/movie-project-grpc/pb"
)

type Person struct {
	Name      string `json:"name"`
	AliasName string `json:"alias_name"`
}

func CreateUser(index int, user Person, stream pb.UserService_CreateUsersClient, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()
	// create user request
	createUserReq := &pb.CreateUserRequest{
		Name:      user.Name,
		AliasName: user.AliasName,
	}

	// send user to the stream
	if err := stream.Send(createUserReq); err != nil {
		errChan <- fmt.Errorf("failed to send user %d: %v", index+1, err)
	}

}
