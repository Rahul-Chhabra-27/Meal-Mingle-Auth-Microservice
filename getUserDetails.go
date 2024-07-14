package main

import (
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"strconv"
)

func (*UserService) GetUserDetails(ctx context.Context, request *userpb.GetUserDetailsRequest) (*userpb.GetUserDetailsResponse, error) {
	logger.Info("GetUserDetails invoked")
	// Extract the fields for context
	userEmail, emailCtxError := ctx.Value("userEmail").(string)
	userRole, roleCtxError := ctx.Value("userRole").(string)
	if !emailCtxError || !roleCtxError {
		logger.Error("Failed to get user email and role from context")
		return &userpb.GetUserDetailsResponse{
			Data:       nil,
			Message:    "Failed to get user email and role from context",
			StatusCode: StatusInternalServerError,
			Error:      "Internal Server Error",
		}, nil
	}
	if userRole != model.AdminRole {
		logger.Warn("Unauthorized access")
		return &userpb.GetUserDetailsResponse{
			Data:       nil,
			Message:    "Unauthorized access",
			StatusCode: StatusUnauthorized,
			Error:      "Unauthorized",
		}, nil
	}
	var user model.User
	var details model.Details

	if err := userDbConnector.Where("email = ?", userEmail).First(&user).Error; err != nil {
		logger.Error("User not found")
		return &userpb.GetUserDetailsResponse{
			Data:       nil,
			Message:    "User not found",
			StatusCode: StatusNotFound,
			Error:      "Not Found",
		}, nil
	}
	if err := userDbConnector.Where("user_id = ?", user.ID).First(&details).Error; err != nil {
		logger.Error("User details not found")
		return &userpb.GetUserDetailsResponse{
			Data:       nil,
			Message:    "User details not found",
			StatusCode: StatusNotFound,
			Error:      "Not Found",
		}, nil
	}
	return &userpb.GetUserDetailsResponse{
		Data: &userpb.GetUserDetailsResponseData{
			User: &userpb.User{
				UserId:    strconv.FormatUint(uint64(user.ID), 10),
				UserName:  user.Name,
				UserEmail: user.Email,
				UserPhone: user.Phone,
			},
			AccountNumber: details.AccountNumber,
			IfscCode:      details.IfscCode,
			BankName:      details.BankName,
			BranchName:    details.BrachName,
			PanNumber:     details.PanNumber,
			GstNumber:     details.GstNumber,
			AdharNumber:   details.AdharNumber,
		},
		Message:    "User details fetched successfully",
		StatusCode: StatusOK,
		Error:      "",
	}, nil
}
