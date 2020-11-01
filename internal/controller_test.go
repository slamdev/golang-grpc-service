package internal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang-grpc-service/api"
	"testing"
)

func TestController_CreateUser(t *testing.T) {
	c := NewController()

	request := &api.CreateUserRequest{User: &api.User{
		Id:   1,
		Name: "test",
	}}

	response, err := c.CreateUser(context.TODO(), request)

	if assert.NoError(t, err) {
		assert.NotNil(t, t, response.User)
	}
}

func TestController_ListUsers(t *testing.T) {
	c := NewController()

	request := &api.ListUsersRequest{UserName: "test"}

	response, err := c.ListUsers(context.TODO(), request)

	if assert.NoError(t, err) {
		assert.NotNil(t, t, response.Users)
	}
}
