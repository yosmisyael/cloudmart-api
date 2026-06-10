package entity

import "time"

type Review struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderItemID uint   `gorm:"uniqueIndex;not null" json:"order_item_id"` // one review per item
	UserID      uint   `gorm:"not null" json:"user_id"`
	ProductID   uint   `gorm:"not null" json:"product_id"`
	Rating      int    `gorm:"not null" json:"rating"` // 1–5
	Comment     string `gorm:"type:text" json:"comment"`
	VideoURL    string `gorm:"type:varchar(500)" json:"video_url"`
	ReplyText   string `gorm:"type:text" json:"reply_text"`
	RepliedAt   *time.Time `json:"replied_at"`

	Images    []ReviewImage `gorm:"foreignKey:ReviewID" json:"images"`
	OrderItem OrderItem     `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
	User      User          `gorm:"foreignKey:UserID" json:"user,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewImage struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewID uint   `gorm:"not null" json:"review_id"`
	ImageURL string `gorm:"type:varchar(500);not null" json:"image_url"`
}
