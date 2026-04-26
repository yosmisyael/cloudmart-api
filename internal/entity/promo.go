package entity

import "time"

type Voucher struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID   uint      `gorm:"not null" json:"store_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"` // "percentage", "price", "free_shipping",
	Amount    float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Max       float64   `gorm:"type:decimal(10,2)" json:"max"`
	ExpiredAt time.Time `json:"expired_at"`

	Users []User `gorm:"many2many:user_vouchers;" json:"users,omitempty"`
}

type PaymentConfiguration struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	StoreID uint   `gorm:"not null" json:"store_id"`
	Name    string `gorm:"type:varchar(100);not null" json:"name"`

	Banks []PaymentBank `gorm:"foreignKey:PaymentConfigID" json:"banks"`
}

type PaymentBank struct {
	ID              uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	PaymentConfigID uint   `gorm:"not null" json:"payment_configuration_id"`
	Name            string `gorm:"type:varchar(100);not null" json:"name"`
	AccountID       string `gorm:"type:varchar(50);not null" json:"account_id"`
	AccountName     string `gorm:"type:varchar(100);not null" json:"account_name"`
}

type Logistic struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(100);not null" json:"name"`

	Services []LogisticService `gorm:"foreignKey:LogisticID" json:"services"`
}

type LogisticService struct {
	ID         uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	LogisticID uint    `gorm:"not null" json:"logistic_id"`
	Name       string  `gorm:"type:varchar(100);not null" json:"name"`
	BasePrice  float64 `gorm:"type:decimal(10,2);not null" json:"base_price"`
}
