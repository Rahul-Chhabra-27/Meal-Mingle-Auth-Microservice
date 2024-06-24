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
	userRole := request.UserRole

	if !config.ValidateFields(userEmail, userPassword, userName, userPhone) {
		return &userpb.AddUserResponse{
			Data: nil,
			Message: "The request contains missing or invalid fields. Make sure Phone number is 10 digits long.", 
			Error: "Invalid Request", 
			StatusCode: int64(400),
		}, nil
	}
	var existingUser model.User
	userNotFoundError := userDbConnector.Where("email = ?", userEmail).First(&existingUser).Error
	if userNotFoundError != nil {
		hashedPassword := config.GenerateHashedPassword(userPassword)
		newUser := &model.User{ Name: userName, Email: userEmail,
			Phone: userPhone, Password: hashedPassword, Role: userRole}

		// Create a new user in the database and return the primary key if successful or an error if it fails
		primaryKey := userDbConnector.Create(newUser)
		if primaryKey.Error != nil {
			return &userpb.AddUserResponse{
				Data: nil,
				Message: "Conflict", 
				Error: "The phone number is already registered.", 
				StatusCode: int64(400),
			}, nil
		}
		// Gennerating the the jwt token.
		token, err := userServiceManager.jwtManager.GenerateToken(newUser)
		if err != nil {
			fmt.Println("Error in generating token")
			return &userpb.AddUserResponse{
				Data: nil,
				Error:      "Internal Server Error",
				StatusCode: int64(500),
				Message:    "Security Issues, Please try again later.",
			}, nil
		}
		return &userpb.AddUserResponse{
			Message: "User created successfully", 
			Error: "", StatusCode: 200, 
			Data: &userpb.Responsedata{ 
				User: &userpb.User{
					UserId: strconv.FormatUint(uint64(newUser.ID), 10), 
					UserName: newUser.Name, 
					UserEmail: newUser.Email,
					UserPhone: newUser.Phone,
				}, 
				Token: token,
			},
		}, nil
	}
	return &userpb.AddUserResponse{
		Data: nil,
		Message: "User Email is already registered", 
		Error: "Conflict", 
		StatusCode: int64(409),
	}, nil
}