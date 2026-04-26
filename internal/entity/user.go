package entity

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Email        string    `gorm:"varchar(100);not null" json:"email"`
	Password     string    `gorm:"type:varchar(255);not null" json:"-"`
	Phone        string    `gorm:"type:varchar(20);not null" json:"phone"`
	Role         string    `gorm:"type:varchar(20);default: 'customer'" json:"role"`
	RefreshToken string    `gorm:"type:text" json:"-"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Addresses    []Address `gorm:"foreignKey:UserID" json:"addresses,omitempty"`
}

type Address struct {
	ID                    uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID                uint    `gorm:"not null" json:"user_id"`
	Address               string  `gorm:"type:text;not null" json:"address"`
	City                  string  `gorm:"type:varchar(100);not null" json:"city"`
	State                 string  `gorm:"type:varchar(100);not null" json:"state"`
	Country               string  `gorm:"type:varchar(100);not null" json:"country"`
	PostalCode            string  `gorm:"type:varchar(20);not null" json:"postal_code"`
	Phone                 string  `gorm:"type:varchar(20);not null" json:"phone"`
	Recipient             string  `gorm:"type:varchar(100);not null" json:"recipient"`
	Type                  string  `gorm:"type:varchar(50);not null" json:"type"` // e.g., "home", "office"
	AdditionalInformation *string `gorm:"type:text" json:"additional_information"`
}

type Store struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	AddressID *uint  `json:"address_id"`
	Name      string `gorm:"type:varchar(100);not null" json:"name"`

	Products []Product `gorm:"foreignKey:StoreID" json:"products,omitempty"`
}
