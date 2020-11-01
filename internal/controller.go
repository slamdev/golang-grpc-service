package internal

import (
	"context"
	"go.uber.org/zap"
	"golang-grpc-service/api"
)

type controller struct {
	api.UnimplementedGolangGrpcServiceServer
}

func (c *controller) ListUsers(ctx context.Context, request *api.ListUsersRequest) (*api.ListUsersResponse, error) {
	zap.L().Info("received list users request", zap.Reflect("req", request))
	return &api.ListUsersResponse{
		Users: []*api.User{{
			Id:   1,
			Name: "test",
		}},
	}, nil
}

func (c *controller) CreateUser(ctx context.Context, request *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	zap.L().Info("received create user request", zap.Reflect("req", request))
	return &api.CreateUserResponse{
		User: &api.User{
			Id:   1,
			Name: "test",
		},
	}, nil
}

func NewController() api.GolangGrpcServiceServer {
	return &controller{}
}
