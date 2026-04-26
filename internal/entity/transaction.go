package entity

import "time"

type Cart struct {
	ID        uint `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint `gorm:"not null" json:"user_id"`
	VariantID uint `gorm:"not null" json:"variant_id"`
	Quantity  int  `gorm:"not null;default:1" json:"quantity"`

	Variant ProductVariant `gorm:"foreignKey:VariantID" json:"variant"`
}

type Order struct {
	ID                uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint    `gorm:"not null" json:"user_id"`
	GrandTotal        float64 `gorm:"type:decimal(15,2);not null" json:"grand_total"`
	PaymentMethod     *string `gorm:"type:varchar(50)" json:"payment_method"`
	PaymentStatus     string  `gorm:"type:varchar(20);default:'pending'" json:"payment_status"`
	ShippingAddress   string  `gorm:"type:text;not null" json:"shipping_address"`
	LogisticService   string  `gorm:"type:varchar(100)" json:"logistic_service"`
	LogisticVoucherID *uint   `json:"logistic_voucher_id"`

	SnapToken  *string `gorm:"type:varchar(255)" json:"snap_token"`
	PaymentURL *string `gorm:"type:varchar(255)" json:"payment_url"`

	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID             uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID        uint    `gorm:"not null" json:"order_id"`
	VariantID      uint    `gorm:"not null" json:"variant_id"`
	VariantDetails string  `gorm:"type:text;not null" json:"variant_details"`
	Price          float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity       int     `gorm:"not null" json:"quantity"`
}
