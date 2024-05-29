package config

import (
	model "auth-microservice/model"
	"context"
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func GoDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
func ConnectDB(dsn string) *gorm.DB {
	// Responsible for connecting to the database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&model.User{})
	return db
}

func GenerateHashedPassword(password string) string {
	// Responsible for generating a hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Couldn't hash password and the error is %s", err)
	}
	return string(hashedPassword)
}
func UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	fmt.Println("--> UnaryInterceptor: ", info.FullMethod)
	return handler(ctx, req)
}
