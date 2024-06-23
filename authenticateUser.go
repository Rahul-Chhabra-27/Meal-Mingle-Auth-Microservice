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
	if userEmail == "" || userPassword == "" {
		return &userpb.AuthenticateUserResponse{
			Data:    nil,
			Message: "The request contains missing or invalid fields.",
			Error:   "Invalid fields!", 
			StatusCode: 400,
		}, nil
	}
	var existingUser model.User
	userNotFoundError := userDbConnector.Where("email = ?", userEmail).First(&existingUser).Error
	// If the user is not found, create a new user with the provided details
	if userNotFoundError != nil || 
		existingUser.Role != request.Role {
		return &userpb.AuthenticateUserResponse{
			Data:       nil,
			Message:    "Authentication Failed, User not found OR Invalid role",
			Error:      "Not Found",
			StatusCode: 404,
		}, nil
	}
	if config.ComparePasswords(existingUser.Password, userPassword) != nil {
		return &userpb.AuthenticateUserResponse{
			Message: "Authentication Failed,Wrong Password",
			Error:   "Unauthorized", StatusCode: 401,
		}, nil
	}
	// Gennerating the the jwt token.
	token, err := UserServiceManager.jwtManager.GenerateToken(&existingUser)
	if err != nil {
		fmt.Println("Error in generating token")
		return &userpb.AuthenticateUserResponse{
			Data: nil,
			Error:      "Internal Server Error",
			StatusCode: 500,
			Message:    "Security Issues, Please try again later.",
		}, nil

	}
	fmt.Println("User authenticated successfully")
	return &userpb.AuthenticateUserResponse{Error: "",
		Message:    "User authenticated successfully",
		StatusCode: 200,
		Data: &userpb.Responsedata{
			User: &userpb.User{
				UserId:    strconv.FormatUint(uint64(existingUser.ID), 10),
				UserName:  existingUser.Name,
				UserEmail: existingUser.Email,
				UserPhone: existingUser.Phone,
			},
			Token: token,
		},
	}, nil
}
