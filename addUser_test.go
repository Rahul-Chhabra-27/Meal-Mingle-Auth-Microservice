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
	"gorm.io/gorm"
)

// MockJWTManager is a mock implementation of the JWTManager interface.
type MockJWTManager struct {
	mock.Mock
}

// MockUserServiceAddUser is a mock implementation of the UserService interface.
type MockUserServiceAddUser struct {
	mock.Mock
	jwtManager *MockJWTManager
}

func (m *MockUserServiceAddUser) AddUser(ctx context.Context, request *userpb.AddUserRequest) (*userpb.AddUserResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*userpb.AddUserResponse), args.Error(1)
}

// MockDBConnector is a mock implementation of the GORM DB connector.
type MockDBConnector struct {
	mock.Mock
}

func (m *MockDBConnector) Where(query string, args ...interface{}) *gorm.DB {
	args = append([]interface{}{query}, args...)
	return m.Called(args...).Get(0).(*gorm.DB)
}

func (m *MockDBConnector) Create(value interface{}) *gorm.DB {
	return m.Called(value).Get(0).(*gorm.DB)
}

type UserServiceTestSuiteAddUser struct {
	suite.Suite
	userService *MockUserServiceAddUser
	jwtManager  *MockJWTManager
	dbConnector *MockDBConnector
}

func (suite *UserServiceTestSuiteAddUser) SetupTest() {
	// Initialize mocks
	suite.jwtManager = new(MockJWTManager)
	suite.dbConnector = new(MockDBConnector)
	suite.userService = &MockUserServiceAddUser{
		jwtManager: suite.jwtManager,
	}
}

func (suite *UserServiceTestSuiteAddUser) TestAddUser_Success() {
	request := &userpb.AddUserRequest{
		UserEmail:    "newuser@example.com",
		UserPassword: "validpassword",
		UserName:     "New User",
		UserPhone:    "1234567890",
		UserRole:     model.UserRole,
	}

	expectedUser := &model.User{
		Name:     "New User",
		Email:    "newuser@example.com",
		Phone:    "1234567890",
		Role:     model.UserRole,
		Password: "hashedpassword",
	}

	token := "mocked_token"

	suite.dbConnector.On("Where", "email = ?", request.UserEmail).Return(suite.dbConnector)
	suite.dbConnector.On("Create", mock.AnythingOfType("*model.User")).Return(&gorm.DB{Error: nil})
	suite.jwtManager.On("GenerateToken", expectedUser).Return(token, nil)

	suite.userService.On("AddUser", context.Background(), request).Return(
		&userpb.AddUserResponse{
			StatusCode: 200,
			Message:    "User created successfully",
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

	response, err := suite.userService.AddUser(context.Background(), request)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(200), response.StatusCode)
	assert.Equal(suite.T(), "User created successfully", response.Message)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), token, response.Data.Token)
}

func (suite *UserServiceTestSuiteAddUser) TestAddUser_UserAlreadyExists() {
	request := &userpb.AddUserRequest{
		UserEmail:    "existinguser@example.com",
		UserPassword: "validpassword",
		UserName:     "Existing User",
		UserPhone:    "1234567890",
		UserRole:     model.UserRole,
	}

	suite.dbConnector.On("Where", "email = ?", request.UserEmail).Return(suite.dbConnector)
	suite.dbConnector.On("Create", mock.AnythingOfType("*model.User")).Return(&gorm.DB{Error: gorm.ErrDuplicatedKey})

	suite.userService.On("AddUser", context.Background(), request).Return(
		&userpb.AddUserResponse{
			StatusCode: 409,
			Message:    "User Email is already registered",
			Error:      "Conflict",
			Data:       nil,
		}, nil)

	response, err := suite.userService.AddUser(context.Background(), request)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(409), response.StatusCode)
	assert.Equal(suite.T(), "User Email is already registered", response.Message)
	assert.Nil(suite.T(), response.Data)
}

