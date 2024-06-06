package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ( UserServiceManager *UserService) AuthenticateUser(ctx context.Context, request *userpb.AuthenticateUserRequest) (*userpb.AuthenticateUserResponse, error) {
	userEmail := request.UserEmail
	userPassword := request.UserPassword
	// Validate the fields
	if userEmail == "" || userPassword == "" {
		return &userpb.AuthenticateUserResponse{Message: "Invalid fields!", StatusCode: 403}, nil
	}

	var existingUser model.User
	userNotFoundError := dbConnector.Where("email = ?", userEmail).First(&existingUser).Error
	// If the user is not found, create a new user with the provided details
	if userNotFoundError != nil {
		return &userpb.AuthenticateUserResponse{Message: "User not found", StatusCode: 404}, nil
	}
	if config.ComparePasswords(existingUser.Password, userPassword) != nil {
		return &userpb.AuthenticateUserResponse{Message: "Wrong Password", StatusCode: 403}, nil
	}
	// Gennerating the the jwt token.
	token, err := UserServiceManager.jwtManager.GenerateToken(&existingUser)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Could not generate token: %s", err),
		)
	}
	return &userpb.AuthenticateUserResponse{Message: "User authenticated successfully", StatusCode: 200, Token: token}, nil
}
