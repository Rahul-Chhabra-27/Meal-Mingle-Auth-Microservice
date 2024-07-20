// user_service_test.go
package main

import (
	"auth-microservice/model"
	userpb "auth-microservice/proto/user"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func (m *MockJWTManager) GenerateToken(user *model.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

// MockUserService is a mock implementation of the UserService interface.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, request *userpb.AuthenticateUserRequest) (*userpb.AuthenticateUserResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*userpb.AuthenticateUserResponse), args.Error(1)
}

type UserServiceTestSuite struct {
	suite.Suite
	userService *MockUserService
	jwtManager  *MockJWTManager
}

func (suite *UserServiceTestSuite) SetupTest() {
	// Initialize mocks
	suite.jwtManager = new(MockJWTManager)
	suite.userService = new(MockUserService)

	// Initialize UserService with the mock JWT manager
	
}

func (suite *UserServiceTestSuite) TestAuthenticateUser_Success() {
	request := &userpb.AuthenticateUserRequest{
		UserEmail:    "validuser@example.com",
		UserPassword: "correctpassword",
		Role:         model.UserRole,
	}

	expectedUser := &model.User{
		Name:     "Valid User",
		Email:    "validuser@example.com",
		Phone:    "9876543210",
		Role:     model.UserRole,
		Password: "hashedpassword", // Assume this is the hashed password
	}

	token := "mocked_token"

	// Set up expectations
	suite.jwtManager.On("GenerateToken", expectedUser).Return(token, nil)

	suite.userService.On("AuthenticateUser", context.Background(), request).Return(
		&userpb.AuthenticateUserResponse{
			StatusCode: 200,
			Message:    "User authenticated successfully",
			Data: &userpb.Responsedata{
				User: &userpb.User{
					UserId:    "1",
					UserName:  expectedUser.Name,
					UserEmail: expectedUser.Email,
					UserPhone: expectedUser.Phone,
				},
				Token: token,
			},
			Error: "",
		}, nil)

	// Call the AuthenticateUser function
	response, err := suite.userService.AuthenticateUser(context.Background(), request)

	// Assert the results
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(200), response.StatusCode)
	assert.Equal(suite.T(), "User authenticated successfully", response.Message)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), token, response.Data.Token)
}

func (suite *UserServiceTestSuite) TestAuthenticateUser_UserNotFound() {
	request := &userpb.AuthenticateUserRequest{
		UserEmail:    "nonexistentuser@example.com",
		UserPassword: "password",
		Role:         model.UserRole,
	}

	// Set up expectations
	suite.userService.On("AuthenticateUser", context.Background(), request).Return(
		&userpb.AuthenticateUserResponse{
			StatusCode: 404,
			Message:    "Authentication Failed, User not found OR Invalid role",
			Error:      "Not Found",
			Data:       nil,
		}, nil)

	// Call the AuthenticateUser function
	response, err := suite.userService.AuthenticateUser(context.Background(), request)

	// Assert the results
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(404), response.StatusCode)
	assert.Equal(suite.T(), "Authentication Failed, User not found OR Invalid role", response.Message)
	assert.Nil(suite.T(), response.Data)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
