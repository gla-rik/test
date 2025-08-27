package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID                uint           `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	OrderUID          string         `json:"order_uid" gorm:"uniqueIndex;not null;size:100"`
	TrackNumber       string         `json:"track_number" gorm:"not null;size:100"`
	Entry             string         `json:"entry" gorm:"not null;size:50"`
	Locale            string         `json:"locale" gorm:"not null;size:10"`
	InternalSignature string         `json:"internal_signature" gorm:"size:255"`
	CustomerID        string         `json:"customer_id" gorm:"not null;size:100"`
	DeliveryService   string         `json:"delivery_service" gorm:"not null;size:100"`
	ShardKey          string         `json:"shardkey" gorm:"not null;size:10"`
	SmID              int            `json:"sm_id" gorm:"not null"`
	DateCreated       time.Time      `json:"date_created" gorm:"not null"`
	OofShard          string         `json:"oof_shard" gorm:"not null;size:10"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Delivery *Delivery   `json:"delivery" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Payment  *Payment    `json:"payment" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Items    []OrderItem `json:"items" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Order) TableName() string {
	return "orders"
}
