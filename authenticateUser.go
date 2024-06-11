package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"fmt"
	"strconv"
)

func (UserServiceManager *UserService) AuthenticateUser(ctx context.Context, request *userpb.AuthenticateUserRequest) (*userpb.AuthenticateUserResponse, error) {
	userEmail := request.UserEmail
	userPassword := request.UserPassword
	// Validate the fields
	if userEmail == "" || userPassword == "" {
		return &userpb.AuthenticateUserResponse{Message: "", Error: "Invalid fields!", StatusCode: 400}, nil
	}

	var existingUser model.User
	userNotFoundError := dbConnector.Where("email = ?", userEmail).First(&existingUser).Error
	// If the user is not found, create a new user with the provided details
	if userNotFoundError != nil {
		return &userpb.AuthenticateUserResponse{Message: "", Error: "Authentication Failed, User not found", StatusCode: 401}, nil
	}
	if config.ComparePasswords(existingUser.Password, userPassword) != nil {
		return &userpb.AuthenticateUserResponse{Message: "", Error: "Authentication Failed,Wrong Password", StatusCode: 401}, nil
	}
	// Gennerating the the jwt token.
	token, err := UserServiceManager.jwtManager.GenerateToken(&existingUser)
	if err != nil {
		return &userpb.AuthenticateUserResponse{
			Error:      "Internal Server Error",
			StatusCode: 500,
			Message:    "",
		}, nil

	}
	fmt.Println("User authenticated successfully")
	return &userpb.AuthenticateUserResponse{Error: "", Message: "User authenticated successfully", StatusCode: 200, Data: &userpb.Data{User: &userpb.User{UserId: strconv.FormatUint(uint64(existingUser.ID), 10), UserName: existingUser.Name, UserEmail: existingUser.Email}, Token: token}}, nil
}
