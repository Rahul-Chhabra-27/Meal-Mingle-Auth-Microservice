package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"strconv"
	"strings"

	"go.uber.org/zap"
)


func (UserServiceManager *UserService) AuthenticateUser(ctx context.Context, request *userpb.AuthenticateUserRequest) (*userpb.AuthenticateUserResponse, error) {
	userEmail := request.UserEmail
	userPassword := request.UserPassword
	logger.Info("Received AuthenticateUser request",
		zap.String("userEmail", userEmail))

	// Check if the request contains the required fields
	if userEmail == "" || userPassword == "" || !strings.Contains(userEmail, "@") || 
	!strings.Contains(userEmail, ".") || len(userPassword) < 6 {

		logger.Warn("Invalid request fields",
			zap.String("userEmail", userEmail),
			zap.String("userPaasword", userPassword))

		return &userpb.AuthenticateUserResponse{
			Data:       nil,
			Message:    "The request contains missing or invalid fields.",
			Error:      "Invalid fields!",
			StatusCode: 400,
		}, nil
	}
	var existingUser model.User
	userNotFoundError := userDbConnector.Where("email = ?", userEmail).First(&existingUser).Error
	// If the user is not found, create a new user with the provided details
	if userNotFoundError != nil || existingUser.Role != request.Role {
		logger.Warn("Authentication failed",
			zap.String("userEmail", userEmail),
			zap.String("userRole", existingUser.Role),
			zap.Error(userNotFoundError))
		return &userpb.AuthenticateUserResponse{
			Data:       nil,
			Message:    "Authentication Failed, User not found OR Invalid role",
			Error:      "Not Found",
			StatusCode: 404,
		}, nil
	}
	if config.ComparePasswords(existingUser.Password, userPassword) != nil {
		logger.Warn("Authentication failed due to wrong password",
			zap.String("userEmail", userEmail))
		return &userpb.AuthenticateUserResponse{
			Message: "Authentication Failed,Wrong Password",
			Error:   "Unauthorized", StatusCode: 401,
		}, nil
	}
	// Gennerating the the jwt token.
	token, err := UserServiceManager.jwtManager.GenerateToken(&existingUser)
	if err != nil {
		logger.Error("Error in generating token",
			zap.String("userEmail", userEmail),
			zap.Error(err))
		return &userpb.AuthenticateUserResponse{
			Data:       nil,
			Error:      "Internal Server Error",
			StatusCode: 500,
			Message:    "Security Issues, Please try again later.",
		}, nil

	}
	logger.Info("User authenticated successfully",
		zap.String("userEmail", existingUser.Email),
		zap.String("userName", existingUser.Name))

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
