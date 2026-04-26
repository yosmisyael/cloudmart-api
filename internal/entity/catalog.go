package entity

import "time"

type Category struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"type:varchar(100);not null" json:"name"`
	UserID    *uint  `json:"user_id"`
	IsDefault bool   `gorm:"default:false" json:"is_default"`
}

type Product struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID     uint   `gorm:"not null" json:"store_id"`
	CategoryID  uint   `gorm:"not null" json:"category_id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`

	Store    Store            `gorm:"foreignKey:StoreID" json:"store"`
	Category Category         `gorm:"foreignKey:CategoryID" json:"category"`
	Variants []ProductVariant `gorm:"foreignKey:ProductID" json:"variants"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductVariant struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	SKU       string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"`
	Color     string  `gorm:"type:varchar(50);not null" json:"color"`
	Size      string  `gorm:"type:varchar(50);not null" json:"size"`
	Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock     int     `gorm:"not null;default:0" json:"stock"`

	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}
