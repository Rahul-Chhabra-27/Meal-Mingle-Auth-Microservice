package config

import (
	model "auth-microservice/model"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func DatabaseDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
}
func ValidatePhone(userPhone string) bool {
	if len(userPhone) != 10 {
		return false
	}

	for _, char := range userPhone {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
func ValidateFields(userEmail string, userPassword string, userName string, userPhone string) bool {
	// Responsible for validating the fields
	if userEmail == "" || userPassword == "" || userName == "" || userPhone == "" {
		return false
	}
	if !strings.Contains(userEmail, "@") || !strings.Contains(userEmail, ".") {
		return false
	}
	// password validation also
	return len(userPassword) >= 6 && ValidatePhone(userPhone)
}
func GoDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
func ConnectDB() (*gorm.DB, *gorm.DB) {
	// Responsible for connecting to the database
	userdb, err := gorm.Open(mysql.Open(DatabaseDsn()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	userdb.AutoMigrate(&model.User{})

	ownerDetailsdb, err := gorm.Open(mysql.Open(DatabaseDsn()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	ownerDetailsdb.AutoMigrate(&model.Details{})
	return userdb, ownerDetailsdb
}

func GenerateHashedPassword(password string) string {
	// Responsible for generating a hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Couldn't hash password and the error is %s", err)
	}
	return string(hashedPassword)
}
func ComparePasswords(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func ValidateOwnerDeatils(AccountNumber string, IFSCCode string,
	BankName string, BranchName string, PanNumber string,
	AdharNumber string, GstNumber string) bool {
	// Responsible for validating the fields
	fmt.Println("AccountNumber", AccountNumber, "IFSCCode", IFSCCode, "BankName", BankName, "BranchName", BranchName, "PanNumber", PanNumber, "AdharNumber", AdharNumber, "GstNumber", GstNumber)
	
	if AccountNumber == "" || IFSCCode == "" || BankName == "" ||
		BranchName == "" || PanNumber == "" || AdharNumber == "" || GstNumber == "" {
		return false
	}

	if len(AccountNumber) != 12 {
		return false
	}
	
	// Check if IFSC code is 11 characters and starts with a letter
	if len(IFSCCode) != 11 || !unicode.IsLetter(rune(IFSCCode[0])) {
		return false
	}
	
	// Check if GST number is 15 characters and matches the pattern
	return true
}
