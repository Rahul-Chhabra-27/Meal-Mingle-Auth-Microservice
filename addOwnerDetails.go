package main

import (
	"auth-microservice/config"
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"strconv"
)

func (*UserService) AddOwnerDetails(ctx context.Context, request *userpb.AddOwnerDetailsRequest) (*userpb.AddOwnerDetailsResponse, error) {
	// get the user email from the context
	userEmail, ok := ctx.Value("userEmail").(string)
	if !ok {
		return &userpb.AddOwnerDetailsResponse{
			Message:    "Failed to get user email from context",
			Error:      "Internal Server Error",
			StatusCode: int64(500),
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
		return &userpb.AddOwnerDetailsResponse{
			Data: 	 nil,
			Message:    "You do not have permission to perform this action. Invalid owner details.",
			Error:      "Invalid Fields",
			StatusCode: int64(400),
		}, nil
	}
	// check if the user is owner or not
	var user model.User
	err := userDbConnector.Where("email = ?", userEmail).First(&user)
	if err.Error != nil || user.Role != model.AdminRole {
		return &userpb.AddOwnerDetailsResponse{
			Data: 	 nil,
			Message:    "You do not have permission to perform this action. Invalid owner details.",
			Error:      "Invalid Fields",
			StatusCode: int64(400),
		}, nil
	}
	ownerDetails.UserId =  strconv.FormatUint(uint64(user.ID),10); 
	// check if the owner details already exists
	var existingOwnerDetails model.Details
	ownerDetailsNotFoundError := ownerDetailsDbConector.Where("user_id = ?", ownerDetails.UserId).First(&existingOwnerDetails).Error
	if ownerDetailsNotFoundError != nil {
		// create a new owner details
		primaryKey := ownerDetailsDbConector.Create(&ownerDetails)
		if primaryKey.Error != nil {
			return &userpb.AddOwnerDetailsResponse{
				Data: 	 nil,
				Message:    "Owner details already exist",
				Error:      "Failed to create owner details",
				StatusCode: 409,
			}, nil
		}
		
		return &userpb.AddOwnerDetailsResponse{
			Data: &userpb.AddOwnerDetailsResponseData{
				UserId:        ownerDetails.UserId,
			},
			Message:    "Owner details added successfully",
			Error:      "",
			StatusCode: int64(200),
		}, nil
	}
	return &userpb.AddOwnerDetailsResponse{
		Message:    "Owner details already exists",
		Error:      "Conflict",
		StatusCode: int64(409),
	}, nil
}
