package config

import (
	model "auth-microservice/model"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
   openapi "github.com/twilio/twilio-go/rest/verify/v2"
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
func SendOtp(to string) error {
	var client = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: GoDotEnvVariable("ACCOUNT_SID"),
		Password: GoDotEnvVariable("AUTH_TOKEN"),
	})
	params := &openapi.CreateVerificationParams{}
	params.SetTo(to)
	params.SetChannel("sms")

	resp, err := client.VerifyV2.CreateVerification(GoDotEnvVariable("VERIFICATION_SID"), params)
	fmt.Println("Verification SID: ", resp)
	if err != nil {
		return err;
	} else {
		fmt.Printf("Sent verification '%s'\n", *resp.Sid)
		return nil;
	}
}

func CheckOtp(to string,code string) (string,error) {
	var client = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: GoDotEnvVariable("ACCOUNT_SID"),
		Password: GoDotEnvVariable("AUTH_TOKEN"),
	})
	fmt.Println("Please check your phone and enter the code:")
	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(to)
	params.SetCode(code)
 
	resp, err := client.VerifyV2.CreateVerificationCheck(GoDotEnvVariable("VERIFICATION_SID"), params)
 
	if err != nil {
		fmt.Println(err.Error())
		return "",err;
	} else if *resp.Status == "approved" {
		fmt.Println("Correct!")
		return *resp.Status,nil;
	} else {
		fmt.Println("Incorrect!")
		return *resp.Status,nil;
	}
 }