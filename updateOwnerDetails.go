package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"strconv"

	"go.uber.org/zap"
)

func (*UserService) UpdateOwnerDetails(ctx context.Context, request *userpb.UpdateOwnerDetailsRequest) (*userpb.UpdateOwnerDetailsResponse, error) {
	// get the user email from the context
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		logger.Error("Failed to get user email from context")
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "Failed to get user email from context",
			StatusCode: StatusInternalServerError,
			Error:      "Internal Server Error",
		}, nil
	}
	logger.Info("Received UpdateOwnerDetails request", zap.String("userEmail", userEmail))

	// get the user email from the database
	var user model.User
	var ownerDetails model.Details
	err := userDbConnector.Where("email = ?", userEmail).First(&user)
	if err.Error != nil {
		logger.Warn("Admin does not exist", zap.String("userEmail", userEmail), zap.Error(err.Error))
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "Admin does not exist",
			Error:      "Not authorized",
			StatusCode: StatusNotFound,
		}, nil
	}
	logger.Info("Retrieved user details successfully", zap.String("userEmail", userEmail))
	// validate fields here
	if !config.ValidateOwnerDeatils(request.AccountNumber, request.IfscCode,
		request.BankName, request.BranchName, request.PanNumber,
		request.AdharNumber, request.GstNumber) {
		logger.Warn("Invalid owner details", zap.String("userEmail", userEmail))
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "You do not have permission to perform this action. Invalid owner details.",
			Error:      "Invalid Fields",
			StatusCode: StatusBadRequest,
		}, nil
	}
	// check if owner details already exists
	ownerDetailsNotFoundError := ownerDetailsDbConector.Where("user_id = ?", user.ID).First(&ownerDetails).Error
	if ownerDetailsNotFoundError != nil || user.Role != model.AdminRole {
		logger.Warn("Owner details not found", zap.String("userEmail", userEmail), zap.Error(ownerDetailsNotFoundError))
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "Owner details not found",
			Error:      "Not authorized",
			StatusCode: StatusNotFound,
		}, nil
	}

	ownerDetails.AccountNumber = request.AccountNumber
	ownerDetails.BankName = request.BankName
	ownerDetails.BrachName = request.BranchName
	ownerDetails.IfscCode = request.IfscCode
	ownerDetails.PanNumber = request.PanNumber
	ownerDetails.AdharNumber = request.AdharNumber
	ownerDetails.GstNumber = request.GstNumber
	ownerDetails.UserId = strconv.FormatUint(uint64(user.ID), 10)

	// Save updated owner details
	if err := ownerDetailsDbConector.Where("user_id = ?", user.ID).Save(&ownerDetails).Error; err != nil {
		logger.Error("Failed to update owner details", zap.String("userEmail", userEmail), zap.Error(err))
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "Failed to update owner details",
			Error:      "Internal Server Error",
			StatusCode: StatusInternalServerError,
		}, nil
	}
	logger.Info("Owner details updated successfully", zap.String("userEmail", userEmail))
	return &userpb.UpdateOwnerDetailsResponse{
		Data: &userpb.UpdateOwnerDetailsResponseData{
			UserId: ownerDetails.UserId,
		},
		Message:    "Owner details updated successfully",
		StatusCode: StatusOK,
		Error:      "",
	}, nil
}
