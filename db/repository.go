// package db provides functionality for interacting with a relational database
package db

import (
	"errors"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type Repository struct {
	users   *UserRepo
	devices *DeviceRepo
}

// AutoMigrate will automatically migrate the database and correct schema errors on startup
func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&UserDB{}, &DeviceDB{})
}

func NewRepository(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &Repository{
		users:   NewUserRepo(db),
		devices: NewDeviceRepo(db),
	}, nil
}

// ConnectDatabase initalizes the sql database connection and gorm
func ConnectDatabase() {
	// Read the DSN from the environment variable
	dsn := os.Getenv("DSN_ENV")
	if dsn == "" {
		log.Fatal("DSN_ENV environment variable is not set")
	}

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	db.AutoMigrate(&UserDB{}, &DeviceDB{})
}

func (r *Repository) Users() *UserRepo {
	return r.users
}

func (r *Repository) Devices() *DeviceRepo {
	return r.devices
}
