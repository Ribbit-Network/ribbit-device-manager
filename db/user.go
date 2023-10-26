package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserDB struct {
	ID        string `gorm:"column:id;primary_key"`
	Email     string `gorm:"column:email"`
	CreatedAt time.Time
	Password  string `gorm:"column:password"` //TODO hash password
}

type UserRepo struct {
	db *gorm.DB
}
type User struct {
	ID       string
	Email    string
	Password string
}

// TableName sets the table name for the UserDB model
func (UserDB) TableName() string {
	return "users"
}

// NewUserRepo initializes a new instance of the [UserRepo] type
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db}
}

// CreateUser will add a single new user to database
func CreateUser(user User) error {
	user.ID = uuid.New().String()
	result := db.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetAllUsers retrieves a list of all users and their info from the database
func GetAllUsers() ([]User, error) {
	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// DeleteUserByEmail deletes a user from the database identified by email
func DeleteUserByEmail(email string) error {
	return db.Delete(&User{}, "email = ?", email).Error
}

func GetUserByEmail(email string) (User, error) {
	var user User
	result := db.First(&user, "email = ?", email)
	if result.Error != nil {
		return User{}, result.Error
	}
	return user, nil
}
