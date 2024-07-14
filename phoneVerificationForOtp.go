package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"

	"go.uber.org/zap"
)

func (*UserService) PhoneVerification(ctx context.Context, request *userpb.PhoneVerificationRequest) (*userpb.PhoneVerificationResponse, error) {
	logger.Info("Received PhoneVerification request", zap.String("phone", request.Phone))
	phone := request.Phone
	if !config.ValidatePhone(phone) {
		logger.Warn("Invalid phone number", zap.String("phone", phone))
		return &userpb.PhoneVerificationResponse{
			Data:       nil,
			Message:    "Invalid phone number. Phone number can only be 10 digits long",
			Error:      "Invalid Phone",
			StatusCode: StatusBadRequest,
		}, nil
	}
	// check if the phone number is already registered
	var existingUser model.User
	userNotFoundError := userDbConnector.Where("phone = ?", phone).First(&existingUser).Error
	if userNotFoundError == nil {
		logger.Warn("Phone number already registered", zap.String("phone", phone))
		return &userpb.PhoneVerificationResponse{
			Data: &userpb.PhoneVerificationResponseData{
				Phone: phone,
			},
			Message:    "The phone number is already registered.",
			Error:      "Phone number already registered",
			StatusCode: StatusOK,
		}, nil
	}
	return &userpb.PhoneVerificationResponse{
		Data:       nil,
		Message:    "Phone number is not registered",
		Error:      "Not Found",
		StatusCode: StatusNotFound,
	}, nil
}
