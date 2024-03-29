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
	Password  string `gorm:"column:password"`
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
	userDB := UserDB{
		ID:        uuid.New().String(),
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: time.Now(),
	}
	result := db.Create(&userDB)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetAllUsers retrieves a list of all users and their info from the database
func GetAllUsers() ([]UserDB, error) {
	var users []UserDB
	result := db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// DeleteUserByEmail deletes a user from the database identified by email
func DeleteUserByEmail(email string) error {
	return db.Delete(&UserDB{}, "email = ?", email).Error
}

func GetUserByEmail(email string) (UserDB, error) {
	var user UserDB
	result := db.First(&user, "email = ?", email)
	if result.Error != nil {
		return UserDB{}, result.Error
	}
	return user, nil
}
