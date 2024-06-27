package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"strconv"

	"go.uber.org/zap"
)

func (*UserService) AddOwnerDetails(ctx context.Context, request *userpb.AddOwnerDetailsRequest) (*userpb.AddOwnerDetailsResponse, error) {
	logger.Info("Received AddOwnerDetails request",
		zap.String("accountNumber", request.AccountNumber),
		zap.String("bankName", request.BankName),
		zap.String("branchName", request.BranchName),
		zap.String("ifscCode", request.IfscCode),
		zap.String("panNumber", request.PanNumber),
		zap.String("adharNumber", request.AdharNumber),
		zap.String("gstNumber", request.GstNumber))

	userEmail, emailCtxError := ctx.Value("userEmail").(string)
	userRole, roleCtxError := ctx.Value("userRole").(string)

	if !emailCtxError || !roleCtxError {
		logger.Error("Failed to get user email and role from context")
		return &userpb.AddOwnerDetailsResponse{
			Message:    "Failed to get user email and role from context",
			Error:      "Internal Server Error",
			StatusCode: int64(500),
		}, nil
	}
	logger.Info("Context values retrieved", zap.String("userEmail", userEmail), zap.String("userRole", userRole))
	if userRole != model.AdminRole {
		logger.Warn("Permission denied", zap.String("userRole", userRole))
		return &userpb.AddOwnerDetailsResponse{
			Data:       nil,
			Message:    "You do not have permission to perform this action. Only admin can add owner details",
			StatusCode: 403,
			Error:      "Forbidden",
		}, nil
	}
	var ownerDetails model.Details
	ownerDetails.AccountNumber = request.AccountNumber
	ownerDetails.BankName = request.BankName
	ownerDetails.BrachName = request.BranchName
	ownerDetails.IfscCode = request.IfscCode
	ownerDetails.PanNumber = request.PanNumber
	ownerDetails.AdharNumber = request.AdharNumber
	ownerDetails.GstNumber = request.GstNumber

	// validate fields here
	if !config.ValidateOwnerDeatils(request.AccountNumber, request.IfscCode,
		request.BankName, request.BranchName, request.PanNumber, request.AdharNumber, request.GstNumber) {
		logger.Warn("Invalid owner details", zap.String("userEmail", userEmail))
		return &userpb.AddOwnerDetailsResponse{
			Data:       nil,
			Message:    "Invalid owner details make sure to use mentioned format.",
			Error:      "Invalid Fields",
			StatusCode: int64(400),
		}, nil
	}
	// check if the user is owner or not
	var user model.User
	userDbConnector.Where("email = ?", userEmail).First(&user)
	ownerDetails.UserId = strconv.FormatUint(uint64(user.ID), 10)

	// check if the owner details already exists
	var existingOwnerDetails model.Details
	ownerDetailsNotFoundError := ownerDetailsDbConector.Where("user_id = ?", ownerDetails.UserId).First(&existingOwnerDetails).Error
	if ownerDetailsNotFoundError != nil {
		// create a new owner details
		primaryKey := ownerDetailsDbConector.Create(&ownerDetails)
		if primaryKey.Error != nil {
			logger.Error("Failed to create owner details", zap.String("userId", ownerDetails.UserId), zap.Error(primaryKey.Error))
			return &userpb.AddOwnerDetailsResponse{
				Data:       nil,
				Message:    "Owner details already exist",
				Error:      "Failed to create owner details",
				StatusCode: 409,
			}, nil
		}
		logger.Info("Owner details added successfully", zap.String("userId", ownerDetails.UserId))
		return &userpb.AddOwnerDetailsResponse{
			Data: &userpb.AddOwnerDetailsResponseData{
				UserId: ownerDetails.UserId,
			},
			Message:    "Owner details added successfully",
			Error:      "",
			StatusCode: int64(200),
		}, nil
	}
	logger.Warn("Owner details already exist", zap.String("userId", ownerDetails.UserId))
	return &userpb.AddOwnerDetailsResponse{
		Message:    "Owner details already exists",
		Error:      "Conflict",
		StatusCode: int64(409),
	}, nil
}
