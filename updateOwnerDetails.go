package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"strconv"
)

func (*UserService) UpdateOwnerDetails(ctx context.Context, request *userpb.UpdateOwnerDetailsRequest) (*userpb.UpdateOwnerDetailsResponse, error) {
	// get the user email from the context
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "Failed to get user email from context",
			StatusCode: 500,
			Error:      "Internal Server Error",
		}, nil
	}
	// get the user email from the database
	var user model.User
	var ownerDetails model.Details
	err := userDbConnector.Where("email = ?", userEmail).First(&user)
	if err.Error != nil {
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "Admin does not exist",
			Error:      "Not authorized",
			StatusCode: int64(404),
		}, nil
	}
	// validate fields here
	if !config.ValidateOwnerDeatils(request.AccountNumber, request.IfscCode,
		request.BankName, request.BranchName, request.PanNumber, request.AdharNumber, request.GstNumber) {
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "You do not have permission to perform this action. Invalid owner details.",
			Error:      "Invalid Fields",
			StatusCode: int64(400),
		}, nil
	}
	// check if owner details already exists
	ownerDetailsNotFoundError := ownerDetailsDbConector.Where("user_id = ?", user.ID).First(&ownerDetails).Error
	if ownerDetailsNotFoundError != nil || user.Role != model.AdminRole {
		return &userpb.UpdateOwnerDetailsResponse{
			Data:       nil,
			Message:    "Owner details not found",
			Error:      "Not authorized",
			StatusCode: int64(404),
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

	ownerDetailsDbConector.Where("user_id = ?", user.ID).Save(&ownerDetails)
	return &userpb.UpdateOwnerDetailsResponse{
		Data: &userpb.UpdateOwnerDetailsResponseData{
			UserId: ownerDetails.UserId,
		},
		Message:    "Owner details updated successfully",
		StatusCode: int64(200),
		Error:      "",
	}, nil
}
