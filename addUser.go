package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	"context"
	"fmt"
	userpb "auth-microservice/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AddUser is a RPC that adds a new user to the database
func (userServiceManager *UserService) AddUser(ctx context.Context, request *userpb.AddUserRequest) (*userpb.AddUserResponse, error) {
	userEmail := request.UserEmail
	userPassword := request.UserPassword
	userName := request.UserName
	userAddress := request.UserAddress
	userCity := request.UserCity
	userPhone := request.UserPhone

	// Validate the fields
	if !config.ValidateFields(userEmail, userPassword, userName, userAddress, userCity, userPhone) {
		return &userpb.AddUserResponse{Message: "Invalid fields!", StatusCode: int64(codes.InvalidArgument)}, nil
	}
	var existingUser model.User
	userNotFoundError := dbConnector.Where("email = ?", userEmail).First(&existingUser).Error
	// If the user is not found, create a new user with the provided details
	if userNotFoundError != nil {
		
		hashedPassword := config.GenerateHashedPassword(userPassword)

		newUser := &model.User{Name: userName, Address: userAddress, Email: userEmail, City: userCity, Phone: userPhone, Password: hashedPassword}

		// Create a new user in the database and return the primary key if successful or an error if it fails
		primaryKey := dbConnector.Create(newUser)
		if primaryKey.Error != nil {
			return &userpb.AddUserResponse{Message: "User is already exist", StatusCode: int64(codes.AlreadyExists)}, nil
		}

		// Gennerating the the jwt token.
		token, err := userServiceManager.jwtManager.GenerateToken(newUser)
		if err != nil {
			fmt.Println("Error in generating token")
			return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Could not generate token: %s", err),
			)
		}
		return &userpb.AddUserResponse{Message: "User created successfully", StatusCode: 200, Token: token}, nil
	}
	return &userpb.AddUserResponse{Message: "User is already exist", StatusCode: int64(codes.AlreadyExists)}, nil
}
