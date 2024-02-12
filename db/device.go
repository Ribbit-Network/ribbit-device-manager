package db

import (
	"time"

	"gorm.io/gorm"
)

type DeviceDB struct {
	DeviceID   string `gorm:"column:device_id;primary_key"`
	DeviceName string `gorm:"column:device_name"`
	DevicePSK  string `gorm:"column:device_psk"`
	UserEmail  string `gorm:"column:user_email"` // forgein key
	ProjectID  string `gorm:"column:project_id"`
	CreatedAt  time.Time
}

type DeviceRepo struct {
	db *gorm.DB
}

// TableName sets the table name for the DeviceDB model
func (DeviceDB) TableName() string {
	return "devices"
}

// NewDeviceRepo initializes a new instance of the [UserRepo] type
func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db}
}

// CreateDevice will add a single new device to database
func CreateDevice(device DeviceDB) error {
	return db.Create(&device).Error
}