func (suite *UserServiceTestSuiteAddUser) TestAddUser_InvalidRole() {
	request := &userpb.AddUserRequest{
		UserEmail:    "validuser@example.com",
		UserPassword: "validpassword",
		UserName:     "Valid User",
		UserPhone:    "1234567890",
		UserRole:     "invalidrole",
	}

	suite.userService.On("AddUser", context.Background(), request).Return(
		&userpb.AddUserResponse{
			StatusCode: 400,
			Message:    "Invalid user role. User role can only be user or admin",
			Error:      "Invalid Role",
			Data:       nil,
		}, nil)

	response, err := suite.userService.AddUser(context.Background(), request)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(400), response.StatusCode)
	assert.Equal(suite.T(), "Invalid user role. User role can only be user or admin", response.Message)
	assert.Nil(suite.T(), response.Data)
}

func (suite *UserServiceTestSuiteAddUser) TestAddUser_InvalidFields() {
	request := &userpb.AddUserRequest{
		UserEmail:    "invalidemail",
		UserPassword: "short",
		UserName:     "",
		UserPhone:    "123",
		UserRole:     model.UserRole,
	}

	suite.userService.On("AddUser", context.Background(), request).Return(
		&userpb.AddUserResponse{
			StatusCode: 400,
			Message:    "The request contains missing or invalid fields. Make sure Phone number is 10 digits long.",
			Error:      "Invalid Request",
			Data:       nil,
		}, nil)

	response, err := suite.userService.AddUser(context.Background(), request)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(400), response.StatusCode)
	assert.Equal(suite.T(), "The request contains missing or invalid fields. Make sure Phone number is 10 digits long.", response.Message)
	assert.Nil(suite.T(), response.Data)
}

func (suite *UserServiceTestSuiteAddUser) TestAddUser_DatabaseError() {
	request := &userpb.AddUserRequest{
		UserEmail:    "newuser@example.com",
		UserPassword: "validpassword",
		UserName:     "New User",
		UserPhone:    "1234567890",
		UserRole:     model.UserRole,
	}

	suite.dbConnector.On("Where", "email = ?", request.UserEmail).Return(suite.dbConnector)
	suite.dbConnector.On("Create", mock.AnythingOfType("*model.User")).Return(&gorm.DB{Error: gorm.ErrInvalidData})

	suite.userService.On("AddUser", context.Background(), request).Return(
		&userpb.AddUserResponse{
			StatusCode: 409,
			Message:    "The phone number is already registered.",
			Error:      "Conflict",
			Data:       nil,
		}, nil)

	response, err := suite.userService.AddUser(context.Background(), request)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(409), response.StatusCode)
	assert.Equal(suite.T(), "The phone number is already registered.", response.Message)
	assert.Nil(suite.T(), response.Data)
}

func (suite *UserServiceTestSuiteAddUser) TestAddUser_TokenGenerationError() {
	request := &userpb.AddUserRequest{
		UserEmail:    "newuser@example.com",
		UserPassword: "validpassword",
		UserName:     "New User",
		UserPhone:    "1234567890",
		UserRole:     model.UserRole,
	}

	expectedUser := &model.User{
		Name:     "New User",
		Email:    "newuser@example.com",
		Phone:    "1234567890",
		Role:     model.UserRole,
		Password: "hashedpassword",
	}

	suite.dbConnector.On("Where", "email = ?", request.UserEmail).Return(suite.dbConnector)
	suite.dbConnector.On("Create", mock.AnythingOfType("*model.User")).Return(&gorm.DB{Error: nil})
	suite.jwtManager.On("GenerateToken", expectedUser).Return("", gorm.ErrInvalidData)

	suite.userService.On("AddUser", context.Background(), request).Return(
		&userpb.AddUserResponse{
			StatusCode: 500,
			Message:    "Security Issues, Please try again later.",
			Error:      "Internal Server Error",
			Data:       nil,
		}, nil)

	response, err := suite.userService.AddUser(context.Background(), request)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), int64(500), response.StatusCode)
	assert.Equal(suite.T(), "Security Issues, Please try again later.", response.Message)
	assert.Nil(suite.T(), response.Data)
}

func TestUserServiceTestSuiteAddUserAdddUser(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuiteAddUser))
}
