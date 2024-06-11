package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"fmt"
	"strconv"
)

// AddUser is a RPC that adds a new user to the database
func (userServiceManager *UserService) AddUser(ctx context.Context, request *userpb.AddUserRequest) (*userpb.AddUserResponse, error) {
	userEmail := request.UserEmail
	userPassword := request.UserPassword
	userName := request.UserName
	userPhone := request.UserPhone
	fmt.Println("AddUser function was invoked with email: ", userEmail, " password: ", userPassword, " name: ", userName, " phone: ", userPhone)
	// Validate the fields
	if !config.ValidateFields(userEmail, userPassword, userName, userPhone) {
		return &userpb.AddUserResponse{Message: "", Error: "Invalid fields!", StatusCode: int64(400)}, nil
	}
	var existingUser model.User
	userNotFoundError := dbConnector.Where("email = ?", userEmail).First(&existingUser).Error
	// If the user is not found, create a new user with the provided details
	if userNotFoundError != nil {

		hashedPassword := config.GenerateHashedPassword(userPassword)

		newUser := &model.User{Name: userName, Email: userEmail, Phone: userPhone, Password: hashedPassword}

		// Create a new user in the database and return the primary key if successful or an error if it fails
		primaryKey := dbConnector.Create(newUser)
		if primaryKey.Error != nil {
			return &userpb.AddUserResponse{Message: "", Error: "Phone number is already registered", StatusCode: int64(400)}, nil
		}

		// Gennerating the the jwt token.
		token, err := userServiceManager.jwtManager.GenerateToken(newUser)
		if err != nil {
			fmt.Println("Error in generating token")
			return &userpb.AddUserResponse{
				Error:      "Internal Server Error",
				StatusCode: int64(500),
				Message:    "",
			}, nil
		}
		return &userpb.AddUserResponse{Message: "User created successfully", Error: "", StatusCode: 200, Data: &userpb.Data{User: &userpb.User{UserId: strconv.FormatUint(uint64(newUser.ID), 10), UserName: newUser.Name, UserEmail: newUser.Email}, Token: token}}, nil
	}
	return &userpb.AddUserResponse{Message: "", Error: "User Email is already registered", StatusCode: int64(400)}, nil
}
