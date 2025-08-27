package models

import (
	"time"
)

type Payment struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	OrderID      uint      `json:"order_id" gorm:"not null;index;type:bigint"`
	Transaction  string    `json:"transaction" gorm:"not null;size:100"`
	RequestID    string    `json:"request_id" gorm:"size:100"`
	Currency     string    `json:"currency" gorm:"not null;size:10"`
	Provider     string    `json:"provider" gorm:"not null;size:100"`
	Amount       float64   `json:"amount" gorm:"not null"`
	PaymentDt    time.Time `json:"payment_dt" gorm:"not null"`
	Bank         string    `json:"bank" gorm:"not null;size:100"`
	DeliveryCost float64   `json:"delivery_cost" gorm:"not null"`
	GoodsTotal   float64   `json:"goods_total" gorm:"not null"`
	CustomFee    float64   `json:"custom_fee" gorm:"not null"`
}

func (Payment) TableName() string {
	return "payments"
}
