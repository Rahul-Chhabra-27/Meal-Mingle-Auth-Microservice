package model
import "gorm.io/gorm"
// role as enum for user
const (
	AdminRole = "admin"
	UserRole = "user"
)
type User struct {
	gorm.Model
	Name string
	Password string
	Email   string  `gorm:"unique"`
	Phone	string  `gorm:"unique"`
	Address string
	City string
	Role string
}

type Details struct {
	AccountNumber string
	IfscCode string
	BankName string
	BrachName string
	PanNumber string
	GstNumber string
	AdharNumber string
	// foreign key for user table
	UserId string `gorm:"foreignKey:UserID"`
}
